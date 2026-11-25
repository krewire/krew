package version

import (
	ver "github.com/krewire/krewire/version"
)

var rawVersion = ""

var Version = ver.MustParseVersion("0.2.0")

func VersionString() string {
	if rawVersion != "" {
		return rawVersion
	}
	return Version.String()
}

var EcosystemRequires = map[string]ver.Version{
	"krewire": ver.MustParseVersion("0.2.0"),
}
