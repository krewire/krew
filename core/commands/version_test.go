package commands

import (
	"strings"
	"testing"

	"github.com/krewire/krew/core/buildinfo"
)

func TestHumanVersion_CollapsesUnknownToDev(t *testing.T) {
	cases := map[string]string{
		"":              "dev",
		"(devel)":       "dev",
		"v(devel)":      "dev",
		"devel":         "dev",
		"0.5.1":         "v0.5.1",
		"v0.5.1":        "v0.5.1",
		"0.2.0-alpha.1": "v0.2.0-alpha.1",
	}
	for in, want := range cases {
		if got := humanVersion(in); got != want {
			t.Errorf("humanVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHumanVersion_SingleLeadingV(t *testing.T) {
	for _, in := range []string{"v0.5.1", "0.5.1", "v(devel)", "(devel)", "devel", ""} {
		got := humanVersion(in)
		if got == "dev" {
			continue
		}
		if len(got) < 2 || got[0] != 'v' || got[1] == 'v' {
			t.Errorf("humanVersion(%q) = %q, want exactly one leading v or dev", in, got)
		}
	}
}

func TestQualifiedVersion_MarksWorkspaceBuilds(t *testing.T) {
	got := qualifiedVersion(buildinfo.ModKrewire)
	if strings.Contains(got, "devel") {
		t.Errorf("qualifiedVersion(%q) = %q, must never print devel", buildinfo.ModKrewire, got)
	}
	if !strings.HasSuffix(got, "(dev)") {
		t.Errorf("qualifiedVersion(%q) = %q, want a (dev) marker in workspace builds", buildinfo.ModKrewire, got)
	}
}
