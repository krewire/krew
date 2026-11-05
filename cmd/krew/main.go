package main

import (
	"fmt"
	"os"

	"github.com/krewire/krew"
)

const banner = `
  _  ______  _______          __
 | |/ /  _ \| ____\ \        / /
 | ' /| |_) |  _|  \ \  /\  / / 
 | . \|  _ <| |___  \ \/  \/ /  
 |_|\_\_| \_\_____|  \_/\_/     
 Autonomous AI Agent Engine — v%s
 Platform-Agnostic, Multi-Transport, Embeddable
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "chat":
		runChat(args)
	case "run":
		runSingleShot(args)
	case "serve":
		runServe(args)
	case "mcp":
		runMCP(args)
	case "sessions":
		runSessions(args)
	case "export":
		runExport(args)
	case "version", "-v", "--version":
		fmt.Printf("krew %s\n", krew.Version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Printf(banner, krew.Version)
	fmt.Print(`
USAGE:
  krew <command> [options]

COMMANDS:
  chat     Start an interactive conversational coding session in the terminal
  run      Execute a single prompt or automated task to completion headless
  serve    Launch HTTP/REST and Server-Sent Events (SSE) server for remote hosting
  mcp      Start Model Context Protocol (stdio) server for IDE integration
  sessions List past conversation sessions with message and token counts
  export   Export conversation session transcript to markdown
  version  Print version information
  help     Show this help screen

OPTIONS (chat):
  --resume=<id>                             Resume past conversation session by ID
  --sessions-dir=<path>                     Custom sessions directory (default: ~/.krew/sessions)
  --persona=<coding|devops|worker|general>  Agent persona (default: coding)
  --workspace=<path>                        Working directory root (default: .)
  --model=<name>                            Target model name override
  --max-steps=<n>                           Maximum reasoning loop turns (default: 25)

OPTIONS (run):
  --persona=<coding|devops|worker|general>  Agent persona (default: coding)
  --workspace=<path>                        Working directory root (default: .)
  --model=<name>                            Target model name override
  --max-steps=<n>                           Maximum reasoning loop turns (default: 25)

OPTIONS (serve):
  --port=<port>                             Port to bind (default: 8080)
  --host=<host>                             Host address to bind (default: 0.0.0.0)
  --persona=<name>                          Default persona (default: coding)
  --no-auth                                 Disable authentication for local loopback development

OPTIONS (sessions & export):
  --dir=<path>                              Sessions directory override (default: ~/.krew/sessions)
  --format=<markdown>                       Export format for 'export' (default: markdown)

ENVIRONMENT:
  GEMINI_API_KEY      Google Gemini API token
  OPENAI_API_KEY      OpenAI API token
  ANTHROPIC_API_KEY   Anthropic Claude API token
  OLLAMA_HOST         Ollama server base URL (e.g. http://localhost:11434)
`)
}
