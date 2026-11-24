package mcp

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPTransport_MaxRequestBodySize_HTTP413 tests that oversized payloads
// are rejected with HTTP 413 without reading beyond the limit.
func TestHTTPTransport_MaxRequestBodySize_HTTP413(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv, SecurityOptions{
		MaxRequestBodySize: 256, // small 256-byte limit for testing
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Small request within limit -> 200 OK
	smallBody := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	respSmall, err := http.Post(server.URL+"/mcp/message", "application/json", strings.NewReader(smallBody))
	if err != nil {
		t.Fatalf("small request failed: %v", err)
	}
	defer respSmall.Body.Close()
	if respSmall.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for small payload, got %d", respSmall.StatusCode)
	}

	// 2. Oversized payload exceeding 256 bytes -> HTTP 413 Payload Too Large
	largePayload := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"ping","params":{"padding":"%s"}}`, strings.Repeat("A", 500))
	respLarge, err := http.Post(server.URL+"/mcp/message", "application/json", strings.NewReader(largePayload))
	if err != nil {
		t.Fatalf("large request failed: %v", err)
	}
	defer respLarge.Body.Close()

	if respLarge.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected HTTP 413 StatusRequestEntityTooLarge, got %d", respLarge.StatusCode)
	}
}

// TestHTTPTransport_UnsupportedMethodsAndContentType verifies that non-standard methods
// and non-JSON content types are rejected with 405 and 415 respectively.
func TestHTTPTransport_UnsupportedMethodsAndContentType(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. Unsupported methods on /mcp/message -> 405
	for _, method := range []string{"GET", "PUT", "DELETE", "PATCH"} {
		req, _ := http.NewRequest(method, server.URL+"/mcp/message", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %s failed: %v", method, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 Method Not Allowed for %s /mcp/message, got %d", method, resp.StatusCode)
		}
	}

	// 2. Unsupported methods on /mcp/sse -> 405
	for _, method := range []string{"POST", "PUT", "DELETE"} {
		req, _ := http.NewRequest(method, server.URL+"/mcp/sse", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %s failed: %v", method, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 Method Not Allowed for %s /mcp/sse, got %d", method, resp.StatusCode)
		}
	}

	// 3. Unsupported Content-Type on POST /mcp/message -> 415
	for _, ct := range []string{"text/plain", "application/x-www-form-urlencoded", "multipart/form-data"} {
		req, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Header.Set("Content-Type", ct)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request with content-type %s failed: %v", ct, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnsupportedMediaType {
			t.Errorf("expected 415 StatusUnsupportedMediaType for %s, got %d", ct, resp.StatusCode)
		}
	}

	// 4. Valid JSON with charset -> 200 OK
	reqValid, _ := http.NewRequest("POST", server.URL+"/mcp/message", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	reqValid.Header.Set("Content-Type", "application/json; charset=utf-8")
	respValid, err := http.DefaultClient.Do(reqValid)
	if err != nil {
		t.Fatalf("valid request failed: %v", err)
	}
	defer respValid.Body.Close()
	if respValid.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for application/json; charset=utf-8, got %d", respValid.StatusCode)
	}
}

// TestHTTPTransport_CustomMountPrefix verifies that a custom mount prefix
// (e.g. /custom/api/mcp) is correctly used in the SSE endpoint event.
func TestHTTPTransport_CustomMountPrefix(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv)
	customPrefix := "/custom/api/mcp"
	transport.Mount(mux, customPrefix)

	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+customPrefix+"/sse", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE connect failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	var endpointLine string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("error reading SSE response: %v", err)
		}
		if strings.HasPrefix(line, "data: ") {
			endpointLine = strings.TrimSpace(line)
			break
		}
	}

	expectedPrefix := fmt.Sprintf("data: %s/message?sessionId=", customPrefix)
	if !strings.HasPrefix(endpointLine, expectedPrefix) {
		t.Errorf("expected endpoint line starting with %q, got %q", expectedPrefix, endpointLine)
	}
}

// TestHTTPTransport_QueueSaturation_Backpressure verifies backpressure when the SSE response
// queue is saturated; it must return 503 Service Unavailable with Retry-After rather than dropping silently.
func TestHTTPTransport_QueueSaturation_Backpressure(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv, SecurityOptions{
		QueueSize: 2, // Tiny buffer: holds 2 items
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	sessionID := "test-backpressure-session"
	ch, ok := transport.registerSession(sessionID)
	if !ok {
		t.Fatalf("failed to register session")
	}
	defer transport.unregisterSession(sessionID)

	msgBody := `{"jsonrpc":"2.0","id":1,"method":"ping"}`

	// 1. Fill buffer with 2 items -> both 202 Accepted
	for i := 1; i <= 2; i++ {
		r, err := http.Post(server.URL+fmt.Sprintf("/mcp/message?sessionId=%s", sessionID), "application/json", strings.NewReader(msgBody))
		if err != nil {
			t.Fatalf("message %d failed: %v", i, err)
		}
		r.Body.Close()
		if r.StatusCode != http.StatusAccepted {
			t.Fatalf("expected 202 Accepted for message %d, got %d", i, r.StatusCode)
		}
	}

	// Channel buffer is now exactly full (2 items in buffer)
	if len(ch) != 2 {
		t.Fatalf("expected channel buffer to have 2 items, got %d", len(ch))
	}

	// 2. 3rd item should hit saturation -> returns 503 Service Unavailable with Retry-After
	rSaturated, err := http.Post(server.URL+fmt.Sprintf("/mcp/message?sessionId=%s", sessionID), "application/json", strings.NewReader(msgBody))
	if err != nil {
		t.Fatalf("saturated message post failed: %v", err)
	}
	defer rSaturated.Body.Close()

	if rSaturated.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable for saturated queue, got %d", rSaturated.StatusCode)
	}
	if rSaturated.Header.Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header on queue saturation response")
	}

	// 3. Drain 1 item from channel -> now 4th item can be accepted
	<-ch

	rAccepted, err := http.Post(server.URL+fmt.Sprintf("/mcp/message?sessionId=%s", sessionID), "application/json", strings.NewReader(msgBody))
	if err != nil {
		t.Fatalf("drained message post failed: %v", err)
	}
	defer rAccepted.Body.Close()
	if rAccepted.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted after draining 1 item, got %d", rAccepted.StatusCode)
	}

	// 4. Post with invalid session ID -> 404 Not Found
	rNotFound, err := http.Post(server.URL+"/mcp/message?sessionId=non-existent-12345", "application/json", strings.NewReader(msgBody))
	if err != nil {
		t.Fatalf("not-found message post failed: %v", err)
	}
	defer rNotFound.Body.Close()
	if rNotFound.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for invalid sessionId, got %d", rNotFound.StatusCode)
	}
}

// TestHTTPTransport_SessionAdmissionLimit verifies that concurrent SSE sessions are bounded
// by MaxSessions and that disconnecting releases session resources.
func TestHTTPTransport_SessionAdmissionLimit(t *testing.T) {
	srv, _ := setupTestServer(t)
	mux := http.NewServeMux()
	transport := NewHTTPTransport(srv, SecurityOptions{
		MaxSessions: 2, // Cap at 2 concurrent sessions
	})
	transport.Mount(mux, "/mcp")

	server := httptest.NewServer(mux)
	defer server.Close()

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	req1, _ := http.NewRequestWithContext(ctx1, "GET", server.URL+"/mcp/sse", nil)
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("session 1 failed: %v", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("session 1: expected 200 OK, got %d", resp1.StatusCode)
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	req2, _ := http.NewRequestWithContext(ctx2, "GET", server.URL+"/mcp/sse", nil)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("session 2 failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("session 2: expected 200 OK, got %d", resp2.StatusCode)
	}

	// 3rd session exceeds limit -> 503 Service Unavailable
	req3, _ := http.NewRequest("GET", server.URL+"/mcp/sse", nil)
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("session 3 request failed: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable for exceeding MaxSessions, got %d", resp3.StatusCode)
	}

	// Close session 1 -> resources must be freed
	cancel1()
	time.Sleep(50 * time.Millisecond) // Allow unregister goroutine to complete

	// Now 4th session should succeed
	ctx4, cancel4 := context.WithCancel(context.Background())
	defer cancel4()
	req4, _ := http.NewRequestWithContext(ctx4, "GET", server.URL+"/mcp/sse", nil)
	resp4, err := http.DefaultClient.Do(req4)
	if err != nil {
		t.Fatalf("session 4 connect failed: %v", err)
	}
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusOK {
		t.Fatalf("expected session 4 to succeed after session 1 closed, got %d", resp4.StatusCode)
	}
}
