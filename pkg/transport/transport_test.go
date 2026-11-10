package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	srv := transport.NewHTTPServer(ag, transport.WithNoAuth(true))

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
	srv := transport.NewHTTPServer(ag, transport.WithNoAuth(true))

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
	if !strings.Contains(body, "event: finish") {
		t.Errorf("expected SSE event: finish in body, got: %s", body)
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

func TestHTTPServer_OversizedBodyRejected(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag,
		transport.WithNoAuth(true),
		transport.WithMaxBodySize(1024),
	)

	largePrompt := strings.Repeat("A", 2048)
	payload := fmt.Sprintf(`{"prompt": "%s"}`, largePrompt)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413 Payload Too Large, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHTTPServer_SecurityHeaders(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag, transport.WithNoAuth(true))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if val := rec.Header().Get("X-Content-Type-Options"); val != "nosniff" {
		t.Errorf("expected X-Content-Type-Options 'nosniff', got %q", val)
	}
	if val := rec.Header().Get("X-Frame-Options"); val != "DENY" {
		t.Errorf("expected X-Frame-Options 'DENY', got %q", val)
	}
	if val := rec.Header().Get("Referrer-Policy"); val != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy 'strict-origin-when-cross-origin', got %q", val)
	}
}

func TestHTTPServer_StaticToken(t *testing.T) {
	ag := setupTestAgent()
	srv := transport.NewHTTPServer(ag, transport.WithStaticToken("secret-token-42", nil))

	// 1. Missing header -> 401
	{
		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hi"}`))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for missing token, got %d", rec.Code)
		}
	}

	// 2. Wrong token -> 401
	{
		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hi"}`))
		req.Header.Set("Authorization", "Bearer wrong-token")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for wrong token, got %d", rec.Code)
		}
	}

	// 3. Correct token -> 200
	{
		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hi"}`))
		req.Header.Set("Authorization", "Bearer secret-token-42")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for valid token, got %d: %s", rec.Code, rec.Body.String())
		}
	}
}
