package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads krewire.yaml from dir. A missing file yields a zero Config without
// an error.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	// Fold project.name or title into top-level Name
	if c.Name == "" {
		c.Name = c.Project.Name
	}
	if c.Name == "" {
		c.Name = c.Title
	}
	// A project declares its version either at the top level or next to the
	// kind. Fold the latter into the single effective value so both pipelines
	// (ssg page data and the book asset cache-busting query) see the same
	// version instead of one silently losing it.
	if c.Version == "" {
		c.Version = c.Project.Version
	}
	// The URL base lives under `build:` in current krewire.yaml files
	// (e.g. krewire.com `build.base: "/"`, portal `build.base:
	// "/contribute/"). Fold the nested form into the single effective
	// value so tooling never silently ignores what the repo declares.
	// Precedence: krew --base flag > build.base > legacy top-level base:.
	if c.Build.Base != "" {
		c.Base = c.Build.Base
	}
	if c.Output == "" && c.Build.Output != "" {
		c.Output = c.Build.Output
	}
	if c.Output == "" {
		c.Output = DefaultOutput
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}
