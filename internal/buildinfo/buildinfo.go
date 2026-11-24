// Package buildinfo resolves module versions for the devtool: from Go build
// metadata first, falling back to the master Krewire version.
package buildinfo

import (
	"path/filepath"
	"runtime/debug"
	"strings"

	krewire "github.com/krewire/krewire"
)

const (
	ModKrewire = "github.com/krewire/krewire"
)

// DevelVersion is the version Go records for modules built from source
// rather than a released tag.
const DevelVersion = "(devel)"

// ModuleVersion returns the version of the module at path as recorded in the
// build info, or "" when the module is not part of the build.
func ModuleVersion(path string) string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	if bi.Main.Path == path {
		return bi.Main.Version
	}
	for _, dep := range bi.Deps {
		if dep.Path == path {
			return dep.Version
		}
	}
	return ""
}

// KnownVersion returns the master version declared by the Krewire monorepo.
func KnownVersion(path string) string {
	if path == ModKrewire || path == "" {
		return krewire.VersionString()
	}
	return ""
}

// ResolveVersion returns the effective version of the module at path.
func ResolveVersion(path string) (version string, fromSource bool) {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		if bi.Main.Path == path {
			if bi.Main.Version != "" && bi.Main.Version != DevelVersion && !strings.HasPrefix(bi.Main.Version, "v0.0.0-") && !strings.Contains(bi.Main.Version, "+dirty") {
				return bi.Main.Version, false
			}
			if known := KnownVersion(path); known != "" {
				return known, true
			}
			return bi.Main.Version, true
		}
		for _, dep := range bi.Deps {
			if dep.Path == path {
				// Detect local source builds (go.work or local directory replace)
				if dep.Replace != nil {
					rep := dep.Replace
					if rep.Version == "" || rep.Version == DevelVersion || strings.HasPrefix(rep.Path, ".") || filepath.IsAbs(rep.Path) {
						if known := KnownVersion(path); known != "" {
							return known, true
						}
						if rep.Version != "" {
							return rep.Version, true
						}
						return dep.Version, true
					}
				}
				if dep.Version != "" && dep.Version != DevelVersion && !strings.HasPrefix(dep.Version, "v0.0.0-") && !strings.Contains(dep.Version, "+dirty") {
					return dep.Version, false
				}
				if known := KnownVersion(path); known != "" {
					return known, true
				}
				return dep.Version, true
			}
		}
	}
	if known := KnownVersion(path); known != "" {
		return known, true
	}
	return "", false
}

// Provenance contains VCS and compilation metadata extracted from debug.ReadBuildInfo.
type Provenance struct {
	Revision string
	Time     string
	Modified bool
	Trimpath bool
	Cgo      string
	GoOS     string
	GoArch   string
}

// ReadProvenance extracts build provenance settings recorded by the Go compiler.
func ReadProvenance() (Provenance, bool) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return Provenance{}, false
	}
	var prov Provenance
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			prov.Revision = s.Value
		case "vcs.time":
			prov.Time = s.Value
		case "vcs.modified":
			prov.Modified = (s.Value == "true")
		case "-trimpath":
			prov.Trimpath = (s.Value == "true")
		case "CGO_ENABLED":
			prov.Cgo = s.Value
		case "GOOS":
			prov.GoOS = s.Value
		case "GOARCH":
			prov.GoArch = s.Value
		}
	}
	return prov, true
}
