package kui

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/krewire/krewire/pkg/ui/components"
)

func actions(rs []Result) map[string]Action {
	m := map[string]Action{}
	for _, r := range rs {
		m[r.Name] = r.Action
	}
	return m
}

func TestInstallLifecycle(t *testing.T) {
	root := t.TempDir()

	// Fresh install of a pack.
	rs, err := Install(root, []string{"ui:minimal"}, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]Action{"Button": ActionAdded, "Card": ActionAdded, "Input": ActionAdded}; !reflect.DeepEqual(actions(rs), want) {
		t.Fatalf("install = %v", actions(rs))
	}
	upstream, _ := components.Get("Button")
	if got, _ := os.ReadFile(Path(root, "Button")); string(got) != string(upstream) {
		t.Error("Button.kiw is not the kui source")
	}
	lock, _ := ReadLock(root)
	if lock.Components["Button"].SHA256 != hash(upstream) || lock.Components["Button"].Kui == "" {
		t.Errorf("lock = %+v", lock.Components["Button"])
	}

	// Re-running is a no-op.
	rs, _ = Install(root, []string{"ui/button"}, false, false)
	if actions(rs)["Button"] != ActionUnchanged {
		t.Errorf("reinstall = %v", rs)
	}

	// A stale but unmodified copy is updated.
	_ = os.WriteFile(Path(root, "Card"), []byte("old upstream"), 0o644)
	lock.Components["Card"] = LockEntry{Kui: "v0", SHA256: hash([]byte("old upstream"))}
	_ = writeLock(root, lock)
	if st, _ := State(root, "Card", lock); st != StatusOutdated {
		t.Errorf("state = %s, want %s", st, StatusOutdated)
	}
	rs, _ = Install(root, []string{"card"}, false, false)
	if actions(rs)["Card"] != ActionUpdated {
		t.Errorf("update = %v", rs)
	}

	// A locally modified copy is kept, reported, and replaced only with force.
	_ = os.WriteFile(Path(root, "Input"), []byte("<input class=\"mine\">"), 0o644)
	lock, _ = ReadLock(root)
	if st, _ := State(root, "Input", lock); st != StatusModified {
		t.Errorf("state = %s, want %s", st, StatusModified)
	}
	rs, err = Install(root, []string{"Input"}, false, false)
	if !errors.Is(err, ErrModified) || actions(rs)["Input"] != ActionKept {
		t.Errorf("modified install = %v, %v", rs, err)
	}
	if got, _ := os.ReadFile(Path(root, "Input")); string(got) != "<input class=\"mine\">" {
		t.Error("modified component was overwritten")
	}
	rs, err = Install(root, []string{"Input"}, true, false)
	if err != nil || actions(rs)["Input"] != ActionReplaced {
		t.Errorf("forced install = %v, %v", rs, err)
	}
}

func TestInstallDependenciesAndDryRun(t *testing.T) {
	root := t.TempDir()
	rs, err := Install(root, []string{"ui/navbar"}, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if a := actions(rs); a["Navbar"] != ActionAdded || a["Brand"] != ActionAdded {
		t.Errorf("dry run = %v", a)
	}
	if _, err := os.Stat(Path(root, "Navbar")); !os.IsNotExist(err) {
		t.Error("dry run wrote files")
	}
	if _, err := Install(root, []string{"ui/nope"}, false, false); err == nil {
		t.Error("unknown component: want error")
	}
}

func TestParseRef(t *testing.T) {
	for in, want := range map[string]string{"ui/button": "button", "ui:dashboard": "pack:dashboard", "pack:base": "pack:base", "Card": "Card"} {
		if got := ParseRef(in); got != want {
			t.Errorf("ParseRef(%q) = %q, want %q", in, got, want)
		}
	}
}
