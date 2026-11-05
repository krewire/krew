package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krewire/pkg/ai"
)

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
			if len(preview) > MaxToolPreviewLength {
				preview = preview[:MaxToolPreviewLength] + "... (truncated)"
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			fmt.Fprintf(w, "\033[1;32m✔  Result:\033[0m %s\n\n", preview)
		}
	}
}

func runChat(args []string) {
	fs := flag.NewFlagSet("chat", flag.ExitOnError)
	personaFlag := fs.String("persona", DefaultPersonaName, "Agent persona")
	workspaceFlag := fs.String("workspace", DefaultWorkspaceDir, "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", DefaultMaxSteps, "Maximum steps per prompt")
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
