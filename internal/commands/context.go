package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/krewire/krew/internal/context"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/model"
)

var (
	contextFlagJSON  bool
	contextFlagLint  bool
	contextFlagScope string
)

// RegisterContext registers flags for the "krew context" command.
func RegisterContext(fs *flag.FlagSet) {
	fs.BoolVar(&contextFlagJSON, "json", false, "output complete project context in structured JSON for agents")
	fs.BoolVar(&contextFlagLint, "lint", false, "run context integrity checks and return error if missing or conflicting")
	fs.StringVar(&contextFlagScope, "scope", "", "filter context rules by architectural scope (workspace, module, domain, service, unit)")
}

// RunContext inspects, exposes, and validates the project context.
func RunContext(fs *flag.FlagSet) errs.ExitCode {
	targetDir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	ctx, err := context.Load(targetDir)
	if err != nil {
		return fail(err)
	}

	// Filter rules by scope if requested
	if contextFlagScope != "" {
		filterScope := model.Scope(strings.ToLower(strings.TrimSpace(contextFlagScope)))
		var filteredRules []model.ContextRule
		for _, r := range ctx.Rules {
			if r.Scope == filterScope {
				filteredRules = append(filteredRules, r)
			}
		}
		ctx.Rules = filteredRules
	}

	// Output structured JSON if requested
	if contextFlagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(ctx); err != nil {
			return fail(err)
		}
		return errs.ExitCodeSuccess
	}

	// If --lint was explicitly requested, check for blocking issues
	if contextFlagLint {
		hasErrors := false
		fmt.Printf("Running context integrity lint for %s...\n\n", ctx.Name)
		if len(ctx.Diagnostics) == 0 {
			fmt.Println("Context integrity verified: 0 issues detected.")
			return errs.ExitCodeSuccess
		}

		for _, d := range ctx.Diagnostics {
			badge := "[INFO]"
			switch d.Severity {
			case model.DiagConflict:
				badge = "[ERROR]"
				hasErrors = true
			case model.DiagMissing:
				if strings.HasSuffix(d.File, "go.mod") || strings.HasSuffix(d.File, "AGENTS.md") {
					badge = "[ERROR]"
					hasErrors = true
				} else {
					badge = "[INFO]"
				}
			case model.DiagMalformed:
				badge = "[ERROR]"
				hasErrors = true
			case model.DiagStale:
				badge = "[WARN]"
			}

			loc := d.File
			if d.Line > 0 {
				loc = fmt.Sprintf("%s:%d", d.File, d.Line)
			}
			fmt.Printf("%s %s (%s)\n  %s\n", badge, d.Message, loc, d.Hint)
		}

		if hasErrors {
			return errs.ExitCodeFailure
		}
		return errs.ExitCodeSuccess
	}

	// Human-readable summary output
	fmt.Printf("Project Context: %s (schema v%s)\n", ctx.Name, ctx.Version)
	fmt.Printf("  Workload:    %s\n", ctx.Workload)
	fmt.Printf("  Tier:        %s\n", ctx.Tier)
	if ctx.GoVersion != "" {
		fmt.Printf("  Go Version:  %s\n", ctx.GoVersion)
	}
	if ctx.ModulePath != "" {
		fmt.Printf("  Module Path: %s\n", ctx.ModulePath)
	}
	if ctx.Workspace != "" {
		fmt.Printf("  Workspace:   %s\n", ctx.Workspace)
	}

	fmt.Println("\nThree Pillars (3S) Invariants:")
	fmt.Println("  • Secure (S1):")
	for _, r := range ctx.Pillars.Secure {
		fmt.Printf("    - [%s] %s\n", r.ID, r.Statement)
	}
	fmt.Println("  • Sustainable (S2):")
	for _, r := range ctx.Pillars.Sustainable {
		fmt.Printf("    - [%s] %s\n", r.ID, r.Statement)
	}
	fmt.Println("  • Scalable (S3):")
	for _, r := range ctx.Pillars.Scalable {
		fmt.Printf("    - [%s] %s\n", r.ID, r.Statement)
	}

	fmt.Println("\nQuality Gates:")
	for _, g := range ctx.Gates {
		req := "Optional"
		if g.Required {
			req = "Required"
		}
		fmt.Printf("  • %s (%s): %s\n", g.Name, req, g.Command)
	}

	if len(ctx.Diagnostics) > 0 {
		fmt.Printf("\nContext Diagnostics (%d issues detected):\n", len(ctx.Diagnostics))
		for _, d := range ctx.Diagnostics {
			fmt.Printf("  • [%s] %s (%s)\n", d.Severity, d.Message, d.File)
		}
	} else {
		fmt.Println("\nContext Diagnostics: 0 issues detected (authoritative state consistent).")
	}

	return errs.ExitCodeSuccess
}
