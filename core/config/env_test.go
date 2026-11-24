package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/env"
	"github.com/krewire/krewire/errs"
)

func writeConfig(t *testing.T, body string) *Config {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLoadParsesEnvAndDebug(t *testing.T) {
	c := writeConfig(t, "env: production\ndebug: true\n")
	if c.Env != "production" || !c.Debug {
		t.Errorf("config = (env %q, debug %v), want (production, true)", c.Env, c.Debug)
	}
	gotEnv, err := c.ResolveEnv("")
	if err != nil || gotEnv != env.EnvProduction {
		t.Errorf("ResolveEnv = (%q, %v), want (production, nil)", gotEnv, err)
	}
}

func TestEmptyConfigResolvesDefaults(t *testing.T) {
	c := writeConfig(t, "title: x\n")
	envVal, err := c.ResolveEnv("")
	if err != nil || envVal != env.DefaultEnv {
		t.Errorf("ResolveEnv = (%q, %v), want (%q, nil)", envVal, err, env.DefaultEnv)
	}
	if c.ResolveDebug("", false) {
		t.Error("ResolveDebug = true, want default false")
	}
}

func TestResolveEnv_Precedence(t *testing.T) {
	yamlTesting := writeConfig(t, "env: testing\n")

	t.Setenv("KIW_ENV", "production")
	if envVal, _ := yamlTesting.ResolveEnv(""); envVal != env.EnvProduction {
		t.Errorf("KIW_ENV should beat krewire.yaml, got %q", envVal)
	}
	if envVal, _ := yamlTesting.ResolveEnv("local"); envVal != env.EnvLocal {
		t.Errorf("flag should beat KIW_ENV and yaml, got %q", envVal)
	}

	t.Setenv("KIW_ENV", "")
	if envVal, _ := yamlTesting.ResolveEnv(""); envVal != env.EnvTesting {
		t.Errorf("yaml should win when no overrides, got %q", envVal)
	}
}

func TestResolveEnv_InvalidYamlValueIsUsageError(t *testing.T) {
	c := writeConfig(t, "env: staging\n")
	_, err := c.ResolveEnv("")
	if err == nil {
		t.Fatal("ResolveEnv(staging) = nil error, want usage error")
	}
	var ce interface{ ExitCode() errs.ExitCode }
	if !errors.As(err, &ce) || ce.ExitCode() != errs.ExitCodeUsage {
		t.Errorf("ResolveEnv(staging) error = %v, want usage exit code", err)
	}
}

func TestResolveDebug_Precedence(t *testing.T) {
	yamlOn := writeConfig(t, "debug: true\n")
	yamlOff := writeConfig(t, "title: x\n")

	if !yamlOn.ResolveDebug("", false) {
		t.Error("yaml debug: true should resolve true")
	}
	if yamlOff.ResolveDebug("", false) {
		t.Error("default should resolve false")
	}

	t.Setenv("KIW_DEBUG", "1")
	if !yamlOff.ResolveDebug("", false) {
		t.Error("KIW_DEBUG=1 should beat yaml default false")
	}
	if yamlOn.ResolveDebug("false", true) {
		t.Error("explicit --debug=false should beat KIW_DEBUG and krewire.yaml")
	}
}

func TestLoadDotEnv_CustomFileAndNamespace(t *testing.T) {
	dir := t.TempDir()
	envContent := "PORT=9090\nAPP_NAME=test-app\n"
	if err := os.WriteFile(filepath.Join(dir, ".env.custom"), []byte(envContent), 0o644); err != nil {
		t.Fatal(err)
	}

	config := &Config{
		Runtime: RuntimeConfig{
			Namespace: "krewire",
			EnvFile:   ".env.custom",
		},
	}

	if err := config.LoadDotEnv(dir); err != nil {
		t.Fatalf("LoadDotEnv failed: %v", err)
	}

	// Unprefixed should be accessible
	if v := os.Getenv("PORT"); v != "9090" {
		t.Errorf("os.Getenv(PORT) = %q, want 9090", v)
	}
	// Prefixed by namespace should be synchronized
	if v := os.Getenv("KREWIRE_PORT"); v != "9090" {
		t.Errorf("os.Getenv(KREWIRE_PORT) = %q, want 9090", v)
	}
}

func TestLoadDotEnv_ExampleWarning(t *testing.T) {
	dir := t.TempDir()
	// Only .env.example exists, .env does not
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	config := &Config{}
	if err := config.LoadDotEnv(dir); err != nil {
		t.Fatalf("LoadDotEnv should not fail when only .env.example exists: %v", err)
	}
}
