package context

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krewire/krewire/pkg/model"
)

// Lint checks a ProjectContext for missing sources, conflicts, and duplicate rules.
func Lint(ctx *model.ProjectContext, targetDir string) []model.ContextDiagnostic {
	var diags []model.ContextDiagnostic

	// 1. Check for missing critical files
	goModPath := filepath.Join(targetDir, "go.mod")
	goWorkPath := filepath.Join(targetDir, "go.work")
	_, modErr := os.Stat(goModPath)
	_, workErr := os.Stat(goWorkPath)
	if os.IsNotExist(modErr) && os.IsNotExist(workErr) {
		diags = append(diags, model.ContextDiagnostic{
			Severity: model.DiagMissing,
			Scope:    model.ScopeModule,
			File:     goModPath,
			Message:  "Missing go.mod or go.work file in target directory",
			Hint:     "Initialize a Go module with 'go mod init <path>' or workspace with 'go work init'.",
		})
	}

	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		wsAgents := filepath.Join(ctx.Workspace, "AGENTS.md")
		if _, err := os.Stat(wsAgents); os.IsNotExist(err) {
			diags = append(diags, model.ContextDiagnostic{
				Severity: model.DiagMissing,
				Scope:    model.ScopeWorkspace,
				File:     agentsPath,
				Message:  "Missing AGENTS.md in project or workspace root",
				Hint:     "Create AGENTS.md to define operational rules and pillar invariants.",
			})
		}
	}

	krewireYamlPath := filepath.Join(targetDir, "krewire.yaml")
	if _, err := os.Stat(krewireYamlPath); os.IsNotExist(err) {
		diags = append(diags, model.ContextDiagnostic{
			Severity: model.DiagMissing,
			Scope:    model.ScopeModule,
			File:     krewireYamlPath,
			Message:  "Optional krewire.yaml configuration file not present",
			Hint:     "Pure Go libraries do not require krewire.yaml. Run 'krew init' if configuring an application or service.",
		})
	}

	// 2. Check for duplicate rule IDs
	seenRules := make(map[string]model.ContextRule)
	for _, rule := range ctx.Rules {
		if existing, exists := seenRules[rule.ID]; exists {
			diags = append(diags, model.ContextDiagnostic{
				Severity: model.DiagConflict,
				Scope:    rule.Scope,
				File:     rule.Provenance.File,
				Line:     rule.Provenance.Line,
				Message:  fmt.Sprintf("Duplicate rule ID '%s' conflicting with declaration in %s", rule.ID, existing.Provenance.File),
				Hint:     "Ensure all ContextRule IDs are globally unique within the workspace.",
			})
		} else {
			seenRules[rule.ID] = rule
		}
	}

	return diags
}
