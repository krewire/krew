package bench

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/markdown"
	"github.com/krewire/krewire/pkg/model"
	"github.com/krewire/krewire/pkg/web/ssg"
	"github.com/krewire/ui/kiw"
)

// RunnerConfig controls the benchmark runner execution.
type RunnerConfig struct {
	WorkspaceRoot string
	KiwBinary     string
	GitCommit     string
	Iterations    int
	IncludeCLI    bool
	IncludeCore   bool
	IncludeWfk    bool
}

// DefaultRunnerConfig returns standard runner settings.
func DefaultRunnerConfig(workspaceRoot, kiwBin, gitCommit string) RunnerConfig {
	return RunnerConfig{
		WorkspaceRoot: workspaceRoot,
		KiwBinary:     kiwBin,
		GitCommit:     gitCommit,
		Iterations:    100,
		IncludeCLI:    true,
		IncludeCore:   true,
		IncludeWfk:    true,
	}
}

// RunSuite executes benchmarks across enabled categories and produces a report.
func RunSuite(cfg RunnerConfig) (*BenchmarkReport, error) {
	if cfg.Iterations <= 0 {
		cfg.Iterations = 50
	}
	report := &BenchmarkReport{
		Environment: CurrentEnvironment(cfg.GitCommit),
		Results:     make([]BenchmarkResult, 0, len(CanonicalInventory)),
	}

	for _, item := range CanonicalInventory {
		switch item.Category {
		case CategoryCorePrimitive:
			if !cfg.IncludeCore {
				continue
			}
		case CategoryCLILatency:
			if !cfg.IncludeCLI {
				continue
			}
		case CategoryWorkflowE2E:
			if !cfg.IncludeWfk {
				continue
			}
		}

		res, err := runSingleBenchmark(item, cfg)
		if err != nil {
			return nil, fmt.Errorf("benchmark %s (%s) failed: %w", item.ID, item.Name, err)
		}
		report.Results = append(report.Results, res)
	}

	report.Summary = FormatReportSummary(report)
	return report, nil
}

