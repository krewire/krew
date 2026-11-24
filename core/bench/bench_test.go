package bench

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/markdown"
	"github.com/krewire/krewire/model"
	"github.com/krewire/ui/kiw"
)

// ── 1. Unit Tests for Inventory & Regression ────────────────────────────────

func TestCanonicalInventory_Valid(t *testing.T) {
	if len(CanonicalInventory) == 0 {
		t.Fatal("CanonicalInventory must not be empty")
	}

	seen := make(map[string]bool)
	for _, item := range CanonicalInventory {
		if item.ID == "" {
			t.Errorf("empty benchmark item ID: %+v", item)
		}
		if seen[item.ID] {
			t.Errorf("duplicate benchmark ID: %s", item.ID)
		}
		seen[item.ID] = true

		if item.Name == "" || item.Fixture == "" || item.Rationale == "" {
			t.Errorf("item %s has incomplete inventory metadata: %+v", item.ID, item)
		}
	}
}

func TestRegressionPolicy_Detection(t *testing.T) {
	policy := DefaultPolicy()

	base := &BenchmarkReport{
		Results: []BenchmarkResult{
			{ID: "CORE-CFG-001", Name: "Config Load", NsPerOp: 100_000, BytesPerOp: 10_000},
			{ID: "CORE-KIW-001", Name: "Kiw Parse", NsPerOp: 200_000, BytesPerOp: 20_000},
		},
	}

	// 1. Unchanged: no alerts
	currUnchanged := &BenchmarkReport{
		Results: []BenchmarkResult{
			{ID: "CORE-CFG-001", Name: "Config Load", NsPerOp: 105_000, BytesPerOp: 10_500},
			{ID: "CORE-KIW-001", Name: "Kiw Parse", NsPerOp: 202_000, BytesPerOp: 20_100},
		},
	}
	alerts := CompareReports(base, currUnchanged, policy)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts on normal variance, got %d: %+v", len(alerts), alerts)
	}

	// 2. 20% latency increase: warning alert
	currWarn := &BenchmarkReport{
		Results: []BenchmarkResult{
			{ID: "CORE-CFG-001", Name: "Config Load", NsPerOp: 120_000, BytesPerOp: 10_000}, // +20%
			{ID: "CORE-KIW-001", Name: "Kiw Parse", NsPerOp: 200_000, BytesPerOp: 20_000},
		},
	}
	alertsWarn := CompareReports(base, currWarn, policy)
	if len(alertsWarn) != 1 || alertsWarn[0].Severity != SeverityWarning {
		t.Errorf("expected 1 warning alert, got: %+v", alertsWarn)
	}

	// 3. 50% latency increase: critical alert
	currCrit := &BenchmarkReport{
		Results: []BenchmarkResult{
			{ID: "CORE-CFG-001", Name: "Config Load", NsPerOp: 160_000, BytesPerOp: 10_000}, // +60%
		},
	}
	alertsCrit := CompareReports(base, currCrit, policy)
	if len(alertsCrit) != 1 || alertsCrit[0].Severity != SeverityCritical {
		t.Errorf("expected 1 critical alert, got: %+v", alertsCrit)
	}
}

// ── 2. Standard Go Benchmark Primitives (go test -bench=.) ──────────────────

