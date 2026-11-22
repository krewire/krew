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
