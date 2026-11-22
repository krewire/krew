package cli

import (
	"flag"
	"fmt"
	"os"

	aicli "github.com/krewire/krewire/pkg/ai/cli"
	"github.com/krewire/krewire/pkg/errs"
)

func registerChatFlags(fs *flag.FlagSet) {
	fs.String("persona", aicli.DefaultPersonaName, "Agent persona")
	fs.String("workspace", aicli.DefaultWorkspaceDir, "Workspace directory")
	fs.String("model", "", "Model override")
	fs.Int("max-steps", aicli.DefaultMaxSteps, "Maximum steps per prompt")
	fs.String("resume", "", "Resume saved session ID")
	fs.String("sessions-dir", "", "Custom sessions directory")
}

func runChatCmd(fs *flag.FlagSet) errs.ExitCode {
	aicli.RunChat(flagArgs(fs))
	return errs.ExitCodeSuccess
}

func registerAgentFlags(fs *flag.FlagSet) {
	fs.String("persona", aicli.DefaultPersonaName, "Agent persona")
	fs.String("workspace", aicli.DefaultWorkspaceDir, "Workspace directory")
	fs.String("model", "", "Model override")
	fs.Int("max-steps", aicli.DefaultMaxSteps, "Maximum steps per prompt")
	fs.String("brief", "", "Path to markdown Task Brief file or GitHub issue")
	fs.Bool("sdlc", false, "Execute governed SDLC workflow (Understand -> Plan -> Implement -> Verify -> Review -> Report)")
}

func runAgentCmd(fs *flag.FlagSet) errs.ExitCode {
	briefVal := ""
	if f := fs.Lookup("brief"); f != nil {
		briefVal = f.Value.String()
	}
	remaining := fs.Args()
	if len(remaining) == 0 && briefVal == "" {
		fmt.Fprintln(os.Stderr, "error: prompt or --brief required for 'krew agent'")
		return errs.ExitCodeUsage
	}
	aicli.RunSingleShot(flagArgs(fs))
	return errs.ExitCodeSuccess
}

func registerDaemonFlags(fs *flag.FlagSet) {
	fs.Int("port", aicli.DefaultServePort, "Port to listen on")
	fs.String("host", aicli.DefaultServeHost, "Host address to listen on")
	fs.String("persona", aicli.DefaultPersonaName, "Default agent persona")
	fs.String("workspace", aicli.DefaultWorkspaceDir, "Workspace directory")
	fs.String("token", "", "Shared secret Bearer token for authentication")
	fs.Bool("no-auth", false, "Disable authentication for local loopback development")
}

func runDaemonCmd(fs *flag.FlagSet) errs.ExitCode {
	aicli.RunServe(flagArgs(fs))
	return errs.ExitCodeSuccess
}

func registerSessionsFlags(fs *flag.FlagSet) {
	fs.String("dir", "", "Custom sessions directory override")
}

func runSessionsCmd(fs *flag.FlagSet) errs.ExitCode {
	aicli.RunSessions(flagArgs(fs))
	return errs.ExitCodeSuccess
}

func registerExportFlags(fs *flag.FlagSet) {
	fs.String("format", aicli.DefaultExportFormat, "Export format (default: markdown)")
	fs.String("dir", "", "Custom sessions directory override")
}

func runExportCmd(fs *flag.FlagSet) errs.ExitCode {
	aicli.RunExport(flagArgs(fs))
	return errs.ExitCodeSuccess
}

func flagArgs(fs *flag.FlagSet) []string {
	var args []string
	fs.Visit(func(f *flag.Flag) {
		args = append(args, fmt.Sprintf("--%s=%s", f.Name, f.Value.String()))
	})
	args = append(args, fs.Args()...)
	return args
}
