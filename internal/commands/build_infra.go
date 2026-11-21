package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/cloud/infra/provider"
	"github.com/krewire/krewire/pkg/cloud/infra/schema"
	"github.com/krewire/krewire/pkg/errs"
)

const (
	// DefaultInfraProvider is the fallback cloud provider backend.
	DefaultInfraProvider = "krewire-cloud"

	// DefaultInfraRegion is the default deployment target region.
	DefaultInfraRegion = "asia-southeast1"

	// DefaultInfraRuntime is the default container orchestration runtime.
	DefaultInfraRuntime = "fargate"

	// DefaultInfraReplicas is the default compute replica factor.
	DefaultInfraReplicas = 2

	// InfraPlanFileName is the file name for generated infrastructure blueprints.
	InfraPlanFileName = "plan.json"
)

// buildInfra compiles and verifies cloud infrastructure topology into plan.json (KWN-BUILD-001, KWF-B7N3D).
func buildInfra(root string, cfg *config.Config, fs *flag.FlagSet) errs.ExitCode {
	providerName := DefaultInfraProvider
	region := DefaultInfraRegion
	if cfg.Infra != nil {
		if cfg.Infra.Provider != "" {
			providerName = cfg.Infra.Provider
		}
		if cfg.Infra.Region != "" {
			region = cfg.Infra.Region
		}
	}
	env := firstNonEmpty(flagValue(fs, "env"), cfg.Env, "production")

	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	infraDir := filepath.Join(outDir, "infra")
	if err := os.MkdirAll(infraDir, 0o755); err != nil {
		return fail(err)
	}
	planFile := filepath.Join(infraDir, InfraPlanFileName)

	// 1. Gather desired resources
	var resources []provider.Resource
	if cfg.Infra != nil && len(cfg.Infra.Resources) > 0 {
		resources = append(resources, cfg.Infra.Resources...)
	} else {
		// Synthesize declarative infrastructure resources from project definition
		appName := firstNonEmpty(cfg.Name, moduleName(root), DefaultBinaryName)
		resources = append(resources, provider.Resource{
			Kind: schema.KindCompute,
			ID:   appName,
			Properties: map[string]any{
				"runtime":  DefaultInfraRuntime,
				"replicas": DefaultInfraReplicas,
			},
		})
		if len(cfg.Services) > 0 {
			for _, svc := range cfg.Services {
				if svc.Name == "" {
					continue
				}
				resources = append(resources, provider.Resource{
					Kind: schema.KindCompute,
					ID:   svc.Name,
					Properties: map[string]any{
						"port": svc.Port,
					},
					DependsOn: []string{appName},
				})
			}
		}
		if cfg.Worker != nil {
			resources = append(resources, provider.Resource{
				Kind: schema.KindCompute,
				ID:   "worker",
				Properties: map[string]any{
					"concurrency": cfg.Worker.Concurrency,
					"queues":      cfg.Worker.Queues,
				},
				DependsOn: []string{appName},
			})
		}
	}

	// 2. Validate topology (no duplicate IDs, all dependsOn exist)
	idMap := make(map[string]bool)
	for _, r := range resources {
		if r.ID == "" {
			return fail(fmt.Errorf("infra topology error: resource missing id (kind: %s)", r.Kind))
		}
		if idMap[r.ID] {
			return fail(fmt.Errorf("infra topology error: duplicate resource id %q", r.ID))
		}
		idMap[r.ID] = true
	}
	for _, r := range resources {
		for _, dep := range r.DependsOn {
			if !idMap[dep] {
				return fail(fmt.Errorf("infra topology error: resource %q depends on non-existent resource %q", r.ID, dep))
			}
		}
	}

	// 3. Assemble execution plan
	plan := provider.Plan{
		Actions: make([]provider.Action, len(resources)),
	}
	for i, r := range resources {
		plan.Actions[i] = provider.Action{
			Op:       provider.OpCreate,
			Resource: r,
			Reason:   "declared in infrastructure blueprint",
		}
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fail(fmt.Errorf("marshal infra plan: %w", err))
	}
	if err := os.WriteFile(planFile, data, 0o644); err != nil {
		return fail(fmt.Errorf("write infra plan: %w", err))
	}
	fmt.Println("created " + planFile)

	if fs != nil && (flagValue(fs, "plan") == "true" || flagProvided(fs, "plan")) {
		fmt.Printf("Infrastructure Plan (%s, env=%s, region=%s):\n", providerName, env, region)
		for _, act := range plan.Actions {
			fmt.Printf("  + [%s] %s (%s)\n", act.Op, act.Resource.Kind, act.Resource.ID)
			if len(act.Resource.DependsOn) > 0 {
				fmt.Printf("      depends on: %s\n", strings.Join(act.Resource.DependsOn, ", "))
			}
		}
		fmt.Printf("Plan: %d to add, 0 to change, 0 to destroy.\n", len(plan.Actions))
	}

	return errs.ExitCodeSuccess
}
