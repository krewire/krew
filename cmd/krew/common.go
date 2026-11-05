package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/persona"
)

const (
	// DefaultServePort is the default TCP listen port for the HTTP/SSE daemon.
	DefaultServePort = 8080

	// DefaultServeHost is the default host address binding for the daemon.
	DefaultServeHost = "0.0.0.0"

	// DefaultPersonaName is the default persona when none is specified.
	DefaultPersonaName = "coding"

	// DefaultWorkspaceDir is the fallback working directory root.
	DefaultWorkspaceDir = "."

	// DefaultMaxSteps is the maximum reasoning turns allocated per instruction.
	DefaultMaxSteps = 25

	// DefaultExportFormat is the default serialization format for transcripts.
	DefaultExportFormat = "markdown"

	// DefaultRunTimeout is the timeout ceiling for headless single-shot tasks.
	DefaultRunTimeout = 10 * time.Minute

	// DefaultShutdownTimeout is the graceful drain timeout on SIGINT/SIGTERM.
	DefaultShutdownTimeout = 5 * time.Second

	// MaxToolPreviewLength bounds the terminal output snippet for tool results.
	MaxToolPreviewLength = 120
)

func parsePersona(name string) persona.Persona {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "devops", "ops", "sre":
		return persona.DevOpsRunner
	case "worker", "autonomous":
		return persona.AutonomousWorker
	case "general", "assistant":
		return persona.GeneralAssistant
	case "coding", "code", "dev":
		fallthrough
	default:
		return persona.CodingAssistant
	}
}

func cliEventHandler(ev agent.Event) {
	switch ev.Type {
	case agent.EventThought:
		fmt.Printf("\033[2m[%s] %s\033[0m\n", ev.Timestamp.Format("15:04:05"), ev.Content)
	case agent.EventToolCall:
		fmt.Printf("\033[1;33m⚙  Tool Call:\033[0m %s(%s)\n", ev.ToolName, ev.ToolArgs)
	case agent.EventToolResult:
		preview := fmt.Sprintf("%v", ev.Result)
		if len(preview) > MaxToolPreviewLength {
			preview = preview[:MaxToolPreviewLength] + "... (truncated)"
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		fmt.Printf("\033[1;32m✔  Result:\033[0m %s\n", preview)
	case agent.EventError:
		fmt.Printf("\033[1;31m✖  Error:\033[0m %s\n", ev.Error)
	}
}
