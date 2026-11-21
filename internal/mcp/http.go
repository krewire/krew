package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// HTTPTransport coordinates SSE and HTTP message endpoints for an MCP Server.
type HTTPTransport struct {
	srv      *Server
	mu       sync.RWMutex
	sessions map[string]chan *Response
}

// NewHTTPTransport creates an HTTP transport handler for the MCP Server.
func NewHTTPTransport(srv *Server) *HTTPTransport {
	return &HTTPTransport{
		srv:      srv,
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

func (h *HTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	sessionID := r.URL.Query().Get("sessionId")
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
