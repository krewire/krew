package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/agent"
)

// HTTPServer exposes the Krew Agent via REST and Server-Sent Events (SSE).
type HTTPServer struct {
	agent *agent.Agent
	mux   *http.ServeMux
	mu    sync.RWMutex
}

// NewHTTPServer initializes an HTTP transport for an Agent.
func NewHTTPServer(ag *agent.Agent) *HTTPServer {
	s := &HTTPServer{
		agent: ag,
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

// Handler returns the underlying http.Handler.
func (s *HTTPServer) Handler() http.Handler {
	return s.mux
}

func (s *HTTPServer) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/agent/run", s.handleRun)
	s.mux.HandleFunc("/api/agent/chat", s.handleRun)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
		"engine": "krew-agent",
		"time":   time.Now().UTC(),
	})
}

type RunRequest struct {
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

func (s *HTTPServer) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Missing prompt field", http.StatusBadRequest)
		return
	}

	// 1. Streaming via SSE
	if req.Stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		eventChan := make(chan agent.Event, 20)

		// Create a run-specific agent configured with event channel
		opts := s.agent
		_ = opts

		go func() {
			defer close(eventChan)
			_, _ = s.agent.Run(r.Context(), req.Prompt)
		}()

		for ev := range eventChan {
			b, _ := json.Marshal(ev)
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(b))
			flusher.Flush()
		}
		return
	}

	// 2. Standard synchronous execution
	res, err := s.agent.Run(r.Context(), req.Prompt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// MountTo attaches the agent API routes under prefix to an existing http.ServeMux.
func MountTo(mux *http.ServeMux, prefix string, ag *agent.Agent) {
	server := NewHTTPServer(ag)
	trimmed := prefix
	if trimmed == "" {
		trimmed = "/"
	}
	mux.Handle(trimmed, http.StripPrefix(trimmed, server.Handler()))
}
