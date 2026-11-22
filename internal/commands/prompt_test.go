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

func TestRunNew_PromptsForName(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	fakeTerminal(t, "docs-site\n")
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	if code := RunNew(fs); code != errs.ExitCodeSuccess {
		t.Fatalf("code %v", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs-site", "krewire.yaml")); err != nil {
		t.Fatalf("expected krewire.yaml: %v", err)
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
