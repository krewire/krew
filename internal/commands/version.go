package commands

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/krewire/krew/internal/buildinfo"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/version"
	"github.com/krewire/tui/term"
)

func RunVersion(_ *flag.FlagSet) errs.ExitCode {
	tm := term.NewTerminal()
	bold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }

	fmt.Printf("%s %s\n", bold("krew"), dim("Unified Developer Toolchain"))
	ver := qualifiedVersion(buildinfo.ModKrewire)
	verColor := green(ver)
	if strings.Contains(ver, "dev") {
		verColor = yellow(ver)
	}
	fmt.Printf("  %-18s %s\n", dim("Krewire Ecosystem"), verColor)
	fmt.Printf("  %-18s %s\n", dim("Signature"), green(version.BuildSignature()))
	fmt.Printf("  %-18s %s\n", dim("Go"), dim(runtime.Version()+" ("+runtime.GOOS+"/"+runtime.GOARCH+")"))
	if prov, ok := buildinfo.ReadProvenance(); ok && prov.Revision != "" {
		rev := prov.Revision
		if len(rev) > 8 {
			rev = rev[:8]
		}
		if prov.Modified {
			rev += "-dirty"
		}
		fmt.Printf("  %-18s %s\n", dim("Commit"), dim(rev))
	}
	return errs.ExitCodeSuccess
}

func humanVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == "devel" || v == buildinfo.DevelVersion {
		return "dev"
	}
	return "v" + v
}

// qualifiedVersion renders the module version for display: "v0.1.0" for a
// released tag, "v0.1.0 (dev)" when resolved from workspace sources, or "dev"
// when unknown.
func qualifiedVersion(path string) string {
	v, fromSource := buildinfo.ResolveVersion(path)
	human := humanVersion(v)
	if fromSource && human != "dev" {
		return human + " (dev)"
	}
	return human
}

func resolveVersion() string {
	v, _ := buildinfo.ResolveVersion(buildinfo.ModKrewire)
	return v
}
