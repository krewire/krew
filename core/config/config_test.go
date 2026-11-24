package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/web/ssg"
	"github.com/krewire/ui/core"
)

func TestLoadMissingFile(t *testing.T) {
	c, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "" || c.Base != "" || c.Input != "" || c.Output != "" || c.Theme != nil || len(c.Nav) != 0 {
		t.Errorf("expected zero config, got %+v", *c)
	}
}

func TestLoadParsesConfig(t *testing.T) {
	dir := t.TempDir()
	body := `title: Krewire Documentation
author: Krewire Contributors
base: /docs/
input: docs
output: site
footer: Krewire Documentation — MIT License
nav:
  - text: Krewire
    url: /
theme:
  default: dark
  light:
    primary: "#7c3aed"
    accent: "#b45309"
  dark:
    primary: "#a78bfa"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Krewire Documentation" || c.Base != "/docs/" || c.Output != "site" {
		t.Errorf("parsed config = %+v", c)
	}
	if len(c.Nav) != 1 || c.Nav[0].URL != "/" {
		t.Errorf("nav = %+v", c.Nav)
	}
	if c.Theme.Default != "dark" || string(c.Theme.Light.Primary) != "#7c3aed" || c.Theme.Dark.Accent != "" {
		t.Errorf("theme = %+v", c.Theme)
	}
}

// A project may declare its version under project: instead of at the top
// level. Both must produce the same effective version, otherwise the ssg half
// stamps its cache-busting query while the book half does not, and a reader
// keeps a stale stylesheet on one page family and a fresh one on the other.
func TestLoadFoldsProjectVersionIntoVersion(t *testing.T) {
	dir := t.TempDir()
	body := `project:
  name: krewire
  version: "v0.1.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != "v0.1.0" {
		t.Errorf("Version = %q, want v0.1.0 from project.version", c.Version)
	}
}

// An explicit top-level version is the effective one: it is the documented
// location and must win over the nested declaration.
func TestLoadPrefersTopLevelVersion(t *testing.T) {
	dir := t.TempDir()
	body := `version: "2.0.0"
project:
  name: krewire
  version: "v0.1.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != "2.0.0" {
		t.Errorf("Version = %q, want the top-level 2.0.0", c.Version)
	}
}
func TestPaletteMerge(t *testing.T) {
	p := &Palette{Primary: "#123456", Accent: "#abcdef"}
	got := p.UI(ui.DefaultLightPalette)
	if got.Primary != "#123456" || got.Accent != "#abcdef" {
		t.Errorf("overrides not applied: %+v", got)
	}
	if got.Base1 != ui.DefaultLightPalette.Base1 {
		t.Errorf("unset tokens must fall back to defaults: %+v", got)
	}
	if (*Palette)(nil).UI(ui.DefaultLightPalette) != ui.DefaultLightPalette {
		t.Error("nil palette must return the defaults unchanged")
	}
}

func TestLoad_BuildDependsOnAndOutputs(t *testing.T) {
	dir := t.TempDir()
	body := `build:
  depends_on:
    - "npm run build"
    - "composer install --no-dev"
  outputs:
    nodejs: "dist"
    laravel: "public/build"
    krewire: ".krewire/build"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Build.DependsOn) != 2 || c.Build.DependsOn[0] != "npm run build" || c.Build.DependsOn[1] != "composer install --no-dev" {
		t.Errorf("Build.DependsOn = %v, want 2 commands", c.Build.DependsOn)
	}
	if len(c.Build.Outputs) != 3 || c.Build.Outputs["nodejs"] != "dist" || c.Build.Outputs["laravel"] != "public/build" {
		t.Errorf("Build.Outputs = %v, want 3 mapped outputs", c.Build.Outputs)
	}
}

func TestAPIEndpoints(t *testing.T) {
	c := &Config{}
	if got := c.APIURL(); got != DefaultAPIURL {
		t.Errorf("APIURL() = %q, want %q", got, DefaultAPIURL)
	}
	if got := c.AuthURL(); got != DefaultAuthAPIURL {
		t.Errorf("AuthURL() = %q, want %q", got, DefaultAuthAPIURL)
	}
	if got := c.CloudURL(); got != DefaultCloudAPIURL {
		t.Errorf("CloudURL() = %q, want %q", got, DefaultCloudAPIURL)
	}
	if got := c.DashURL(); got != DefaultDashAPIURL {
		t.Errorf("DashURL() = %q, want %q", got, DefaultDashAPIURL)
	}

	// Environment variable override
	t.Setenv("KREWIRE_API_URL", "https://api.staging.krewire.com/")
	if got := c.APIURL(); got != "https://api.staging.krewire.com" {
		t.Errorf("APIURL() with env override = %q, want %q", got, "https://api.staging.krewire.com")
	}
}

func TestConfig_SSGAndDeclaredPort(t *testing.T) {
	var nilCfg *Config
	if nilCfg.IsSSG() {
		t.Errorf("nilCfg.IsSSG() should be false")
	}
	if nilCfg.DeclaredPort() != 0 {
		t.Errorf("nilCfg.DeclaredPort() = %d, want 0", nilCfg.DeclaredPort())
	}

	c := &Config{}
	if c.IsSSG() {
		t.Errorf("c.IsSSG() should be false when SSG is nil")
	}
	if c.DeclaredPort() != 0 {
		t.Errorf("c.DeclaredPort() = %d, want 0", c.DeclaredPort())
	}

	c.Runtime.Port = 8081
	if c.DeclaredPort() != 8081 {
		t.Errorf("DeclaredPort() = %d, want 8081", c.DeclaredPort())
	}

	c.Runtime.Port = 0
	c.Dev.Port = 8082
	if c.DeclaredPort() != 8082 {
		t.Errorf("DeclaredPort() = %d, want 8082", c.DeclaredPort())
	}

	c.Dev.Port = 0
	c.Server.Port = 8083
	if c.DeclaredPort() != 8083 {
		t.Errorf("DeclaredPort() = %d, want 8083", c.DeclaredPort())
	}

	c.Server.Port = 0
	c.Services = []ServiceConfig{{Name: "svc", Port: 8084}}
	if c.DeclaredPort() != 8084 {
		t.Errorf("DeclaredPort() = %d, want 8084", c.DeclaredPort())
	}

	// SSG config conversion
	bTrue := true
	c.SSG = &SSGConfig{
		Description: "test site",
		AutoAssets:  &ssg.AutoAssetConfig{Enabled: &bTrue},
	}
	if !c.IsSSG() {
		t.Errorf("c.IsSSG() should be true")
	}
	ssgCfg := c.ToSSGConfig()
	if ssgCfg.Description != "test site" || ssgCfg.AutoAssets == nil || !*ssgCfg.AutoAssets.Enabled {
		t.Errorf("unexpected ssgCfg: %+v", ssgCfg)
	}
}
