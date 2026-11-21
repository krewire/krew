package context_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/internal/context"
	"github.com/krewire/krewire/pkg/model"
)

func TestLoad_Success(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create go.mod
	goModContent := `module example.com/testapp

go 1.23.0
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// 2. Create krewire.yaml
	krewireYamlContent := `project:
  name: testapp
  kind: app
`
	if err := os.WriteFile(filepath.Join(tempDir, "krewire.yaml"), []byte(krewireYamlContent), 0644); err != nil {
		t.Fatalf("failed to write krewire.yaml: %v", err)
	}

	// 3. Create AGENTS.md
	agentsContent := `# Operational Guide
- **Code Style**: Pure-Go without CGO.
- **Language**: English for all stored files.
`
	if err := os.WriteFile(filepath.Join(tempDir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatalf("failed to write AGENTS.md: %v", err)
	}

	// Load context
	ctx, err := context.Load(tempDir)
	if err != nil {
		t.Fatalf("context.Load failed: %v", err)
	}

	if ctx.Name != "testapp" {
		t.Errorf("expected name 'testapp', got %q", ctx.Name)
	}
	if ctx.Workload != model.KindApp {
		t.Errorf("expected workload 'app', got %q", ctx.Workload)
	}
	if ctx.ModulePath != "example.com/testapp" {
		t.Errorf("expected module path 'example.com/testapp', got %q", ctx.ModulePath)
	}
	if ctx.GoVersion != "1.23.0" {
		t.Errorf("expected Go version '1.23.0', got %q", ctx.GoVersion)
	}

	// Check Three Pillars
	if len(ctx.Pillars.Secure) == 0 {
		t.Errorf("expected Secure pillar rules to be populated")
	}
	if len(ctx.Pillars.Sustainable) == 0 {
		t.Errorf("expected Sustainable pillar rules to be populated")
	}
	if len(ctx.Pillars.Scalable) == 0 {
		t.Errorf("expected Scalable pillar rules to be populated")
	}

	// Check Quality Gates
	if len(ctx.Gates) != 4 {
		t.Errorf("expected 4 default quality gates, got %d", len(ctx.Gates))
	}

	// Check that extracted AGENTS.md rules exist
	foundStyleRule := false
	for _, r := range ctx.Rules {
		if r.ID == "AGNT-CODE-STYLE" {
			foundStyleRule = true
			if r.Provenance.Precedence != model.PrecedenceWorkspaceRoot {
				t.Errorf("expected PrecedenceWorkspaceRoot, got %v", r.Provenance.Precedence)
			}
			break
		}
	}
	if !foundStyleRule {
		t.Errorf("expected AGNT-CODE-STYLE rule extracted from AGENTS.md")
	}

	// Ensure no lint diagnostics
	if len(ctx.Diagnostics) > 0 {
		t.Errorf("expected 0 diagnostics, got %d: %+v", len(ctx.Diagnostics), ctx.Diagnostics)
	}
}

func TestLint_MissingFiles(t *testing.T) {
	tempDir := t.TempDir()

	ctx, err := context.Load(tempDir)
	if err != nil {
		t.Fatalf("context.Load failed: %v", err)
	}

	// Expect diagnostics for missing go.mod and krewire.yaml
	var missingMod, missingYaml bool
	for _, d := range ctx.Diagnostics {
		if d.Severity == model.DiagMissing {
			if filepath.Base(d.File) == "go.mod" {
				missingMod = true
			}
			if filepath.Base(d.File) == "krewire.yaml" {
				missingYaml = true
			}
		}
	}

	if !missingMod {
		t.Errorf("expected diagnostic for missing go.mod")
	}
	if !missingYaml {
		t.Errorf("expected diagnostic for missing krewire.yaml")
	}
}

func TestLint_DuplicateRules(t *testing.T) {
	ctx := &model.ProjectContext{
		Rules: []model.ContextRule{
			{
				ID:        "RULE-001",
				Statement: "First declaration",
				Provenance: model.Provenance{
					File: "file_a.go",
				},
			},
			{
				ID:        "RULE-001",
				Statement: "Conflicting duplicate declaration",
				Provenance: model.Provenance{
					File: "file_b.go",
				},
			},
		},
	}

	diags := context.Lint(ctx, t.TempDir())
	foundConflict := false
	for _, d := range diags {
		if d.Severity == model.DiagConflict {
			foundConflict = true
			break
		}
	}

	if !foundConflict {
		t.Errorf("expected DiagConflict diagnostic for duplicate rule ID")
	}
}

func TestPrecedenceTier_Ordering(t *testing.T) {
	tiers := []model.PrecedenceTier{
		model.PrecedenceToolchain,
		model.PrecedenceCoreKernel,
		model.PrecedenceWorkspaceRoot,
		model.PrecedenceProjectConfig,
		model.PrecedenceArchitecture,
		model.PrecedenceActiveTask,
	}

	for i := 1; i < len(tiers); i++ {
		if tiers[i] <= tiers[i-1] {
			t.Errorf("expected tier %v > tier %v", tiers[i], tiers[i-1])
		}
	}
}
