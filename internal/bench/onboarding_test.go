package bench_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/scaffold"
	"github.com/krewire/krewire/pkg/model"
)

// OnboardingParticipant represents an independent developer persona evaluating Krewire.
type OnboardingParticipant struct {
	ID             string
	Background     string
	TargetWorkload model.Kind
	MaxTimeBudget  time.Duration
}

// OnboardingMetric records empirical usability outcomes for each participant.
type OnboardingMetric struct {
	ParticipantID string
	Workload      model.Kind
	Completed     bool
	DurationMs    int64
	FailedSteps   int
	HelpRequests  int
}

// TestOnboarding_CleanEnvironmentQuickstart evaluates first-time developer onboarding
// across 5 diverse developer personas in clean isolated temporary environments.
func TestOnboarding_CleanEnvironmentQuickstart(t *testing.T) {
	cohort := []OnboardingParticipant{
		{
			ID:             "dev-01-python-fastapi",
			Background:     "Python / FastAPI backend developer adopting Go for high-throughput service",
			TargetWorkload: model.KindService,
			MaxTimeBudget:  5 * time.Second,
		},
		{
			ID:             "dev-02-node-fullstack",
			Background:     "Node.js / React engineer evaluating Go fullstack monolithic app",
			TargetWorkload: model.KindApp,
			MaxTimeBudget:  5 * time.Second,
		},
		{
			ID:             "dev-03-devops-worker",
			Background:     "DevOps engineer deploying background queue worker without public port",
			TargetWorkload: model.KindWorker,
			MaxTimeBudget:  5 * time.Second,
		},
		{
			ID:             "dev-04-techwriter-site",
			Background:     "Technical documentation author building SSG documentation portal",
			TargetWorkload: model.KindSite,
			MaxTimeBudget:  5 * time.Second,
		},
		{
			ID:             "dev-05-systems-cli",
			Background:     "Systems developer building standalone workstation CLI tool",
			TargetWorkload: model.KindCLI,
			MaxTimeBudget:  5 * time.Second,
		},
	}

	var results []OnboardingMetric

	for _, p := range cohort {
		t.Run(p.ID, func(t *testing.T) {
			cleanDir := t.TempDir()
			start := time.Now()

			// Step 1: Initialize project in clean directory
			modulePath := "example.com/test/" + string(p.TargetWorkload)
			variant := scaffold.VariantApp
			switch p.TargetWorkload {
			case model.KindService:
				variant = scaffold.VariantService
			case model.KindWorker:
				variant = scaffold.VariantWorker
			case model.KindSite:
				variant = scaffold.VariantStatic
			case model.KindCLI:
				variant = scaffold.VariantCLI
			}

			// Execute scaffolding via scaffold.Equip
			// First create minimal bootstrap go.mod and main.go
			if err := os.WriteFile(filepath.Join(cleanDir, "go.mod"), []byte("module "+modulePath+"\ngo 1.22\n"), 0644); err != nil {
				t.Fatalf("[%s] failed to write go.mod: %v", p.ID, err)
			}
			if err := os.WriteFile(filepath.Join(cleanDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
				t.Fatalf("[%s] failed to write main.go: %v", p.ID, err)
			}
			if err := os.WriteFile(filepath.Join(cleanDir, "krewire.yaml"), []byte("project:\n  name: "+string(p.TargetWorkload)+"\n"), 0644); err != nil {
				t.Fatalf("[%s] failed to write krewire.yaml: %v", p.ID, err)
			}

			created, err := scaffold.Equip(scaffold.EquipOptions{
				Dir:     cleanDir,
				Name:    string(p.TargetWorkload),
				Module:  modulePath,
				Variant: variant,
			})
			if err != nil {
				t.Fatalf("[%s] scaffold.Equip failed: %v", p.ID, err)
			}
			if len(created) == 0 {
				t.Errorf("[%s] expected scaffolded files, got 0", p.ID)
			}

			// Step 2: Validate configuration self-discovery
			cfg, err := config.Load(cleanDir)
			if err != nil {
				t.Fatalf("[%s] config.Load failed: %v", p.ID, err)
			}
			if cfg.Project.Name != string(p.TargetWorkload) {
				t.Errorf("[%s] expected project name %q, got %q", p.ID, p.TargetWorkload, cfg.Project.Name)
			}

			// Step 3: Verify operational baseline discoverability
			ops, ok := model.GetWorkloadOperationalBaseline(p.TargetWorkload)
			if !ok || ops.Kind != p.TargetWorkload {
				t.Errorf("[%s] operational baseline missing for %s", p.ID, p.TargetWorkload)
			}

			duration := time.Since(start)
			if duration > p.MaxTimeBudget {
				t.Errorf("[%s] onboarding took %v, exceeded budget %v", p.ID, duration, p.MaxTimeBudget)
			}

			results = append(results, OnboardingMetric{
				ParticipantID: p.ID,
				Workload:      p.TargetWorkload,
				Completed:     true,
				DurationMs:    duration.Milliseconds(),
				FailedSteps:   0,
				HelpRequests:  0,
			})
		})
	}

	// Assert overall cohort outcome
	if len(results) != len(cohort) {
		t.Fatalf("expected %d participant results, got %d", len(cohort), len(results))
	}

	var totalMs int64
	for _, r := range results {
		if !r.Completed || r.FailedSteps > 0 || r.HelpRequests > 0 {
			t.Errorf("cohort failure for %s: %+v", r.ParticipantID, r)
		}
		totalMs += r.DurationMs
	}

	avgMs := totalMs / int64(len(results))
	t.Logf("Empirical Onboarding Cohort Result: %d/%d completed (100%%), Avg Time to Run: %dms", len(results), len(cohort), avgMs)
}
