package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/hub/resolver"
	"github.com/krewire/hub/svelte"
	"github.com/krewire/hub/tailwind"
	"github.com/krewire/hub/vue"
	"github.com/krewire/krewire/errs"
)

func TestRunAdd_EmptyArgsReturnsUsage(t *testing.T) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	RegisterAdd(fs)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("fs.Parse: %v", err)
	}
	code := RunAdd(fs)
	if code != errs.ExitCodeUsage {
		t.Errorf("RunAdd(empty) = %v, want ExitCodeUsage", code)
	}
}

func TestRunRemove_EmptyArgsReturnsUsage(t *testing.T) {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	RegisterRemove(fs)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("fs.Parse: %v", err)
	}
	code := RunRemove(fs)
	if code != errs.ExitCodeUsage {
		t.Errorf("RunRemove(empty) = %v, want ExitCodeUsage", code)
	}
}

func TestRunAdd_UnsafePackageReturnsUsage(t *testing.T) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	RegisterAdd(fs)
	if err := fs.Parse([]string{"file://../malicious"}); err != nil {
		t.Fatalf("fs.Parse: %v", err)
	}
	code := RunAdd(fs)
	if code != errs.ExitCodeUsage {
		t.Errorf("RunAdd(unsafe) = %v, want ExitCodeUsage", code)
	}
}

func TestTailwindPlugin_AddAndRemove(t *testing.T) {
	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	tw := &tailwind.Tailwind{}
	if err := tw.Add(dir, "latest"); err != nil {
		t.Logf("tw.Add note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tailwind.config.js")); err != nil {
		t.Errorf("expected tailwind.config.js to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "assets", "tailwind.css")); err != nil {
		t.Errorf("expected assets/tailwind.css to exist: %v", err)
	}

	if err := tw.Remove(dir); err != nil {
		t.Logf("tw.Remove note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tailwind.config.js")); !os.IsNotExist(err) {
		t.Errorf("expected tailwind.config.js to be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "assets", "tailwind.css")); !os.IsNotExist(err) {
		t.Errorf("expected unmodified assets/tailwind.css to be removed")
	}
}

func TestSveltePlugin_AddAndRemove(t *testing.T) {
	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	sv := &svelte.Svelte{}
	if err := sv.Add(dir, "latest"); err != nil {
		t.Logf("sv.Add note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "svelte.config.js")); err != nil {
		t.Errorf("expected svelte.config.js to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Counter.svelte")); err != nil {
		t.Errorf("expected components/Counter.svelte to exist: %v", err)
	}

	if err := sv.Remove(dir); err != nil {
		t.Logf("sv.Remove note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "svelte.config.js")); !os.IsNotExist(err) {
		t.Errorf("expected svelte.config.js to be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Counter.svelte")); !os.IsNotExist(err) {
		t.Errorf("expected components/Counter.svelte to be removed")
	}
}

func TestVuePlugin_AddAndRemove(t *testing.T) {
	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	vue := &vue.Vue{}
	if err := vue.Add(dir, "latest"); err != nil {
		t.Logf("vue.Add note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "vite.config.js")); err != nil {
		t.Errorf("expected vite.config.js to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Counter.vue")); err != nil {
		t.Errorf("expected components/Counter.vue to exist: %v", err)
	}

	if err := vue.Remove(dir); err != nil {
		t.Logf("vue.Remove note: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "components", "Counter.vue")); !os.IsNotExist(err) {
		t.Errorf("expected components/Counter.vue to be removed")
	}
}

func TestResolverChainOrder(t *testing.T) {
	chain := resolver.DefaultChain()

	// 1. Plugin alias matches plugin
	res, err := chain.Resolve(resolver.Spec{Name: "twcss", Raw: "twcss"})
	if err != nil {
		t.Fatalf("resolve twcss: %v", err)
	}
	if res.Kind != "plugin" {
		t.Errorf("twcss kind = %q, want plugin", res.Kind)
	}

	// 2. Go module path matches gomod
	resGo, err := chain.Resolve(resolver.Spec{Name: "github.com/foo/bar", Raw: "github.com/foo/bar"})
	if err != nil {
		t.Fatalf("resolve go module: %v", err)
	}
	if resGo.Kind != "gomod" {
		t.Errorf("go module kind = %q, want gomod", resGo.Kind)
	}

	// 3. Npm package fallback
	resNpm, err := chain.Resolve(resolver.Spec{Name: "lodash", Raw: "lodash"})
	if err != nil {
		t.Fatalf("resolve npm: %v", err)
	}
	if resNpm.Kind != "npm" {
		t.Errorf("npm package kind = %q, want npm", resNpm.Kind)
	}

	// 4. Krewire UI shortcut resolves to gomod (github.com/krewire/ui)
	resUI, err := chain.Resolve(resolver.Spec{Name: "@krewire/ui", Raw: "@krewire/ui"})
	if err != nil {
		t.Fatalf("resolve @krewire/ui: %v", err)
	}
	if resUI.Kind != "gomod" {
		t.Errorf("@krewire/ui kind = %q, want gomod", resUI.Kind)
	}
	if resUI.Installer.Describe() != "go:github.com/krewire/ui" {
		t.Errorf("@krewire/ui describe = %q, want go:github.com/krewire/ui", resUI.Installer.Describe())
	}
}
