package krew_test

import (
	"testing"

	"github.com/krewire/krew"
)

func TestVersion(t *testing.T) {
	if krew.VersionString() != "0.2.0" {
		t.Errorf("expected version 0.2.0, got %s", krew.VersionString())
	}
}
