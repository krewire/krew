package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/auth/pkg/identity"
	"github.com/krewire/auth/pkg/rbac"
	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krew/pkg/transport"
	"github.com/krewire/krewire/pkg/ai"
)

type mockTokenVerifier struct {
	sessions map[string]*identity.Session
}

func newMockTokenVerifier() *mockTokenVerifier {
	return &mockTokenVerifier{
		sessions: make(map[string]*identity.Session),
	}
}

func (m *mockTokenVerifier) AddSession(token string, sess *identity.Session) {
	m.sessions[token] = sess
}

func (m *mockTokenVerifier) VerifyToken(_ context.Context, token string) (*identity.Session, error) {
	sess, ok := m.sessions[token]
	if !ok {
		return nil, errors.New("invalid or expired token")
	}
	return sess, nil
}

func TestHTTPServer_Auth_UnauthorizedReturns401(t *testing.T) {
	verifier := newMockTokenVerifier()
	srv := transport.NewHTTPServer(setupTestAgent(), transport.WithVerifier(verifier))

	// Case 1: Missing Authorization header
	{
		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hello"}`))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for missing token, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// Case 2: Invalid bearer token
	{
		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hello"}`))
		req.Header.Set("Authorization", "Bearer invalid-token-xyz")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for invalid token, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// Case 3: Valid bearer token succeeds
	{
		verifier.AddSession("valid-token", &identity.Session{
			ID:        "sess_1",
			TenantID:  "tenant_abc",
			UserID:    "user_123",
			Roles:     []string{"developer"},
			ExpiresAt: time.Now().Add(1 * time.Hour),
		})

		req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hello"}`))
		req.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for valid token, got %d: %s", rec.Code, rec.Body.String())
		}
	}
}

func TestHTTPServer_Auth_NoAuthBypass(t *testing.T) {
	verifier := newMockTokenVerifier()
	srv := transport.NewHTTPServer(setupTestAgent(),
		transport.WithVerifier(verifier),
		transport.WithNoAuth(true),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"hello"}`))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when WithNoAuth(true), got %d: %s", rec.Code, rec.Body.String())
	}
}

// mockPathEscapeProvider simulates agent calling read_file on prohibited paths
type mockPathEscapeProvider struct {
	escapePath string
}

func (p *mockPathEscapeProvider) Name() string { return "escape-mock" }
func (p *mockPathEscapeProvider) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	return &ai.CompletionResponse{
		Content: "Attempting to read path",
		ToolCalls: []ai.ToolCall{
			{
				ID:        "call_read",
				Name:      "read_file",
				Arguments: `{"path":"` + p.escapePath + `"}`,
			},
		},
	}, nil
}
func (p *mockPathEscapeProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestHTTPServer_Auth_WorkspaceSandboxingAndPathTrapped(t *testing.T) {
	tenantBase := t.TempDir()
	tenant1Dir := filepath.Join(tenantBase, "tenant_1")
	tenant2Dir := filepath.Join(tenantBase, "tenant_2")
	_ = os.MkdirAll(tenant1Dir, 0o755)
	_ = os.MkdirAll(tenant2Dir, 0o755)

	secretFile := filepath.Join(tenant2Dir, "secret.key")
	if err := os.WriteFile(secretFile, []byte("super-secret-data"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}

	verifier := newMockTokenVerifier()
	verifier.AddSession("t1-token", &identity.Session{
		ID:        "sess_t1",
		TenantID:  "tenant_1",
		UserID:    "user_t1",
		Roles:     []string{"developer"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	testCases := []struct {
		name       string
		targetPath string
	}{
		{"Host root escape", "/etc/passwd"},
		{"Relative path traversal to sibling tenant", "../tenant_2/secret.key"},
		{"Absolute path to sibling tenant", secretFile},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			escapeMock := &mockPathEscapeProvider{escapePath: tc.targetPath}
			ag, err := agent.New(
				agent.WithProvider(escapeMock),
				agent.WithMaxSteps(2),
			)
			if err != nil {
				t.Fatalf("agent.New: %v", err)
			}

			eval := rbac.NewEvaluator()
			srv := transport.NewHTTPServer(ag,
				transport.WithVerifier(verifier),
				transport.WithTenantBaseDir(tenantBase),
				transport.WithRBAC(eval, nil),
			)

			req := httptest.NewRequest(http.MethodPost, "/api/agent/run", strings.NewReader(`{"prompt":"read file"}`))
			req.Header.Set("Authorization", "Bearer t1-token")
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)

			// The response should indicate the tool result trapped the path escape error
			var res agent.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("parse response: %v, body: %s", err, rec.Body.String())
			}

			// Verify via SafePath directly that the traversal is rejected with ErrPathEscape
			_, err = tools.SafePath(tenant1Dir, tc.targetPath)
			if err == nil {
				t.Fatalf("expected SafePath to fail for %s", tc.targetPath)
			}
			if !errors.Is(err, tools.ErrPathEscape) {
				t.Errorf("expected ErrPathEscape, got: %v", err)
			}
		})
	}
}

type contextCheckingProvider struct {
	tenantObserved string
	userObserved   string
}

func (p *contextCheckingProvider) Name() string { return "context-checker" }
func (p *contextCheckingProvider) Complete(ctx context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	p.tenantObserved = transport.TenantFromContext(ctx)
	p.userObserved = transport.UserFromContext(ctx)
	return &ai.CompletionResponse{Content: "context confirmed", FinishReason: "stop"}, nil
}
func (p *contextCheckingProvider) Stream(ctx context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	p.tenantObserved = transport.TenantFromContext(ctx)
	p.userObserved = transport.UserFromContext(ctx)
	ch := make(chan ai.StreamChunk, 2)
	ch <- ai.StreamChunk{Delta: "token"}
	ch <- ai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestHTTPServer_Auth_SSEStreamingRetainsAuthContext(t *testing.T) {
	checker := &contextCheckingProvider{}
	ag, err := agent.New(
		agent.WithProvider(checker),
		agent.WithStreaming(true),
	)
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}

	verifier := newMockTokenVerifier()
	verifier.AddSession("stream-token", &identity.Session{
		ID:        "sess_stream",
		TenantID:  "tenant_stream_99",
		UserID:    "user_stream_42",
		Roles:     []string{"developer"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	srv := transport.NewHTTPServer(ag, transport.WithVerifier(verifier))

	frames := streamSSE(t, srv.Handler(), "/api/agent/run", `{"prompt":"test stream","stream":true}`, map[string]string{
		"Authorization": "Bearer stream-token",
	})

	if len(frames) == 0 {
		t.Fatal("expected SSE frames")
	}

	if checker.tenantObserved != "tenant_stream_99" {
		t.Errorf("expected TenantID tenant_stream_99 in stream context, got %q", checker.tenantObserved)
	}
	if checker.userObserved != "user_stream_42" {
		t.Errorf("expected UserID user_stream_42 in stream context, got %q", checker.userObserved)
	}
}
