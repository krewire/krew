package commands

import (
	"flag"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestClientCatcallsHasAtLeastTenEntries(t *testing.T) {
	if len(ClientCatcalls) < 10 {
		t.Fatalf("expected at least 10 catcalls, got %d", len(ClientCatcalls))
	}
	for i, c := range ClientCatcalls {
		if c == "" {
			t.Errorf("catcall entry %d is empty", i)
		}
	}
}

func TestRunKiwSuccess(t *testing.T) {
	fs := flag.NewFlagSet("kiw", flag.ContinueOnError)
	RegisterKiw(fs)

	// Test default random execution
	if code := RunKiw(fs); code != errs.ExitCodeSuccess {
		t.Errorf("RunKiw() = %v, want %v", code, errs.ExitCodeSuccess)
	}

	// Test --all flag
	kiwFlagAll = true
	if code := RunKiw(fs); code != errs.ExitCodeSuccess {
		t.Errorf("RunKiw(--all) = %v, want %v", code, errs.ExitCodeSuccess)
	}
	kiwFlagAll = false

	// Test --json flag
	kiwFlagJSON = true
	if code := RunKiw(fs); code != errs.ExitCodeSuccess {
		t.Errorf("RunKiw(--json) = %v, want %v", code, errs.ExitCodeSuccess)
	}
	kiwFlagAll = true
	if code := RunKiw(fs); code != errs.ExitCodeSuccess {
		t.Errorf("RunKiw(--json --all) = %v, want %v", code, errs.ExitCodeSuccess)
	}
	kiwFlagAll = false
	kiwFlagJSON = false
}
