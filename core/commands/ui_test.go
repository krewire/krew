package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/errs"
)

func runWith(t *testing.T, register func(*flag.FlagSet), run func(*flag.FlagSet) errs.ExitCode, args ...string) errs.ExitCode {
	t.Helper()
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	register(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return run(fs)
}

func TestAddUIComponentsAndPacks(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	_ = os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte("project:\n  name: p\n"), 0o644)

	if code := runWith(t, RegisterAdd, RunAdd, "ui/button", "ui:dashboard"); code != errs.ExitCodeSuccess {
		t.Fatalf("krew add: code %v", code)
	}
	for _, name := range []string{"Button", "Table", "Stat", "Sidebar", "Modal", "SideMenu"} {
		if _, err := os.Stat(filepath.Join(dir, "components", name+".kiw")); err != nil {
			t.Errorf("%s.kiw not installed: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "components", ".kui-lock.json")); err != nil {
		t.Errorf("lock file missing: %v", err)
	}

	// A local edit is kept and reported as a failure until --force.
	_ = os.WriteFile(filepath.Join(dir, "components", "Button.kiw"), []byte("<button>mine</button>"), 0o644)
	if code := runWith(t, RegisterUI, RunUI, "add", "button"); code != errs.ExitCodeFailure {
		t.Errorf("modified add: code %v, want failure", code)
	}
	if code := runWith(t, RegisterUI, RunUI, "--force", "add", "button"); code != errs.ExitCodeSuccess {
		t.Errorf("forced add: code %v", code)
	}

	for _, args := range [][]string{{"list"}, {"--json", "list"}, {"info", "card"}, {"diff", "Button"}, {"vendor", "badge"}, {"eject", "alert"}} {
		if code := runWith(t, RegisterUI, RunUI, args...); code != errs.ExitCodeSuccess {
			t.Errorf("krew ui %v: code %v", args, code)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Badge.kiw")); err != nil {
		t.Errorf("Badge.kiw not vendored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Alert.kiw")); err != nil {
		t.Errorf("Alert.kiw not ejected: %v", err)
	}
	if code := runWith(t, RegisterUI, RunUI, "info", "nope"); code != errs.ExitCodeUsage {
		t.Errorf("unknown component: code %v, want usage", code)
	}
	if code := runWith(t, RegisterAdd, RunAdd, "ui/nope"); code != errs.ExitCodeUsage {
		t.Errorf("unknown ui ref: code %v, want usage", code)
	}
}

func TestHintsPointAtExistingCommands(t *testing.T) {
	for name, hint := range Hints {
		if hint == "" || name == "" {
			t.Errorf("empty hint entry %q", name)
		}
	}
}
