package version

import (
	"github.com/krewire/krewire/version"
)

var rawVersion = ""

// Version is the semantic release version of the Krew AI Agent engine.
var Version = version.MustParseVersion("0.2.0")

// EcosystemRequires declares the minimum version of each ecosystem module this one
// was built against.
var EcosystemRequires = map[string]version.Version{
	"krewire": version.MustParseVersion("0.2.0"),
}

// VersionString returns the formatted version identifier.
func VersionString() string {
	if rawVersion != "" {
		return rawVersion
	}
	return Version.String()
}
