package transport

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/tools"
)

type contextKey string

const (
	tenantContextKey  contextKey = "krewire_tenant_id"
	userContextKey    contextKey = "krewire_user_id"
	sessionContextKey contextKey = "krewire_session"
)

// Session represents an authenticated user/tenant session.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Roles     []string  `json:"roles"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Session) IsExpired() bool {
	return !s.ExpiresAt.IsZero() && time.Now().After(s.ExpiresAt)
}

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
func SessionFromContext(ctx context.Context) *Session {
	if s, ok := ctx.Value(sessionContextKey).(*Session); ok {
		return s
	}
	return nil
}

// TokenVerifier verifies bearer tokens and returns the associated Session.
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*Session, error)
}

// SessionStore provides retrieval of Sessions by token.
type SessionStore interface {
	Get(ctx context.Context, token string) (*Session, error)
}

// SessionStoreVerifier adapts a SessionStore to the TokenVerifier interface.
type SessionStoreVerifier struct {
	Store SessionStore
}

func (v *SessionStoreVerifier) VerifyToken(ctx context.Context, token string) (*Session, error) {
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

// StaticTokenVerifier validates a bearer token against an expected secret using constant-time comparison.
type StaticTokenVerifier struct {
	ExpectedToken string
	Session       *Session
}

// NewStaticTokenVerifier creates a verifier that validates bearer tokens against expectedToken using subtle.ConstantTimeCompare.
func NewStaticTokenVerifier(expectedToken string, session *Session) *StaticTokenVerifier {
	if session == nil {
		session = &Session{
			ID:       "static-session",
			UserID:   "admin",
			TenantID: "default",
			Roles:    []string{"admin"},
		}
	}
	return &StaticTokenVerifier{
		ExpectedToken: expectedToken,
		Session:       session,
	}
}

func (v *StaticTokenVerifier) VerifyToken(_ context.Context, token string) (*Session, error) {
	if v.ExpectedToken == "" {
		return nil, errors.New("empty static token configured")
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(v.ExpectedToken)) != 1 {
		return nil, errors.New("invalid static bearer token")
	}
	return v.Session, nil
}

// DefaultMaxBodySize is the default ceiling for incoming request bodies (1MB).
const DefaultMaxBodySize int64 = 1 << 20

// HTTPServer exposes the Krew Agent via REST and Server-Sent Events (SSE).
type HTTPServer struct {
	agent          *agent.Agent
	mux            *http.ServeMux
	verifier       TokenVerifier
	noAuth         bool
	tenantBaseDir  string
	authorizer     agent.Authorizer
	auditLogger    agent.Auditor
	allowedOrigins []string
	maxBodySize    int64
	mu             sync.RWMutex
}

// ServerOption configures an HTTPServer instance.
type ServerOption func(*HTTPServer)

// WithMaxBodySize configures the maximum allowed request body size in bytes.
func WithMaxBodySize(size int64) ServerOption {
	return func(s *HTTPServer) {
		if size > 0 {
			s.maxBodySize = size
		}
	}
}

// WithAllowedOrigins configures the allowed CORS origins for agent endpoints.
func WithAllowedOrigins(origins ...string) ServerOption {
	return func(s *HTTPServer) { s.allowedOrigins = origins }
}

// WithVerifier configures a token verifier for Bearer authentication.
func WithVerifier(v TokenVerifier) ServerOption {
	return func(s *HTTPServer) { s.verifier = v }
}

// WithSessionStore configures SessionStore for Bearer authentication.
func WithSessionStore(store SessionStore) ServerOption {
	return func(s *HTTPServer) { s.verifier = &SessionStoreVerifier{Store: store} }
}

// WithStaticToken configures a static bearer token verifier with constant-time comparison.
func WithStaticToken(token string, session *Session) ServerOption {
	return func(s *HTTPServer) {
		s.verifier = NewStaticTokenVerifier(token, session)
	}
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
func WithRBAC(auth agent.Authorizer, auditLedger agent.Auditor) ServerOption {
	return func(s *HTTPServer) {
		s.authorizer = auth
		s.auditLogger = auditLedger
	}
}

// NewHTTPServer initializes an HTTP transport for an Agent.
func NewHTTPServer(ag *agent.Agent, opts ...ServerOption) *HTTPServer {
	s := &HTTPServer{
		agent:         ag,
		mux:           http.NewServeMux(),
		tenantBaseDir: "/var/krewire/tenants",
		maxBodySize:   DefaultMaxBodySize,
	}
	for _, opt := range opts {
		opt(s)
	}
	s.routes()
	return s
}

// Handler returns the underlying http.Handler wrapped with baseline defense-in-depth security headers.
func (s *HTTPServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		s.mux.ServeHTTP(w, r)
	})
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

func (s *HTTPServer) authenticate(r *http.Request) (context.Context, *Session, error) {
	if s.noAuth {
		return r.Context(), nil, nil
	}
	if s.verifier == nil {
		return nil, nil, errors.New("authentication verifier unconfigured: fail-closed")
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

func (s *HTTPServer) scopedAgent(sess *Session) *agent.Agent {
	if sess == nil || sess.TenantID == "" || s.tenantBaseDir == "" {
		return s.agent
	}
	tenantWorkspace := filepath.Join(s.tenantBaseDir, sess.TenantID)
	_ = os.MkdirAll(tenantWorkspace, 0o755)

	tenantTools := tools.DefaultWorkspaceTools(tenantWorkspace)
	opts := []agent.Option{agent.WithTools(tenantTools)}

	if s.authorizer != nil {
		approver := agent.NewAgentApprover(s.authorizer, sess.Roles, sess.TenantID, sess.UserID, s.auditLogger)
		opts = append(opts, agent.WithApprover(approver), agent.WithStrictApprovals(true))
	}
	return s.agent.Clone(opts...)
}

func (s *HTTPServer) applyCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}
	for _, allowed := range s.allowedOrigins {
		if allowed == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			return
		}
		if allowed == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			return
		}
	}
}

func (s *HTTPServer) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		s.applyCORS(w, r)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.applyCORS(w, r)

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

	limit := s.maxBodySize
	if limit <= 0 {
		limit = DefaultMaxBodySize
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Payload too large", http.StatusRequestEntityTooLarge)
			return
		}
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
	s.applyCORS(w, r)
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events := make(chan agent.Event, 32)
	done := make(chan struct{})

	go func() {
		defer close(events)
		defer close(done)
		_, _ = execAgent.Clone(agent.WithStreaming(true)).RunWithHandler(ctx, prompt, func(ev agent.Event) {
			select {
			case events <- ev:
			case <-ctx.Done():
			}
		})
	}()

	for {
		select {
		case <-ctx.Done():
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
