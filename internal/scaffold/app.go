// Package-level file for the fullstack monolith (VariantApp): the app
// template set and equip wiring (Domain: app). krew init dispatches here
// whenever no variant flag is given. Templates are split per bounded context:
// - app_templates.go: go.mod, krewire.yaml, main.go, README, local paths
// - app_internal.go: internal/app/config/http
// - app_web.go: web layouts/pages/theme
// - app_assets.go: assets/embed, public css/js
package scaffold

import (
	"fmt"
	"os"
	"os/exec"
)

// equipApp shapes the bootstrap project into a fullstack monolith. The root main.go
// becomes the entry point (canonical layout KWF-CCI0N) and go.mod is
// upgraded in place with the pinned ecosystem module require.
func equipApp(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip app: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip app: project name is required")
	}
	version := opts.Version
	if version == "" {
		version = "latest"
	}

	cwd, _ := os.Getwd()
	localPath := detectLocalModulePath(cwd)

	files := []file{
		{goModFile, goModTemplate(opts.Module, version, localPath)},
		{krewireYaml, krewireYamlTemplate(opts.Name)},
		{mainGo, cmdMainTemplate(opts.Module)},
		{"internal/app/app.go", internalAppTemplate(opts.Module)},
		{"internal/config/config.go", internalConfigTemplate()},
		{"internal/http/http.go", internalHttpTemplate()},
		{"web/layouts/shell.go", webLayoutsShellTemplate()},
		{"web/pages/pages.go", webPagesTemplate()},
		{"web/components/components.go", webComponentsTemplate()},
		{"web/theme/theme.go", webThemeTemplate()},
		{"assets/embed.go", assetsEmbedTemplate()},
		{"assets/public/app.css", publicCssTemplate()},
		{"assets/public/app.js", publicJsTemplate()},
		{"README.md", readmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
		{".env.example", dotenvExampleTemplate(opts.Name)},
		{".env", dotenvExampleTemplate(opts.Name)},
	}
	report, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = opts.Dir
	_ = cmd.Run()
	return report, nil
}
