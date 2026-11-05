package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/krewire/krew/pkg/embed"
)

func runSingleShot(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	personaFlag := fs.String("persona", DefaultPersonaName, "Agent persona")
	workspaceFlag := fs.String("workspace", DefaultWorkspaceDir, "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", DefaultMaxSteps, "Maximum steps per prompt")
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

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRunTimeout)
	defer cancel()

	res, err := inst.Run(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", res.Content)
}
