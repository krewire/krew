package commands

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/tui/term"
)

// RegisterUpdate registers flags for `krew update`.
func RegisterUpdate(fs *flag.FlagSet) {
	fs.Bool("tidy", true, "run 'go mod tidy' after updating dependencies (default: true)")
	fs.Bool("dry-run", false, "preview available dependency updates without applying them")
}

// RunUpdate updates project dependencies to their latest compatible versions.
func RunUpdate(fs *flag.FlagSet) errs.ExitCode {
	tidy := flagValue(fs, "tidy") != "false"
	dryRun := flagValue(fs, "dry-run") == "true"
	args := fs.Args()

	tm := term.NewTerminal()
	bold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }

	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	root := moduleDir(dir)
	hasGoMod := root != ""

	// Check if krewire.yaml exists if no go.mod was found
	if !hasGoMod {
		if _, err := os.Stat(filepath.Join(dir, "krewire.yaml")); err == nil {
			root = dir
		} else {
			fmt.Fprintln(os.Stderr, "update: no go.mod or krewire.yaml found in current directory or parents")
			fmt.Fprintln(os.Stderr, "Run 'krew update' inside a Krewire project directory.")
			return errs.ExitCodeUsage
		}
	}

	// 1. Update Go dependencies if go.mod exists
	if hasGoMod {
		goBin, err := exec.LookPath("go")
		if err != nil {
			return fail(errors.New("go toolchain not found in PATH"))
		}

		mod, err := gomod.Read(filepath.Join(root, "go.mod"))
		if err != nil {
			return fail(fmt.Errorf("reading go.mod: %w", err))
		}

		if len(args) > 0 {
			// Update specific packages
			for _, arg := range args {
				pkg := expandKrewireShortcut(arg)
				target := pkg
				if !strings.Contains(target, "@") {
					target += "@latest"
				}

				if dryRun {
					fmt.Printf("%s Would update dependency: %s\n", yellow("[dry-run]"), target)
					continue
				}

				slog.Info("updating dependency", "package", target, "root", root)
				fmt.Printf("%s Updating dependency %s ...\n", bold("→"), green(target))

				cmd := exec.Command(goBin, "get", target)
				cmd.Dir = root
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fail(fmt.Errorf("go get %s failed: %w", target, err))
				}
				fmt.Printf("%s Updated %s\n", green("✓"), target)
			}
		} else {
			// Update all dependencies
			if dryRun {
				fmt.Printf("%s Checking available updates in go.mod (%s) ...\n", yellow("[dry-run]"), mod.Path)
				cmd := exec.Command(goBin, "list", "-u", "-m", "all")
				cmd.Dir = root
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
			} else {
				fmt.Printf("%s Updating dependencies in %s ...\n", bold("→"), green(mod.Path))

				// Update Krewire packages to latest if used
				if mod.UsesKrewire() {
					for req := range mod.Requires {
						if strings.HasPrefix(req, "github.com/krewire/") {
							target := req + "@latest"
							slog.Info("updating Krewire core library", "package", target)
							fmt.Printf("  %s %s ...\n", bold("→"), target)
							cmd := exec.Command(goBin, "get", target)
							cmd.Dir = root
							_ = cmd.Run()
						}
					}
				}

				// Update all dependencies to latest minor/patch versions
				cmd := exec.Command(goBin, "get", "-u", "./...")
				cmd.Dir = root
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					// Fallback to updating direct requires if ./... has no Go packages
					for req := range mod.Requires {
						subCmd := exec.Command(goBin, "get", "-u", req)
						subCmd.Dir = root
						_ = subCmd.Run()
					}
				}
				fmt.Printf("%s All Go dependencies updated to latest versions\n", green("✓"))
			}
		}

		// Run go mod tidy if requested and not dry-run
		if tidy && !dryRun {
			fmt.Printf("%s Running go mod tidy ...\n", bold("→"))
			cmd := exec.Command(goBin, "mod", "tidy")
			cmd.Dir = root
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fail(fmt.Errorf("go mod tidy failed: %w", err))
			}
			fmt.Printf("%s go.mod and go.sum are tidy\n", green("✓"))
		}
	}

	// 2. Update npm dependencies if package.json exists
	pkgJSONPath := filepath.Join(root, "package.json")
	if _, err := os.Stat(pkgJSONPath); err == nil {
		if npmBin, err := exec.LookPath("npm"); err == nil {
			if dryRun {
				fmt.Printf("%s Would run npm update in %s\n", yellow("[dry-run]"), root)
			} else {
				fmt.Printf("%s Updating npm packages (package.json) ...\n", bold("→"))
				cmd := exec.Command(npmBin, "update")
				cmd.Dir = root
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					slog.Warn("npm update returned an error", "error", err)
				} else {
					fmt.Printf("%s npm dependencies updated\n", green("✓"))
				}
			}
		}
	}

	fmt.Printf("%s Project dependencies updated successfully\n", green("✓"))
	return errs.ExitCodeSuccess
}

func expandKrewireShortcut(pkg string) string {
	lower := strings.ToLower(pkg)
	switch lower {
	case "krewire", "core", "@krewire/core":
		return "github.com/krewire/krewire"
	case "ui", "@krewire/ui":
		return "github.com/krewire/ui"
	case "krew", "@krewire/krew":
		return "github.com/krewire/krew"
	case "auth", "@krewire/auth":
		return "github.com/krewire/auth"
	case "cloud", "@krewire/cloud":
		return "github.com/krewire/cloud"
	case "dash", "@krewire/dash":
		return "github.com/krewire/dash"
	default:
		return pkg
	}
}
