// Package ecosystem detects foreign host runtime ecosystems (Node.js,
// Laravel, PHP) and discovers co-located Krewire projects.
//
// It serves as the single source of truth for cross-ecosystem coexistence,
// manifest detection, and process/bridge metadata (KWN-NDJ5S, KWN-L4RVE).
package ecosystem

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	kiwconfig "github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krew/internal/shape"
)

// Manifest file names for foreign ecosystems.
const (
	ManifestPackageJSON  = "package.json"
	ManifestComposerJSON = "composer.json"
	FileArtisan          = "artisan"
)

// HostType identifies a foreign host runtime platform.
type HostType string

const (
	HostNone    HostType = ""
	HostNodeJS  HostType = "nodejs"
	HostLaravel HostType = "laravel"
	HostPHP     HostType = "php"
)

// HostProject describes a detected foreign project hosting the current environment.
type HostProject struct {
	Type           HostType
	Name           string
	Manifest       string
	PackageManager string
	Framework      string
}

// DisplayTitle returns a human-readable name for the host platform.
func (h *HostProject) DisplayTitle() string {
	if h == nil {
		return "None"
	}
	switch h.Type {
	case HostNodeJS:
		return "Node.js"
	case HostLaravel:
		return "Laravel"
	case HostPHP:
		return "PHP / Composer"
	default:
		return string(h.Type)
	}
}

// ColocatedProject describes a Krewire project discovered in a subdirectory.
type ColocatedProject struct {
	Name       string
	Path       string // Relative path from scan root, e.g. "api"
	Kind       string // e.g. "app", "service", "cli"
	ModulePath string
}

// DetectAllHosts inspects dir for all foreign ecosystem manifests (package.json, composer.json).
func DetectAllHosts(dir string) []*HostProject {
	var hosts []*HostProject

	// 1. Check Node.js (package.json)
	pkgJSONPath := filepath.Join(dir, ManifestPackageJSON)
	if data, err := os.ReadFile(pkgJSONPath); err == nil {
		if nodeHost := parsePackageJSON(dir, data); nodeHost != nil {
			hosts = append(hosts, nodeHost)
		}
	}

	// 2. Check Laravel / PHP (composer.json)
	composerPath := filepath.Join(dir, ManifestComposerJSON)
	if data, err := os.ReadFile(composerPath); err == nil {
		if phpHost := parseComposerJSON(dir, data); phpHost != nil {
			hosts = append(hosts, phpHost)
		}
	}

	return hosts
}

// DetectHost inspects dir for foreign ecosystem manifests (package.json, composer.json).
// If both exist, the more specialized or primary project is returned.
func DetectHost(dir string) *HostProject {
	hosts := DetectAllHosts(dir)
	if len(hosts) == 0 {
		return nil
	}
	for _, h := range hosts {
		if h.Type == HostLaravel {
			return h
		}
	}
	return hosts[0]
}

// parsePackageJSON parses package.json and infers package manager & framework.
func parsePackageJSON(dir string, data []byte) *HostProject {
	var pkg struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return &HostProject{
			Type:     HostNodeJS,
			Manifest: ManifestPackageJSON,
		}
	}

	host := &HostProject{
		Type:           HostNodeJS,
		Name:           pkg.Name,
		Manifest:       ManifestPackageJSON,
		PackageManager: detectNodePackageManager(dir),
		Framework:      detectNodeFramework(pkg.Dependencies, pkg.DevDependencies),
	}
	return host
}

func detectNodePackageManager(dir string) string {
	switch {
	case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
		return "pnpm"
	case fileExists(filepath.Join(dir, "yarn.lock")):
		return "yarn"
	case fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")):
		return "bun"
	case fileExists(filepath.Join(dir, "package-lock.json")):
		return "npm"
	default:
		return "npm"
	}
}

func detectNodeFramework(deps, devDeps map[string]string) string {
	all := make(map[string]bool)
	for k := range deps {
		all[k] = true
	}
	for k := range devDeps {
		all[k] = true
	}

	switch {
	case all["@sveltejs/kit"]:
		return "SvelteKit"
	case all["svelte"]:
		return "Svelte"
	case all["nuxt"]:
		return "Nuxt"
	case all["vue"] || all["@vue/cli-service"]:
		return "Vue"
	case all["next"]:
		return "Next.js"
	case all["@remix-run/react"] || all["@remix-run/node"]:
		return "Remix"
	case all["astro"]:
		return "Astro"
	case all["solid-start"]:
		return "SolidStart"
	case all["@nestjs/core"]:
		return "NestJS"
	case all["fastify"]:
		return "Fastify"
	case all["express"]:
		return "Express"
	default:
		return ""
	}
}

// parseComposerJSON parses composer.json and checks for Laravel or generic PHP.
func parseComposerJSON(dir string, data []byte) *HostProject {
	var comp struct {
		Name    string            `json:"name"`
		Require map[string]string `json:"require"`
	}
	_ = json.Unmarshal(data, &comp)

	isLaravel := fileExists(filepath.Join(dir, FileArtisan))
	if !isLaravel && comp.Require != nil {
		if _, hasLaravel := comp.Require["laravel/framework"]; hasLaravel {
			isLaravel = true
		}
	}

	ht := HostPHP
	framework := ""
	if isLaravel {
		ht = HostLaravel
		framework = "Laravel"
	}

	return &HostProject{
		Type:           ht,
		Name:           comp.Name,
		Manifest:       ManifestComposerJSON,
		PackageManager: "composer",
		Framework:      framework,
	}
}

var ignoredDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	".krewire":     true,
	"storage":      true,
	"dist":         true,
	"build":        true,
	".agents":      true,
	"internal":     false, // allowed if it contains a sub-module
}

// FindColocated scans immediate subdirectories of dir for Krewire projects.
func FindColocated(dir string) []ColocatedProject {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var results []ColocatedProject
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoredDirs[name] {
			continue
		}

		subDir := filepath.Join(dir, name)
		hasMod := fileExists(filepath.Join(subDir, gomod.GoModFileName))
		hasYaml := fileExists(filepath.Join(subDir, kiwconfig.ConfigFileName))

		if !hasMod && !hasYaml {
			continue
		}

		config, _ := kiwconfig.Load(subDir)
		res, _ := shape.Detect(subDir, config.Kind())
		mod, _ := gomod.Find(subDir)

		kindStr := res.String()
		if res.IsPreInit() {
			kindStr = shape.MarkerPreInit
		}

		modPath := ""
		if mod != nil {
			modPath = mod.Path
		}

		results = append(results, ColocatedProject{
			Name:       name,
			Path:       name + "/",
			Kind:       kindStr,
			ModulePath: modPath,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
