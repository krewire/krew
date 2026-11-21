package commands

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

// fakeTerminal makes the prompt helpers read answers from input.
func fakeTerminal(t *testing.T, input string) {
	t.Helper()
	oldIn, oldOut, oldInteractive := promptIn, promptOut, interactive
	promptIn, promptOut, interactive = strings.NewReader(input), io.Discard, func() bool { return true }
	t.Cleanup(func() { promptIn, promptOut, interactive = oldIn, oldOut, oldInteractive })
}

func readKind(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "krewire.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, ok := strings.CutPrefix(strings.TrimSpace(line), "kind: "); ok {
			return k
		}
	}
	return ""
}

func TestRunNew_PromptsForNameAndKind(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	fakeTerminal(t, "docs-site\n2\n")
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	if code := RunNew(fs); code != errs.ExitCodeSuccess {
		t.Fatalf("code %v", code)
	}
	if k := readKind(t, filepath.Join(dir, "docs-site")); k != "site" {
		t.Errorf("kind = %q, want site", k)
	}
}

func TestRunNew_NoInputKeepsUsageError(t *testing.T) {
	chdir(t, t.TempDir())
	fakeTerminal(t, "ignored\n")
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Parse([]string{"--no-input"})
	if code := RunNew(fs); code != errs.ExitCodeUsage {
		t.Errorf("code %v, want usage", code)
	}
}

func TestRunInit_PromptsForKind(t *testing.T) {
	for input, want := range map[string]string{"\n": "", "book\n": "book", "bogus\n4\n": "cli"} {
		dir := t.TempDir()
		fakeTerminal(t, input)
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		RegisterInit(fs)
		_ = fs.Parse([]string{dir})
		if code := RunInit(fs); code != errs.ExitCodeSuccess {
			t.Fatalf("%q: code %v", input, code)
		}
		if k := readKind(t, dir); k != want {
			t.Errorf("%q: kind = %q, want %q", input, k, want)
		}
	}
}

func TestRunInit_FlagSkipsPrompt(t *testing.T) {
	dir := t.TempDir()
	fakeTerminal(t, "") // a prompt would read EOF and fall back to none
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	RegisterInit(fs)
	_ = fs.Parse([]string{"--site", dir})
	if code := RunInit(fs); code != errs.ExitCodeSuccess {
		t.Fatalf("code %v", code)
	}
	if k := readKind(t, dir); k != "site" {
		t.Errorf("kind = %q, want site", k)
	}
}
