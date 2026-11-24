package ecosystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectHost_NodeJS(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{
		"name": "my-svelte-app",
		"dependencies": {
			"@sveltejs/kit": "^2.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, ManifestPackageJSON), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	host := DetectHost(dir)
	if host == nil {
		t.Fatal("expected host to be detected")
	}
	if host.Type != HostNodeJS {
		t.Errorf("expected type %q, got %q", HostNodeJS, host.Type)
	}
	if host.Name != "my-svelte-app" {
		t.Errorf("expected name %q, got %q", "my-svelte-app", host.Name)
	}
	if host.PackageManager != "pnpm" {
		t.Errorf("expected pnpm, got %q", host.PackageManager)
	}
	if host.Framework != "SvelteKit" {
		t.Errorf("expected SvelteKit, got %q", host.Framework)
	}
	if host.DisplayTitle() != "Node.js" {
		t.Errorf("expected Node.js, got %q", host.DisplayTitle())
	}
}

func TestDetectHost_Laravel(t *testing.T) {
	dir := t.TempDir()

	compJSON := `{
		"name": "laravel/laravel",
		"require": {
			"php": "^8.2",
			"laravel/framework": "^11.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, ManifestComposerJSON), []byte(compJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, FileArtisan), []byte("#!/usr/bin/env php\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	host := DetectHost(dir)
	if host == nil {
		t.Fatal("expected host to be detected")
	}
	if host.Type != HostLaravel {
		t.Errorf("expected type %q, got %q", HostLaravel, host.Type)
	}
	if host.Framework != "Laravel" {
		t.Errorf("expected framework Laravel, got %q", host.Framework)
	}
	if host.PackageManager != "composer" {
		t.Errorf("expected composer, got %q", host.PackageManager)
	}
	if host.DisplayTitle() != "Laravel" {
		t.Errorf("expected Laravel, got %q", host.DisplayTitle())
	}
}

func TestFindColocated(t *testing.T) {
	dir := t.TempDir()

	// 1. Host project files
	if err := os.WriteFile(filepath.Join(dir, ManifestPackageJSON), []byte(`{"name":"host"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Co-located Krewire service in api/
	apiDir := filepath.Join(dir, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte("module myapp/api\n\ngo 1.27.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 3. node_modules should be ignored
	nmDir := filepath.Join(dir, "node_modules", "somepkg")
	if err := os.MkdirAll(nmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nmDir, "go.mod"), []byte("module fake\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	colocated := FindColocated(dir)
	if len(colocated) != 1 {
		t.Fatalf("expected 1 colocated project, got %d", len(colocated))
	}
	if colocated[0].Name != "api" {
		t.Errorf("expected name api, got %q", colocated[0].Name)
	}
	if colocated[0].Path != "api/" {
		t.Errorf("expected path api/, got %q", colocated[0].Path)
	}
	if colocated[0].ModulePath != "myapp/api" {
		t.Errorf("expected module myapp/api, got %q", colocated[0].ModulePath)
	}
}

func TestDetectHost_SvelteAndVue(t *testing.T) {
	// 1. Svelte (Vite + Svelte)
	dirSvelte := t.TempDir()
	pkgSvelte := `{"name": "svelte-app", "dependencies": {"svelte": "^5.0.0"}}`
	_ = os.WriteFile(filepath.Join(dirSvelte, ManifestPackageJSON), []byte(pkgSvelte), 0o644)
	hostSvelte := DetectHost(dirSvelte)
	if hostSvelte == nil || hostSvelte.Framework != "Svelte" {
		t.Errorf("expected Svelte framework, got %+v", hostSvelte)
	}

	// 2. Vue (Vite + Vue)
	dirVue := t.TempDir()
	pkgVue := `{"name": "vue-app", "dependencies": {"vue": "^3.5.0"}}`
	_ = os.WriteFile(filepath.Join(dirVue, ManifestPackageJSON), []byte(pkgVue), 0o644)
	hostVue := DetectHost(dirVue)
	if hostVue == nil || hostVue.Framework != "Vue" {
		t.Errorf("expected Vue framework, got %+v", hostVue)
	}
}

func TestEcosystem_ExtraCoverage(t *testing.T) {
	var nilHost *HostProject
	if nilHost.DisplayTitle() != "None" {
		t.Errorf("nilHost.DisplayTitle() = %s, want None", nilHost.DisplayTitle())
	}

	phpHost := &HostProject{Type: HostPHP}
	if phpHost.DisplayTitle() != "PHP / Composer" {
		t.Errorf("phpHost.DisplayTitle() = %s, want PHP / Composer", phpHost.DisplayTitle())
	}

	customHost := &HostProject{Type: HostType("custom")}
	if customHost.DisplayTitle() != "custom" {
		t.Errorf("customHost.DisplayTitle() = %s, want custom", customHost.DisplayTitle())
	}

	// Package manager detection
	dirYarn := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirYarn, "yarn.lock"), []byte(""), 0o644)
	if detectNodePackageManager(dirYarn) != "yarn" {
		t.Errorf("expected yarn")
	}

	dirBun := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirBun, "bun.lock"), []byte(""), 0o644)
	if detectNodePackageManager(dirBun) != "bun" {
		t.Errorf("expected bun")
	}

	dirNpm := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirNpm, "package-lock.json"), []byte(""), 0o644)
	if detectNodePackageManager(dirNpm) != "npm" {
		t.Errorf("expected npm")
	}

	// Framework detection
	frameworkTests := []struct {
		dep  string
		want string
	}{
		{"nuxt", "Nuxt"},
		{"next", "Next.js"},
		{"@remix-run/react", "Remix"},
		{"@remix-run/node", "Remix"},
		{"@vue/cli-service", "Vue"},
		{"astro", "Astro"},
		{"solid-start", "SolidStart"},
		{"@nestjs/core", "NestJS"},
		{"fastify", "Fastify"},
		{"express", "Express"},
		{"unknown-lib", ""},
	}
	for _, ft := range frameworkTests {
		got := detectNodeFramework(map[string]string{ft.dep: "1.0.0"}, nil)
		if got != ft.want {
			t.Errorf("detectNodeFramework(%s) = %s, want %s", ft.dep, got, ft.want)
		}
	}

	// Plain PHP (composer.json without laravel)
	dirPHP := t.TempDir()
	compJSON := `{"name": "vendor/package", "require": {"php": "^8.2"}}`
	_ = os.WriteFile(filepath.Join(dirPHP, ManifestComposerJSON), []byte(compJSON), 0o644)
	hostPHP := DetectHost(dirPHP)
	if hostPHP == nil || hostPHP.Type != HostPHP {
		t.Errorf("expected HostPHP, got %+v", hostPHP)
	}

	// Bad composer.json returns fallback
	dirBadPHP := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirBadPHP, ManifestComposerJSON), []byte("{bad"), 0o644)
	if h := DetectHost(dirBadPHP); h == nil || h.Type != HostPHP {
		t.Errorf("expected HostPHP fallback for bad composer.json")
	}

	// Invalid package.json and composer.json fallback handling
	dirInvalid := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirInvalid, ManifestPackageJSON), []byte("{bad"), 0o644)
	if h := DetectHost(dirInvalid); h == nil || h.Type != HostNodeJS {
		t.Errorf("expected HostNodeJS fallback for invalid package.json")
	}

	dirEmpty := t.TempDir()
	if h := DetectHost(dirEmpty); h != nil {
		t.Errorf("expected nil host when no manifests exist")
	}
}
