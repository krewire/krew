package commands

import (
	"flag"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunCompat(t *testing.T) {
	// With the current version.go declarations aligned, the whole ecosystem must
	// be mutually compatible. This guards against future version drift across
	// modules: a regression here means a module's version.go contract broke.
	if got := RunCompat(flag.NewFlagSet("compat", flag.ContinueOnError)); got != errs.ExitCodeSuccess {
		t.Errorf("RunCompat() = %v, want ExitCodeSuccess (ecosystem should be compatible)", got)
	}
}
