package buildinfo

import (
	"strings"
	"testing"
)

func TestLDFlags(t *testing.T) {
	flags := LDFlags(false)
	if !strings.Contains(flags, "build with krewire") {
		t.Errorf("LDFlags(false) = %q, want containing 'build with krewire'", flags)
	}
	if strings.Contains(flags, "-s -w") {
		t.Errorf("LDFlags(false) should not contain -s -w, got %q", flags)
	}

	stripped := LDFlags(true)
	if !strings.HasPrefix(stripped, "-s -w") {
		t.Errorf("LDFlags(true) should begin with -s -w, got %q", stripped)
	}
	if !strings.Contains(stripped, "github.com/krewire/krewire/pkg/version.Signature") {
		t.Errorf("LDFlags(true) missing version.Signature target: %q", stripped)
	}
}
