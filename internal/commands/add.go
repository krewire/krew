package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/hub/resolver"
	kiwdsl "github.com/krewire/krewire/pkg/kiw"
)

func RegisterAdd(fs *flag.FlagSet) {
	// Package versions are encoded in package@version syntax.
	fs.Bool("force", false, "ui/<name>: overwrite locally modified components")
	fs.Bool("dry-run", false, "ui/<name>: show what would change without writing")
}

func RunAdd(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: krew add <package[@version]|ui/<component>|ui:<pack>> ...")
		fmt.Fprintln(os.Stderr, "  package@latest  latest version (e.g. github.com/krewire/ui, twcss)")
		fmt.Fprintln(os.Stderr, "  package@1.2.3   specific version")
		fmt.Fprintln(os.Stderr, "  ui/<component>  copy a kui component into ./components (e.g. ui/button)")
		fmt.Fprintln(os.Stderr, "  ui:<pack>       copy a starter pack (ui:minimal, ui:base, ui:dashboard)")
		fmt.Fprintln(os.Stderr, "examples: krew add github.com/krewire/ui, krew add twcss, krew add ui/card")
		fmt.Fprintln(os.Stderr, "hint: browse components with 'krew ui list'")
		return errs.ExitCodeUsage
	}
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	var uiRefs, pkgs []string
	for _, a := range args {
		if strings.HasSuffix(a, kiwdsl.PackageExtension) || strings.HasPrefix(a, "pkg:") {
			if code := pkgAdd(root, strings.TrimPrefix(a, "pkg:")); code != errs.ExitCodeSuccess {
				return code
			}
			continue
		}
		if isUIRef(a) {
			uiRefs = append(uiRefs, a)
		} else {
			pkgs = append(pkgs, a)
		}
	}
	if len(uiRefs) > 0 {
		uiRoot := root
		if r, err := findRoot(); err == nil {
			uiRoot = r
		}
		if code := uiAdd(uiRoot, uiRefs, flagBool(fs, "force"), flagBool(fs, "dry-run")); code != errs.ExitCodeSuccess || len(pkgs) == 0 {
			return code
		}
	}
	args = pkgs
	chain := resolver.DefaultChain()
	for _, raw := range args {
		spec, err := resolver.ParseSpec(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid package %q: %v\n", raw, err)
			return errs.ExitCodeUsage
		}
		resolved, err := chain.Resolve(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot resolve %q: %v\n", raw, err)
			return errs.ExitCodeFailure
		}
		version := spec.EffectiveVersion()
		slog.Info("adding package", "package", spec.Name, "version", version, "kind", resolved.Kind, "resolver", resolved.Installer.Describe())
		fmt.Printf("→ adding %s (%s) ...\n", spec.Name+"@"+version, resolved.Kind)
		if err := resolved.Installer.Add(root, version); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add %q: %v\n", raw, err)
			return errs.ExitCodeFailure
		}
		fmt.Printf("✓ added %s\n", spec.Name+"@"+version)
	}
	return errs.ExitCodeSuccess
}
