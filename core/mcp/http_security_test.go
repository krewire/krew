package mcp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHTTPTransport_WildcardCORS_NeverPresent ensures that neither SSE nor message endpoints
// ever respond with wildcard CORS (*), regardless of the request.
func TestHTTPTransport_WildcardCORS_NeverPresent(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	testCases := []struct {
		name   string
		method string
		url    string
		origin string
	}{
		{"GET SSE without origin", "GET", server.URL + "/mcp/sse", ""},
		{"OPTIONS SSE without origin", "OPTIONS", server.URL + "/mcp/sse", ""},
		{"POST message without origin", "POST", server.URL + "/mcp/message", ""},
		{"OPTIONS message without origin", "OPTIONS", server.URL + "/mcp/message", ""},
		{"OPTIONS message with untrusted origin", "OPTIONS", server.URL + "/mcp/message", "http://evil.com"},
		{"OPTIONS message with trusted origin", "OPTIONS", server.URL + "/mcp/message", "http://localhost:3000"},
		{"POST message with untrusted origin", "POST", server.URL + "/mcp/message", "http://attacker.com"},
		{"POST message with trusted origin", "POST", server.URL + "/mcp/message", "http://localhost:3000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			if allowOrigin == "*" {
				t.Fatalf("endpoint responded with wildcard CORS (*)")
			}
		})
	}
}

// TestHTTPTransport_UntrustedOrigin_Rejected verifies that untrusted browser origins
// are rejected before any tool execution occurs (protecting inspect, scaffold, and test execution).
func TestHTTPTransport_UntrustedOrigin_Rejected(t *testing.T) {
	srv, root := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Inspect project via untrusted origin -> 403 Forbidden, no project data leaked
	inspectPayload := `{"jsonrpc":"2.0","id":"inspect-1","method":"tools/call","params":{"name":"kiw_inspect_project","arguments":{}}}`
	req, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(inspectPayload))
	req.Header.Set("Origin", "http://evil.com")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "untrusted origin") {
		t.Errorf("expected untrusted origin error message, got: %s", string(body))
	}
	if strings.Contains(string(body), "TestApp") {
		t.Errorf("project metadata was leaked to untrusted origin: %s", string(body))
	}

	// 2. Scaffold file via untrusted origin -> 403 Forbidden, no file created
	scaffoldTarget := filepath.Join(root, "components", "EvilComponent.kiw")
	scaffoldPayload := `{"jsonrpc":"2.0","id":"scaffold-1","method":"tools/call","params":{"name":"kiw_make_scaffold","arguments":{"kind":"component","name":"EvilComponent"}}}`
	req2, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(scaffoldPayload))
	req2.Header.Set("Origin", "https://malicious-site.net")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("scaffold request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for scaffolding from untrusted origin, got %d", resp2.StatusCode)
	}
	if _, err := os.Stat(scaffoldTarget); !os.IsNotExist(err) {
		t.Fatalf("scaffold file was created despite untrusted origin rejection: %s", scaffoldTarget)
	}

	// 3. Verification / Test execution via untrusted origin -> 403 Forbidden
	verifyPayload := `{"jsonrpc":"2.0","id":"verify-1","method":"tools/call","params":{"name":"kiw_verify","arguments":{"task_id":"csrf-test"}}}`
	req3, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(verifyPayload))
	req3.Header.Set("Origin", "http://attacker.internal:8080")
	req3.Header.Set("Content-Type", "application/json")

	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("verify request failed: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for verification from untrusted origin, got %d", resp3.StatusCode)
	}

	// 4. SSE subscription via untrusted origin -> 403 Forbidden
	reqSSE, _ := http.NewRequest("GET", server.URL+"/mcp/sse", nil)
	reqSSE.Header.Set("Origin", "http://evil.com")
	respSSE, err := http.DefaultClient.Do(reqSSE)
	if err != nil {
		t.Fatalf("SSE request failed: %v", err)
	}
	defer respSSE.Body.Close()

	if respSSE.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for SSE from untrusted origin, got %d", respSSE.StatusCode)
	}
}

