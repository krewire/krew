package commands

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestExecuteBuildDeps_Success(t *testing.T) {
	dir := t.TempDir()
	out1 := filepath.Join(dir, "step1.txt")
	out2 := filepath.Join(dir, "step2.txt")

	deps := []string{
		"echo step1 > " + filepath.Base(out1),
		"echo step2 > " + filepath.Base(out2),
	}

	code := executeBuildDeps(dir, deps, os.Environ())
	if code != errs.ExitCodeSuccess {
		t.Fatalf("executeBuildDeps() = %v, want ExitCodeSuccess (0)", code)
	}

	if _, err := os.Stat(out1); err != nil {
		t.Fatalf("step1 output file not found: %v", err)
	}
	if _, err := os.Stat(out2); err != nil {
		t.Fatalf("step2 output file not found: %v", err)
	}
}

func TestExecuteBuildDeps_AbortsOnFailure(t *testing.T) {
	dir := t.TempDir()
	out2 := filepath.Join(dir, "step2.txt")

	deps := []string{
		"exit 42",
		"echo step2 > " + filepath.Base(out2),
	}

	code := executeBuildDeps(dir, deps, os.Environ())
	if code != errs.ExitCode(42) && code != errs.ExitCodeFailure {
		t.Fatalf("executeBuildDeps() = %v, want non-zero exit code", code)
	}

	// step 2 should NOT have run
	if _, err := os.Stat(out2); !os.IsNotExist(err) {
		t.Errorf("step2 should have been aborted, but file was created")
	}
}

func TestExecuteBuildDeps_InheritsEnvironment(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "env.txt")

	env := append(os.Environ(), "CUSTOM_TEST_VAR=krewire_power")
	deps := []string{
		"echo $CUSTOM_TEST_VAR > " + filepath.Base(outFile),
	}

	code := executeBuildDeps(dir, deps, env)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("executeBuildDeps() = %v, want ExitCodeSuccess (0)", code)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Errorf("expected env.txt to have content")
	}
}

func TestDetectRuntimeBuildSuggestions_NodeAndComposer(t *testing.T) {
	dir := t.TempDir()

	// Initial empty dir
	if suggestions := detectRuntimeBuildSuggestions(dir); len(suggestions) != 0 {
		t.Fatalf("expected 0 suggestions for empty dir, got %v", suggestions)
	}

	// Add package.json with scripts.build
	pkgJSON := `{"scripts": {"build": "vite build"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	sug := detectRuntimeBuildSuggestions(dir)
	if len(sug) != 1 || sug[0].Command != "npm run build" {
		t.Fatalf("expected npm run build suggestion, got %v", sug)
	}

	// Add pnpm-lock.yaml: should suggest pnpm
	if err := os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	sugPnpm := detectRuntimeBuildSuggestions(dir)
	if len(sugPnpm) != 1 || sugPnpm[0].Command != "pnpm run build" {
		t.Fatalf("expected pnpm run build suggestion, got %v", sugPnpm)
	}

	// Add composer.json
	if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	sugBoth := detectRuntimeBuildSuggestions(dir)
	if len(sugBoth) != 2 {
		t.Fatalf("expected 2 suggestions, got %v", sugBoth)
	}
}

func TestRunBuild_SkipDepsAndDepsOnly(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	// Create krewire.yaml with build.depends_on and content/
	configContent := `
name: test-orchestrator
build:
  depends_on:
    - "echo dep-executed > dep.txt"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "content", "intro.md"), []byte("# Intro"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test 1: deps-only flag
	fsDepsOnly := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fsDepsOnly)
	_ = fsDepsOnly.Set("deps-only", "true")

	code := RunBuild(fsDepsOnly)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild(deps-only) code = %v, want ExitCodeSuccess (0)", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "dep.txt")); err != nil {
		t.Fatalf("expected dep.txt to be created by deps-only")
	}

	// Remove dep.txt
	_ = os.Remove(filepath.Join(dir, "dep.txt"))

	// Test 2: skip-deps flag
	fsSkipDeps := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fsSkipDeps)
	_ = fsSkipDeps.Set("skip-deps", "true")

	code = RunBuild(fsSkipDeps)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild(skip-deps) code = %v, want ExitCodeSuccess (0)", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "dep.txt")); !os.IsNotExist(err) {
		t.Errorf("dep.txt should not have been created when skip-deps is true")
	}
}

func TestVirtualRootFS_MultiRootResolution(t *testing.T) {
	root1 := t.TempDir()
	root2 := t.TempDir()

	// File in root1
	if err := os.WriteFile(filepath.Join(root1, "app.js"), []byte("console.log('root1');"), 0o644); err != nil {
		t.Fatal(err)
	}
	// File in root2
	if err := os.WriteFile(filepath.Join(root2, "style.css"), []byte("body { color: red; }"), 0o644); err != nil {
		t.Fatal(err)
	}

	handler := virtualRootFS(root1, root2)

	// Test request for app.js from root1
	req1 := httptest.NewRequest("GET", "/app.js", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK || rec1.Body.String() != "console.log('root1');" {
		t.Fatalf("request /app.js got %d %q", rec1.Code, rec1.Body.String())
	}

	// Test request for style.css from root2
	req2 := httptest.NewRequest("GET", "/style.css", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK || rec2.Body.String() != "body { color: red; }" {
		t.Fatalf("request /style.css got %d %q", rec2.Code, rec2.Body.String())
	}

	// Test 404 for non-existent file
	req3 := httptest.NewRequest("GET", "/missing.png", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusNotFound {
		t.Fatalf("request /missing.png got %d, want 404", rec3.Code)
	}
}
