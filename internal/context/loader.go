package context

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/model"
)

// SchemaVersion defines the specification version of the ProjectContext schema.
const SchemaVersion = "1.0.0"

// Load resolves the authoritative ProjectContext for the project at targetDir.
// It never executes arbitrary code or scripts; all discovery is purely declarative.
func Load(targetDir string) (*model.ProjectContext, error) {
	wsRoot := findWorkspaceRoot(targetDir)
	cfg, _ := config.Load(targetDir)

	ctx := &model.ProjectContext{
		Version:   SchemaVersion,
		Workspace: wsRoot,
		Tier:      model.TierFree,
		Pillars: model.PillarInvariants{
			Secure:      make([]model.ContextRule, 0),
			Sustainable: make([]model.ContextRule, 0),
			Scalable:    make([]model.ContextRule, 0),
		},
		Rules: make([]model.ContextRule, 0),
		Gates: defaultQualityGates(),
	}

	// 1. Resolve Project Identity from krewire.yaml (Precedence 3)
	if cfg != nil {
		if cfg.Project.Name != "" {
			ctx.Name = cfg.Project.Name
		}
		if cfg.Project.Kind != "" {
			ctx.Workload = model.Kind(cfg.Project.Kind)
		}
	}
	if ctx.Name == "" {
		ctx.Name = filepath.Base(targetDir)
	}

	// 2. Resolve Module & Toolchain from go.mod (Precedence 3 / 0)
	goModPath := filepath.Join(targetDir, "go.mod")
	if info, err := gomod.Read(goModPath); err == nil && info != nil {
		ctx.ModulePath = info.Path
		ctx.GoVersion = info.GoVersion
	} else if _, err := os.Stat(filepath.Join(targetDir, "go.work")); err == nil {
		ctx.ModulePath = "workspace"
		ctx.Workload = "workspace"
	}

	// 3. Populate Canonical Three Pillars (3S) Invariants (Precedence 2)
	populatePillarRules(ctx, wsRoot, targetDir)

	// 4. Extract Operational Rules from AGENTS.md (Precedence 2)
	extractAgentsRules(ctx, wsRoot, targetDir)

	// 5. Run Integrity Linter to detect missing, conflict, or stale context
	ctx.Diagnostics = Lint(ctx, targetDir)

	return ctx, nil
}

func findWorkspaceRoot(startDir string) string {
	curr := startDir
	for {
		if _, err := os.Stat(filepath.Join(curr, "go.work")); err == nil {
			return curr
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			return startDir
		}
		curr = parent
	}
}

func defaultQualityGates() []model.QualityGateCommand {
	return []model.QualityGateCommand{
		{
			Name:        "formatting",
			Command:     "gofmt -l .",
			Required:    true,
			Description: "All Go code must be formatted according to standard gofmt rules (must return 0 lines).",
		},
		{
			Name:        "static-analysis",
			Command:     "go vet ./...",
			Required:    true,
			Description: "Compiler static analysis must exit with code 0 without unhandled vet warnings.",
		},
		{
			Name:        "test-suite",
			Command:     "go test -v ./...",
			Required:    true,
			Description: "All unit tests, integration tests, and layering guard tests must pass cleanly.",
		},
		{
			Name:        "pure-go-build",
			Command:     "CGO_ENABLED=0 go build ./...",
			Required:    true,
			Description: "Pure-Go invariant: applications and packages must compile cleanly with CGO disabled.",
		},
	}
}

func populatePillarRules(ctx *model.ProjectContext, wsRoot, targetDir string) {
	agentsFile := filepath.Join(wsRoot, "AGENTS.md")
	if _, err := os.Stat(agentsFile); err != nil {
		agentsFile = filepath.Join(targetDir, "AGENTS.md")
	}

	// S1: Secure
	ctx.Pillars.Secure = []model.ContextRule{
		{
			ID:        "PIL-SEC-001",
			Scope:     model.ScopeModule,
			Category:  "pillar",
			Statement: "Stdlib-First: Always prefer Go standard library packages over external dependencies.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 1: Secure",
			},
			EnforcedBy: "layering_test.go",
		},
		{
			ID:        "PIL-SEC-002",
			Scope:     model.ScopeModule,
			Category:  "pillar",
			Statement: "Path-Traversal Prevention (SafePath): Sanitize all file system paths before access.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 1: Secure",
			},
			EnforcedBy: "pkg/sec",
		},
		{
			ID:        "PIL-SEC-003",
			Scope:     model.ScopeWorkspace,
			Category:  "pillar",
			Statement: "Zero Committed Secrets: Never commit secrets, passwords, or tokens; use environment variables.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 1: Secure",
			},
			EnforcedBy: "git-hooks / CI",
		},
	}

	// S2: Sustainable
	ctx.Pillars.Sustainable = []model.ContextRule{
		{
			ID:        "PIL-SUS-001",
			Scope:     model.ScopeModule,
			Category:  "pillar",
			Statement: "Near-Zero Cost: Maintain 10-30 MB RAM footprint and single static binary embed (CGO_ENABLED=0).",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 2: Sustainable",
			},
			EnforcedBy: "CI compilation quality gate",
		},
		{
			ID:        "PIL-SUS-002",
			Scope:     model.ScopeModule,
			Category:  "pillar",
			Statement: "Zero npm / Node.js Runtime: Core libraries and UI components must compile purely in Go.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 2: Sustainable",
			},
			EnforcedBy: "krew build & conformance suite",
		},
	}

	// S3: Scalable
	ctx.Pillars.Scalable = []model.ContextRule{
		{
			ID:        "PIL-SCA-001",
			Scope:     model.ScopeWorkspace,
			Category:  "pillar",
			Statement: "Strict Downward-Only Dependencies: Layer N may import only Layer < N. Higher layers must never be imported.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 3: Scalable",
			},
			EnforcedBy: "layering_test.go",
		},
		{
			ID:        "PIL-SCA-002",
			Scope:     model.ScopeWorkspace,
			Category:  "pillar",
			Statement: "Zero Peer Coupling: Lateral platform services (cloud, auth, dash, krew) never import peer source trees directly.",
			Provenance: model.Provenance{
				File:       agentsFile,
				Precedence: model.PrecedenceWorkspaceRoot,
				SourceKey:  "Pillar 3: Scalable",
			},
			EnforcedBy: "layering_test.go / qualify",
		},
	}

	// Also append all pillar rules to global Rules slice
	ctx.Rules = append(ctx.Rules, ctx.Pillars.Secure...)
	ctx.Rules = append(ctx.Rules, ctx.Pillars.Sustainable...)
	ctx.Rules = append(ctx.Rules, ctx.Pillars.Scalable...)
}

func extractAgentsRules(ctx *model.ProjectContext, wsRoot, targetDir string) {
	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		agentsPath = filepath.Join(wsRoot, "AGENTS.md")
	}

	f, err := os.Open(agentsPath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		text := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(text, "- **") && strings.Contains(text, "**: ") {
			parts := strings.SplitN(text[4:], "**: ", 2)
			ruleName := parts[0]
			ruleStatement := parts[1]

			ruleID := "AGNT-" + strings.ToUpper(strings.ReplaceAll(ruleName, " ", "-"))
			ctx.Rules = append(ctx.Rules, model.ContextRule{
				ID:        ruleID,
				Scope:     model.ScopeWorkspace,
				Category:  "conventions",
				Statement: ruleName + ": " + ruleStatement,
				Provenance: model.Provenance{
					File:       agentsPath,
					Line:       lineNum,
					Precedence: model.PrecedenceWorkspaceRoot,
					SourceKey:  ruleName,
				},
			})
		}
	}
}
