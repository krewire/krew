package bench

import (
	"fmt"
	"runtime"
	"time"
)

// BenchmarkCategory groups related benchmark operations.
type BenchmarkCategory string

const (
	CategoryCorePrimitive BenchmarkCategory = "core"
	CategoryCLILatency    BenchmarkCategory = "cli"
	CategoryWorkflowE2E   BenchmarkCategory = "workflow"
)

// BenchmarkItem defines a documented benchmark inventory target.
type BenchmarkItem struct {
	ID           string            `json:"id"`
	Category     BenchmarkCategory `json:"category"`
	Name         string            `json:"name"`
	Fixture      string            `json:"fixture"`
	TargetNs     int64             `json:"target_ns"`
	TargetBytes  int64             `json:"target_bytes"`
	TargetAllocs int64             `json:"target_allocs"`
	Rationale    string            `json:"rationale"`
}

// EnvironmentMetadata records the hardware and toolchain context for reproducible results.
type EnvironmentMetadata struct {
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	NumCPU    int       `json:"num_cpu"`
	GoVersion string    `json:"go_version"`
	GitCommit string    `json:"git_commit"`
	Timestamp time.Time `json:"timestamp"`
}

// BenchmarkResult stores the measured performance metrics of an operation.
type BenchmarkResult struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    BenchmarkCategory `json:"category"`
	Iterations  int               `json:"iterations"`
	NsPerOp     int64             `json:"ns_per_op"`
	BytesPerOp  int64             `json:"bytes_per_op"`
	AllocsPerOp int64             `json:"allocs_per_op"`
	TargetNs    int64             `json:"target_ns"`
	Passed      bool              `json:"passed"`
}

// BenchmarkReport encapsulates a full benchmark execution run.
type BenchmarkReport struct {
	Environment EnvironmentMetadata `json:"environment"`
	Results     []BenchmarkResult   `json:"results"`
	Summary     string              `json:"summary"`
}

// CanonicalInventory holds the documented set of high-traffic primitives and workflows.
var CanonicalInventory = []BenchmarkItem{
	// ── 1. Core Primitives ──────────────────────────────────────────────────
	{
		ID:           "CORE-CFG-001",
		Category:     CategoryCorePrimitive,
		Name:         "Project Configuration Loading",
		Fixture:      "krewire.yaml (standard site configuration)",
		TargetNs:     500_000, // 0.5 ms
		TargetBytes:  50_000,  // 50 KB
		TargetAllocs: 200,
		Rationale:    "Invoked on every single CLI command dispatch to determine root configuration and active plugins.",
	},
	{
		ID:           "CORE-MOD-001",
		Category:     CategoryCorePrimitive,
		Name:         "Model & Kind Dependency Resolution",
		Fixture:      "Multi-workload dependency graph",
		TargetNs:     1_000_000, // 1 ms
		TargetBytes:  100_000,   // 100 KB
		TargetAllocs: 500,
		Rationale:    "Validates project graph topology, downward dependency rules, and opt-in batteries.",
	},
	{
		ID:           "CORE-KIW-001",
		Category:     CategoryCorePrimitive,
		Name:         "Single-File Component Parsing (.kiw)",
		Fixture:      "Landing page component with frontmatter, scoped styles, script blocks, and components",
		TargetNs:     350_000, // 0.35 ms
		TargetBytes:  40_000,  // 40 KB
		TargetAllocs: 350,
		Rationale:    "High-frequency leaf concern: executed for every page, layout, and component during live dev and build.",
	},
	{
		ID:           "CORE-DSG-001",
		Category:     CategoryCorePrimitive,
		Name:         "Template Desugaring & Clean Interpolation",
		Fixture:      "Complex nested loops, conditionals, slots, and props",
		TargetNs:     200_000, // 0.20 ms
		TargetBytes:  30_000,  // 30 KB
		TargetAllocs: 120,
		Rationale:    "Transforms JSX-like authoring grammar into pure Go html/template trees without CGO.",
	},
	{
		ID:           "CORE-MDN-001",
		Category:     CategoryCorePrimitive,
		Name:         "Documentation Markdown Rendering",
		Fixture:      "10 KB multi-section technical document with code blocks and callouts",
		TargetNs:     800_000, // 0.8 ms
		TargetBytes:  150_000, // 150 KB
		TargetAllocs: 600,
		Rationale:    "Powers book and documentation site generation pipelines across thousands of pages.",
	},

	// ── 2. CLI Latency & Responsiveness ─────────────────────────────────────
	{
		ID:           "CLI-LAT-001",
		Category:     CategoryCLILatency,
		Name:         "krew version Cold-Start Latency",
		Fixture:      "krew version standalone process",
		TargetNs:     30_000_000, // 30 ms
		TargetBytes:  0,
		TargetAllocs: 0,
		Rationale:    "Immediate developer feedback benchmark for bare CLI binary execution overhead.",
	},
	{
		ID:           "CLI-LAT-002",
		Category:     CategoryCLILatency,
		Name:         "krew help Command Discovery Latency",
		Fixture:      "krew help top-level overview",
		TargetNs:     35_000_000, // 35 ms
		TargetBytes:  0,
		TargetAllocs: 0,
		Rationale:    "Measures command registration, flag parsing, and formatting responsiveness.",
	},
	{
		ID:           "CLI-LAT-003",
		Category:     CategoryCLILatency,
		Name:         "krew tier Pre-Commit Check Latency",
		Fixture:      "krew tier --check free against repository root",
		TargetNs:     60_000_000, // 60 ms
		TargetBytes:  0,
		TargetAllocs: 0,
		Rationale:    "Pre-commit gate speed: developers must never feel friction waiting for tier verification.",
	},

	// ── 3. End-to-End Workflows ─────────────────────────────────────────────
	{
		ID:           "WFK-SSG-001",
		Category:     CategoryWorkflowE2E,
		Name:         "SSG Multi-Page Site Compilation",
		Fixture:      "Complete 5-page site with layout, components, assets, and client islands",
		TargetNs:     150_000_000, // 150 ms
		TargetBytes:  10_000_000,  // 10 MB
		TargetAllocs: 50_000,
		Rationale:    "Production build throughput: defines how quickly deployments and preview environments compile.",
	},
}

// CurrentEnvironment captures the running machine's parameters.
func CurrentEnvironment(gitCommit string) EnvironmentMetadata {
	return EnvironmentMetadata{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		GoVersion: runtime.Version(),
		GitCommit: gitCommit,
		Timestamp: time.Now().UTC(),
	}
}

// FormatReportSummary generates an executive summary of the benchmark results.
func FormatReportSummary(report *BenchmarkReport) string {
	var passed, total int
	for _, r := range report.Results {
		total++
		if r.Passed {
			passed++
		}
	}
	return fmt.Sprintf("Krewire Benchmark Run: %d/%d within budget (%s %s, %s, %s)",
		passed, total, report.Environment.OS, report.Environment.Arch,
		report.Environment.GoVersion, report.Environment.Timestamp.Format("2006-01-02 15:04:05 UTC"))
}
