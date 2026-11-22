package release

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	krewire "github.com/krewire/krewire"
	"github.com/krewire/krewire/pkg/model"
	"github.com/krewire/krewire/pkg/version"
)

// Manifest describes a module's location and version file within the workspace.
// It is the single source of truth for the release topology.
type Manifest struct {
	Name        string
	Dir         string // workspace-relative directory
	VersionFile string // module-relative path to the file declaring Version
	Tier        model.Tier
}

// Modules lists the ecosystem modules in dependency order (dependencies first).
var Modules = []Manifest{
	{Name: "krewire", Dir: "krewire", VersionFile: "version.go", Tier: model.TierFree},
	{Name: "krew", Dir: "krew", VersionFile: "version.go", Tier: model.TierEnterprise},
	{Name: "auth", Dir: "auth", VersionFile: "version.go", Tier: model.TierFree},
	{Name: "cloud", Dir: "cloud", VersionFile: "version.go", Tier: model.TierFree},
	{Name: "dash", Dir: "dash", VersionFile: "version.go", Tier: model.TierFree},
}

// dependents maps each module to the modules that require it.
var dependents = map[string][]string{
	"krewire": {"krew", "auth", "cloud", "dash"},
	"krew":    {},
	"auth":    {"cloud", "dash"},
	"cloud":   {},
	"dash":    {},
}

// DefaultRequirements maps each module to the static default versions of ecosystem modules it requires.
var DefaultRequirements = map[string]map[string]version.Version{
	"krewire": {},
	"krew": {
		"krewire": krewire.Version,
	},
	"auth": {
		"krewire": krewire.Version,
	},
	"cloud": {
		"krewire": krewire.Version,
		"auth":    krewire.Version,
	},
	"dash": {
		"krewire": krewire.Version,
		"auth":    krewire.Version,
	},
}

var (
	versionDeclRe = regexp.MustCompile(`var\s+Version\s*=\s*version\.MustParseVersion\("([^"]+)"\)`)
	reqItemRe     = regexp.MustCompile(`"([a-zA-Z0-9_-]+)"\s*:\s*version\.MustParseVersion\("([^"]+)"\)`)
)

// FindWorkspaceRoot walks up from working directory or executable location to locate go.work.
func FindWorkspaceRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.work not found")
}

// resolveModuleVersionPath finds the path to the version file for the named module.
func resolveModuleVersionPath(root, module string) (string, error) {
	m := manifest(module)
	if m.Name == "" {
		return "", fmt.Errorf("unknown module %q", module)
	}
	candidates := []string{
		filepath.Join(root, m.Dir, m.VersionFile),
		filepath.Join(root, m.Name, m.VersionFile),
		filepath.Join(root, m.VersionFile),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("version file not found for module %q in workspace %s", module, root)
}

// ReadModuleVersion parses and returns the Version declared in module's version.go.
func ReadModuleVersion(root, module string) (version.Version, error) {
	path, err := resolveModuleVersionPath(root, module)
	if err != nil {
		return version.Version{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return version.Version{}, err
	}
	matches := versionDeclRe.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return version.Version{}, fmt.Errorf("could not find Version declaration in %s", path)
	}
	return version.ParseVersion(matches[1])
}

// ReadModuleRequirements parses EcosystemRequires from module's version.go.
func ReadModuleRequirements(root, module string) (map[string]version.Version, error) {
	path, err := resolveModuleVersionPath(root, module)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	reqs := make(map[string]version.Version)
	matches := reqItemRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			v, err := version.ParseVersion(m[2])
			if err == nil {
				reqs[m[1]] = v
			}
		}
	}
	return reqs, nil
}

// ReadEcosystemActualVersions parses and returns the current versions of all workspace modules.
func ReadEcosystemActualVersions(root string) (map[string]version.Version, error) {
	actual := make(map[string]version.Version)
	for _, m := range Modules {
		v, err := ReadModuleVersion(root, m.Name)
		if err == nil {
			actual[m.Name] = v
		} else {
			if cv, cerr := CurrentVersion(m.Name); cerr == nil {
				actual[m.Name] = cv
			}
		}
	}
	return actual, nil
}

