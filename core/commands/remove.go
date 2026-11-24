package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krewire/hub/resolver"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/version"
)

func RegisterRemove(fs *flag.FlagSet) {}

func RunRemove(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: krew remove <package> [package ...]")
		fmt.Fprintln(os.Stderr, "examples: krew remove twcss, krew remove tailwindcss, krew remove github.com/foo/bar")
		return errs.ExitCodeUsage
	}
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	chain := resolver.DefaultChain()
	for _, raw := range args {
		spec, err := resolver.ParseSpec(raw)
		if err != nil {
			// For remove, allow bare names without version parsing strictness
			spec = resolver.Spec{Name: raw, Raw: raw}
		}
		// Strip version for remove; only name matters
		spec.RawVer = ""
		spec.Version = version.Version{}
		resolved, err := chain.Resolve(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot resolve %q: %v\n", raw, err)
			return errs.ExitCodeFailure
		}
		slog.Info("removing package", "package", spec.Name, "kind", resolved.Kind)
		fmt.Printf("→ removing %s (%s) ...\n", spec.Name, resolved.Kind)
		if err := resolved.Installer.Remove(root); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove %q: %v\n", raw, err)
			return errs.ExitCodeFailure
		}
		fmt.Printf("✓ removed %s\n", spec.Name)
	}
	return errs.ExitCodeSuccess
}
