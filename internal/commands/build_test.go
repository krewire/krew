package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/book"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/web/ssg"
)

func TestVersionedAssetURLs(t *testing.T) {
	cases := []struct {
		name    string
		urls    []string
		version string
		want    []string
	}{
		{"stamps version", []string{"/assets/a.css"}, "0.1.0", []string{"/assets/a.css?v=0.1.0"}},
		{"strips v prefix", []string{"/assets/a.css"}, "v1.2.3", []string{"/assets/a.css?v=1.2.3"}},
		{"no version leaves urls untouched", []string{"/assets/a.css"}, "", []string{"/assets/a.css"}},
		{"empty urls", nil, "1.0.0", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := versionedAssetURLs(c.urls, c.version)
			if len(got) != len(c.want) {
				t.Fatalf("got %d urls %v, want %d %v", len(got), got, len(c.want), c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("url[%d] = %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestWithoutAssetDropsMatchingEntry(t *testing.T) {
	plan := []string{"/assets/style.css", "/assets/tailwind.css", "/assets/krewire.css"}
	got := withoutAsset(plan, book.StylesheetName)
	if len(got) != 2 {
		t.Fatalf("got %v, want 2 entries", got)
	}
	for _, u := range got {
		if strings.Contains(u, "krewire.css") {
			t.Errorf("krewire.css must be dropped, got %q", u)
		}
	}
}

// TestBookStylesheetIsPartOfTheSitePlan pins the coupling between pkg/book and
// pkg/web/ssg: the site's asset plan lists the book's stylesheet, and the
// devtool filters it out because the book links it itself. If either side renames
// it, this test fails instead of the page loading the stylesheet twice.
func TestBookStylesheetIsPartOfTheSitePlan(t *testing.T) {
	s := ssg.New().
		Asset("assets/style.css", "").
		Asset(book.StylesheetName, "")
	css, _ := s.InjectedAssets()
	found := false
	for _, u := range css {
		if strings.HasSuffix(u, book.StylesheetName) {
			found = true
		}
	}
	if !found {
		t.Fatalf("site plan %v does not contain %q", css, book.StylesheetName)
	}
	if got := withoutAsset(css, book.StylesheetName); len(got) != len(css)-1 {
		t.Errorf("filtering removed %d entries, want 1", len(css)-len(got))
	}
}

func TestWithoutAssetKeepsUnrelatedEntries(t *testing.T) {
	plan := []string{"/assets/style.css", "/assets/app.js"}
	got := withoutAsset(plan, "assets/not-present.css")
	if len(got) != len(plan) {
		t.Errorf("got %v, want the plan unchanged", got)
	}
}

func TestProgressiveBuild_BothManuscriptAndSSG(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, ".krewire", "build")

	configContent := `
name: test-progressive
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	layoutsDir := filepath.Join(dir, "layouts")
	if err := os.MkdirAll(layoutsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layoutsDir, "Base.kiw"), []byte("<html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pagesDir := filepath.Join(dir, "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pageContent := "---\ntitle: Home\nlayout: Base\n---\n<h1>Landing Page</h1>\n"
	if err := os.WriteFile(filepath.Join(pagesDir, "index.kiw"), []byte(pageContent), 0o644); err != nil {
		t.Fatal(err)
	}

	bookDir := filepath.Join(dir, "content")
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "intro.md"), []byte("# Intro\nBook content here.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	code := RunBuild(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild failed with exit code %v", code)
	}

	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Errorf("expected ssg index.html in output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "intro.html")); err != nil {
		t.Errorf("expected book intro.html in output: %v", err)
	}
}

func TestRunBuild_AppBinary(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, ".krewire", "build")

	for name, content := range map[string]string{
		"go.mod":       "module testapp\n\ngo 1.22\n",
		"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"ok\") }\n",
		"krewire.yaml": "project:\n  name: testapp\n  kind: app\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)

	code := RunBuild(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild failed with exit code %v", code)
	}

	binPath := filepath.Join(outDir, "bin", "testapp")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("expected binary %s to be created: %v", binPath, err)
	}
	if _, err := os.Stat(filepath.Join(outDir, ".kiw-build-manifest")); err != nil {
		t.Errorf("expected .kiw-build-manifest: %v", err)
	}
}

func TestRunBuild_InfraPlan(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, ".krewire", "build")

	for name, content := range map[string]string{
		"krewire.yaml": "project:\n  name: my-cloud\n  kind: infra\ninfra:\n  provider: krewire-cloud\n  region: asia-southeast1\nserver:\n  port: 8080\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)
	_ = fs.Set("plan", "true")

	code := RunBuild(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild failed with exit code %v", code)
	}

	planFile := filepath.Join(outDir, "infra", "plan.json")
	data, err := os.ReadFile(planFile)
	if err != nil {
		t.Fatalf("expected plan.json to be created: %v", err)
	}
	if !strings.Contains(string(data), "my-cloud") {
		t.Errorf("plan.json missing my-cloud resource: %s", string(data))
	}
}

func TestRunBuild_PreInitRequiresInit(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":       "module starter\n\ngo 1.22\n",
		"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello, krewire\") }\n",
		"krewire.yaml": "title: Starter\n",
		".gitignore":   ".krewire/\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)

	code := RunBuild(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("RunBuild on pre-init = %v, want ExitCodeUsage (2)", code)
	}
}

func TestRunBuild_WorkerAndService(t *testing.T) {
	for _, kind := range []string{"worker", "service"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range map[string]string{
				"go.mod":       "module testmod\n\ngo 1.22\n",
				"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"ok\") }\n",
				"krewire.yaml": "project:\n  name: " + kind + "-app\n  kind: " + kind + "\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			chdir(t, dir)
			fs := flag.NewFlagSet("build", flag.ContinueOnError)
			RegisterBuild(fs)

			code := RunBuild(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunBuild on %s failed: %v", kind, code)
			}
			binPath := filepath.Join(dir, ".krewire", "build", "bin", kind+"-app")
			if _, err := os.Stat(binPath); err != nil {
				t.Fatalf("expected binary at %s: %v", binPath, err)
			}
		})
	}
}

func TestRunBuild_WasmAutomaticSideBySide(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, ".krewire", "build")

	for name, content := range map[string]string{
		"go.mod":           "module testwasm\n\ngo 1.22\n",
		"krewire.yaml":     "project:\n  kind: site\nwasm:\n  entry: ./wasm\n  name: client\n",
		"pages/index.html": "<!DOCTYPE html><html><body><h1>Hello</h1></body></html>",
		"wasm/main.go":     "package main\nfunc main() {}\n",
	} {
		p := filepath.Join(dir, name)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)

	code := RunBuild(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild failed: %v", code)
	}

	wasmFile := filepath.Join(outDir, "client.wasm")
	if _, err := os.Stat(wasmFile); err != nil {
		t.Fatalf("expected client.wasm at %s: %v", wasmFile, err)
	}
	wasmAsset := filepath.Join(outDir, "assets", "client.wasm")
	if _, err := os.Stat(wasmAsset); err != nil {
		t.Fatalf("expected assets/client.wasm at %s: %v", wasmAsset, err)
	}
}

func TestRunBuild_FluidHybridProject(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "krewire.yaml"), `
name: hybrid-app
output: .krewire/build
`)
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/hybrid\n\ngo 1.24\n")
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("hybrid server running")
}
`)
	writeFile(t, filepath.Join(dir, "layouts", "Base.kiw"), `<html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>`)
	writeFile(t, filepath.Join(dir, "pages", "index.kiw"), `---
title: Welcome
---
<h1>Welcome to Hybrid App</h1>
`)

	chdir(t, dir)
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fs)

	code := RunBuild(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild failed on fluid hybrid project: %v", code)
	}

	binPath := filepath.Join(dir, ".krewire", "build", "bin", "hybrid-app")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected binary at %s: %v", binPath, err)
	}

	htmlPath := filepath.Join(dir, ".krewire", "build", "index.html")
	if _, err := os.Stat(htmlPath); err != nil {
		t.Fatalf("expected SSG page at %s: %v", htmlPath, err)
	}
}