// ReadEcosystemRequirements parses and returns the declared EcosystemRequires for all workspace modules.
func ReadEcosystemRequirements(root string) (map[string]map[string]version.Version, error) {
	allReqs := make(map[string]map[string]version.Version)
	for _, m := range Modules {
		reqs, err := ReadModuleRequirements(root, m.Name)
		if err == nil {
			allReqs[m.Name] = reqs
		} else if def, ok := DefaultRequirements[m.Name]; ok {
			allReqs[m.Name] = def
		} else {
			allReqs[m.Name] = make(map[string]version.Version)
		}
	}
	return allReqs, nil
}

// ManifestFor returns the manifest for name.
func ManifestFor(name string) Manifest { return manifest(name) }

func manifest(name string) Manifest {
	for _, m := range Modules {
		if m.Name == name {
			return m
		}
	}
	return Manifest{}
}

// CurrentTier returns the declared Tier of module name.
func CurrentTier(name string) (model.Tier, error) {
	switch ModuleName(name) {
	case ModuleKrewire:
		return krewire.Tier, nil
	case ModuleKrew:
		return model.TierEnterprise, nil
	case ModuleCloud, ModuleAuth, ModuleDash:
		return model.TierFree, nil
	default:
		return "", fmt.Errorf("unknown module %q", name)
	}
}

// CheckTierCompliance validates licensing and tier gates for workspace modules.
func CheckTierCompliance(root string) error {
	for _, m := range Modules {
		if !m.Tier.IsValid() {
			return fmt.Errorf("module %q has invalid tier %q: want free, pro, team, or enterprise", m.Name, m.Tier)
		}
		curTier, err := CurrentTier(m.Name)
		if err == nil && curTier != m.Tier {
			return fmt.Errorf("module %q manifest tier %q does not match compiled tier %q", m.Name, m.Tier, curTier)
		}
		if root != "" {
			licPath := filepath.Join(root, m.Dir, "LICENSE")
			if _, err := os.Stat(licPath); os.IsNotExist(err) {
				altPath := filepath.Join(root, m.Name, "LICENSE")
				if _, aerr := os.Stat(altPath); aerr == nil {
					licPath = altPath
				}
			}
			if data, err := os.ReadFile(licPath); err == nil {
				content := string(data)
				if m.Tier == model.TierFree {
					if !strings.Contains(content, "MIT License") && !strings.Contains(content, "Permission is hereby granted") {
						return fmt.Errorf("module %q license at %s must be open source", m.Name, licPath)
					}
				} else {
					if !strings.Contains(content, "Proprietary License") && !strings.Contains(content, "All rights reserved") {
						return fmt.Errorf("module %q license at %s must be proprietary", m.Name, licPath)
					}
				}
			}
		}
	}
	return nil
}

// CurrentVersion returns the currently compiled or declared Version of module name.
func CurrentVersion(name string) (version.Version, error) {
	m := manifest(name)
	if m.Name == "" {
		return version.Version{}, fmt.Errorf("unknown module %q", name)
	}
	if root, err := FindWorkspaceRoot(); err == nil {
		if v, err := ReadModuleVersion(root, name); err == nil {
			return v, nil
		}
	}
	switch ModuleName(name) {
	case ModuleKrewire:
		return krewire.Version, nil
	case ModuleKrew, ModuleCloud, ModuleAuth, ModuleDash:
		return krewire.Version, nil
	default:
		return version.Version{}, fmt.Errorf("unknown module %q", name)
	}
}

// RequiredVersion returns the minimum version that dependent currently requires of name.
func RequiredVersion(dependent, name string) (version.Version, bool) {
	if root, err := FindWorkspaceRoot(); err == nil {
		if reqs, err := ReadModuleRequirements(root, dependent); err == nil {
			if reqVer, ok := reqs[name]; ok {
				return reqVer, true
			}
		}
	}
	if reqs, ok := DefaultRequirements[dependent]; ok {
		if reqVer, ok := reqs[name]; ok {
			return reqVer, true
		}
	}
	return version.Version{}, false
}
