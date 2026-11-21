package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func runSeed(t *testing.T, args ...string) (string, errs.ExitCode) {
	t.Helper()
	return captureStdout(t, func() errs.ExitCode {
		fs := flag.NewFlagSet("seed", flag.ContinueOnError)
		RegisterSeed(fs)
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return RunSeed(fs)
	})
}

func TestSeed_NoSeedsDir(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Minimal krewire.yaml
	yaml := `project:
  name: demo
kind: app
database:
  driver: sqlite
  dsn: test.db
`
	writeFile(t, filepath.Join(dir, "krewire.yaml"), yaml)
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/demo\n\ngo 1.22\n")

	_, code := runSeed(t)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected ExitCodeUsage when seeds directory missing, got %v", code)
	}
}

func TestSeed_EmptySeedsDir(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	yaml := `project:
  name: demo
kind: app
database:
  driver: sqlite
  dsn: test.db
`
	writeFile(t, filepath.Join(dir, "krewire.yaml"), yaml)
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/demo\n\ngo 1.22\n")
	if err := os.MkdirAll(filepath.Join(dir, "seeds"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, code := runSeed(t)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected ExitCodeUsage when seeds directory is empty, got %v", code)
	}
}
