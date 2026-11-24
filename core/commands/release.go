package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/release"
	"github.com/krewire/krew/core/shape"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/version"
	"github.com/krewire/tui/term"
)

type releaseModuleSlice []string

func (s *releaseModuleSlice) String() string { return strings.Join(*s, ",") }
func (s *releaseModuleSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var (
	releaseModules    releaseModuleSlice
	releaseAll        bool
	releaseBump       string
	releaseTrack      string
	releaseApply      bool
	releaseNotes      bool
	releaseQualify    bool
	releaseJSON       bool
	releaseOut        string
	releaseIncludeExp bool
)

// RegisterRelease declares the flags for the release command.
func RegisterRelease(fs *flag.FlagSet) {
	fs.Var(&releaseModules, "module", "module to release; repeatable ("+releaseModuleNames()+"); use with --all for every module")
	fs.BoolVar(&releaseAll, "all", false, "release every Krewire module (maintainer mode, run inside the workspace)")
	fs.StringVar(&releaseBump, "bump", "patch", "version component to increment: patch|minor|major")
	fs.StringVar(&releaseTrack, "track", "stable", "release track: lts, stable, rc, beta, alpha, nightly, experimental, dev")
	fs.BoolVar(&releaseApply, "apply", false, "write the changes (default is a dry run)")
	fs.BoolVar(&releaseNotes, "notes", false, "print release notes from git log since the last tag")
	fs.BoolVar(&releaseQualify, "qualify", false, "run cross-repository release qualification checks against compatibility matrix")
	fs.BoolVar(&releaseJSON, "json", false, "output qualification report in machine-readable JSON format")
	fs.StringVar(&releaseOut, "out", "", "output file path to write qualification report artifact")
	fs.BoolVar(&releaseIncludeExp, "include-experimental", false, "include experimental and canary boundaries in release qualification")
}

// RunRelease stages a release. There are two audiences:
//
//   - Maintainer mode: inside the Krewire workspace with --module/--all, it bumps
//     the released modules' versions and propagates the new minimum versions into
//     every dependent module's EcosystemRequires.
//   - Project mode: inside a Krewire project (with krewire.yaml) with
//     no --module flag, it bumps the project version in krewire.yaml and builds the
//     project, so ecosystem developers can release the software they build.
//
// It is a dry run unless --apply is given.
func RunRelease(fs *flag.FlagSet) errs.ExitCode {
	if releaseQualify || (fs != nil && len(fs.Args()) > 0 && fs.Args()[0] == "qualify") {
		return runReleaseQualification(releaseJSON, releaseOut, releaseIncludeExp)
	}

	bump, err := release.ParseBump(releaseBump)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return errs.ExitCodeUsage
	}

	track, err := release.ParseTrack(releaseTrack)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return errs.ExitCodeUsage
	}

	if releaseAll || len(releaseModules) > 0 {
		return runModuleRelease(bump, track)
	}
	return runProjectRelease(bump, track)
}

// runModuleRelease implements maintainer mode (Krewire's own modules).
func runModuleRelease(bump release.BumpKind, track release.Track) errs.ExitCode {
	var names []string
	if releaseAll {
		for _, m := range release.Modules {
			names = append(names, m.Name)
		}
	} else {
		for _, s := range releaseModules {
			if !validReleaseModule(s) {
				fmt.Fprintf(os.Stderr, "release: unknown module %q\n", s)
				return errs.ExitCodeUsage
			}
			names = append(names, s)
		}
	}

	edits, err := release.PlanTrack(names, bump, track)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return errs.ExitCodeFailure
	}

	var root string
	if releaseApply || releaseNotes {
		if root, err = findReleaseWorkspaceRoot(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errs.ExitCodeFailure
		}
	}

	// Tier compliance gate: a module's LICENSE must agree with the tier the
	// registry assigns it. Relicensing a free module or shipping MIT on a paid
	// one is a licensing bug, so it blocks the release rather than shipping.
	if root != "" {
		if err := release.CheckTierCompliance(root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errs.ExitCodeFailure
		}
	}

	tm := term.NewTerminal()
	label := "DRY RUN"
	if releaseApply {
		label = "APPLY"
	}
	fmt.Printf("%s: %d version.go change(s) for this module release\n", tm.Paint(label, term.ColorDefault, []term.Style{term.StyleBold}), len(edits))
	for _, e := range edits {
		fmt.Printf("  • %s\n    %s\n", tm.Paint(e.File, term.ColorCyan, nil), e.Summary)
	}

	if releaseApply {
		modified, err := release.Apply(edits, root, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errs.ExitCodeFailure
		}
		fmt.Printf("modified %d file(s)\n", len(modified))
	}

	fmt.Println("\nNext steps (run manually, or automate with a release workflow):")
	for _, n := range names {
		cur, _ := release.CurrentVersion(n)
		nv := release.Bump(cur, bump)
		m := release.ManifestFor(n)
		fmt.Printf("  cd %s && git add -A && git commit -m \"release: %s %s\" && git tag v%s\n", m.Dir, n, nv.String(), nv.String())
	}
	fmt.Println("  • remove local replace directives in kiw/go.mod before publishing (see AGENTS.md release rule)")
	fmt.Println("  • run `krew compat` to confirm the ecosystem is still mutually compatible")

	if releaseNotes {
		printReleaseNotes(root, names)
	}
	return errs.ExitCodeSuccess
}

