package mcp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SecurityOptions defines origin allowlists, host validation, authentication, and rate limiting
// for the local MCP HTTP transport.
type SecurityOptions struct {
	// AllowedOrigins lists explicit origins permitted to make cross-origin requests.
	// Supported IDE schemes (vscode-webview://, cursor://, jetbrains://, antigravity://)
	// and localhost origins are permitted by default.
	AllowedOrigins []string

	// AllowedHosts lists valid Host header values permitted to access the server.
	// Standard loopback addresses (127.0.0.1, localhost, [::1]) are permitted by default.
	AllowedHosts []string

	// AuthToken specifies the bearer token required for requests.
	// Mandatory when binding to non-loopback addresses; optional on loopback.
	AuthToken string

	// RateLimit specifies the maximum requests per window per IP (default: 120).
	RateLimit int

	// RateWindow specifies the sliding window duration for rate limiting (default: 1 minute).
	RateWindow time.Duration
}

// ipRateLimiter implements an in-memory sliding window rate limiter per client IP.
type ipRateLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
	limit   int
	window  time.Duration
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	if limit <= 0 {
		limit = 120
	}
	if window <= 0 {
		window = time.Minute
	}
	return &ipRateLimiter{
		records: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

func (rl *ipRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	timestamps := rl.records[ip]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.records[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.records[ip] = valid

	if len(rl.records) > 1000 {
		for k, v := range rl.records {
			if len(v) == 0 || v[len(v)-1].Before(cutoff) {
				delete(rl.records, k)
			}
		}
	}
	return true
}

// HTTPTransport coordinates SSE and HTTP message endpoints for an MCP Server with strict security controls.
type HTTPTransport struct {
	srv      *Server
	sec      SecurityOptions
	limiter  *ipRateLimiter
	mu       sync.RWMutex
	sessions map[string]chan *Response
}

// NewHTTPTransport creates an HTTP transport handler for the MCP Server with security controls.
func NewHTTPTransport(srv *Server, opts ...SecurityOptions) *HTTPTransport {
	var sec SecurityOptions
	if len(opts) > 0 {
		sec = opts[0]
	}
	if sec.RateLimit <= 0 {
		sec.RateLimit = 120
	}
	if sec.RateWindow <= 0 {
		sec.RateWindow = time.Minute
	}
	return &HTTPTransport{
		srv:      srv,
		sec:      sec,
		limiter:  newIPRateLimiter(sec.RateLimit, sec.RateWindow),
		sessions: make(map[string]chan *Response),
	}
}

// Mount attaches MCP routes (/sse and /message) under prefix to the given mux.
func (h *HTTPTransport) Mount(mux *http.ServeMux, prefix string) {
	prefix = stringsTrimSlash(prefix)
	ssePath := prefix + "/sse"
	msgPath := prefix + "/message"

	mux.HandleFunc(ssePath, h.handleSSE)
	mux.HandleFunc(msgPath, h.handleMessage)
}

func stringsTrimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func (h *HTTPTransport) registerSession(id string) chan *Response {
	ch := make(chan *Response, 32)
	h.mu.Lock()
	h.sessions[id] = ch
	h.mu.Unlock()
	return ch
}

func (h *HTTPTransport) unregisterSession(id string) {
	h.mu.Lock()
	if ch, ok := h.sessions[id]; ok {
		close(ch)
		delete(h.sessions, id)
	}
	h.mu.Unlock()
}

func (h *HTTPTransport) getSession(id string) (chan *Response, bool) {
	h.mu.RLock()
	ch, ok := h.sessions[id]
	h.mu.RUnlock()
	return ch, ok
}

// isHostAllowed validates the Host header against loopback addresses and explicitly configured hosts
// to protect against DNS rebinding attacks.
func (h *HTTPTransport) isHostAllowed(hostHeader string) bool {
	host := strings.TrimSpace(hostHeader)
	if host == "" {
		return false
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.Trim(strings.ToLower(host), "[]")

	// Standard loopback addresses are always permitted
	if host == "127.0.0.1" || host == "localhost" || host == "::1" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}

	// Explicitly configured allowed hosts
	for _, allowed := range h.sec.AllowedHosts {
		a := strings.TrimSpace(allowed)
		if parsedA, _, err := net.SplitHostPort(a); err == nil {
			a = parsedA
		}
		a = strings.Trim(strings.ToLower(a), "[]")
		if a == host {
			return true
		}
		// If bound to 0.0.0.0 or ::, any IP literal is safe from DNS rebinding
		if (a == "0.0.0.0" || a == "::") && net.ParseIP(host) != nil {
			return true
		}
	}

	return false
}

// isOriginAllowed validates browser Origin headers against supported IDE webview schemes,
// local loopback origins, and explicitly configured allowed origins.
func (h *HTTPTransport) isOriginAllowed(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		// Non-browser clients (curl, IDE language servers, native stdio proxies)
		// do not send an Origin header and are not subject to browser CORS vulnerabilities.
		return true
	}

	lower := strings.ToLower(origin)

	// 1. Supported IDE webview schemes
	if strings.HasPrefix(lower, "vscode-webview://") ||
		strings.HasPrefix(lower, "vscode-file://") ||
		strings.HasPrefix(lower, "cursor://") ||
		strings.HasPrefix(lower, "jetbrains://") ||
		strings.HasPrefix(lower, "antigravity://") {
		return true
	}

	// 2. Local browser loopback origins (e.g. http://localhost:3000, http://127.0.0.1:8080)
	if u, err := url.Parse(origin); err == nil && u.Host != "" {
		host := u.Hostname()
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return true
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return true
		}
	}

	// 3. User-configured allowed origins (supports exact match or prefix* matching)
	for _, allowed := range h.sec.AllowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(lower, strings.ToLower(prefix)) {
				return true
			}
		} else if strings.EqualFold(allowed, origin) {
			return true
		}
	}

	return false
}

