package version_test

import (
	"testing"

	"github.com/krewire/krew/version"
)

func TestVersion(t *testing.T) {
	if version.VersionString() != "0.2.0" {
		t.Errorf("expected version 0.2.0, got %s", version.VersionString())
	}
}
