package cli_test

import (
	"testing"

	"github.com/krewire/krew/internal/cli"
	"github.com/krewire/krewire/pkg/errs"
)

func TestBuildAppAgentCommands(t *testing.T) {
	app := cli.BuildApp("krew")
	if app == nil {
		t.Fatal("BuildApp returned nil")
	}

	agentCommands := []string{"chat", "agent", "daemon", "sessions", "export", "auth"}
	for _, name := range agentCommands {
		code := app.Run([]string{"help", name})
		if code != errs.ExitCodeSuccess {
			t.Errorf("expected command %q to be registered in app, got exit code %v", name, code)
		}
	}
}
