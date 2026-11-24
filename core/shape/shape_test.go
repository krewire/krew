package shape

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectAppByRootMainGo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "main.go", `package main

func main() {}
`)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities.HasApp || got.Marker != "main.go" {
		t.Errorf("got %v (%s), want HasApp=true (main.go)", got.Capabilities.HasApp, got.Marker)
	}
}

func TestDetectAppByCmdPackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "cmd/server/main.go", `package main

func main() {}
`)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities.HasApp || got.Marker != "cmd/*" {
		t.Errorf("got %v (%s), want HasApp=true (cmd/*)", got.Capabilities.HasApp, got.Marker)
	}
}

func TestDetectNonMainGoIsNotApp(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/lib\n")
	writeFile(t, dir, "main.go", `package lib

func Main() {}
`)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Capabilities.HasApp {
		t.Errorf("got HasApp=true, want false")
	}
}

func TestDetectSiteByPagesDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pages/index.kiw", "---\ntitle: Hi\n---\n<h1>Hi</h1>\n")
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities.HasPages || got.Marker != "pages/" {
		t.Errorf("got %v (%s), want HasPages=true (pages/)", got.Capabilities.HasPages, got.Marker)
	}
}

func TestDetectBook(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "content/index.md", "# Hi\n")
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities.HasBook || got.Marker != "content/" {
		t.Errorf("got %v (%s), want HasBook=true (content/)", got.Capabilities.HasBook, got.Marker)
	}
}

func TestDetectPreInitBootstrap(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/starter\n\ngo 1.27.1\n")
	writeFile(t, dir, "krewire.yaml", "project:\n  name: starter\n")
	writeFile(t, dir, ".gitignore", ".krewire/\n")
	writeFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello, krewire")
}
`)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsPreInit() || got.Marker != MarkerPreInit {
		t.Errorf("got %v (%s), want IsPreInit=true (pre-init)", got.IsPreInit(), got.Marker)
	}
}

func TestResult_String(t *testing.T) {
	r1 := Result{Marker: "pages/"}
	if r1.String() != "pages/" {
		t.Errorf("r1.String() = %s, want pages/", r1.String())
	}
	r2 := Result{Marker: ""}
	if r2.String() != "krewire" {
		t.Errorf("r2.String() = %s, want krewire", r2.String())
	}
}

func TestIsApp(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	if !IsApp(dir) {
		t.Errorf("IsApp should be true")
	}
	if IsApp(t.TempDir()) {
		t.Errorf("IsApp should be false for empty dir")
	}
}

func TestDetectCapabilities(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", `
ssg:
  enabled: true
worker:
  concurrency: 5
service:
  port: 8080
gateway:
  port: 8000
infra:
  provider: cloud
`)
	got, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Capabilities.HasSSG || !got.Capabilities.HasWorker || !got.Capabilities.HasService || !got.Capabilities.HasGateway || !got.Capabilities.HasInfra {
		t.Errorf("expected all capabilities to be detected, got %+v", got.Capabilities)
	}
}

func TestShape_EdgeCases(t *testing.T) {
	// isPreInit false when main.go has other content
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/starter\n")
	writeFile(t, dir, "krewire.yaml", "project:\n  name: starter\n")
	writeFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	if isPreInit(dir) {
		t.Errorf("isPreInit should be false without hello, krewire marker")
	}

	// packageMain non-existent
	if packageMain("/nonexistent/path") {
		t.Errorf("packageMain should be false for non-existent dir")
	}

	// packageMain syntax error
	badDir := t.TempDir()
	writeFile(t, badDir, "bad.go", "package main\nfunc bad {")
	writeFile(t, badDir, "good.go", "package main\nfunc main() {}")
	if !packageMain(badDir) {
		t.Errorf("packageMain should tolerate parse error on bad file if good file has main")
	}

	// hasYamlKey non-existent
	if hasYamlKey(t.TempDir(), "worker") {
		t.Errorf("hasYamlKey should be false when krewire.yaml is missing")
	}
}
