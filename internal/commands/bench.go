package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/krewire/krew/internal/bench"
	"github.com/krewire/krewire/pkg/errs"
)

var (
	benchFlagJSON       bool
	benchFlagSave       string
	benchFlagCompare    string
	benchFlagCore       bool
	benchFlagCLI        bool
	benchFlagWorkflow   bool
	benchFlagIterations int
)

// RegisterBench registers CLI flags for the "krew bench" command.
func RegisterBench(fs *flag.FlagSet) {
	fs.BoolVar(&benchFlagJSON, "json", false, "output benchmark results in JSON format")
	fs.StringVar(&benchFlagSave, "save", "", "save benchmark report to file path (e.g. --save baseline.json)")
	fs.StringVar(&benchFlagCompare, "compare", "", "compare results against baseline report and report regressions")
	fs.BoolVar(&benchFlagCore, "core", false, "run only Core primitive benchmarks")
	fs.BoolVar(&benchFlagCLI, "cli", false, "run only CLI latency benchmarks")
	fs.BoolVar(&benchFlagWorkflow, "workflow", false, "run only End-to-End workflow benchmarks")
	fs.IntVar(&benchFlagIterations, "iterations", 50, "number of iterations per benchmark (default: 50)")
}

// RunBench executes the performance benchmark suite.
func RunBench(fs *flag.FlagSet) errs.ExitCode {
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	// Locate workspace root
	wsRoot := root
	for {
		if _, err := os.Stat(filepath.Join(wsRoot, "go.work")); err == nil {
			break
		}
		parent := filepath.Dir(wsRoot)
		if parent == wsRoot {
			wsRoot = root
			break
		}
		wsRoot = parent
	}

	kiwBin, err := os.Executable()
	if err != nil {
		kiwBin = filepath.Join(wsRoot, "bin", "krew")
	}

	includeAll := !benchFlagCore && !benchFlagCLI && !benchFlagWorkflow
	cfg := bench.RunnerConfig{
		WorkspaceRoot: wsRoot,
		KiwBinary:     kiwBin,
		GitCommit:     "local",
		Iterations:    benchFlagIterations,
		IncludeCore:   includeAll || benchFlagCore,
		IncludeCLI:    includeAll || benchFlagCLI,
		IncludeWfk:    includeAll || benchFlagWorkflow,
	}

	if !benchFlagJSON {
		fmt.Println("================================================================================")
		fmt.Println("                     KREWIRE PERFORMANCE BENCHMARK SUITE                        ")
		fmt.Println("================================================================================")
		fmt.Printf("Running benchmarks (%d iterations per target)...\n\n", cfg.Iterations)
	}

	report, err := bench.RunSuite(cfg)
	if err != nil {
		return fail(fmt.Errorf("benchmark execution failed: %w", err))
	}

	if benchFlagJSON {
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println(bench.FormatMarkdownTable(report))
	}

	// Optional: save report
	if benchFlagSave != "" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fail(err)
		}
		if err := os.WriteFile(benchFlagSave, data, 0o644); err != nil {
			return fail(fmt.Errorf("failed to save benchmark report to %s: %w", benchFlagSave, err))
		}
		if !benchFlagJSON {
			fmt.Printf("✓ Saved baseline report to %s\n", benchFlagSave)
		}
	}

	// Optional: compare against baseline
	if benchFlagCompare != "" {
		baseData, err := os.ReadFile(benchFlagCompare)
		if err != nil {
			return fail(fmt.Errorf("failed to read baseline %s: %w", benchFlagCompare, err))
		}
		var baseReport bench.BenchmarkReport
		if err := json.Unmarshal(baseData, &baseReport); err != nil {
			return fail(fmt.Errorf("invalid baseline JSON in %s: %w", benchFlagCompare, err))
		}

		alerts := bench.CompareReports(&baseReport, report, bench.DefaultPolicy())
		if len(alerts) > 0 {
			if !benchFlagJSON {
				fmt.Println("--------------------------------------------------------------------------------")
				fmt.Println("REGRESSION ANALYSIS:")
				for _, a := range alerts {
					if a.Severity == bench.SeverityCritical {
						fmt.Printf("  ✗ [%s] %s\n", a.Severity, a.Message)
					} else {
						fmt.Printf("  ! [%s] %s\n", a.Severity, a.Message)
					}
				}
				fmt.Println("--------------------------------------------------------------------------------")
			}
			for _, a := range alerts {
				if a.Severity == bench.SeverityCritical {
					return errs.ExitCodeFailure
				}
			}
		} else if !benchFlagJSON {
			fmt.Println("✓ No performance regressions detected against baseline!")
		}
	}

	return errs.ExitCodeSuccess
}
