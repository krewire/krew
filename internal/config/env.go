package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krewire/pkg/config"
	"github.com/krewire/krewire/pkg/env"
)

// ResolveEnv resolves the effective environment with strict precedence:
// override (--env flag) > KIW_ENV > krewire.yaml > local
// (KWL-K4T7W KWL-ENVV-004). An invalid resolved value is a usage error.
func (c *Config) ResolveEnv(override string) (env.Env, error) {
	v := firstNonEmpty(override, os.Getenv("KIW_ENV"))
	if v == "" && c != nil {
		v = c.Env
	}
	return env.ParseEnv(v)
}

// ResolveDebug resolves the effective debug switch with the same precedence:
// an explicitly set override wins, then KIW_DEBUG ("1"/"true"/"yes"), then
// krewire.yaml, then false.
func (c *Config) ResolveDebug(override string, overrideSet bool) bool {
	if overrideSet {
		return isTruthy(override)
	}
	switch strings.ToLower(os.Getenv("KIW_DEBUG")) {
	case "1", "true", "yes":
		return true
	}
	return c != nil && c.Debug
}

// LoadDotEnv parses the project environment file into the process environment
// without overwriting existing variables (KWN-Q7X4M KWL-CFGV-006).
// It honors runtime.env_file (FRK-CONF-012), warns if target env file is missing while
// .env.example exists (FRK-CONF-013), and handles runtime.namespace (FRK-CONF-001, FRK-CONF-002).
func (c *Config) LoadDotEnv(root string) error {
	envFile := c.EffectiveEnvFile()
	targetPath := envFile
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, envFile)
	}

	// FRK-CONF-013: Warn if target env file does not exist but .env.example does
	if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
		examplePath := filepath.Join(root, ".env.example")
		if _, exErr := os.Stat(examplePath); exErr == nil {
			fmt.Fprintf(os.Stderr, "warning: %s not found, but .env.example exists; copy .env.example to %s to configure your environment\n", envFile, envFile)
		}
		return nil
	}

	if err := config.LoadDotEnv(targetPath); err != nil {
		return err
	}

	// FRK-CONF-001 & FRK-CONF-002: Handle runtime.namespace prefixing
	if c != nil && c.Runtime.Namespace != "" {
		syncNamespaceEnv(c.Runtime.Namespace)
	}

	return nil
}

func syncNamespaceEnv(ns string) {
	prefix := strings.ToUpper(strings.TrimSpace(ns)) + "_"
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := parts[0], parts[1]
		if strings.HasPrefix(k, prefix) {
			unprefixed := strings.TrimPrefix(k, prefix)
			if _, exists := os.LookupEnv(unprefixed); !exists {
				_ = os.Setenv(unprefixed, v)
			}
		} else {
			prefixed := prefix + k
			if _, exists := os.LookupEnv(prefixed); !exists {
				_ = os.Setenv(prefixed, v)
			}
		}
	}
}

// FirstNonEmpty returns the first non-empty string among the arguments (SSOT).
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string { return FirstNonEmpty(vals...) }

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