// runProjectRelease implements developer mode: release the Krewire project in
// the current directory by bumping its version in krewire.yaml and building it.
func runProjectRelease(bump release.BumpKind, track release.Track) errs.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return errs.ExitCodeFailure
	}
	if _, err := os.Stat(filepath.Join(dir, config.ConfigFileName)); err != nil {
		fmt.Fprintln(os.Stderr, "release: not a Krewire project (no krewire.yaml found); run inside your project, or use --module/--all inside the Krewire workspace")
		return errs.ExitCodeUsage
	}
	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return errs.ExitCodeFailure
	}
	res, _ := shape.Detect(dir)
	_ = res

	cur := cfg.Version
	prefix := ""
	stripped := strings.TrimPrefix(cur, "v")
	if cur != stripped {
		prefix = "v"
	}
	cv := version.Version{}
	if stripped == "" {
		cv = version.MustParseVersion("0.1.0")
	} else if cv, err = version.ParseVersion(stripped); err != nil {
		fmt.Fprintf(os.Stderr, "release: invalid project version %q in krewire.yaml: %v\n", cur, err)
		return errs.ExitCodeUsage
	}
	nv := release.BumpTrack(cv, bump, track)
	newVer := prefix + nv.String()

	tm := term.NewTerminal()
	label := "DRY RUN"
	if releaseApply {
		label = "APPLY"
	}
	fmt.Printf("%s: release project %q %s -> %s\n", tm.Paint(label, term.ColorDefault, []term.Style{term.StyleBold}), dir, cur, newVer)
	fmt.Printf("  • %s\n    %s\n", tm.Paint("krewire.yaml", term.ColorCyan, nil), fmt.Sprintf("version %s -> %s", cur, newVer))

	if releaseApply {
		if err := bumpKrewireVersion(dir, newVer); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errs.ExitCodeFailure
		}
		fmt.Println("  • building the project with `krew build`")
		bfs := flag.NewFlagSet("build", flag.ContinueOnError)
		RegisterBuild(bfs)
		if code := RunBuild(bfs); code != errs.ExitCodeSuccess {
			return code
		}
	}

	fmt.Printf("\nNext steps (run manually, or automate with a release workflow):\n")
	fmt.Printf("  git add krewire.yaml && git commit -m \"release: %s\" && git tag %s\n", newVer, newVer)
	if releaseNotes {
		if notes := gitLogNotes(dir); notes != "" {
			fmt.Printf("\nRelease notes:\n%s\n", notes)
		}
	}
	return errs.ExitCodeSuccess
}

// bumpKrewireVersion sets the top-level `version:` field in dir/krewire.yaml to
// newVer, preserving the file's other content. It adds the line when missing.
func bumpKrewireVersion(dir, newVer string) error {
	p := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("release: read %s: %w", p, err)
	}
	re := regexp.MustCompile(`(?m)^(version:\s*).*$`)
	updated := re.ReplaceAllString(string(data), `$1"`+newVer+`"`)
	if updated == string(data) {
		// No version: line present; append one.
		updated = string(data)
		if !strings.HasSuffix(updated, "\n") {
			updated += "\n"
		}
		updated += fmt.Sprintf("version: \"%s\"\n", newVer)
	}
	return os.WriteFile(p, []byte(updated), 0o644)
}

func validReleaseModule(n string) bool {
	for _, m := range release.Modules {
		if m.Name == n {
			return true
		}
	}
	return false
}

// releaseModuleNames lists the releasable module names from the release
// manifest, so the flag help can never drift from release.Modules.
func releaseModuleNames() string {
	names := make([]string, 0, len(release.Modules))
	for _, m := range release.Modules {
		names = append(names, m.Name)
	}
	return strings.Join(names, ", ")
}

// findReleaseWorkspaceRoot walks up from the working directory to locate the ecosystem workspace root.
func findReleaseWorkspaceRoot() (string, error) {
	return release.FindWorkspaceRoot()
}

// printReleaseNotes prints Conventional-Commits-style notes per released module.
func printReleaseNotes(root string, names []string) {
	for _, n := range names {
		m := release.ManifestFor(n)
		if notes := gitLogNotes(filepath.Join(root, m.Dir)); notes != "" {
			fmt.Printf("\nRelease notes for %s:\n%s\n", n, notes)
		}
	}
}

func gitLogNotes(dir string) string {
	last := lastGitTag(dir)
	rev := "HEAD"
	if last != "" {
		rev = last + "..HEAD"
	}
	out, err := exec.Command("git", "-C", dir, "log", rev, "--pretty=format:- %s").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func lastGitTag(dir string) string {
	out, err := exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runReleaseQualification(asJSON bool, outFile string, includeExp bool) errs.ExitCode {
	ctx := &release.QualificationContext{
		IncludeExperimental: includeExp,
	}
	report, err := release.RunQualification(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "release qualification error: %v\n", err)
		return errs.ExitCodeFailure
	}

	if asJSON {
		if err := release.ExportJSON(report, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "failed to export JSON: %v\n", err)
			return errs.ExitCodeFailure
		}
	} else {
		fmt.Print(release.FormatHumanReadable(report))
	}

	if outFile != "" {
		if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err == nil {
			if f, err := os.Create(outFile); err == nil {
				defer f.Close()
				_ = release.ExportJSON(report, f)
			}
		}
	}

	if !report.OverallPassed {
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}
