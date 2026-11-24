package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/model"
	"github.com/krewire/tui/term"
)

var (
	tierFlagJSON  bool
	tierFlagCheck string
)

// RegisterTier registers flags for the "krew tier" command.
func RegisterTier(fs *flag.FlagSet) {
	fs.BoolVar(&tierFlagJSON, "json", false, "output tier posture in structured JSON")
	fs.StringVar(&tierFlagCheck, "check", "", "assert project does not exceed specified tier (e.g. --check free)")
}

// TierReport summarizes the detected open-core tier for a project.
type TierReport struct {
	ProjectName   string       `json:"project_name"`
	EffectiveTier model.Tier   `json:"effective_tier"`
	AllTiers      []model.Tier `json:"ladder"`
	Modules       []string     `json:"modules,omitempty"`
	OpenSource    bool         `json:"open_source"`
	License       string       `json:"license"`
	Compliant     bool         `json:"compliant"`
}

// RunTier inspects and validates the open-core tier of the current project.
func RunTier(fs *flag.FlagSet) errs.ExitCode {
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	cfg, _ := config.Load(root)
	effectiveTier := model.TierFree

	projectName := "workspace"
	if cfg != nil && cfg.Project.Name != "" {
		projectName = cfg.Project.Name
	}

	// Detect if project uses higher tier features
	var detectedModules []string
	if cfg != nil {
		if cfg.Worker != nil && (cfg.Worker.Concurrency > 0 || len(cfg.Worker.Queues) > 0) {
			detectedModules = append(detectedModules, "worker")
		}
		if cfg.Service != nil && (cfg.Service.Name != "" || cfg.Service.Port > 0) {
			detectedModules = append(detectedModules, "service")
		}
		if cfg.Infra != nil && cfg.Infra.Provider != "" {
			detectedModules = append(detectedModules, "infra")
		}
	}

	report := TierReport{
		ProjectName:   projectName,
		EffectiveTier: effectiveTier,
		AllTiers:      model.AllTiers,
		Modules:       detectedModules,
		OpenSource:    true,
		License:       "MIT (100% Open Source)",
		Compliant:     true,
	}

	if tierFlagCheck != "" {
		maxAllowed, err := model.ParseTier(tierFlagCheck)
		if err != nil {
			return usageMessage(fmt.Sprintf("invalid --check tier %q (choose: free, pro, team, enterprise)", tierFlagCheck))
		}
		if effectiveTier.Rank() > maxAllowed.Rank() {
			report.Compliant = false
			if tierFlagJSON {
				_ = json.NewEncoder(os.Stdout).Encode(report)
			} else {
				fmt.Fprintf(os.Stderr, "krew tier error: project tier %q exceeds maximum allowed %q\n", effectiveTier, maxAllowed)
			}
			return errs.ExitCodeFailure
		}
	}

	if tierFlagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fail(err)
		}
		return errs.ExitCodeSuccess
	}

	tm := term.NewTerminal()
	boldCyan := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, []term.Style{term.StyleBold}) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }

	fmt.Printf("\n  %s — %s\n\n", boldCyan("Krewire Open Core Tier"), dim("single source of truth"))
	fmt.Printf("  %s %s\n", dim("Project:       "), projectName)
	fmt.Printf("  %s %s\n", dim("Effective Tier:"), green(string(effectiveTier)))
	fmt.Printf("  %s %s\n", dim("Licensing:     "), green("MIT (100% Open Source, Forever)"))
	fmt.Printf("  %s %s\n\n", dim("Boundary Gate: "), green("Compliant — No paid prerequisites for free tier"))

	fmt.Printf("  %s\n", boldCyan("TIER LADDER:"))
	ladderDesc := map[model.Tier]string{
		model.TierFree:       "Complete for students, UMKM, hobbyists & evaluation (Forever free, MIT)",
		model.TierPro:        "Individual professional developers (Hardened reliability & primitives)",
		model.TierTeam:       "Multi-contributor teams & agencies (Shared config, policy & fleet governance)",
		model.TierEnterprise: "Custom B2B contracts (Negotiated compliance, dedicated SLA & audit)",
	}

	for _, t := range model.AllTiers {
		prefix := "  "
		tierDisplay := string(t)
		if t == effectiveTier {
			prefix = "• "
			tierDisplay = green(tierDisplay + " [active]")
		} else {
			tierDisplay = yellow(tierDisplay)
		}
		fmt.Printf("  %s%-20s %s\n", prefix, tierDisplay, dim(ladderDesc[t]))
	}

	fmt.Println()
	return errs.ExitCodeSuccess
}

func printTierUsage() errs.ExitCode {
	fmt.Println(`usage: krew tier [--check=<tier>] [--json]`)
	return errs.ExitCodeSuccess
}

func parseTierName(s string) (model.Tier, error) {
	return model.ParseTier(strings.TrimSpace(s))
}
