package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/transport"
	"github.com/krewire/krewire/pkg/ai/providers"
)

func setupTestAgent() *agent.Agent {
	mock := providers.NewMock("Server response ok")
	ag, _ := agent.New(
		agent.WithProvider(mock),
	)
	return ag
}

func TestHTTPServer_Health(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("unexpected status: %v", body["status"])
	}
}

func TestHTTPServer_Run(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag)

	payload := `{"prompt": "Hello server", "stream": false}`
	req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res agent.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.Content != "Server response ok" {
		t.Errorf("unexpected content: %q", res.Content)
	}
}

func TestHTTPServer_RunStreamSSE(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag)

	payload := `{"prompt": "Stream to me", "stream": true}`
	req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: start") {
		t.Errorf("expected SSE event: start in body, got: %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Errorf("expected SSE event: done in body, got: %s", body)
	}
}

func TestMCPServer_InitializeAndListTools(t *testing.T) {
	ag := setupTestAgent()

	var in bytes.Buffer
	var out bytes.Buffer

	// 1. Send initialize request
	in.WriteString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n")
	// 2. Send tools/list request
	in.WriteString(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n")

	srv := transport.NewMCPServer(ag, &in, &out)
	_ = srv.Serve(context.Background())

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 JSON-RPC responses, got: %s", output)
	}

	var resp1, resp2 transport.MCPResponse
	if err := json.Unmarshal([]byte(lines[0]), &resp1); err != nil {
		t.Fatalf("decode resp1: %v", err)
	}
	if resp1.ID != float64(1) {
		t.Errorf("unexpected resp1 ID: %v", resp1.ID)
	}

	if err := json.Unmarshal([]byte(lines[1]), &resp2); err != nil {
		t.Fatalf("decode resp2: %v", err)
	}
	if resp2.ID != float64(2) {
		t.Errorf("unexpected resp2 ID: %v", resp2.ID)
	}
}
