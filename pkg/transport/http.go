package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/krewire/auth/pkg/audit"
	"github.com/krewire/auth/pkg/identity"
	"github.com/krewire/auth/pkg/rbac"
	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/tools"
)

type contextKey string

const (
	tenantContextKey  contextKey = "krewire_tenant_id"
	userContextKey    contextKey = "krewire_user_id"
	sessionContextKey contextKey = "krewire_session"
)

// TenantFromContext extracts the active TenantID from request context.
func TenantFromContext(ctx context.Context) string {
	if t, ok := ctx.Value(tenantContextKey).(string); ok {
		return t
	}
	return ""
}

// UserFromContext extracts the active UserID from request context.
func UserFromContext(ctx context.Context) string {
	if u, ok := ctx.Value(userContextKey).(string); ok {
		return u
	}
	return ""
}

// SessionFromContext extracts the authenticated Session from request context.
func SessionFromContext(ctx context.Context) *identity.Session {
	if s, ok := ctx.Value(sessionContextKey).(*identity.Session); ok {
		return s
	}
	return nil
}

// TokenVerifier verifies bearer tokens and returns the associated identity.Session.
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*identity.Session, error)
}

// SessionStoreVerifier adapts an identity.SessionStore to the TokenVerifier interface.
type SessionStoreVerifier struct {
	Store identity.SessionStore
}

func (v *SessionStoreVerifier) VerifyToken(ctx context.Context, token string) (*identity.Session, error) {
	if v.Store == nil {
		return nil, errors.New("no session store configured")
	}
	sess, err := v.Store.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if sess.IsExpired() {
		return nil, errors.New("session expired")
	}
	return sess, nil
}

// HTTPServer exposes the Krew Agent via REST and Server-Sent Events (SSE).
type HTTPServer struct {
	agent         *agent.Agent
	mux           *http.ServeMux
	verifier      TokenVerifier
	noAuth        bool
	tenantBaseDir string
	evaluator     *rbac.Evaluator
	auditLedger   audit.AuditLogger
	mu            sync.RWMutex
}

// ServerOption configures an HTTPServer instance.
type ServerOption func(*HTTPServer)

// WithVerifier configures a token verifier for Bearer authentication.
func WithVerifier(v TokenVerifier) ServerOption {
	return func(s *HTTPServer) { s.verifier = v }
}

// WithSessionStore configures identity.SessionStore for Bearer authentication.
func WithSessionStore(store identity.SessionStore) ServerOption {
	return func(s *HTTPServer) { s.verifier = &SessionStoreVerifier{Store: store} }
}

// WithNoAuth bypasses token authentication for local development and loopback usage.
func WithNoAuth(noAuth bool) ServerOption {
	return func(s *HTTPServer) { s.noAuth = noAuth }
}

// WithTenantBaseDir specifies the root directory for tenant workspace jails.
func WithTenantBaseDir(dir string) ServerOption {
	return func(s *HTTPServer) { s.tenantBaseDir = dir }
}

// WithRBAC enables role-based tool gating and audit logging.
func WithRBAC(eval *rbac.Evaluator, auditLedger audit.AuditLogger) ServerOption {
	return func(s *HTTPServer) {
		s.evaluator = eval
		s.auditLedger = auditLedger
	}
}

// NewHTTPServer initializes an HTTP transport for an Agent.
func NewHTTPServer(ag *agent.Agent, opts ...ServerOption) *HTTPServer {
	s := &HTTPServer{
		agent:         ag,
		mux:           http.NewServeMux(),
		tenantBaseDir: "/var/krewire/tenants",
	}
	for _, opt := range opts {
		opt(s)
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

func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
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

func (s *HTTPServer) authenticate(r *http.Request) (context.Context, *identity.Session, error) {
	if s.noAuth || s.verifier == nil {
		return r.Context(), nil, nil
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil, errors.New("missing or malformed Authorization: Bearer token")
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		return nil, nil, errors.New("empty bearer token")
	}

	sess, err := s.verifier.VerifyToken(r.Context(), token)
	if err != nil {
		return nil, nil, fmt.Errorf("unauthorized: %w", err)
	}

	ctx := context.WithValue(r.Context(), sessionContextKey, sess)
	ctx = context.WithValue(ctx, tenantContextKey, sess.TenantID)
	ctx = context.WithValue(ctx, userContextKey, sess.UserID)
	return ctx, sess, nil
}

func (s *HTTPServer) scopedAgent(sess *identity.Session) *agent.Agent {
	if sess == nil || sess.TenantID == "" || s.tenantBaseDir == "" {
		return s.agent
	}
	tenantWorkspace := filepath.Join(s.tenantBaseDir, sess.TenantID)
	_ = os.MkdirAll(tenantWorkspace, 0o755)

	tenantTools := tools.DefaultWorkspaceTools(tenantWorkspace)
	opts := []agent.Option{agent.WithTools(tenantTools)}

	if s.evaluator != nil {
		approver := agent.NewAgentApprover(s.evaluator, sess.Roles, sess.TenantID, sess.UserID, s.auditLedger)
		opts = append(opts, agent.WithApprover(approver), agent.WithStrictApprovals(true))
	}
	return s.agent.Clone(opts...)
}

func (s *HTTPServer) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, sess, err := s.authenticate(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	r = r.WithContext(ctx)

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Missing prompt field", http.StatusBadRequest)
		return
	}

	execAgent := s.scopedAgent(sess)

	// Streaming is opted in either via the body flag or the standard Accept
	// negotiation header.
	if req.Stream || strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		s.streamRun(w, r, execAgent, req.Prompt)
		return
	}

	// Standard synchronous execution
	res, err := execAgent.Run(r.Context(), req.Prompt)
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
func (s *HTTPServer) streamRun(w http.ResponseWriter, r *http.Request, execAgent *agent.Agent, prompt string) {
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
		_, _ = execAgent.Clone(agent.WithStreaming(true)).RunWithHandler(ctx, prompt, func(ev agent.Event) {
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
func MountTo(mux *http.ServeMux, prefix string, ag *agent.Agent, opts ...ServerOption) {
	server := NewHTTPServer(ag, opts...)
	trimmed := prefix
	if trimmed == "" {
		trimmed = "/"
	}
	mux.Handle(trimmed, http.StripPrefix(trimmed, server.Handler()))
}
