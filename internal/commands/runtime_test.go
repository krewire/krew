package commands

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/env"
	"github.com/krewire/krewire/pkg/errs"
)

// chdir moves into dir for the duration of the test, restoring afterwards
// (testing.Chdir needs a newer language version than this module declares).
func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestBootRuntime_ResolvesConfigDotenvEnvAndDebug(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":       "module demo\n\ngo 1.26\n",
		"krewire.yaml": "title: Demo\nenv: production\ndebug: true\n",
		".env":         "BOOT_PROOF=from-dotenv\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)

	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	RegisterServe(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("bootRuntime code = %d, want success", code)
	}
	if rt.env != env.EnvProduction {
		t.Errorf("rt.env = %q, want production from krewire.yaml", rt.env)
	}
	if !rt.debug {
		t.Error("rt.debug = false, want true from krewire.yaml")
	}
	if got := os.Getenv("BOOT_PROOF"); got != "from-dotenv" {
		t.Errorf(".env BOOT_PROOF = %q, want %q", got, "from-dotenv")
	}
}

func TestRunServe_OnCLIKind_PassesArgsAndPropagatesExit(t *testing.T) {
	dir := t.TempDir()
	main := `package main

import "os"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "ok" {
		return
	}
	os.Exit(3)
}
`
	for name, body := range map[string]string{
		"go.mod":       "module demo\n\ngo 1.26\n",
		"main.go":      main,
		"krewire.yaml": "title: Demo\nproject:\n  kind: cli\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)

	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	run := func(args ...string) errs.ExitCode {
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		RegisterServe(fs)
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return RunServe(fs)
	}
	if code := run("ok"); code != errs.ExitCodeSuccess {
		t.Errorf("RunServe(cli, ok) = %d, want 0 (arg passthrough)", code)
	}
	if code := run(); code != errs.ExitCodeFailure {
		t.Errorf("RunServe(cli, no args) = %d, want child failure mapped to canonical exit 1 (G5)", code)
	}
}
