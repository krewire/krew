package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/sdlc"
)

func runSingleShot(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	personaFlag := fs.String("persona", DefaultPersonaName, "Agent persona")
	workspaceFlag := fs.String("workspace", DefaultWorkspaceDir, "Workspace directory")
	modelFlag := fs.String("model", "", "Model override")
	maxStepsFlag := fs.Int("max-steps", DefaultMaxSteps, "Maximum steps per prompt")
	briefFlag := fs.String("brief", "", "Path to markdown Task Brief file or GitHub issue")
	sdlcFlag := fs.Bool("sdlc", false, "Execute governed SDLC workflow (Understand -> Plan -> Implement -> Verify -> Review -> Report)")
	_ = fs.Parse(args)

	remaining := fs.Args()
	prompt := strings.Join(remaining, " ")

	if *briefFlag != "" || *sdlcFlag {
		runSDLCWorkflow(*workspaceFlag, *briefFlag, prompt, *modelFlag)
		return
	}

	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "error: prompt or --brief required for 'krew agent'")
		os.Exit(2)
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

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRunTimeout)
	defer cancel()

	res, err := inst.Run(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", res.Content)
}

func runSDLCWorkflow(workspace, briefPath, prompt, modelOverride string) {
	var brief sdlc.TaskBrief
	var err error

	if briefPath != "" {
		data, err := os.ReadFile(briefPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading brief file %s: %v\n", briefPath, err)
			os.Exit(1)
		}
		title := filepath.Base(briefPath)
		brief, err = sdlc.ParseTaskBriefFromMarkdown(briefPath, title, string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing task brief: %v\n", err)
			os.Exit(1)
		}
	} else {
		brief = sdlc.TaskBrief{
			ID:            fmt.Sprintf("task-%d", time.Now().Unix()),
			Title:         prompt,
			Objective:     prompt,
			Scope:         []string{"."},
			RequirePlan:   true,
			RequireReview: true,
			Timeout:       15 * time.Minute,
			MaxRetries:    2,
		}
	}

	wf, err := sdlc.NewWorkflow(sdlc.WorkflowOptions{
		Workspace: workspace,
		Brief:     brief,
		Model:     modelOverride,
		OnEvent: func(ev sdlc.WorkflowEvent) {
			statusTag := fmt.Sprintf("[%s:%s]", ev.Stage, ev.Status)
			fmt.Printf("\033[1;36m%-24s\033[0m %s\n", statusTag, ev.Message)
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating SDLC workflow: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	result, err := wf.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nWorkflow ended with status: %v\n", err)
		if result != nil && result.Error != "" {
			fmt.Fprintf(os.Stderr, "Reason: %s\n", result.Error)
		}
		os.Exit(1)
	}

	fmt.Printf("\n── SDLC Governed Workflow Completed ─────────────────────\n")
	fmt.Printf("  • Task ID:   %s\n", result.TaskID)
	fmt.Printf("  • Status:    %s\n", result.Status)
	fmt.Printf("  • Duration:  %s\n", result.EndTime.Sub(result.StartTime).Round(time.Millisecond))
	if result.Ledger != nil {
		if ok, err := result.Ledger.VerifyIntegrity(); ok && err == nil {
			fmt.Printf("  • Ledger:    ✓ Verified (%d chained blocks)\n", len(result.Ledger.Blocks))
			ledgerFile := filepath.Join(workspace, ".krewire", "ledger.json")
			_ = result.Ledger.SaveFile(ledgerFile)
		}
	}
	if result.Review != nil && result.Review.ChangeSummary != "" {
		fmt.Printf("\n%s\n", result.Review.ChangeSummary)
	}
}
