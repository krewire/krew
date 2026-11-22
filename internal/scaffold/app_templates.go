package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/buildinfo"
)

func detectLocalModulePath(startDir string) string {
	dir := startDir
	for {
		modPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modPath); err == nil {
			if strings.Contains(string(data), "module "+buildinfo.ModKrewire) {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func goModTemplate(module, version, localPath string) string {
	mod := fmt.Sprintf("module %s\n\ngo %s\n\nrequire (\n\t%s %s\n)",
		module, GoVersion, buildinfo.ModKrewire, pinnedVersion(version))
	if localPath != "" {
		mod += fmt.Sprintf("\n\nreplace %s => %s", buildinfo.ModKrewire, localPath)
	}
	return mod + "\n"
}

// pinnedVersion normalizes a version into the v-prefixed form that go.mod
// require directives mandate ("0.1.0" -> "v0.1.0").
func pinnedVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

func krewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  version: 0.1.0

dev:
  port: 8080
`, name)
}

func dotenvExampleTemplate(name string) string {
	return fmt.Sprintf(`# %s environment configuration (KWN-Q7X4M KWL-CFGV-005)
APP_NAME=%s
APP_ENV=local
APP_DEBUG=true
APP_PORT=8080
`, name, name)
}

func cmdMainTemplate(module string) string {
	return fmt.Sprintf(`// Thin entry point: load config, build app, run.
package main

import (
	"fmt"
	"os"

	"%s/internal/app"
	"%s/internal/config"
)

// Signature is the ecosystem watermark embedded into the binary at compile time via -ldflags.
var Signature = "build with krewire"

func main() {
	meta, err := config.LoadMetadata("krewire.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	a, err := app.New(meta)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := a.Run(meta.Addr()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, module, module)
}

func readmeTemplate(name string) string {
	return fmt.Sprintf(`# %s

A Krewire project scaffolded by the [krew CLI](https://github.com/krewire/krew) (binary `+"`krew`"+`).

Built on the [Krewire Ecosystem](https://github.com/krewire/krewire).

## Project Structure

This is a Krewire fullstack monolith:

	main.go       # thin entry point
	internal/     # app assembly, config, HTTP handlers
	web/          # UI sources: layouts, pages, theme
	public/       # static assets (embedded)

## Getting Started

	go run .

## Testing

	go test ./...
`, name)
}