func runSingleBenchmark(item BenchmarkItem, cfg RunnerConfig) (BenchmarkResult, error) {
	res := BenchmarkResult{
		ID:       item.ID,
		Name:     item.Name,
		Category: item.Category,
		TargetNs: item.TargetNs,
	}

	switch item.ID {
	// ── 1. Core Primitives ──────────────────────────────────────────────────
	case "CORE-CFG-001":
		// Benchmark config loading
		dir, _ := os.MkdirTemp("", "bench-cfg-*")
		defer os.RemoveAll(dir)
		_ = os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte("project:\n  name: benchmark-site\n  kind: site\n  module: example.com/site\n"), 0o644)

		iters := cfg.Iterations * 10
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			_, _ = config.Load(dir)
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)

	case "CORE-MOD-001":
		// Benchmark model validation
		p := model.Project{
			Name:       "benchmark-app",
			Kind:       model.KindSite,
			ModulePath: "github.com/krewire/krewire",
		}
		iters := cfg.Iterations * 100
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			_ = p.Validate()
			_, _ = model.ParseKind("site")
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)

	case "CORE-KIW-001":
		// Benchmark krew parsing
		fixture := `---
title: Landing
layout: Base
props:
  headline: { type: string, default: "Hello" }
  count: { type: number, default: 42 }
---
<script lang="ts">export let headline: string = "Hello";</script>
<style scoped>.h1 { font-weight: bold; }</style>
<div class="h1">
  <Navbar Brand="Krewire" />
  <h1>{headline}</h1>
  {#if count > 0}<p>Count: {count}</p>{/if}
</div>`
		iters := cfg.Iterations * 20
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			_, _ = kiw.ParseKiw(fixture)
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)

	case "CORE-DSG-001":
		// Benchmark template desugaring
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
		iters := cfg.Iterations * 20
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			_, _ = kiw.DesugarTemplate(fixture)
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)

	case "CORE-MDN-001":
		// Benchmark markdown rendering
		mdDoc := []byte(`# Architecture Overview
Krewire is an end-to-end digital SDLC platform built on three immutable pillars:
- **Secure**: Safety by default.
- **Sustainable**: Near-zero cost engineered.
- **Scalable**: Additive progression without rewrite.

## Code Example
` + "```go\nfunc main() {\n\tprintln(\"Hello\")\n}\n```\n" + `
> [!NOTE]
> Architecture rules are enforced by tests.`)

		iters := cfg.Iterations * 10
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			_, _ = markdown.Render(mdDoc)
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)

	// ── 2. CLI Latency ──────────────────────────────────────────────────────
	case "CLI-LAT-001":
		// Measure krew version
		iters := 10
		start := time.Now()
		for i := 0; i < iters; i++ {
			cmd := exec.Command(cfg.KiwBinary, "version")
			cmd.Dir = cfg.WorkspaceRoot
			if err := cmd.Run(); err != nil {
				return res, fmt.Errorf("krew version failed: %w", err)
			}
		}
		elapsed := time.Since(start)
		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)

	case "CLI-LAT-002":
		// Measure krew help
		iters := 10
		start := time.Now()
		for i := 0; i < iters; i++ {
			cmd := exec.Command(cfg.KiwBinary, "help")
			cmd.Dir = cfg.WorkspaceRoot
			if err := cmd.Run(); err != nil {
				return res, fmt.Errorf("krew help failed: %w", err)
			}
		}
		elapsed := time.Since(start)
		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)

	case "CLI-LAT-003":
		// Measure krew tier --check free
		iters := 5
		start := time.Now()
		for i := 0; i < iters; i++ {
			cmd := exec.Command(cfg.KiwBinary, "tier", "--check", "free")
			cmd.Dir = filepath.Join(cfg.WorkspaceRoot, "krewire")
			if err := cmd.Run(); err != nil {
				return res, fmt.Errorf("krew tier failed: %w", err)
			}
		}
		elapsed := time.Since(start)
		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)

	// ── 3. Workflow End-to-End ──────────────────────────────────────────────
	case "WFK-SSG-001":
		// Build a temporary 5-page SSG site
		tmpSite, err := os.MkdirTemp("", "bench-site-*")
		if err != nil {
			return res, err
		}
		defer os.RemoveAll(tmpSite)

		_ = os.WriteFile(filepath.Join(tmpSite, "krewire.yaml"), []byte("project:\n  kind: site\n"), 0o644)
		_ = os.MkdirAll(filepath.Join(tmpSite, "layouts"), 0o755)
		_ = os.WriteFile(filepath.Join(tmpSite, "layouts", "Base.kiw"), []byte("<!DOCTYPE html><html><body>{{.Content}}</body></html>"), 0o644)
		_ = os.MkdirAll(filepath.Join(tmpSite, "pages"), 0o755)

		for i := 1; i <= 5; i++ {
			p := fmt.Sprintf("---\ntitle: Page %d\nlayout: Base\n---\n<h1>Page %d</h1>\n<p>Content for page %d</p>", i, i, i)
			filename := "index.kiw"
			if i > 1 {
				filename = fmt.Sprintf("page%d.kiw", i)
			}
			_ = os.WriteFile(filepath.Join(tmpSite, "pages", filename), []byte(p), 0o644)
		}

		site, err := ssg.LoadFromDir(tmpSite)
		if err != nil {
			return res, err
		}

		iters := 5
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		start := time.Now()
		for i := 0; i < iters; i++ {
			outDir, _ := os.MkdirTemp("", "bench-out-*")
			_, _ = site.Build(outDir)
			_ = os.RemoveAll(outDir)
		}
		elapsed := time.Since(start)
		runtime.ReadMemStats(&m2)

		res.Iterations = iters
		res.NsPerOp = elapsed.Nanoseconds() / int64(iters)
		res.BytesPerOp = int64(m2.TotalAlloc-m1.TotalAlloc) / int64(iters)
		res.AllocsPerOp = int64(m2.Mallocs-m1.Mallocs) / int64(iters)
	}

	res.Passed = (res.TargetNs == 0 || res.NsPerOp <= res.TargetNs)
	return res, nil
}

// FormatMarkdownTable renders the report as a clean GitHub-flavored Markdown document.
func FormatMarkdownTable(report *BenchmarkReport) string {
	var b bytes.Buffer
	b.WriteString("# Krewire Performance Benchmark Baseline Report\n\n")
	b.WriteString(fmt.Sprintf("**Environment:** %s/%s · %s · %d CPUs\n",
		report.Environment.OS, report.Environment.Arch, report.Environment.GoVersion, report.Environment.NumCPU))
	b.WriteString(fmt.Sprintf("**Git Commit:** `%s` · **Generated:** %s\n\n",
		report.Environment.GitCommit, report.Environment.Timestamp.Format("2006-01-02 15:04:05 UTC")))

	b.WriteString("## Benchmark Measurements\n\n")
	b.WriteString("| ID | Operation | Category | Ns/Op | Latency | Bytes/Op | Allocs/Op | Target | Status |\n")
	b.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")

	for _, r := range report.Results {
		latStr := formatDuration(time.Duration(r.NsPerOp))
		targetStr := "N/A"
		if r.TargetNs > 0 {
			targetStr = formatDuration(time.Duration(r.TargetNs))
		}
		status := "✅ PASS"
		if !r.Passed {
			status = "⚠️ OVER"
		}
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %d | %s | %d B | %d | %s | %s |\n",
			r.ID, r.Name, r.Category, r.NsPerOp, latStr, r.BytesPerOp, r.AllocsPerOp, targetStr, status))
	}

	b.WriteString("\n## Summary\n\n")
	b.WriteString(report.Summary + "\n")
	return b.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1_000_000.0)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}
