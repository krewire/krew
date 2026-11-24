package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var validNamespaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Validate checks the configuration for semantic and structural validity
// (FRK-CONF-032, FRK-CONF-041).
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}

	// FRK-CONF-041: validate runtime.namespace
	if c.Runtime.Namespace != "" {
		if !validNamespaceRegex.MatchString(c.Runtime.Namespace) {
			return fmt.Errorf("config: invalid runtime.namespace %q: must contain only letters, numbers, and underscores", c.Runtime.Namespace)
		}
	}

	// FRK-CONF-041: validate runtime.env_file
	if c.Runtime.EnvFile != "" {
		clean := filepath.Clean(c.Runtime.EnvFile)
		if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
			return fmt.Errorf("config: invalid runtime.env_file %q: must be a relative path within project root", c.Runtime.EnvFile)
		}
	}

	// Validate ports
	if c.Runtime.Port < 0 || c.Runtime.Port > 65535 {
		return fmt.Errorf("config: invalid runtime.port %d: must be between 1 and 65535", c.Runtime.Port)
	}
	if c.Dev.Port < 0 || c.Dev.Port > 65535 {
		return fmt.Errorf("config: invalid dev.port %d: must be between 1 and 65535", c.Dev.Port)
	}
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("config: invalid server.port %d: must be between 1 and 65535", c.Server.Port)
	}

	// FRK-CONF-032: multiple services must declare unique ports
	if len(c.Services) > 1 {
		seenPorts := make(map[int]string)
		for _, svc := range c.Services {
			if svc.Port <= 0 {
				continue
			}
			if svc.Port > 65535 {
				return fmt.Errorf("config: invalid service %q port %d: must be between 1 and 65535", svc.Name, svc.Port)
			}
			if existing, seen := seenPorts[svc.Port]; seen {
				return fmt.Errorf("config: port conflict between services %q and %q: both declare port %d", existing, svc.Name, svc.Port)
			}
			seenPorts[svc.Port] = svc.Name
		}
	}

	return nil
}
