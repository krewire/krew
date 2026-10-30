package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	// Streaming is opted in either via the body flag or the standard Accept
	// negotiation header.
	if req.Stream || strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		s.streamRun(w, r, req.Prompt)
		return
	}

	// Standard synchronous execution
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

// streamRun executes the agent with token-level streaming enabled and relays
// every event as a discrete Server-Sent Event.
func (s *HTTPServer) streamRun(w http.ResponseWriter, r *http.Request, prompt string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events := make(chan agent.Event, 32)
	done := make(chan struct{})

	go func() {
		defer close(events)
		defer close(done)
		_, _ = s.agent.Clone(agent.WithStreaming(true)).RunWithHandler(r.Context(), prompt, func(ev agent.Event) {
			// Never block the agent loop on a slow or vanished client:
			// dropping events on context cancellation interrupts the run.
			select {
			case events <- ev:
			case <-ctx.Done():
			}
		})
	}()

	for {
		select {
		case <-ctx.Done():
			// Client vanished: wait for the producer goroutine to settle.
			<-done
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			b, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(b))
			flusher.Flush()
		}
	}
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
