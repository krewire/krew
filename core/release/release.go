// Package release implements the Krewire ecosystem release topology and the
// version-bump propagation used by `kiw release`.
package release

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/krewire/krewire/version"
)

// ModuleName is a typed name of an ecosystem module.
type ModuleName string

const (
	ModuleKrewire ModuleName = "krewire"
	ModuleKrew    ModuleName = "krew"
	ModuleCloud   ModuleName = "cloud"
	ModuleAuth    ModuleName = "auth"
	ModuleDash    ModuleName = "dash"
)

func ident(name string) string {
	switch ModuleName(name) {
	case ModuleKrewire:
		return "ModuleKrewire"
	case ModuleKrew:
		return "ModuleKrew"
	case ModuleCloud:
		return "ModuleCloud"
	case ModuleAuth:
		return "ModuleAuth"
	case ModuleDash:
		return "ModuleDash"
	}
	return ""
}

// Edit describes one replacement to be made in a version file.
type Edit struct {
	Module  string
	File    string
	Summary string
	From    string
	To      string
}

// Plan computes the exact file edits required to release the named modules on the stable track.
func Plan(targets []string, b BumpType) ([]Edit, error) {
	return PlanTrack(targets, b, TrackStable)
}

// PlanTrack computes the exact file edits required to release the named modules on the specified release track.
func PlanTrack(targets []string, b BumpType, t Track) ([]Edit, error) {
	var edits []Edit
	seen := map[string]bool{}

	for _, name := range targets {
		cur, err := CurrentVersion(name)
		if err != nil {
			return nil, err
		}
		next := BumpTrack(cur, b, t)

		m := manifest(name)
		if m.Name == "" {
			return nil, fmt.Errorf("no manifest for %q", name)
		}
		from := versionDecl(name, cur.String())
		to := versionDecl(name, next.String())
		edits = append(edits, Edit{
			Module:  name,
			File:    m.VersionFile,
			Summary: fmt.Sprintf("bump %s %s -> %s [%s]", name, cur, next, t),
			From:    from,
			To:      to,
		})

		for _, dep := range dependents[name] {
			key := dep + ":" + name
			if seen[key] {
				continue
			}
			reqVer, ok := RequiredVersion(dep, name)
			if !ok {
				continue
			}
			if next.IsCompatible(reqVer) {
				continue
			}
			depM := manifest(dep)
			if depM.Name == "" {
				continue
			}
			seen[key] = true
			edits = append(edits, Edit{
				Module:  dep,
				File:    depM.VersionFile,
				Summary: fmt.Sprintf("raise %s requires %s -> %s", dep, name, next),
				From:    reqDecl(name, reqVer.String()),
				To:      reqDecl(name, next.String()),
			})
		}
	}

	sort.Slice(edits, func(i, j int) bool {
		if edits[i].File != edits[j].File {
			return edits[i].File < edits[j].File
		}
		return edits[i].Summary < edits[j].Summary
	})

	return edits, nil
}

func versionDecl(name, ver string) string {
	return fmt.Sprintf(`var Version = version.MustParseVersion("%s")`, ver)
}

func reqDecl(name, ver string) string {
	return fmt.Sprintf(`"%s": version.MustParseVersion("%s"),`, name, ver)
}

// Apply executes a list of Edits against the filesystem.
func Apply(edits []Edit, root string, dryRun bool) ([]string, error) {
	var modified []string
	for _, e := range edits {
		targetPath := e.File
		if root != "" && !filepath.IsAbs(targetPath) {
			m := manifest(e.Module)
			candidate := filepath.Join(root, targetPath)
			if m.Dir != "" {
				dirCandidate := filepath.Join(root, m.Dir, e.File)
				if _, err := os.Stat(dirCandidate); err == nil {
					candidate = dirCandidate
				}
			}
			if _, err := os.Stat(candidate); os.IsNotExist(err) && e.Module != "" {
				altPath := filepath.Join(root, e.Module, e.File)
				if _, aerr := os.Stat(altPath); aerr == nil {
					candidate = altPath
				}
			}
			targetPath = candidate
		}
		if dryRun {
			modified = append(modified, targetPath)
			continue
		}
		data, err := os.ReadFile(targetPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", targetPath, err)
		}
		content := string(data)
		var replaced string
		if strings.Contains(content, e.From) {
			replaced = strings.Replace(content, e.From, e.To, 1)
		} else {
			// Try flexible matching for map entries where gofmt aligned colons with multiple spaces/tabs
			words := strings.Fields(e.From)
			if len(words) >= 2 {
				pattern := regexp.QuoteMeta(words[0]) + `\s+` + regexp.QuoteMeta(strings.Join(words[1:], " "))
				re, err := regexp.Compile(pattern)
				if err == nil && re.MatchString(content) {
					replaced = re.ReplaceAllString(content, e.To)
				}
			}
			if replaced == "" {
				return nil, fmt.Errorf("%s: target text not found: %q", targetPath, e.From)
			}
		}
		if err := os.WriteFile(targetPath, []byte(replaced), 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", targetPath, err)
		}
		modified = append(modified, targetPath)
	}
	return modified, nil
}

func mustCur(name string) version.Version {
	v, err := CurrentVersion(name)
	if err != nil {
		panic(err)
	}
	return v
}
