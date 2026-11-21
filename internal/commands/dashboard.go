package commands

import (
	"flag"
	"time"

	"github.com/krewire/krewire/pkg/errs"
)

const (
	// DefaultDashboardPort is the default HTTP listen port for the developer dashboard.
	DefaultDashboardPort = "4000"

	// DefaultDashboardHost is the default host address binding for the developer dashboard.
	DefaultDashboardHost = "127.0.0.1"

	// DefaultDashboardShutdownTimeout is the maximum duration allocated for server drainage.
	DefaultDashboardShutdownTimeout = 3 * time.Second

	// DefaultDashboardPulseInterval is the SSE telemetry push cadence.
	DefaultDashboardPulseInterval = 2 * time.Second
)

// RegisterDashboard registers flags for the dashboard command (alias for RegisterDash).
func RegisterDashboard(fs *flag.FlagSet) {
	RegisterDash(fs)
}

// RunDashboard runs the local dev dashboard (alias for RunDash).
func RunDashboard(fs *flag.FlagSet) errs.ExitCode {
	return RunDash(fs)
}
