package commands

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func newBoostFlagSet(args ...string) *flag.FlagSet {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	fs.Parse(append([]string{"install"}, args...))
	return fs
}

func TestRunBoostInstallFresh(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader(""), &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	for _, want := range []string{"AGENTS.md", "opencode.json"} {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if !strings.Contains(out.String(), "Installed Boost") {
		t.Errorf("missing next-steps banner: %q", out.String())
	}
}

func TestRunBoostInstallConflictDeclines(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader("n\n"), &out)
	if code != errs.ExitCodeUsage {
		t.Fatalf("want usage on decline, got %d", code.Int())
	}
	if data, err := os.ReadFile(filepath.Join(target, "AGENTS.md")); err != nil || string(data) != "mine" {
		t.Errorf("decline must not overwrite: %v %q", err, data)
	}
}

func TestRunBoostInstallConflictAccepts(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader("y\n"), &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	if data, err := os.ReadFile(filepath.Join(target, "AGENTS.md")); err != nil || !strings.Contains(string(data), "Agent Constitution") {
		t.Errorf("accept must overwrite: %v %q", err, data)
	}
}

func TestRunBoostInstallForceSkipsPrompt(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target, "--force"), strings.NewReader(""), &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
}

func TestRunBoostInstallDryRunWritesNothing(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target, "--dry-run"), strings.NewReader(""), &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run must write nothing, found %d entries", len(entries))
	}
	if !strings.Contains(out.String(), "Dry run") {
		t.Errorf("missing dry-run banner: %q", out.String())
	}
}

func TestRunBoostInstallPromptsForTarget(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(), strings.NewReader(target+"\n"), &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); err != nil {
		t.Errorf("interactive install missing AGENTS.md: %v", err)
	}
}

func TestRunBoostMissingSubcommand(t *testing.T) {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	if code := RunBoost(fs); code != errs.ExitCodeUsage {
		t.Fatalf("want usage, got %d", code.Int())
	}
}

func TestRunBoostUnknownSubcommand(t *testing.T) {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	fs.Parse([]string{"frobnicate"})
	if code := RunBoost(fs); code != errs.ExitCodeUsage {
		t.Fatalf("want usage, got %d", code.Int())
	}
}

func TestRunBoostInstallMissingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nope")
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader(""), &out)
	if code != errs.ExitCodeUsage {
		t.Fatalf("want usage for missing target, got %d", code.Int())
	}
}

func TestIsYes(t *testing.T) {
	for _, yes := range []string{"y", "Y", "yes", "YES", " yes "} {
		if !isYes(yes) {
			t.Errorf("isYes(%q) = false, want true", yes)
		}
	}
	for _, no := range []string{"", "n", "no", "x"} {
		if isYes(no) {
			t.Errorf("isYes(%q) = true, want false", no)
		}
	}
}

func TestRunBoostInstall_InNodeJSProject(t *testing.T) {
	dir := t.TempDir()

	// 1. Setup Node.js project
	pkgJSON := `{"name": "my-node-app", "version": "1.0.0", "scripts": {"dev": "vite"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	jsFile := `console.log("hello node");`
	if err := os.WriteFile(filepath.Join(srcDir, "index.js"), []byte(jsFile), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Run krew boost install <nodejs-project-dir>
	code := RunBoost(newBoostFlagSet(dir))
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected success, got %d", code)
	}

	// 3. Verify Boost files exist
	agentsPath := filepath.Join(dir, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be installed: %v", err)
	}
	if !strings.Contains(string(data), "If the current project is not a Krewire project") {
		t.Errorf("expected AGENTS.md to contain project-agnostic disclaimer")
	}
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); err != nil {
		t.Errorf("expected opencode.json to be installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents")); err != nil {
		t.Errorf("expected .agents/ directory to be installed: %v", err)
	}

	// 4. Verify Node.js files remain completely untouched
	currentPkg, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	if string(currentPkg) != pkgJSON {
		t.Errorf("package.json was modified: got %s", string(currentPkg))
	}
	currentJS, _ := os.ReadFile(filepath.Join(srcDir, "index.js"))
	if string(currentJS) != jsFile {
		t.Errorf("src/index.js was modified: got %s", string(currentJS))
	}
}

func TestRunBoostInstall_InLaravelProject(t *testing.T) {
	dir := t.TempDir()

	// 1. Setup Laravel project
	compJSON := `{"name": "laravel/laravel", "require": {"php": "^8.2"}}`
	if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(compJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	appDir := filepath.Join(dir, "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	phpFile := "<?php echo 'hello laravel';\n"
	if err := os.WriteFile(filepath.Join(appDir, "User.php"), []byte(phpFile), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Run krew boost install <laravel-project-dir>
	code := RunBoost(newBoostFlagSet(dir))
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected success, got %d", code)
	}

	// 3. Verify Boost files exist
	agentsPath := filepath.Join(dir, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md to be installed: %v", err)
	}
	if !strings.Contains(string(data), "If the current project is not a Krewire project") {
		t.Errorf("expected AGENTS.md to contain project-agnostic disclaimer")
	}

	// 4. Verify Laravel files remain completely untouched
	currentComp, _ := os.ReadFile(filepath.Join(dir, "composer.json"))
	if string(currentComp) != compJSON {
		t.Errorf("composer.json was modified: got %s", string(currentComp))
	}
	currentPHP, _ := os.ReadFile(filepath.Join(appDir, "User.php"))
	if string(currentPHP) != phpFile {
		t.Errorf("app/User.php was modified: got %s", string(currentPHP))
	}
}

func TestRunBoostStatus(t *testing.T) {
	dir := t.TempDir()

	// 1. Status when not installed
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	_ = fs.Parse([]string{"status", dir})

	var out bytes.Buffer
	code := runBoostStatus(fs, &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected success, got %v", code)
	}
	if !strings.Contains(out.String(), "Not installed") {
		t.Errorf("expected Not installed in output, got: %s", out.String())
	}

	// 2. Install Boost
	installCode := runBoostInstall(newBoostFlagSet(dir), strings.NewReader(""), io.Discard)
	if installCode != errs.ExitCodeSuccess {
		t.Fatalf("boost install failed: %v", installCode)
	}

	// 3. Status when installed
	out.Reset()
	code = runBoostStatus(fs, &out)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected success, got %v", code)
	}
	if !strings.Contains(out.String(), "fully installed") {
		t.Errorf("expected fully installed in output, got: %s", out.String())
	}
}
