package commands

import (
	"flag"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunTier_Basic(t *testing.T) {
	fs := flag.NewFlagSet("tier", flag.ContinueOnError)
	RegisterTier(fs)
	_ = fs.Parse([]string{})

	code := RunTier(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for krew tier, got %v", code)
	}
}

func TestRunTier_JSON(t *testing.T) {
	fs := flag.NewFlagSet("tier", flag.ContinueOnError)
	RegisterTier(fs)
	_ = fs.Parse([]string{"--json"})

	code := RunTier(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for krew tier --json, got %v", code)
	}
}

func TestRunTier_CheckSuccess(t *testing.T) {
	fs := flag.NewFlagSet("tier", flag.ContinueOnError)
	RegisterTier(fs)
	_ = fs.Parse([]string{"--check", "free"})

	code := RunTier(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for --check free on free workspace, got %v", code)
	}
}

func TestRunTier_CheckInvalidTier(t *testing.T) {
	fs := flag.NewFlagSet("tier", flag.ContinueOnError)
	RegisterTier(fs)
	_ = fs.Parse([]string{"--check", "invalid-tier"})

	code := RunTier(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected ExitCodeUsage for invalid tier, got %v", code)
	}
}
