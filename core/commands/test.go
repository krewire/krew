package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/gomod"
	"github.com/krewire/krewire/errs"
)

// RegisterTest configures flags for the `krew test` command.
func RegisterTest(fs *flag.FlagSet) {
	fs.String("filter", "", "filter tests by name or file pattern (like vitest --filter, pest --filter)")
	fs.Bool("watch", false, "watch files and rerun tests on change (like vitest --watch)")
	fs.Bool("w", false, "alias for --watch")
	fs.Bool("coverage", false, "run with coverage (go test -cover)")
	fs.Bool("cover", false, "alias for --coverage")
	fs.String("coverprofile", "", "write coverage profile to file (go test -coverprofile)")
	fs.String("html", "", "generate HTML coverage report to file (e.g. coverage.html)")
	fs.String("cover-html", "", "alias for --html")
	fs.Float64("threshold", 0, "fail if total coverage percentage is below threshold (e.g. 80.0)")
	fs.Float64("min", 0, "alias for --threshold")
	fs.Bool("summary", false, "display detailed per-package coverage summary table")
	fs.Bool("breakdown", false, "alias for --summary")
	fs.Bool("verbose", false, "verbose output (go test -v)")
	fs.Bool("v", false, "alias for --verbose")
	fs.Bool("json", false, "json output (go test -json)")
	fs.Bool("update", false, "update snapshots/golden files (UPDATE_GOLDEN=1)")
	fs.Bool("u", false, "alias for --update")
	fs.String("run", "", "run only tests matching regexp (go test -run)")
	fs.String("count", "", "go test -count value (e.g., 1 to disable cache)")
	fs.Bool("list", false, "list tests instead of running (go test -list)")
	fs.String("reporter", "", "reporter: verbose, dot, json (like vitest --reporter)")
}

// RunTest executes tests for the current Go module project.
func RunTest(fs *flag.FlagSet) errs.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, err := gomod.Find(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: not inside a Go module — run 'krew new <project>' first")
		return errs.ExitCodeUsage
	}
	slog.Info("running tests", "module", mod.Path)

	filter := flagValue(fs, "filter")
	if v := flagValue(fs, "run"); v != "" && filter == "" {
		filter = v
	}

	watch := flagValue(fs, "watch") == "true" || flagValue(fs, "w") == "true"
	covOpts, err := resolveCoverageOptions(fs)
	if err != nil {
		return fail(err)
	}
	if covOpts.TempProfile {
		defer func() {
			_ = os.Remove(covOpts.ProfilePath)
		}()
	}

	verbose := flagValue(fs, "verbose") == "true" || flagValue(fs, "v") == "true"
	jsonOut := flagValue(fs, "json") == "true"
	update := flagValue(fs, "update") == "true" || flagValue(fs, "u") == "true"
	count := flagValue(fs, "count")
	list := flagValue(fs, "list") == "true"
	reporter := flagValue(fs, "reporter")
	if reporter == "json" {
		jsonOut = true
	}
	if reporter == "verbose" {
		verbose = true
	}

	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if jsonOut {
		args = append(args, "-json")
	}
	if covOpts.Enabled {
		args = append(args, "-coverprofile="+covOpts.ProfilePath)
	}
	if count != "" {
		args = append(args, "-count="+count)
	}
	if list {
		pattern := filter
		if pattern == "" {
			pattern = ".*"
		}
		if isPackageFilter(pattern) {
			pattern = ".*"
		}
		args = append(args, "-list", pattern)
		filter = ""
	} else if filter != "" {
		if !isPackageFilter(filter) {
			args = append(args, "-run", filter)
		}
	}

	// collect positional package args (after --)
	var pkgs []string
	if fs != nil {
		for _, a := range fs.Args() {
			if strings.HasPrefix(a, "-") {
				continue
			}
			if a == filter {
				continue
			}
			pkgs = append(pkgs, a)
		}
	}

	if filter != "" && isPackageFilter(filter) && len(pkgs) == 0 {
		pkgs = []string{filter}
		filter = ""
	}

	if len(pkgs) == 0 {
		if fs != nil && containsGoPackageArg(fs.Args()) {
			// already has package, don't add ./...
		} else {
			pkgs = []string{"./..."}
		}
	}
	args = append(args, pkgs...)

	if fs != nil {
		for _, a := range fs.Args() {
			if strings.HasPrefix(a, "-") && !isKnownTestFlag(a) {
				args = append(args, a)
			}
		}
	}

	env := os.Environ()
	if update {
		env = append(env, "UPDATE_GOLDEN=1")
	}

	if watch {
		return runTestWatch(dir, args, env, covOpts)
	}

	code := runTestOnce(dir, args, env)
	if code != errs.ExitCodeSuccess {
		return code
	}
	if covOpts.Enabled {
		return processCoverageResults(covOpts, os.Stdout, os.Stderr)
	}
	return errs.ExitCodeSuccess
}

func runTestOnce(dir string, args []string, env []string) errs.ExitCode {
	slog.Info("go test", "args", strings.Join(args, " "))
	cmd := exec.Command("go", args...) // #nosec G204 — user-invoked devtool
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}

func runTestWatch(dir string, args []string, env []string, covOpts CoverageOptions) errs.ExitCode {
	root := dir
	if mod, err := gomod.Find(dir); err == nil {
		if mdir := moduleDir(dir); mdir != "" {
			root = mdir
		} else {
			_ = mod
		}
	}
	cfg, _ := config.Load(filepath.Join(root, "krewire.yaml"))
	w := newWatcher(root, 500*time.Millisecond, cfg)
	defer func() { close(w.done) }()

	code := runTestOnce(dir, args, env)
	if code != errs.ExitCodeSuccess {
		slog.Info("watch: initial run failed, waiting for changes...")
	} else if covOpts.Enabled {
		_ = processCoverageResults(covOpts, os.Stdout, os.Stderr)
	}
	fmt.Fprintln(os.Stdout, "watching for changes (Ctrl+C to stop)...")

	for {
		select {
		case <-w.Changed():
			fmt.Fprintln(os.Stdout, "\n--- change detected, rerunning ---")
			w.Reset()
			code := runTestOnce(dir, args, env)
			if code == errs.ExitCodeSuccess && covOpts.Enabled {
				_ = processCoverageResults(covOpts, os.Stdout, os.Stderr)
			}
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func isPackageFilter(s string) bool {
	if strings.Contains(s, "/") || strings.HasPrefix(s, "./") || s == "." || s == "..." {
		return true
	}
	if strings.HasSuffix(s, ".go") {
		return true
	}
	return false
}

func isKnownTestFlag(a string) bool {
	known := []string{
		"-filter", "--filter",
		"-watch", "--watch", "-w",
		"-coverage", "--coverage",
		"-cover", "--cover",
		"-coverprofile", "--coverprofile",
		"-html", "--html",
		"-cover-html", "--cover-html",
		"-threshold", "--threshold",
		"-min", "--min",
		"-summary", "--summary",
		"-breakdown", "--breakdown",
		"-verbose", "--verbose", "-v",
		"-json", "--json",
		"-update", "--update", "-u",
		"-run", "--run",
		"-count", "--count",
		"-list", "--list",
		"-reporter", "--reporter",
	}
	for _, k := range known {
		if a == k || strings.HasPrefix(a, k+"=") {
			return true
		}
	}
	return false
}
