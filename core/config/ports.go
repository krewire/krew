package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// DetectedPort records a port declared by a specific runtime (FRK-CONF-030).
type DetectedPort struct {
	Runtime string // "krewire", "nodejs", "laravel", "env"
	Source  string // "krewire.yaml", "package.json", ".env", "composer.json"
	Port    int
	Detail  string
}

var (
	portFlagRegex = regexp.MustCompile(`(?:--port(?:=|\s+)|-p\s+|PORT=)(\d{2,5})`)
	envPortRegex  = regexp.MustCompile(`(?m)^(?:APP_PORT|SERVER_PORT|PORT|VITE_PORT)=(\d{2,5})`)
)

// DetectRuntimePorts scans the project root and krewire.yaml to discover declared
// ports across runtimes and flags potential conflicts (FRK-CONF-003, FRK-CONF-030).
func DetectRuntimePorts(root string, config *Config) ([]DetectedPort, []string) {
	var detected []DetectedPort

	detected = append(detected, detectKrewirePorts(config)...)
	detected = append(detected, detectNodePorts(root)...)
	detected = append(detected, detectLaravelPorts(root)...)
	detected = append(detected, detectDotenvPorts(root, detected)...)

	warnings := detectPortConflicts(detected)
	return detected, warnings
}

func detectKrewirePorts(config *Config) []DetectedPort {
	if config == nil {
		return nil
	}
	var ports []DetectedPort
	if config.Runtime.Port > 0 {
		ports = append(ports, DetectedPort{
			Runtime: "krewire",
			Source:  ConfigFileName,
			Port:    config.Runtime.Port,
			Detail:  "runtime.port",
		})
	}
	if config.Dev.Port > 0 && config.Dev.Port != config.Runtime.Port {
		ports = append(ports, DetectedPort{
			Runtime: "krewire",
			Source:  ConfigFileName,
			Port:    config.Dev.Port,
			Detail:  "dev.port",
		})
	}
	if config.Server.Port > 0 && config.Server.Port != config.Runtime.Port && config.Server.Port != config.Dev.Port {
		ports = append(ports, DetectedPort{
			Runtime: "krewire",
			Source:  ConfigFileName,
			Port:    config.Server.Port,
			Detail:  "server.port",
		})
	}
	for _, svc := range config.Services {
		if svc.Port > 0 {
			ports = append(ports, DetectedPort{
				Runtime: "krewire",
				Source:  ConfigFileName,
				Port:    svc.Port,
				Detail:  "service: " + svc.Name,
			})
		}
	}
	return ports
}

func detectNodePorts(root string) []DetectedPort {
	pkgPath := filepath.Join(root, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}
	var pkg struct {
		Port    any               `json:"port"`
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}
	var ports []DetectedPort
	if p, ok := pkg.Port.(float64); ok && p > 0 && p <= 65535 {
		ports = append(ports, DetectedPort{
			Runtime: "nodejs",
			Source:  "package.json",
			Port:    int(p),
			Detail:  "port config",
		})
	}
	for scriptName, cmd := range pkg.Scripts {
		matches := portFlagRegex.FindStringSubmatch(cmd)
		if len(matches) == 2 {
			if portNum, err := strconv.Atoi(matches[1]); err == nil && portNum > 0 && portNum <= 65535 {
				ports = append(ports, DetectedPort{
					Runtime: "nodejs",
					Source:  "package.json",
					Port:    portNum,
					Detail:  fmt.Sprintf("scripts.%s: %q", scriptName, cmd),
				})
			}
		}
	}
	return ports
}

func detectLaravelPorts(root string) []DetectedPort {
	composerPath := filepath.Join(root, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return nil
	}
	var comp struct {
		Scripts map[string]any `json:"scripts"`
	}
	if json.Unmarshal(data, &comp) != nil {
		return nil
	}
	var ports []DetectedPort
	for scriptName, val := range comp.Scripts {
		if cmd, ok := val.(string); ok {
			matches := portFlagRegex.FindStringSubmatch(cmd)
			if len(matches) == 2 {
				if portNum, err := strconv.Atoi(matches[1]); err == nil && portNum > 0 && portNum <= 65535 {
					ports = append(ports, DetectedPort{
						Runtime: "laravel",
						Source:  "composer.json",
						Port:    portNum,
						Detail:  fmt.Sprintf("scripts.%s", scriptName),
					})
				}
			}
		}
	}
	return ports
}

func detectDotenvPorts(root string, existing []DetectedPort) []DetectedPort {
	var ports []DetectedPort
	for _, envFileName := range []string{".env", ".env.example"} {
		envPath := filepath.Join(root, envFileName)
		data, err := os.ReadFile(envPath)
		if err != nil {
			continue
		}
		matches := envPortRegex.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			if len(m) == 2 {
				if portNum, err := strconv.Atoi(m[1]); err == nil && portNum > 0 && portNum <= 65535 {
					exists := false
					for _, d := range existing {
						if d.Port == portNum && d.Source == envFileName {
							exists = true
							break
						}
					}
					for _, d := range ports {
						if d.Port == portNum && d.Source == envFileName {
							exists = true
							break
						}
					}
					if !exists {
						ports = append(ports, DetectedPort{
							Runtime: "dotenv",
							Source:  envFileName,
							Port:    portNum,
							Detail:  m[0],
						})
					}
				}
			}
		}
	}
	return ports
}

func detectPortConflicts(detected []DetectedPort) []string {
	var warnings []string
	krewirePorts := make(map[int]DetectedPort)
	for _, d := range detected {
		if d.Runtime == "krewire" {
			krewirePorts[d.Port] = d
		}
	}

	for _, d := range detected {
		if d.Runtime != "krewire" {
			if kw, hasConflict := krewirePorts[d.Port]; hasConflict {
				warnings = append(warnings, fmt.Sprintf(
					"port conflict detected: port %d is declared by Krewire (%s %s) and %s (%s)",
					d.Port, kw.Source, kw.Detail, d.Runtime, d.Source,
				))
			}
		}
	}
	return warnings
}
