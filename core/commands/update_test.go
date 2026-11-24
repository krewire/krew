package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/errs"
)

func TestRegisterUpdate(t *testing.T) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	RegisterUpdate(fs)

	if fs.Lookup("tidy") == nil {
		t.Error("expected --tidy flag to be registered")
	}
	if fs.Lookup("dry-run") == nil {
		t.Error("expected --dry-run flag to be registered")
	}
}

func TestExpandKrewireShortcut(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"krewire", "github.com/krewire/krewire"},
		{"KREWIRE", "github.com/krewire/krewire"},
		{"core", "github.com/krewire/krewire"},
		{"ui", "github.com/krewire/ui/core"},
		{"krew", "github.com/krewire/krew"},
		{"auth", "github.com/krewire/auth"},
		{"cloud", "github.com/krewire/cloud"},
		{"dash", "github.com/krewire/dash"},
		{"github.com/foo/bar", "github.com/foo/bar"},
	}

	for _, tt := range tests {
		got := expandKrewireShortcut(tt.input)
		if got != tt.want {
			t.Errorf("expandKrewireShortcut(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRunUpdateNoProject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kiw-test-update-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(origWd) }()

	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	RegisterUpdate(fs)

	code := RunUpdate(fs)
	if code != errs.ExitCodeUsage {
		t.Errorf("got exit code %d, want %d", code, errs.ExitCodeUsage)
	}
}

func TestRunUpdateDryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kiw-test-update-dryrun-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	goModContent := "module testmod\n\ngo 1.21\n\nrequire github.com/krewire/krewire v0.1.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(origWd) }()

	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	RegisterUpdate(fs)
	_ = fs.Parse([]string{"--dry-run"})

	code := RunUpdate(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("got exit code %d, want %d", code, errs.ExitCodeSuccess)
	}
}
