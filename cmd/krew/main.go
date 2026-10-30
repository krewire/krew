package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/krewire/krew"
	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/transport"
	"github.com/krewire/krewire/pkg/ai"
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

func replaySession(w io.Writer, msgs []ai.Message) {
	for _, msg := range msgs {
		switch msg.Role {
		case ai.RoleUser:
			fmt.Fprintf(w, "\033[1;32myou>\033[0m %s\n\n", msg.Content)
		case ai.RoleAssistant:
			for _, tc := range msg.ToolCalls {
				fmt.Fprintf(w, "\033[1;33m⚙  Tool Call:\033[0m %s(%s)\n", tc.Name, tc.Arguments)
			}
			if msg.Content != "" {
				fmt.Fprintf(w, "\033[1;34m[krew]\033[0m %s\n\n", msg.Content)
			}
		case ai.RoleTool:
			preview := msg.Content
			if len(preview) > 120 {
				preview = preview[:120] + "... (truncated)"
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Fprintf(w, "\033[1;32m✔  Result:\033[0m %s\n\n", preview)
		}
	}
}

func printSessionList(w io.Writer, sessions []memory.SessionInfo) {
	fmt.Fprintln(w, "Saved Sessions:")
	for _, s := range sessions {
		fmt.Fprintf(w, "  • %-36s  %s  (%d msgs, %d tokens)\n",
			s.ID,
			s.ModTime.Format("2006-01-02 15:04:05"),
			s.Messages,
			s.Tokens,
		)
	}
}

func runChat(args []string) {
	fs := flag.NewFlagSet("chat", flag.ExitOnError)
	personaFlag := fs.String("persona", "coding", "Agent persona")
	workspaceFlag := fs.String("workspace", ".", "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", 25, "Maximum steps per prompt")
	resumeFlag := fs.String("resume", "", "Resume saved session ID")
	sessionsDirFlag := fs.String("sessions-dir", "", "Custom sessions directory")
	_ = fs.Parse(args)

	sessDir := *sessionsDirFlag
	if sessDir == "" {
		var err error
		sessDir, err = memory.DefaultSessionsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error locating sessions directory: %v\n", err)
			os.Exit(1)
		}
	}

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

	var sessLog *memory.SessionLog
	var activeSessionID string

	if *resumeFlag != "" {
		history, err := memory.LoadSession(sessDir, *resumeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading session %q: %v\n", *resumeFlag, err)
			os.Exit(1)
		}
		inst.Memory().Add(history...)

		log, err := memory.OpenSessionLog(sessDir, *resumeFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening session log %q: %v\n", *resumeFlag, err)
			os.Exit(1)
		}
		sessLog = log
		activeSessionID = *resumeFlag
		defer sessLog.Close()

		fmt.Printf("\033[1;36m[Krew Agent]\033[0m Resumed session \033[1m%s\033[0m (%d messages)\n\n", *resumeFlag, len(history))
		replaySession(os.Stdout, history)
	} else {
		fmt.Printf("\033[1;36m[Krew Agent]\033[0m Initialized persona \033[1m%s\033[0m in workspace \033[1m%s\033[0m\n", pers.Name, *workspaceFlag)
		fmt.Println("Type your instruction or message. Press Ctrl+C or type 'exit' to quit.")
		fmt.Println()
	}

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

		if sessLog == nil {
			log, id, err := memory.StartSession(sessDir, line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to start session log: %v\n", err)
			} else {
				sessLog = log
				activeSessionID = id
				defer sessLog.Close()
				fmt.Printf("\033[2m[Session saved: %s]\033[0m\n\n", activeSessionID)
			}
		}

		if sessLog != nil {
			_ = sessLog.Append(ai.UserMessage(line))
		}

		fmt.Println()
		res, err := inst.Chat(ctx, line)
		if err != nil {
			fmt.Printf("\033[1;31mError:\033[0m %v\n\n", err)
			continue
		}

		if sessLog != nil {
			_ = sessLog.Append(ai.AssistantMessage(res.Content))
		}

		fmt.Printf("\n\033[1;34m[krew]\033[0m %s\n\n", res.Content)
	}
}

func runSessions(args []string) {
	fs := flag.NewFlagSet("sessions", flag.ExitOnError)
	dirFlag := fs.String("dir", "", "Custom sessions directory")
	_ = fs.Parse(args)

	sessDir := *dirFlag
	if sessDir == "" {
		var err error
		sessDir, err = memory.DefaultSessionsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error locating sessions directory: %v\n", err)
			os.Exit(1)
		}
	}

	sessions, err := memory.ListSessions(sessDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No recorded sessions found.")
		return
	}

	printSessionList(os.Stdout, sessions)
}

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	formatFlag := fs.String("format", "markdown", "Export format (default: markdown)")
	dirFlag := fs.String("dir", "", "Custom sessions directory")
	_ = fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "error: session ID required for 'krew export'")
		os.Exit(2)
	}
	sessionID := remaining[0]

	sessDir := *dirFlag
	if sessDir == "" {
		var err error
		sessDir, err = memory.DefaultSessionsDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error locating sessions directory: %v\n", err)
			os.Exit(1)
		}
	}

	if *formatFlag != "markdown" && *formatFlag != "md" {
		fmt.Fprintf(os.Stderr, "unsupported export format %q (supported: markdown)\n", *formatFlag)
		os.Exit(2)
	}

	rendered, err := memory.RenderTranscriptMarkdown(sessDir, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting session %s: %v\n", sessionID, err)
		os.Exit(1)
	}

	fmt.Print(rendered)
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
	noAuthFlag := fs.Bool("no-auth", false, "Disable authentication for local loopback development")
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

	server := transport.NewHTTPServer(inst.Agent(), transport.WithNoAuth(*noAuthFlag))
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
