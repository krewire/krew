package release

import (
	"fmt"
	"strings"
	"time"

	"github.com/krewire/krewire/version"
)

// BumpType selects which semver component to increment.
type BumpType string

type BumpKind = BumpType

const (
	BumpPatch BumpType = "patch"
	BumpMinor BumpType = "minor"
	BumpMajor BumpType = "major"
)

// Track represents an ecosystem version release track.
type Track string

const (
	TrackStable       Track = "stable"
	TrackLTS          Track = "lts"
	TrackRC           Track = "rc"
	TrackBeta         Track = "beta"
	TrackAlpha        Track = "alpha"
	TrackNightly      Track = "nightly"
	TrackExperimental Track = "experimental"
	TrackDev          Track = "dev"
)

// ParseTrack parses a string into a release Track.
func ParseTrack(s string) (Track, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "stable", "ga", "":
		return TrackStable, nil
	case "lts":
		return TrackLTS, nil
	case "rc":
		return TrackRC, nil
	case "beta":
		return TrackBeta, nil
	case "alpha":
		return TrackAlpha, nil
	case "nightly":
		return TrackNightly, nil
	case "experimental", "exp":
		return TrackExperimental, nil
	case "development", "dev":
		return TrackDev, nil
	default:
		return "", fmt.Errorf("invalid release track %q: want lts, stable, rc, beta, alpha, nightly, experimental, or dev", s)
	}
}

// ParseBump parses a string into a BumpType.
func ParseBump(s string) (BumpType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "patch", "p", "":
		return BumpPatch, nil
	case "minor", "m":
		return BumpMinor, nil
	case "major":
		return BumpMajor, nil
	default:
		return "", fmt.Errorf("invalid bump kind %q: want patch, minor, or major", s)
	}
}

// Bump returns a new Version with the chosen semver component incremented.
func Bump(v version.Version, b BumpType) version.Version {
	return BumpTrack(v, b, TrackStable)
}

// BumpTrack returns a new Version with the chosen semver bump and release track applied.
func BumpTrack(v version.Version, b BumpType, t Track) version.Version {
	var next version.Version
	switch b {
	case BumpMajor:
		next = version.Version{Major: v.Major + 1, Minor: 0, Patch: 0}
	case BumpMinor:
		next = version.Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
	case BumpPatch:
		next = version.Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	default:
		next = v
	}

	switch t {
	case TrackLTS:
		next.PreRelease = "lts"
	case TrackRC:
		next.PreRelease = "rc.1"
	case TrackBeta:
		next.PreRelease = "beta.1"
	case TrackAlpha:
		next.PreRelease = "alpha.1"
	case TrackNightly:
		next.PreRelease = fmt.Sprintf("nightly.%s", time.Now().UTC().Format("20060102"))
	case TrackExperimental:
		next.PreRelease = "exp"
	case TrackDev:
		next.PreRelease = "dev"
	case TrackStable:
		next.PreRelease = ""
	}
	return next
}