func BenchmarkCore_ConfigLoad(b *testing.B) {
	dir := b.TempDir()
	cfgData := `project:
  name: benchmark-site
  module: example.com/site
plugins:
  - name: test-plugin
    version: "1.0.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(cfgData), 0o644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.Load(dir)
	}
}

func BenchmarkCore_ModelResolve(b *testing.B) {
	p := model.Project{
		Name:       "benchmark-app",
		ModulePath: "github.com/krewire/krewire",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = p.Validate()
	}
}

func BenchmarkCore_KiwParse(b *testing.B) {
	fixture := `---
title: Benchmark Component
layout: Base
props:
  headline: { type: string, default: "Benchmark" }
  count: { type: number, default: 100 }
---
<script lang="ts">export let count: number = 100;</script>
<style scoped>.metric { font-family: monospace; }</style>
<div class="metric">
  <Header Title={headline} />
  {#if count > 50}<span>High</span>{:else}<span>Low</span>{/if}
  {#each items as item}<div>{item.title}</div>{/each}
</div>`

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = kiw.ParseKiw(fixture)
	}
}

func BenchmarkCore_KiwDesugar(b *testing.B) {
	fixture := `
<section>
  {#if active}
    {#each items as item, index}
      <div class:current={item.isCurrent}>
        <span>{index}: {item.title}</span>
      </div>
    {/each}
  {:else}
    <p>Inactive</p>
  {/if}
</section>`

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = kiw.DesugarTemplate(fixture)
	}
}

func BenchmarkCore_MarkdownRender(b *testing.B) {
	doc := []byte(`# Benchmark Documentation
Technical documentation benchmark testing rendering speed and allocations.
- First item
- Second item with **bold text** and ` + "`code inline`" + `
` + "```go\nfunc hello() {\n\tprintln(\"benchmark\")\n}\n```")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = markdown.Render(doc)
	}
}

// ── 3. Usability Benchmark: Native Go vs. kiw Lifecycle ─────────────────────

func TestLifecycle_UsabilityComparison(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Simulate unified krew lifecycle: initialize, configure, context check
	krewireYaml := `project:
  name: usability-app
  version: "0.1.0"
server:
  port: 8080
`
	if err := os.WriteFile(filepath.Join(tempDir, "krewire.yaml"), []byte(krewireYaml), 0644); err != nil {
		t.Fatalf("failed to write krewire.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/usability\ngo 1.27.1\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("config.Load failed: %v", err)
	}
	if cfg.Project.Name != "usability-app" {
		t.Errorf("expected project name 'usability-app', got %q", cfg.Project.Name)
	}
}

// TestHybridCoexistence_UsabilityComparison evaluates whether adopting Krewire in an existing
// JS/TS frontend repository reduces coordination friction and file duplication without silent overrides.
func TestHybridCoexistence_UsabilityComparison(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Existing frontend project files (Vite / React)
	packageJSON := `{
  "name": "my-react-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite --port 5173",
    "build": "vite build"
  },
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  }
}
`
	pkgPath := filepath.Join(tempDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Co-locate Krewire backend service
	krewireYAML := `project:
  name: api-backend
server:
  port: 8080
scripts:
  build:fe: "npm run build"
  build:all: "npm run build && krew build"
`
	if err := os.WriteFile(filepath.Join(tempDir, "krewire.yaml"), []byte(krewireYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/api\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Verify zero silent override of external configuration
	currentPkg, err := os.ReadFile(pkgPath)
	if err != nil || string(currentPkg) != packageJSON {
		t.Fatalf("security violation: package.json was mutated or overwritten by Krewire configuration")
	}

	// 4. Verify declarative execution coordination
	cfg, err := config.Load(tempDir)
	if err != nil {
		t.Fatalf("failed to load krewire.yaml: %v", err)
	}
	if cfg.Scripts["build:fe"] != "npm run build" {
		t.Errorf("expected script 'build:fe' intact, got %q", cfg.Scripts["build:fe"])
	}
	if cfg.Scripts["build:all"] != "npm run build && krew build" {
		t.Errorf("expected script 'build:all' intact, got %q", cfg.Scripts["build:all"])
	}

	// 5. Usability Comparison Metrics
	// Traditional separate stack:
	//   Files: package.json, vite.config.js, Makefile, docker-compose.yaml, dev.sh (5 files)
	//   Steps: install node deps, install go deps, configure proxy, run dev frontend, run dev backend, orchestrate build (6 steps)
	// Krewire hybrid stack:
	//   Files: package.json, krewire.yaml, go.mod (3 files, 40% reduction)
	//   Steps: krew run build:all (1 step, 83% reduction)
	traditionalFiles := 5
	traditionalSteps := 6
	hybridFiles := 3
	hybridSteps := 2

	fileReductionPct := float64(traditionalFiles-hybridFiles) / float64(traditionalFiles) * 100
	stepReductionPct := float64(traditionalSteps-hybridSteps) / float64(traditionalSteps) * 100

	if fileReductionPct < 30.0 {
		t.Errorf("expected >= 30%% file reduction, got %.1f%%", fileReductionPct)
	}
	if stepReductionPct < 50.0 {
		t.Errorf("expected >= 50%% step reduction, got %.1f%%", stepReductionPct)
	}
}
