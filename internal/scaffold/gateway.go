// Package-level file for the API gateway application (VariantGateway).
package scaffold

import (
	"fmt"
	"os"
	"os/exec"
)

// equipGateway shapes the bootstrap project into an API gateway that proxies
// requests to upstream services discovered through the service registry.
func equipGateway(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip gateway: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip gateway: project name is required")
	}
	version := opts.Version
	if version == "" {
		version = "latest"
	}

	cwd, _ := os.Getwd()
	localPath := detectLocalModulePath(cwd)

	files := []file{
		{goModFile, goModTemplate(opts.Module, version, localPath)},
		{krewireYaml, gatewayKrewireYamlTemplate(opts.Name)},
		{mainGo, gatewayMainTemplate(opts.Module)},
		{"internal/config/config.go", gatewayConfigTemplate()},
		{"README.md", gatewayReadmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
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

func gatewayKrewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  version: 0.1.0

gateway:
  host: ""
  port: 8080
  routes:
    - path: /api
      prefix: true
      strip_prefix: true
      target_url: http://localhost:8081
      timeout: 10s
`, name)
}

func gatewayMainTemplate(module string) string {
	return fmt.Sprintf(`// Entry point: run the API gateway.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krewire/krewire/pkg/cloud/service/registry"
	"github.com/krewire/krewire/pkg/web/gateway"
	"%s/internal/config"
)

func main() {
	gatewayConfig, err := config.Load("krewire.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gateway: load config: %%v\n", err)
		os.Exit(1)
	}

	routes, err := gatewayConfig.Routes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gateway: %%v\n", err)
		os.Exit(1)
	}

	gw := gateway.New(
		gateway.WithRegistry(registry.NewMemoryRegistry()),
		gateway.WithMiddleware("cors", gateway.CORSMiddleware("*")),
	)
	for _, route := range routes {
		gw.AddRoute(route)
	}

	server := &http.Server{
		Addr:              gatewayConfig.Addr(),
		Handler:           gw,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("gateway listening", "addr", server.Addr, "routes", len(routes))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "gateway error: %%v\n", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	slog.Info("gateway stopped")
}
`, module)
}

func gatewayConfigTemplate() string {
	return `// Package config loads the gateway route table from krewire.yaml.
package config

import (
	"fmt"
	"os"
	"time"

	rconfig "github.com/krewire/krewire/pkg/config"
	"github.com/krewire/krewire/pkg/web/gateway"
)

// Config mirrors the krewire.yaml sections the gateway needs.
type Config struct {
	Project Project ` + "`" + `yaml:"project"` + "`" + `
	Gateway Gateway ` + "`" + `yaml:"gateway"` + "`" + `
}

// Project holds the project section of krewire.yaml.
type Project struct {
	Name    string ` + "`" + `yaml:"name"` + "`" + `
	Kind    string ` + "`" + `yaml:"kind"` + "`" + `
	Version string ` + "`" + `yaml:"version"` + "`" + `
}

// Gateway holds the gateway section of krewire.yaml.
type Gateway struct {
	Host   string  ` + "`" + `yaml:"host"` + "`" + `
	Port   int     ` + "`" + `yaml:"port"` + "`" + `
	Routes []Route ` + "`" + `yaml:"routes"` + "`" + `
}

// Route is one entry of the gateway route table.
type Route struct {
	Path        string   ` + "`" + `yaml:"path"` + "`" + `
	Method      string   ` + "`" + `yaml:"method"` + "`" + `
	Service     string   ` + "`" + `yaml:"service"` + "`" + `
	TargetURL   string   ` + "`" + `yaml:"target_url"` + "`" + `
	Prefix      bool     ` + "`" + `yaml:"prefix"` + "`" + `
	StripPrefix bool     ` + "`" + `yaml:"strip_prefix"` + "`" + `
	Middleware  []string ` + "`" + `yaml:"middleware"` + "`" + `
	RateLimit   int      ` + "`" + `yaml:"rate_limit"` + "`" + `
	Timeout     string   ` + "`" + `yaml:"timeout"` + "`" + `
}

// Load reads krewire.yaml at path into a Config.
func Load(path string) (*Config, error) {
	var config Config
	if err := rconfig.Load(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Addr returns the listen address, honoring GATEWAY_ADDR over krewire.yaml.
func (c *Config) Addr() string {
	if addr := os.Getenv("GATEWAY_ADDR"); addr != "" {
		return addr
	}
	port := c.Gateway.Port
	if port == 0 {
		port = 8080
	}
	if c.Gateway.Host == "" {
		return fmt.Sprintf(":%d", port)
	}
	return fmt.Sprintf("%s:%d", c.Gateway.Host, port)
}

// Routes converts the YAML route table into gateway routes.
func (c *Config) Routes() ([]gateway.Route, error) {
	routes := make([]gateway.Route, 0, len(c.Gateway.Routes))
	for _, r := range c.Gateway.Routes {
		var timeout time.Duration
		if r.Timeout != "" {
			d, err := time.ParseDuration(r.Timeout)
			if err != nil {
				return nil, fmt.Errorf("route %s: invalid timeout %q: %w", r.Path, r.Timeout, err)
			}
			timeout = d
		}
		routes = append(routes, gateway.Route{
			Path:        r.Path,
			Method:      r.Method,
			Service:     r.Service,
			TargetURL:   r.TargetURL,
			Prefix:      r.Prefix,
			StripPrefix: r.StripPrefix,
			Middleware:  r.Middleware,
			RateLimit:   r.RateLimit,
			Timeout:     timeout,
		})
	}
	return routes, nil
}
`
}

func gatewayReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s (API Gateway)

API gateway powered by Krewire (`+"`pkg/web/gateway`"+`). It proxies requests
along the route table declared under the `+"`gateway:`"+` key in
`+"`krewire.yaml`"+` and applies middleware such as CORS, rate limiting, and
tracing.

## Development

Run the gateway locally:

`+"```bash"+`
krew run
`+"```"+`

Run with hot reload:

`+"```bash"+`
krew dev
`+"```"+`

## Building & Deploying

Compile into a production binary:

`+"```bash"+`
krew build
`+"```"+`

Stage deployable artifacts:

`+"```bash"+`
krew deploy
`+"```"+`
`, name)
}
