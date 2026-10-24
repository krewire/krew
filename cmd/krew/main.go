package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/krewire/krew"
	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/transport"
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
  version  Print version information
  help     Show this help screen

OPTIONS (chat & run):
  --persona=<coding|devops|worker|general>  Agent persona (default: coding)
  --workspace=<path>                        Working directory root (default: .)
  --model=<name>                            Target model name override
  --max-steps=<n>                           Maximum reasoning loop turns (default: 25)

OPTIONS (serve):
  --port=<port>                             Port to bind (default: 8080)
  --host=<host>                             Host address to bind (default: 0.0.0.0)
  --persona=<name>                          Default persona (default: coding)

ENVIRONMENT:
  GEMINI_API_KEY      Google Gemini API token
  OPENAI_API_KEY      OpenAI API token
  ANTHROPIC_API_KEY   Anthropic Claude API token
  OLLAMA_HOST         Ollama server base URL (e.g. http://localhost:11434)
`)
}

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

func runChat(args []string) {
	fs := flag.NewFlagSet("chat", flag.ExitOnError)
	personaFlag := fs.String("persona", "coding", "Agent persona")
	workspaceFlag := fs.String("workspace", ".", "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", 25, "Maximum steps per prompt")
	_ = fs.Parse(args)

	pers := parsePersona(*personaFlag)

	inst, err := embed.New(embed.Options{
		Workspace: *workspaceFlag,
		Persona:   &pers,
		Model:     *modelFlag,
		MaxSteps:  *maxStepsFlag,
		OnEvent:   cliEventHandler,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\033[1;36m[Krew Agent]\033[0m Initialized persona \033[1m%s\033[0m in workspace \033[1m%s\033[0m\n", pers.Name, *workspaceFlag)
	fmt.Println("Type your instruction or message. Press Ctrl+C or type 'exit' to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("\033[1;32myou>\033[0m ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		fmt.Println()
		res, err := inst.Chat(ctx, line)
		if err != nil {
			fmt.Printf("\033[1;31mError:\033[0m %v\n\n", err)
			continue
		}

		fmt.Printf("\n\033[1;34m[krew]\033[0m %s\n\n", res.Content)
	}
}

func runSingleShot(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	personaFlag := fs.String("persona", "coding", "Agent persona")
	workspaceFlag := fs.String("workspace", ".", "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", 25, "Maximum steps per prompt")
	_ = fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "error: prompt required for 'krew run'")
		os.Exit(2)
	}
	prompt := strings.Join(remaining, " ")

	pers := parsePersona(*personaFlag)

	inst, err := embed.New(embed.Options{
		Workspace: *workspaceFlag,
		Persona:   &pers,
		Model:     *modelFlag,
		MaxSteps:  *maxStepsFlag,
		OnEvent:   cliEventHandler,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing agent: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	res, err := inst.Run(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", res.Content)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.Int("port", 8080, "Port to listen on")
	hostFlag := fs.String("host", "0.0.0.0", "Host address to listen on")
	personaFlag := fs.String("persona", "coding", "Default agent persona")
	workspaceFlag := fs.String("workspace", ".", "Workspace directory")
	_ = fs.Parse(args)

	pers := parsePersona(*personaFlag)

	inst, err := embed.New(embed.Options{
		Workspace: *workspaceFlag,
		Persona:   &pers,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing agent: %v\n", err)
		os.Exit(1)
	}

	server := transport.NewHTTPServer(inst.Agent())
	addr := fmt.Sprintf("%s:%d", *hostFlag, *portFlag)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Handler(),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("\033[1;36m[Krew Server]\033[0m Listening on http://%s (persona: %s)\n", addr, pers.Name)
		fmt.Println("Available endpoints:")
		fmt.Printf("  GET  http://%s/health\n", addr)
		fmt.Printf("  POST http://%s/api/agent/run  (stream=true for SSE)\n", addr)
		fmt.Printf("  POST http://%s/api/agent/chat\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	<-stop
	fmt.Println("\nShutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	fmt.Println("Server stopped.")
}

func runMCP(args []string) {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	personaFlag := fs.String("persona", "coding", "Agent persona")
	workspaceFlag := fs.String("workspace", ".", "Workspace directory")
	_ = fs.Parse(args)

	pers := parsePersona(*personaFlag)

	inst, err := embed.New(embed.Options{
		Workspace: *workspaceFlag,
		Persona:   &pers,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing agent for MCP: %v\n", err)
		os.Exit(1)
	}

	mcpServer := transport.NewMCPServer(inst.Agent(), os.Stdin, os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := mcpServer.Serve(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "MCP server exited: %v\n", err)
		os.Exit(1)
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
		if len(preview) > 120 {
			preview = preview[:120] + "... (truncated)"
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		fmt.Printf("\033[1;32m✔  Result:\033[0m %s\n", preview)
	case agent.EventError:
		fmt.Printf("\033[1;31m✖  Error:\033[0m %s\n", ev.Error)
	}
}
