package commands

import (
	"errors"
	"slices"
	"testing"

	"github.com/krewire/krewire/pkg/env"
	"github.com/krewire/krewire/pkg/errs"
)

func TestChildEnviron_ExportsRVEnvAndDebug(t *testing.T) {
	base := []string{"PATH=/bin"}
	got := childEnviron(base, ":8080", env.EnvTesting, true)
	want := []string{"PATH=/bin", "APP_ADDR=:8080", "KIW_ENV=testing", "KIW_DEBUG=1"}
	if !slices.Equal(got, want) {
		t.Errorf("childEnviron = %v, want %v", got, want)
	}

	noAddr := childEnviron(nil, "", env.EnvProduction, false)
	wantNoAddr := []string{"KIW_ENV=production", "KIW_DEBUG=0"}
	if !slices.Equal(noAddr, wantNoAddr) {
		t.Errorf("childEnviron(no addr) = %v, want %v", noAddr, wantNoAddr)
	}
}

func TestUsageOrFail_MapsUsageErrorsToExit2(t *testing.T) {
	if got := usageOrFail(errs.UsageError("bad")); got != errs.ExitCodeUsage {
		t.Errorf("usageOrFail(UsageError) = %d, want %d", got, errs.ExitCodeUsage)
	}
	if got := usageOrFail(errors.New("boom")); got != errs.ExitCodeFailure {
		t.Errorf("usageOrFail(plain error) = %d, want %d", got, errs.ExitCodeFailure)
	}
}
