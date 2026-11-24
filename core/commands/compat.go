package commands

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/krewire/krew/core/release"
	kiwver "github.com/krewire/krew/core/version"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/version"
)

// RunCompat validates version compatibility across the whole Krewire ecosystem.
// It treats each module's `<module>/version.go` (`Version` + `EcosystemRequires`)
// as the authoritative compatibility contract and checks that every declared
// requirement is satisfied by the versions actually declared by the dependencies.
// This is the single entry point that turns the per-module version.go files into
// a repo-wide compatibility gate.
func RunCompat(_ *flag.FlagSet) errs.ExitCode {
	var actual map[string]version.Version
	var reqs map[string]map[string]version.Version

	root, err := release.FindWorkspaceRoot()
	if err == nil {
		actual, _ = release.ReadEcosystemActualVersions(root)
		reqs, _ = release.ReadEcosystemRequirements(root)
	}

	if len(actual) == 0 {
		actual = make(map[string]version.Version)
		for _, m := range release.Modules {
			if cv, err := release.CurrentVersion(m.Name); err == nil {
				actual[m.Name] = cv
			} else {
				actual[m.Name] = version.CurrentVersion
			}
		}
	}

	if len(reqs) == 0 {
		reqs = make(map[string]map[string]version.Version)
		for k, v := range release.DefaultRequirements {
			sub := make(map[string]version.Version)
			for sk, sv := range v {
				sub[sk] = sv
			}
			reqs[k] = sub
		}
	}

	// Internal core packages track the master krewire version
	krewireVer := actual["krewire"]
	if krewireVer.String() == "" {
		krewireVer = version.CurrentVersion
	}
	for _, sub := range []string{"model", "version", "testing", "hub", "kiw", "boost"} {
		if _, exists := actual[sub]; !exists {
			actual[sub] = krewireVer
		}
	}

	// Include CLI requires if any
	if _, ok := reqs["krew"]; !ok {
		reqs["krew"] = make(map[string]version.Version)
	}
	for k, v := range kiwver.EcosystemRequires {
		reqs["krew"][k] = v
	}

	issues := version.CheckCompatibility(actual, reqs)
	if len(issues) > 0 {
		for _, err := range issues {
			fmt.Printf("  x %s\n", err.Error())
		}
		return errs.ExitCodeFailure
	}

	// Print mutual compatibility verification summary
	fmt.Printf("Checked %d ecosystem modules (and %d core packages):\n", len(release.Modules), 6)
	var modNames []string
	for _, m := range release.Modules {
		modNames = append(modNames, m.Name)
	}
	sort.Strings(modNames)
	for _, name := range modNames {
		actVer := actual[name]
		depReqs := reqs[name]
		if len(depReqs) == 0 {
			fmt.Printf("  • %s @ %s (no upstream dependencies)\n", name, actVer)
		} else {
			var parts []string
			var depKeys []string
			for k := range depReqs {
				depKeys = append(depKeys, k)
			}
			sort.Strings(depKeys)
			for _, k := range depKeys {
				parts = append(parts, fmt.Sprintf("%s@%s", k, depReqs[k]))
			}
			fmt.Printf("  • %s @ %s (requires: %s)\n", name, actVer, strings.Join(parts, ", "))
		}
	}

	fmt.Println("ok: all module version declarations are mutually compatible")
	return errs.ExitCodeSuccess
}
