package krew

import (
	"github.com/krewire/krewire/pkg/version"
)

// Version is the semantic release version of the Krew AI Agent engine.
var Version = version.MustParseVersion("0.1.0")

// VersionString returns the formatted version identifier.
func VersionString() string {
	return Version.String()
}
