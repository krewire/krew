package commands

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRegisterGateway(t *testing.T) {
	fs := flag.NewFlagSet("gateway", flag.ContinueOnError)
	RegisterGateway(fs)

	if fs.Lookup("port") == nil {
		t.Error("missing port flag")
	}
	if fs.Lookup("host") == nil {
		t.Error("missing host flag")
	}
	if fs.Lookup("routes") == nil {
		t.Error("missing routes flag")
	}
}

func TestRunGateway_WithRoutesFlag(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gateway upstream ok: " + r.URL.Path))
	}))
	defer upstream.Close()

	port := 18989
	fs := flag.NewFlagSet("gateway", flag.ContinueOnError)
	RegisterGateway(fs)
	_ = fs.Set("port", fmt.Sprintf("%d", port))
	_ = fs.Set("host", "127.0.0.1")
	_ = fs.Set("routes", fmt.Sprintf("/api=%s", upstream.URL))

	done := make(chan errs.ExitCode, 1)
	go func() {
		done <- RunGateway(fs)
	}()

	// Wait for server to start
	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := fmt.Sprintf("http://127.0.0.1:%d/api/test", port)
	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, err = client.Get(url)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("failed to query gateway: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "gateway upstream ok") {
		t.Errorf("unexpected body: %s", string(body))
	}

	// Trigger shutdown via signal
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case code := <-done:
		if code != errs.ExitCodeSuccess {
			t.Fatalf("expected success exit code, got %v", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for gateway shutdown")
	}
}

func TestRunGateway_ConfigRoutes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("config route ok"))
	}))
	defer upstream.Close()

	dir := t.TempDir()
	yamlContent := fmt.Sprintf(`project:
  name: test-gw
  kind: app
gateway:
  port: 18990
  host: 127.0.0.1
  routes:
    - path: /configured
      target_url: %s
      prefix: true
`, upstream.URL)

	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("gateway", flag.ContinueOnError)
	RegisterGateway(fs)

	done := make(chan errs.ExitCode, 1)
	go func() {
		done <- RunGateway(fs)
	}()

	// Query gateway
	client := &http.Client{Timeout: 500 * time.Millisecond}
	url := "http://127.0.0.1:18990/configured/subpath"
	var resp *http.Response
	var err error
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, err = client.Get(url)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("failed to query gateway: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Shutdown
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case code := <-done:
		if code != errs.ExitCodeSuccess {
			t.Fatalf("expected success, got %v", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for gateway shutdown")
	}
}
