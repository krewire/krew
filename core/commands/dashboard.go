package commands

import (
	"flag"

	"github.com/krewire/krew/core/dashboard"
	"github.com/krewire/krewire/errs"
)

const (
	// DefaultDashboardPort is the default HTTP listen port for the developer dashboard.
	DefaultDashboardPort = dashboard.DefaultDashboardPort

	// DefaultDashboardHost is the default host address binding for the developer dashboard.
	DefaultDashboardHost = dashboard.DefaultDashboardHost

	// DefaultDashboardShutdownTimeout is the maximum duration allocated for server drainage.
	DefaultDashboardShutdownTimeout = dashboard.DefaultDashboardShutdownTimeout

	// DefaultDashboardPulseInterval is the SSE telemetry push cadence.
	DefaultDashboardPulseInterval = dashboard.DefaultDashboardPulseInterval
)

// RegisterDashboard registers flags for the dashboard command (alias for RegisterDash).
func RegisterDashboard(fs *flag.FlagSet) {
	RegisterDash(fs)
}

// RunDashboard runs the local dev dashboard (alias for RunDash).
func RunDashboard(fs *flag.FlagSet) errs.ExitCode {
	return RunDash(fs)
}
