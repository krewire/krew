package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

// TestWorkflow_CLIApplication validates end-to-end initialization and build of a CLI application.
func TestWorkflow_CLIApplication(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	cliDir := filepath.Join(dir, "my-cli-tool")
	writeFile(t, filepath.Join(cliDir, "krewire.yaml"), `
name: my-cli-tool
kind: cli
output: .krewire/build
`)
	writeFile(t, filepath.Join(cliDir, "go.mod"), "module example.com/cli\n\ngo 1.24\n")
	writeFile(t, filepath.Join(cliDir, "main.go"), `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("my-cli-tool v1.0.0")
		return
	}
	fmt.Println("CLI tool execution success")
}
`)

	chdir(t, cliDir)
	fsBuild := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fsBuild)

	code := RunBuild(fsBuild)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild on CLI tool failed: %v", code)
	}

	binPath := filepath.Join(cliDir, ".krewire", "build", "bin", "my-cli-tool")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected compiled binary at %s: %v", binPath, err)
	}
}

// TestWorkflow_WebApplication validates end-to-end build of a web/SSG application.
func TestWorkflow_WebApplication(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	writeFile(t, filepath.Join(dir, "krewire.yaml"), `
name: portal
kind: site
output: .krewire/build
`)
	writeFile(t, filepath.Join(dir, "layouts", "Base.kiw"), `<!DOCTYPE html><html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>`)
	writeFile(t, filepath.Join(dir, "pages", "index.kiw"), `---
title: Krewire Portal
---
<h1>Welcome to Portal</h1>
`)
	writeFile(t, filepath.Join(dir, "pages", "about.kiw"), `---
title: About Us
---
<p>About information.</p>
`)

	fsBuild := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fsBuild)

	code := RunBuild(fsBuild)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild on web application failed: %v", code)
	}

	indexPath := filepath.Join(dir, ".krewire", "build", "index.html")
	if data, err := os.ReadFile(indexPath); err != nil || !strings.Contains(string(data), "Welcome to Portal") {
		t.Fatalf("expected valid index.html, err=%v, content=%s", err, string(data))
	}

	aboutPath := filepath.Join(dir, ".krewire", "build", "about.html")
	if data, err := os.ReadFile(aboutPath); err != nil || !strings.Contains(string(data), "About information") {
		t.Fatalf("expected valid about.html, err=%v, content=%s", err, string(data))
	}
}

// TestWorkflow_HybridApplication validates multi-workload execution where Go binary,
// SSG pages, and content co-occur within a single fluid workspace.
func TestWorkflow_HybridApplication(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	writeFile(t, filepath.Join(dir, "krewire.yaml"), `
name: full-stack-app
output: .krewire/build
`)
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/fullstack\n\ngo 1.24\n")
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "fmt"

func main() {
	fmt.Println("fullstack server active")
}
`)
	writeFile(t, filepath.Join(dir, "layouts", "Base.kiw"), `<html><body>{{.Content}}</body></html>`)
	writeFile(t, filepath.Join(dir, "pages", "index.kiw"), `---
title: App Landing
---
<h2>App Landing Page</h2>
`)

	fsBuild := flag.NewFlagSet("build", flag.ContinueOnError)
	RegisterBuild(fsBuild)

	code := RunBuild(fsBuild)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunBuild on hybrid failed: %v", code)
	}

	binPath := filepath.Join(dir, ".krewire", "build", "bin", "full-stack-app")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected compiled binary at %s: %v", binPath, err)
	}

	htmlPath := filepath.Join(dir, ".krewire", "build", "index.html")
	if _, err := os.Stat(htmlPath); err != nil {
		t.Fatalf("expected landing page at %s: %v", htmlPath, err)
	}
}

// TestWorkflow_FailureModes verifies that erroneous configurations produce explicit, safe errors.
func TestWorkflow_FailureModes(t *testing.T) {
	t.Run("InvalidYAML", func(t *testing.T) {
		dir := t.TempDir()
		chdir(t, dir)
		writeFile(t, filepath.Join(dir, "krewire.yaml"), `project: [invalid: yaml:`)

		fsBuild := flag.NewFlagSet("build", flag.ContinueOnError)
		RegisterBuild(fsBuild)

		code := RunBuild(fsBuild)
		if code == errs.ExitCodeSuccess {
			t.Fatalf("expected failure on invalid YAML, got success")
		}
	})

	t.Run("MissingLayoutInSSG", func(t *testing.T) {
		dir := t.TempDir()
		chdir(t, dir)
		writeFile(t, filepath.Join(dir, "krewire.yaml"), `kind: site`)
		writeFile(t, filepath.Join(dir, "pages", "index.kiw"), `---
title: No Layout
---
<p>hello</p>
`)
		fsBuild := flag.NewFlagSet("build", flag.ContinueOnError)
		RegisterBuild(fsBuild)

		code := RunBuild(fsBuild)
		if code == errs.ExitCodeSuccess {
			t.Fatalf("expected failure when layout is missing, got success")
		}
	})
}
