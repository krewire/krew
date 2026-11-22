package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/krewire/krew/internal/dashboard"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/tui/term"
)

// RegisterDash registers flags for the dash command group.
func RegisterDash(fs *flag.FlagSet) {
	fs.String("port", DefaultDashboardPort, "port for local dev dashboard or console HTTP server")
	fs.String("host", DefaultDashboardHost, "host address to bind HTTP server")
	fs.String("sandbox", "", "sandbox image/runtime environment (e.g. go-1.27, node-22, python-3.13)")
	fs.Bool("pty", true, "attach interactive pseudo-terminal WebSocket")
	fs.Bool("dry-run", false, "simulate sandbox execution or server startup without running")
	registerRuntimeFlags(fs)
}

const dashUsage = `krew dash — AI-native virtual workspace, execution sandboxes & local dev dashboard

Usage:
  krew dash [--port=4000] [--host=127.0.0.1]                  start local dev dashboard, metrics & MCP server
  krew dash start [--sandbox=<image>] [--port=9000] [--pty]  start or attach to a virtual workspace sandbox
  krew dash stop [id]                                         stop an active virtual workspace or sandbox
  krew dash list                                              list running virtual workspaces and sandboxes
  krew dash status                                            inspect Dash server connectivity, quota & health
  krew dash console [--port=9000]                            open or print Web IDE console URL

Examples:
  krew dash
  krew dash --port=4000
  krew dash start --sandbox=go-1.27
  krew dash start --sandbox=node-22 --dry-run
  krew dash list
  krew dash status
  krew dash console
`

// RunDash dispatches krew dash subcommands or starts the local dev dashboard.
func RunDash(fs *flag.FlagSet) errs.ExitCode {
	sub := ""
	if fs != nil && fs.NArg() > 0 {
		sub = fs.Arg(0)
	}

	switch sub {
	case "start":
		return runDashStart(fs)
	case "stop":
		return runDashStop(fs)
	case "list", "ls":
		return runDashList(fs)
	case "status":
		return runDashStatus(fs)
	case "console":
		return runDashConsole(fs)
	case "help", "-h", "--help":
		fmt.Print(dashUsage)
		return errs.ExitCodeSuccess
	case "", "server", "dev", "dashboard":
		return runDashServer(fs)
	default:
		fmt.Fprintf(os.Stderr, "krew: unknown dash sub-command %q\n\n", sub)
		fmt.Print(dashUsage)
		return errs.ExitCodeUsage
	}
}

func runDashServer(fs *flag.FlagSet) errs.ExitCode {
	dryRun := flagValue(fs, "dry-run") == "true"
	port := flagValue(fs, "port")
	if port == "" {
		port = DefaultDashboardPort
	}
	host := flagValue(fs, "host")
	if host == "" {
		host = DefaultDashboardHost
	}

	if dryRun {
		fmt.Printf("[dry-run] would start dev dashboard on http://%s:%s\n", host, port)
		return errs.ExitCodeSuccess
	}

	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		return code
	}

	explicit := firstNonEmpty(flagValue(fs, "kind"), rt.cfg.Kind())
	res, err := shape.Detect(rt.root, explicit)
	var kind string
	if err == nil && res.Kind != shape.KindNone {
		kind = string(res.Kind)
	} else if rt.cfg.Kind() != "" {
		kind = rt.cfg.Kind()
	} else {
		kind = "app"
	}

	srv := dashboard.NewServer(rt.root, rt.cfg, kind, string(rt.env), host, port)
	if err := srv.Start(); err != nil {
		return fail(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultDashboardShutdownTimeout)
		defer cancel()
		_ = srv.Stop(ctx)
	}()

	tm := term.NewTerminal()
	boldCyan := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }

	addr := srv.Addr()
	info := srv.CollectInfo()

	fmt.Printf("\n  %s\n", boldCyan("Krewire Dev Dashboard"))
	fmt.Printf("  %s %s (kind: %s, env: %s)\n", dim("Project:    "), info.Name, info.Kind, info.Env)
	fmt.Printf("  %s http://%s\n", dim("Dashboard:  "), green(addr))
	fmt.Printf("  %s http://%s/api/health\n", dim("Health API: "), addr)
	fmt.Printf("  %s http://%s/api/metrics\n", dim("Metrics API:"), addr)
	fmt.Printf("  %s http://%s/mcp/sse\n\n", dim("MCP Server: "), addr)
	fmt.Printf("  %s\n\n", dim("Press Ctrl+C to stop."))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down dev dashboard...")
	return errs.ExitCodeSuccess
}
