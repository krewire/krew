package transport_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/transport"
	"github.com/krewire/krewire/pkg/ai"
)

type sseFrame struct {
	event string
	data  map[string]any
}

// streamSSE performs a streaming run against the handler and returns decoded
// SSE frames in arrival order.
func streamSSE(t *testing.T, handler http.Handler, target, payload string, headers map[string]string) []sseFrame {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %q", rec.Header().Get("Content-Type"))
	}

	var frames []sseFrame
	scanner := bufio.NewScanner(bytes.NewReader(rec.Body.Bytes()))
	var current *sseFrame
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			current = &sseFrame{event: strings.TrimPrefix(line, "event: ")}
		case strings.HasPrefix(line, "data: "):
			if current == nil {
				t.Fatal("data line without a preceding event line")
			}
			var data map[string]any
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &data); err != nil {
				t.Fatalf("decode SSE payload: %v", err)
			}
			current.data = data
			frames = append(frames, *current)
			current = nil
		}
	}
	return frames
}

// asText renders a decoded JSON value for assertions.
func asText(v any) string {
	switch n := v.(type) {
	case string:
		return n
	case float64:
		return fmt.Sprintf("%v", n)
	default:
		return ""
	}
}

func TestHTTPServer_StreamEventSequence(t *testing.T) {
	srv := transport.NewHTTPServer(setupTestAgent())

	frames := streamSSE(t, srv.Handler(), "/api/agent/run", `{"prompt":"Stream to me","stream":true}`, nil)
	if len(frames) == 0 {
		t.Fatal("expected SSE frames")
	}

	// Sequence: start first, finish last, only delta/step events between.
	if frames[0].event != "start" {
		t.Errorf("expected start first, got %q", frames[0].event)
	}
	last := frames[len(frames)-1]
	if last.event != "finish" {
		t.Errorf("expected finish last, got %q", last.event)
	}

	var deltas strings.Builder
	for i, frame := range frames {
		switch frame.event {
		case "text_delta":
			if i == 0 {
				t.Error("text_delta cannot precede start")
			}
			if frames[i-1].event == "finish" {
				t.Error("text_delta arrived after finish")
			}
			deltas.WriteString(frame.data["content"].(string))
		case "finish":
			usage, _ := frame.data["usage"].(map[string]any)
			if usage == nil {
				t.Errorf("expected usage stats on finish event, got %v", frame.data)
				continue
			}
			if asText(usage["total_tokens"]) == "" || usage["total_tokens"].(float64) <= 0 {
				t.Errorf("expected non-zero total tokens on finish, got %v", usage)
			}
		}
	}
	if deltas.String() != "Server response ok" {
		t.Errorf("expected assembled deltas equal to mock content, got %q", deltas.String())
	}
}

// TestHTTPServer_AcceptHeaderTriggersSSE verifies Accept: text/event-stream
// negotiation without the JSON body flag.
func TestHTTPServer_AcceptHeaderTriggersSSE(t *testing.T) {
	srv := transport.NewHTTPServer(setupTestAgent())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"prompt":"hi"}`))
	req.Header.Set("Accept", "text/event-stream")
	srv.Handler().ServeHTTP(rec, req)

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected SSE content type via Accept header, got %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "event: finish") {
		t.Errorf("expected finish event, got %s", rec.Body.String())
	}
}

// holdingStreamProvider keeps its Stream open emitting tokens until the
// caller's context is cancelled, simulating a long generation interrupted by
// a client disconnect.
type holdingStreamProvider struct {
	started        chan struct{}
	cancelObserved chan struct{}
}

func (p *holdingStreamProvider) Name() string { return "holding" }

func (p *holdingStreamProvider) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	return &ai.CompletionResponse{Content: "ok", FinishReason: "stop"}, nil
}

func (p *holdingStreamProvider) Stream(ctx context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk)
	go func() {
		defer close(ch)
		close(p.started)
		for {
			select {
			case <-ctx.Done():
				close(p.cancelObserved)
				return
			case ch <- ai.StreamChunk{Delta: "token "}:
			}
		}
	}()
	return ch, nil
}

// TestHTTPServer_ClientDisconnectStopsAgent mirrors the checklist: terminate
// the client midway and confirm the agent stops producing immediately.
func TestHTTPServer_ClientDisconnectStopsAgent(t *testing.T) {
	prov := &holdingStreamProvider{started: make(chan struct{}), cancelObserved: make(chan struct{})}
	ag, err := agent.New(agent.WithProvider(prov), agent.WithStreaming(true))
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	server := httptest.NewServer(transport.NewHTTPServer(ag).Handler())
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		server.URL+"/api/agent/run", strings.NewReader(`{"prompt":"long run"}`))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer res.Body.Close()

	// Observe the first arriving token, then cancel the client context.
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "data: ") {
			cancel()
			break
		}
	}

	select {
	case <-prov.cancelObserved:
	case <-time.After(5 * time.Second):
		t.Fatal("provider stream was not cancelled after client disconnect")
	}
}
