package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/version"
)

func TestBump(t *testing.T) {
	v := version.MustParseVersion("0.3.1")
	if got := Bump(v, BumpPatch); got.String() != "0.3.2" {
		t.Errorf("Bump patch = %s, want 0.3.2", got)
	}
	if got := Bump(v, BumpMinor); got.String() != "0.4.0" {
		t.Errorf("Bump minor = %s, want 0.4.0", got)
	}
	if got := Bump(v, BumpMajor); got.String() != "1.0.0" {
		t.Errorf("Bump major = %s, want 1.0.0", got)
	}
}

func TestBumpTrack(t *testing.T) {
	v := version.MustParseVersion("1.0.0")

	// LTS
	if got := BumpTrack(v, BumpPatch, TrackLTS); got.String() != "1.0.1-lts" {
		t.Errorf("BumpTrack LTS = %s, want 1.0.1-lts", got)
	}

	// Stable
	if got := BumpTrack(v, BumpMinor, TrackStable); got.String() != "1.1.0" {
		t.Errorf("BumpTrack Stable = %s, want 1.1.0", got)
	}

	// Pre-release: RC
	if got := BumpTrack(v, BumpMinor, TrackRC); got.String() != "1.1.0-rc.1" {
		t.Errorf("BumpTrack RC = %s, want 1.1.0-rc.1", got)
	}

	// Pre-release: Beta
	if got := BumpTrack(v, BumpMinor, TrackBeta); got.String() != "1.1.0-beta.1" {
		t.Errorf("BumpTrack Beta = %s, want 1.1.0-beta.1", got)
	}

	// Pre-release: Alpha
	if got := BumpTrack(v, BumpMajor, TrackAlpha); got.String() != "2.0.0-alpha.1" {
		t.Errorf("BumpTrack Alpha = %s, want 2.0.0-alpha.1", got)
	}

	// Experimental
	if got := BumpTrack(v, BumpMinor, TrackExperimental); got.String() != "1.1.0-exp" {
		t.Errorf("BumpTrack Experimental = %s, want 1.1.0-exp", got)
	}

	// Dev
	if got := BumpTrack(v, BumpPatch, TrackDev); got.String() != "1.0.1-dev" {
		t.Errorf("BumpTrack Dev = %s, want 1.0.1-dev", got)
	}

	// Nightly
	nightly := BumpTrack(v, BumpMinor, TrackNightly)
	if nightly.PreRelease == "" || nightly.Major != 1 || nightly.Minor != 1 {
		t.Errorf("BumpTrack Nightly invalid: %s", nightly)
	}
}

func TestPlanReleasingKrewire(t *testing.T) {
	edits, err := Plan([]string{string(ModuleKrewire)}, BumpPatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits) == 0 {
		t.Fatal("expected edits when planning krewire release")
	}
}

func TestPlanAllModules(t *testing.T) {
	var all []string
	for _, m := range Modules {
		all = append(all, m.Name)
	}
	edits, err := Plan(all, BumpPatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits) < len(Modules) {
		t.Fatalf("got %d edits, want at least %d", len(edits), len(Modules))
	}
}

func TestApplyWritesAndValidates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "v.go")
	orig := `var Version = version.MustParseVersion("0.1.0")`
	if err := os.WriteFile(p, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	edits := []Edit{{
		Module:  "krewire",
		File:    p,
		Summary: "bump",
		From:    orig,
		To:      `var Version = version.MustParseVersion("0.1.1")`,
	}}
	if _, err := Apply(edits, "", false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `var Version = version.MustParseVersion("0.1.1")` {
		t.Errorf("got %q, want updated", string(got))
	}
}

func TestCheckTierCompliance(t *testing.T) {
	// Verify current ecosystem modules comply with TierFree and valid open-source licensing
	if err := CheckTierCompliance(""); err != nil {
		t.Fatalf("CheckTierCompliance failed: %v", err)
	}

	for _, m := range Modules {
		tier, err := CurrentTier(m.Name)
		if err != nil {
			t.Fatalf("CurrentTier(%s) err = %v", m.Name, err)
		}
		if tier != m.Tier {
			t.Errorf("expected module %s to be tier %s, got %s", m.Name, m.Tier, tier)
		}
	}
}

func TestReadEcosystemVersionsAndRequirements(t *testing.T) {
	root, err := FindWorkspaceRoot()
	if err != nil {
		t.Skip("skipping workspace test outside go.work")
	}

	actual, err := ReadEcosystemActualVersions(root)
	if err != nil {
		t.Fatalf("ReadEcosystemActualVersions failed: %v", err)
	}
	for _, m := range Modules {
		if v, ok := actual[m.Name]; !ok || v.String() == "" {
			t.Errorf("module %s version not read", m.Name)
		}
	}

	reqs, err := ReadEcosystemRequirements(root)
	if err != nil {
		t.Fatalf("ReadEcosystemRequirements failed: %v", err)
	}
	if len(reqs["dash"]) < 2 {
		t.Errorf("dash requirements = %v, want at least 2", reqs["dash"])
	}
}

func TestPlanMinorBumpPropagatesRequirements(t *testing.T) {
	// When krewire is bumped with BumpMinor, its dependents (krew, auth, cloud, dash)
	// must receive requirement updates because 0.y caret requires identical minor.
	edits, err := Plan([]string{string(ModuleKrewire)}, BumpMinor)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	if len(edits) <= 1 {
		t.Errorf("got %d edits, want edits for krewire plus dependents", len(edits))
	}
}
