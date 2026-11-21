// Package version defines the kiw CLI version aliased to the master krewire version.
package version

import (
	krewire "github.com/krewire/krewire"
	"github.com/krewire/krewire/pkg/version"
)

// Version is the semantic version of the Krewire monorepo.
var Version = krewire.Version

// VersionString returns the version as a string for UI display.
func VersionString() string { return krewire.VersionString() }

// EcosystemRequires declares the minimum version of each external module this one
// was built against.
var EcosystemRequires = map[string]version.Version{}
