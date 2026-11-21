package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func runConfigArgs(t *testing.T, args ...string) errs.ExitCode {
	t.Helper()
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	RegisterConfig(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return RunConfig(fs)
}

func TestRunConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "krewire.yaml")

	_ = os.WriteFile(path, []byte("project:\n  name: ok\n  kind: site\n"), 0o644)
	if code := runConfigArgs(t, "validate", path); code != errs.ExitCodeSuccess {
		t.Errorf("valid file: code %v", code)
	}

	_ = os.WriteFile(path, []byte("project:\n  name: ok\nbuild:\n  outdir: dist\n"), 0o644)
	if code := runConfigArgs(t, dir); code != errs.ExitCodeSuccess {
		t.Errorf("warnings only: code %v, want success", code)
	}
	if code := runConfigArgs(t, "--strict", path); code != errs.ExitCodeFailure {
		t.Errorf("--strict with warnings: code %v, want failure", code)
	}

	_ = os.WriteFile(path, []byte("dev:\n  port: nope\n"), 0o644)
	if code := runConfigArgs(t, path); code != errs.ExitCodeFailure {
		t.Errorf("type error: code %v, want failure", code)
	}

	if code := runConfigArgs(t, filepath.Join(dir, "missing.yaml")); code != errs.ExitCodeUsage {
		t.Errorf("missing file: code %v, want usage", code)
	}
}

func TestDoctorReportsSchemaIssues(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte("project:\n  name: x\noutdir: dist\n"), 0o644)
	report := DiagnoseEnvironment(dir)
	for _, c := range report.Checks {
		if c.Name == "Configuration" {
			if c.Status != StatusWarn || !strings.Contains(c.Message, `krewire.yaml:3:1: warning: unknown key "outdir"; did you mean "output"?`) {
				t.Errorf("configuration check = %+v", c)
			}
			return
		}
	}
	t.Fatal("no Configuration check")
}