// applyCORS checks the Host and Origin headers, rejecting untrusted origins or invalid hosts.
// For valid cross-origin requests, it echoes the specific trusted origin (never wildcard *).
func (h *HTTPTransport) applyCORS(w http.ResponseWriter, r *http.Request) bool {
	// Validate Host header to protect against DNS rebinding attacks
	if !h.isHostAllowed(r.Host) {
		http.Error(w, "invalid or unexpected Host header", http.StatusForbidden)
		return false
	}

	// Validate Origin header to protect against untrusted browser cross-origin requests
	origin := r.Header.Get("Origin")
	if origin != "" {
		if !h.isOriginAllowed(origin) {
			http.Error(w, fmt.Sprintf("untrusted origin: %s", origin), http.StatusForbidden)
			return false
		}
		// Echo exact matching origin; never emit wildcard "*"
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	return true
}

// isAuthorized verifies the Bearer capability token when authentication is configured.
func (h *HTTPTransport) isAuthorized(r *http.Request) bool {
	if h.sec.AuthToken == "" {
		return true
	}

	token := ""
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = strings.TrimSpace(authHeader[7:])
	} else if tok := r.Header.Get("X-MCP-Token"); tok != "" {
		token = tok
	} else if tok := r.URL.Query().Get("token"); tok != "" {
		token = tok
	}

	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(h.sec.AuthToken)) == 1
}

// checkRateLimit enforces per-IP sliding window rate limiting.
func (h *HTTPTransport) checkRateLimit(w http.ResponseWriter, r *http.Request) bool {
	if h.limiter == nil {
		return true
	}
	clientIP := r.RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		clientIP = ip
	}
	if !h.limiter.Allow(clientIP) {
		w.Header().Set("Retry-After", "60")
		http.Error(w, "rate limit exceeded: too many requests", http.StatusTooManyRequests)
		return false
	}
	return true
}

func (h *HTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	if !h.applyCORS(w, r) {
		return
	}

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-MCP-Token, X-MCP-Session-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.isAuthorized(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="krew-mcp"`)
		http.Error(w, "unauthorized: valid authentication token required", http.StatusUnauthorized)
		return
	}

	if !h.checkRateLimit(w, r) {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	sessionID := generateSessionID()
	respCh := h.registerSession(sessionID)
	defer h.unregisterSession(sessionID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Send endpoint event for MCP clients
	endpointURI := fmt.Sprintf("/mcp/message?sessionId=%s", sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURI)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-respCh:
			if !ok {
				return
			}
			if resp == nil {
				continue
			}
			b, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(b))
			flusher.Flush()
		}
	}
}

func (h *HTTPTransport) handleMessage(w http.ResponseWriter, r *http.Request) {
	if !h.applyCORS(w, r) {
		return
	}

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-MCP-Token, X-MCP-Session-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.isAuthorized(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="krew-mcp"`)
		http.Error(w, "unauthorized: valid authentication token required", http.StatusUnauthorized)
		return
	}

	if !h.checkRateLimit(w, r) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(&Response{
			JSONRPC: "2.0",
			Error: &Error{
				Code:    CodeParseError,
				Message: fmt.Sprintf("invalid JSON payload: %v", err),
			},
		})
		return
	}

	resp := h.srv.HandleRequest(r.Context(), req)

	// Check if this request is part of an active SSE session
	sessionID := r.Header.Get("X-MCP-Session-ID")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}

	if sessionID != "" {
		if ch, ok := h.getSession(sessionID); ok {
			if resp != nil {
				select {
				case ch <- resp:
				default:
					// buffer full, avoid blocking
				}
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}

	// Direct synchronous HTTP response
	w.Header().Set("Content-Type", "application/json")
	if resp != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
