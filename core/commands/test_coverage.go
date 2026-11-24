package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/krewire/krewire/errs"
	pkgtesting "github.com/krewire/krewire/testing"
)

// CoverageOptions defines the configuration for test coverage execution and reporting.
type CoverageOptions struct {
	Enabled      bool
	ProfilePath  string
	HTMLPath     string
	Threshold    float64
	HasThreshold bool
	Summary      bool
	TempProfile  bool
}

// resolveCoverageOptions inspects the command-line flags and builds the coverage options.
func resolveCoverageOptions(fs *flag.FlagSet) (CoverageOptions, error) {
	if fs == nil {
		return CoverageOptions{}, nil
	}

	cover := flagValue(fs, "cover") == "true" || flagValue(fs, "coverage") == "true"
	profile := flagValue(fs, "coverprofile")
	html := flagValue(fs, "html")
	if html == "" {
		html = flagValue(fs, "cover-html")
	}
	threshold := flagFloat(fs, "threshold")
	if threshold == 0 {
		threshold = flagFloat(fs, "min")
	}
	hasThreshold := flagProvided(fs, "threshold") || flagProvided(fs, "min")
	summary := flagValue(fs, "summary") == "true" || flagValue(fs, "breakdown") == "true"

	enabled := cover || profile != "" || html != "" || hasThreshold || summary
	if !enabled {
		return CoverageOptions{}, nil
	}

	opts := CoverageOptions{
		Enabled:      true,
		ProfilePath:  profile,
		HTMLPath:     html,
		Threshold:    threshold,
		HasThreshold: hasThreshold,
		Summary:      summary,
	}

	if opts.ProfilePath == "" {
		tmp, err := os.CreateTemp("", "kiw-coverage-*.out")
		if err != nil {
			return opts, fmt.Errorf("create temp coverage profile: %w", err)
		}
		opts.ProfilePath = tmp.Name()
		opts.TempProfile = true
		_ = tmp.Close()
	}

	return opts, nil
}

// processCoverageResults processes coverage data after tests pass:
// - Parses the coverage profile
// - Displays summary line or per-package table
// - Generates HTML report when requested
// - Enforces threshold requirements
func processCoverageResults(opts CoverageOptions, stdout, stderr io.Writer) errs.ExitCode {
	if !opts.Enabled || opts.ProfilePath == "" {
		return errs.ExitCodeSuccess
	}

	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	report, err := pkgtesting.ParseCoverageProfileFile(opts.ProfilePath)
	if err != nil {
		fmt.Fprintf(stderr, "krew: error reading coverage profile: %v\n", err)
		return errs.ExitCodeFailure
	}

	if opts.Summary {
		table := pkgtesting.FormatCoverageTable(report)
		fmt.Fprintln(stdout, table)
	} else {
		summary := pkgtesting.FormatCoverageSummary(report)
		fmt.Fprintln(stdout, summary)
	}

	if opts.HTMLPath != "" {
		if dir := filepath.Dir(opts.HTMLPath); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(stderr, "krew: error creating html output directory: %v\n", err)
				return errs.ExitCodeFailure
			}
		}
		cmd := exec.Command("go", "tool", "cover", "-html="+opts.ProfilePath, "-o", opts.HTMLPath)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(stderr, "krew: error generating html coverage report: %v\n", err)
			return errs.ExitCodeFailure
		}
		fmt.Fprintf(stdout, "Coverage HTML report generated at %s\n", opts.HTMLPath)
	}

	if opts.HasThreshold {
		if !report.MeetsThreshold(opts.Threshold) {
			fmt.Fprintf(stderr, "FAIL: test coverage %.1f%% does not meet required threshold of %.1f%%\n", report.Percentage(), opts.Threshold)
			return errs.ExitCodeFailure
		}
		fmt.Fprintf(stdout, "PASS: test coverage %.1f%% meets required threshold of %.1f%%\n", report.Percentage(), opts.Threshold)
	}

	return errs.ExitCodeSuccess
}
