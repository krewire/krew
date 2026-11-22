package dashboard_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/dashboard"
)

func TestDashboardServer_Endpoints(t *testing.T) {
	dir := t.TempDir()

	// 1. Create krewire.yaml with services
	cfgContent := `
project:
  name: test-hub
title: "Krewire Test Hub"
version: "1.0.0"
runtime:
  port: 8080
services:
  - name: auth-service
    port: 8081
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("failed to write krewire.yaml: %v", err)
	}

	// 2. Create pages, content, and handlers
	pagesDir := filepath.Join(dir, "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		t.Fatalf("failed to create pages dir: %v", err)
	}
	_ = os.WriteFile(filepath.Join(pagesDir, "index.kiw"), []byte("<h1>Home</h1>"), 0o644)
	_ = os.WriteFile(filepath.Join(pagesDir, "about.kiw"), []byte("<h1>About</h1>"), 0o644)

	contentDir := filepath.Join(dir, "content", "docs")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		t.Fatalf("failed to create content dir: %v", err)
	}
	_ = os.WriteFile(filepath.Join(contentDir, "guide.md"), []byte("# Guide"), 0o644)

	handlersDir := filepath.Join(dir, "internal", "handlers")
	if err := os.MkdirAll(handlersDir, 0o755); err != nil {
		t.Fatalf("failed to create handlers dir: %v", err)
	}
	_ = os.WriteFile(filepath.Join(handlersDir, "user.go"), []byte("package handlers"), 0o644)
	_ = os.WriteFile(filepath.Join(handlersDir, "user_test.go"), []byte("package handlers"), 0o644)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	srv := dashboard.NewServer(dir, cfg, "local", "127.0.0.1", "4000")
	handler := srv.Handler()

	// Subtest 1: GET / (HTML Dev Dashboard)
	t.Run("GET / HTML Dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("expected text/html content-type, got %s", ct)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Krewire Dev Dashboard") {
			t.Errorf("body missing dashboard header")
		}
		if !strings.Contains(body, "Krewire Test Hub") {
			t.Errorf("body missing project title")
		}
		if !strings.Contains(body, "panel-routes") {
			t.Errorf("body missing routes panel")
		}

		headReq := httptest.NewRequest(http.MethodHead, "/", nil)
		headRec := httptest.NewRecorder()
		handler.ServeHTTP(headRec, headReq)
		if headRec.Code != http.StatusOK {
			t.Errorf("expected status 200 for HEAD, got %d", headRec.Code)
		}
	})

	// Subtest 2: Method Not Allowed & 404
	t.Run("Invalid Methods and Paths", func(t *testing.T) {
		postReq := httptest.NewRequest(http.MethodPost, "/", nil)
		postRec := httptest.NewRecorder()
		handler.ServeHTTP(postRec, postReq)
		if postRec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 Method Not Allowed, got %d", postRec.Code)
		}

		notFoundReq := httptest.NewRequest(http.MethodGet, "/non-existent", nil)
		notFoundRec := httptest.NewRecorder()
		handler.ServeHTTP(notFoundRec, notFoundReq)
		if notFoundRec.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", notFoundRec.Code)
		}
	})

	// Subtest 3: GET /api/info
	t.Run("GET /api/info", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var info dashboard.ProjectInfo
		if err := json.Unmarshal(rec.Body.Bytes(), &info); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}
		if info.Name != "Krewire Test Hub" {
			t.Errorf("expected Name 'Krewire Test Hub', got %q", info.Name)
		}
		if info.Env != "local" {
			t.Errorf("expected Env 'local', got %q", info.Env)
		}
		if len(info.Services) < 2 {
			t.Errorf("expected at least 2 services (runtime + auth-service), got %d", len(info.Services))
		}
		if len(info.Routes) < 4 {
			t.Errorf("expected at least 4 routes, got %d", len(info.Routes))
		}

		routePaths := make(map[string]bool)
		for _, r := range info.Routes {
			routePaths[r.Path] = true
		}
		for _, expected := range []string{"/", "/about", "/docs/guide", "/api/user"} {
			if !routePaths[expected] {
				t.Errorf("expected route %q not found in routes", expected)
			}
		}
	})

	// Subtest 4: GET /api/metrics
	t.Run("GET /api/metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var m dashboard.RuntimeMetrics
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatalf("failed to decode metrics JSON: %v", err)
		}
		if m.Goroutines <= 0 {
			t.Errorf("expected positive goroutines, got %d", m.Goroutines)
		}
		if m.Timestamp == "" {
			t.Errorf("expected non-empty timestamp")
		}
	})

	// Subtest 5: GET /api/health
	t.Run("GET /api/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var h dashboard.HealthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
			t.Fatalf("failed to decode health JSON: %v", err)
		}
		if h.Status != "ok" {
			t.Errorf("expected status 'ok', got %q", h.Status)
		}
	})

	// Subtest 6: GET /api/events (SSE)
	t.Run("GET /api/events SSE Stream", func(t *testing.T) {
		ts := httptest.NewServer(handler)
		defer ts.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/events", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to connect to SSE: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
			t.Fatalf("expected Content-Type text/event-stream, got %s", ct)
		}

		buf := make([]byte, 512)
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "context") {
			t.Fatalf("failed reading SSE stream: %v", err)
		}
		chunk := string(buf[:n])
		if !strings.Contains(chunk, "event: pulse") {
			t.Errorf("expected 'event: pulse' in SSE response, got %q", chunk)
		}
	})
}

func TestDashboardServer_Lifecycle(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Title: "Lifecycle Test",
	}

	srv := dashboard.NewServer(dir, cfg, "local", "127.0.0.1", "0")
	if err := srv.Start(); err != nil {
		t.Fatalf("failed to start dashboard: %v", err)
	}

	addr := srv.Addr()
	if addr == "" || !strings.Contains(addr, "127.0.0.1:") {
		t.Fatalf("invalid listener address: %q", addr)
	}

	// Verify server responds
	resp, err := http.Get("http://" + addr + "/api/health")
	if err != nil {
		t.Fatalf("failed to reach dashboard: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Stop server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Errorf("failed to cleanly stop dashboard: %v", err)
	}
}
