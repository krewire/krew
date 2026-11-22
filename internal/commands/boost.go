package commands

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krewire/pkg/ai/boost"
	"github.com/krewire/krewire/pkg/errs"
)

// RegisterBoost registers flags for the boost command group.
func RegisterBoost(fs *flag.FlagSet) {
	fs.Bool("force", false, "overwrite existing managed files without prompting")
	fs.Bool("dry-run", false, "report the files that would be written without writing")
}

// RunBoost dispatches the boost sub-commands. Supports "install" and "status" (or "check").
func RunBoost(fs *flag.FlagSet) errs.ExitCode {
	sub := fs.Arg(0)
	switch sub {
	case "install":
		return runBoostInstall(fs, os.Stdin, os.Stdout)
	case "status", "check":
		return runBoostStatus(fs, os.Stdout)
	case "", "help", "-h", "--help":
		return usageMessage("usage: krew boost <install|status> [target] [--force] [--dry-run]")
	default:
		return usageMessage(fmt.Sprintf("unknown boost sub-command %q (supported: install, status)", sub))
	}
}

// runBoostStatus checks the installation status of Boost AI in target directory.
func runBoostStatus(fs *flag.FlagSet, out io.Writer) errs.ExitCode {
	target := fs.Arg(1)
	if target == "" {
		target = "."
	}
	target, err := absPath(target)
	if err != nil {
		return fail(err)
	}

	conflicts, err := detectManagedConflicts(target)
	if err != nil {
		return fail(err)
	}

	allManaged := boost.Managed()
	installedCount := len(conflicts)
	totalCount := len(allManaged)

	fmt.Fprintf(out, "Krewire Boost Status in %s:\n", target)
	if installedCount == 0 {
		fmt.Fprintln(out, "  Status: Not installed")
		fmt.Fprintln(out, "\nTo install Boost AI agent template, run:")
		fmt.Fprintf(out, "  krew boost install %s\n", target)
		return errs.ExitCodeSuccess
	}

	for _, m := range allManaged {
		path := filepath.Join(target, m)
		if _, err := os.Lstat(path); err == nil {
			fmt.Fprintf(out, "  ✓ %-24s (installed)\n", m)
		} else {
			fmt.Fprintf(out, "  ✗ %-24s (missing)\n", m)
		}
	}

	fmt.Fprintln(out)
	if installedCount == totalCount {
		fmt.Fprintf(out, "Boost AI is fully installed (%d/%d assets ready).\n", installedCount, totalCount)
	} else {
		fmt.Fprintf(out, "Boost AI is partially installed (%d/%d assets present).\n", installedCount, totalCount)
		fmt.Fprintln(out, "Run 'krew boost install --force' to repair missing assets.")
	}
	return errs.ExitCodeSuccess
}

// runBoostInstall installs the Boost template into a target directory. When
// the target is not supplied it is asked for interactively; existing managed
// files prompt for confirmation unless --force is given.
func runBoostInstall(fs *flag.FlagSet, in io.Reader, out io.Writer) errs.ExitCode {
	target := fs.Arg(1)
	if target == "" {
		line, err := promptLine(in, out, "Target directory (enter `.` for current): ")
		if err != nil {
			return fail(err)
		}
		target = line
	}
	target = strings.TrimSpace(target)
	if target == "" {
		target = "."
	}

	force := boolFlag(fs, "force")
	dryRun := boolFlag(fs, "dry-run")
	for _, arg := range fs.Args()[1:] {
		switch arg {
		case "--force":
			force = true
		case "--dry-run":
			dryRun = true
		}
	}

	if !force && !dryRun {
		conflicts, err := detectManagedConflicts(target)
		if err != nil {
			return fail(err)
		}
		if len(conflicts) > 0 {
			fmt.Fprintln(out, "The following managed files already exist:")
			for _, c := range conflicts {
				fmt.Fprintf(out, "  %s\n", c)
			}
			answer, err := promptLine(in, out, "Overwrite them? [y/N]: ")
			if err != nil {
				return fail(err)
			}
			if !isYes(answer) {
				fmt.Fprintln(out, "Aborted.")
				return errs.ExitCodeUsage
			}
			force = true
		}
	}

	opts := []boost.Option{}
	if force {
		opts = append(opts, boost.WithForce())
	}
	if dryRun {
		opts = append(opts, boost.WithDryRun())
	}

	target, err := absPath(target)
	if err != nil {
		return fail(err)
	}
	created, err := boost.Install(target, opts...)
	if err != nil {
		return boostInstallError(err)
	}

	if dryRun {
		fmt.Fprintln(out, "Dry run — would install Boost into "+target)
	} else {
		slog.Info("installed boost template", "dir", target, "files", len(created))
		fmt.Fprintln(out, "Installed Boost into "+target)
	}
	for _, path := range created {
		fmt.Fprintln(out, "created "+path)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintln(out, "  1. cd "+target)
	fmt.Fprintln(out, "  2. Open opencode there.")
	fmt.Fprintln(out, "  3. Run /kickoff so the agent maps the project.")
	return errs.ExitCodeSuccess
}

// absPath returns the absolute form of target.
func absPath(target string) (string, error) {
	if strings.HasPrefix(target, "~") {
		return "", errs.UsageError("shell expansion is not applied; use an absolute or relative path")
	}
	return filepath.Abs(target)
}

// detectManagedConflicts reports managed boost paths that already exist under
// target, mirroring the library's conflict detection for prompting.
func detectManagedConflicts(target string) ([]string, error) {
	var conflicts []string
	for _, rel := range boost.Managed() {
		p := filepath.Join(target, rel)
		if _, err := os.Lstat(p); err == nil {
			conflicts = append(conflicts, p)
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return conflicts, nil
}

// boostInstallError maps the library's sentinel errors onto exit codes.
func boostInstallError(err error) errs.ExitCode {
	fmt.Fprintln(os.Stderr, "krew:", err)
	switch {
	case errors.Is(err, boost.ErrTargetMissing), errors.Is(err, boost.ErrConflicts):
		return errs.ExitCodeUsage
	default:
		return errs.ExitCodeFailure
	}
}

// promptLine writes question and reads one trimmed input line.
func promptLine(in io.Reader, out io.Writer, question string) (string, error) {
	fmt.Fprint(out, question)
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", nil
	}
	return strings.TrimSpace(sc.Text()), nil
}

// isYes reports whether a confirmation answer is affirmative.
func isYes(answer string) bool {
	a := strings.ToLower(strings.TrimSpace(answer))
	return a == "y" || a == "yes"
}

// usageMessage prints a usage line and returns ExitCodeUsage.
func usageMessage(msg string) errs.ExitCode {
	fmt.Fprintln(os.Stderr, "krew: "+msg)
	return errs.ExitCodeUsage
}

func boolFlag(fs *flag.FlagSet, name string) bool {
	if f := fs.Lookup(name); f != nil {
		return f.Value.String() == "true"
	}
	return false
}