// TestHTTPTransport_TrustedOrigins_Allowed ensures allowed IDE and local origins can complete requests.
func TestHTTPTransport_TrustedOrigins_Allowed(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv, SecurityOptions{
		AllowedOrigins: []string{"https://my-internal-ide.company.corp"},
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	trustedOrigins := []string{
		"vscode-webview://1234abcd-5678-ef01",
		"vscode-file://vscode-app",
		"cursor://anysession",
		"jetbrains://project",
		"antigravity://session-1",
		"http://localhost:3000",
		"http://127.0.0.1:5173",
		"https://my-internal-ide.company.corp",
	}

	for _, origin := range trustedOrigins {
		t.Run(origin, func(t *testing.T) {
			// Preflight OPTIONS
			reqOpt, _ := http.NewRequest("OPTIONS", server.URL+"/mcp/message", nil)
			reqOpt.Header.Set("Origin", origin)
			respOpt, err := http.DefaultClient.Do(reqOpt)
			if err != nil {
				t.Fatalf("OPTIONS request failed: %v", err)
			}
			defer respOpt.Body.Close()

			if respOpt.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 OK for OPTIONS with trusted origin %s, got %d", origin, respOpt.StatusCode)
			}
			if respOpt.Header.Get("Access-Control-Allow-Origin") != origin {
				t.Errorf("expected Access-Control-Allow-Origin %s, got %s", origin, respOpt.Header.Get("Access-Control-Allow-Origin"))
			}

			// POST Ping
			reqPost, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
			reqPost.Header.Set("Origin", origin)
			reqPost.Header.Set("Content-Type", "application/json")
			respPost, err := http.DefaultClient.Do(reqPost)
			if err != nil {
				t.Fatalf("POST request failed: %v", err)
			}
			defer respPost.Body.Close()

			if respPost.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 OK for POST with trusted origin %s, got %d", origin, respPost.StatusCode)
			}
			if respPost.Header.Get("Access-Control-Allow-Origin") != origin {
				t.Errorf("expected Access-Control-Allow-Origin %s, got %s", origin, respPost.Header.Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

// TestHTTPTransport_HostHeader_Validation ensures unexpected Host headers (DNS rebinding) are rejected.
func TestHTTPTransport_HostHeader_Validation(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Attack simulation: Host set to evil.com (DNS rebinding)
	reqEvil, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	reqEvil.Host = "evil.com:4000"
	reqEvil.Header.Set("Content-Type", "application/json")

	respEvil, err := http.DefaultClient.Do(reqEvil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respEvil.Body.Close()

	if respEvil.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for unexpected Host header, got %d", respEvil.StatusCode)
	}

	// 2. Normal Host: 127.0.0.1 or localhost -> 200 OK
	reqGood, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	reqGood.Host = "127.0.0.1:4000"
	reqGood.Header.Set("Content-Type", "application/json")

	respGood, err := http.DefaultClient.Do(reqGood)
	if err != nil {
		t.Fatalf("good request failed: %v", err)
	}
	defer respGood.Body.Close()

	if respGood.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for valid Host header, got %d", respGood.StatusCode)
	}
}

// TestHTTPTransport_AuthToken_Enforcement verifies capability / bearer token enforcement.
func TestHTTPTransport_AuthToken_Enforcement(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	secretToken := "krew-secret-test-token-12345"
	transport := NewHTTPTransport(srv, SecurityOptions{
		AuthToken: secretToken,
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	pingBody := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	// 1. Request without token -> 401 Unauthorized
	reqNoAuth, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
	reqNoAuth.Header.Set("Content-Type", "application/json")
	respNoAuth, err := http.DefaultClient.Do(reqNoAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respNoAuth.Body.Close()

	if respNoAuth.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without token, got %d", respNoAuth.StatusCode)
	}
	if !strings.Contains(respNoAuth.Header.Get("WWW-Authenticate"), "Bearer") {
		t.Errorf("expected WWW-Authenticate: Bearer header, got: %s", respNoAuth.Header.Get("WWW-Authenticate"))
	}

	// 2. Request with invalid token -> 401 Unauthorized
	reqBadAuth, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
	reqBadAuth.Header.Set("Content-Type", "application/json")
	reqBadAuth.Header.Set("Authorization", "Bearer invalid-token")
	respBadAuth, err := http.DefaultClient.Do(reqBadAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respBadAuth.Body.Close()

	if respBadAuth.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized with wrong token, got %d", respBadAuth.StatusCode)
	}

	// 3. Request with valid Authorization: Bearer <token> -> 200 OK
	reqGoodAuth, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
	reqGoodAuth.Header.Set("Content-Type", "application/json")
	reqGoodAuth.Header.Set("Authorization", "Bearer "+secretToken)
	respGoodAuth, err := http.DefaultClient.Do(reqGoodAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respGoodAuth.Body.Close()

	if respGoodAuth.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK with valid bearer token, got %d", respGoodAuth.StatusCode)
	}

	// 4. Request with valid X-MCP-Token header -> 200 OK
	reqHeaderAuth, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
	reqHeaderAuth.Header.Set("Content-Type", "application/json")
	reqHeaderAuth.Header.Set("X-MCP-Token", secretToken)
	respHeaderAuth, err := http.DefaultClient.Do(reqHeaderAuth)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer respHeaderAuth.Body.Close()

	if respHeaderAuth.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK with valid X-MCP-Token, got %d", respHeaderAuth.StatusCode)
	}
}

// TestHTTPTransport_RateLimiting verifies sliding-window rate limiting on the MCP transport.
func TestHTTPTransport_RateLimiting(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv, SecurityOptions{
		RateLimit:  3,
		RateWindow: 5 * time.Second,
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	pingBody := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	// Send 3 requests (within limit) -> all 200 OK
	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200 OK, got %d", i, resp.StatusCode)
		}
	}

	// 4th request -> 429 Too Many Requests
	reqExceed, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(pingBody))
	reqExceed.Header.Set("Content-Type", "application/json")
	respExceed, err := http.DefaultClient.Do(reqExceed)
	if err != nil {
		t.Fatalf("exceeded request failed: %v", err)
	}
	defer respExceed.Body.Close()

	if respExceed.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", respExceed.StatusCode)
	}
	if respExceed.Header.Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}
}
