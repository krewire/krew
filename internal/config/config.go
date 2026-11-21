// Package config reads project configuration from krewire.yaml so that `kiw`
// commands can drive any project kind without a project-specific cmd.
// A krewire.yaml with an `ssg:` key triggers SSG mode; one with a content/
// directory triggers book mode. All top-level fields are shared across modes.
package config

import (
	"os"
	"strings"

	"github.com/krewire/krewire/pkg/cloud/infra/provider"
	"github.com/krewire/krewire/pkg/model"
	"github.com/krewire/krewire/pkg/web/ssg"
)

// ConfigFileName is the canonical file name for project configuration (SSOT).
const ConfigFileName = model.ConfigFileName

const DefaultOutput = ".krewire/build"

// DefaultAPIURL is the canonical endpoint for the unified Krewire API Gateway.
const DefaultAPIURL = "https://api.krewire.com"

// DefaultAuthAPIURL is the endpoint for central identity and auth routing.
const DefaultAuthAPIURL = "https://api.krewire.com/auth"

// DefaultCloudAPIURL is the endpoint for hybrid PaaS and workload dispatch.
const DefaultCloudAPIURL = "https://api.krewire.com/cloud"

// DefaultDashAPIURL is the endpoint for virtual workspaces and dashboard sessions.
const DefaultDashAPIURL = "https://api.krewire.com/dash"

// DefaultInput is the default content directory consumed by the book
// build, mirroring the ssg content layout.
const DefaultInput = "content"

// Config is the krewire.yaml schema consumed by the build command.
type Config struct {
	// Name holds the project name when declared at top level.
	Name string `yaml:"name"`
	// Project pins the project kind when discovery is ambiguous. Empty means
	// auto-detection (app > site > book precedence).
	Project Project `yaml:"project"`
	// Env pins the target environment: "local", "production", or "testing".
	// Empty resolves to local; KIW_ENV and --env override (KWL-K4T7W).
	Env string `yaml:"env"`
	// Debug enables debug mode. Defaults to false; KIW_DEBUG and --debug
	// override.
	Debug bool `yaml:"debug"`
	// Title is the site title.
	Title string `yaml:"title"`
	// Description is the site description (read by pkg/web/ssg for page data).
	Description string `yaml:"description"`
	// Author is the site author.
	Author string `yaml:"author"`
	// Base is the URL base the site is served under, e.g. "/docs/".
	Base string `yaml:"base"`
	// Version is the product version injected into page data (.Version) for
	// both ssg and book pipelines, so web content never hardcodes it.
	Version string `yaml:"version"`
	// Input is the content directory, relative to the project root.
	// Defaults to "content" when empty; may be overridden by krewire.yaml
	// or the --input flag.
	Input string `yaml:"input"`
	// Output is the output directory, relative to the project root.
	// Defaults to .krewire/build when empty; may be overridden by
	// krewire.yaml or the --output/-o flag.
	Output string `yaml:"output"`
	// Build tunes the build pipeline: include/exclude content path globs.
	Build BuildConfig `yaml:"build"`
	// Nav holds optional navbar links.
	Nav []Link `yaml:"nav"`
	// Footer is optional footer text.
	Footer string `yaml:"footer"`
	// Theme configures the light/dark theme switcher.
	Theme *Theme `yaml:"theme"`
	// Book holds book-specific settings for the content/ shape.
	Book BookConfig `yaml:"book"`
	// SSG holds the declarative SSG config. When non-nil, krew build uses
	// the web/ssg generator instead of the standalone book builder. Top-level fields (Title,
	// Author, Base, Output, Theme) are injected into the SSG config.
	SSG *SSGConfig `yaml:"ssg"`
	// AutoAssets configures asset injection and bundling at top-level.
	AutoAssets *ssg.AutoAssetConfig `yaml:"auto_assets"`
	// Wasm holds WASM build configuration (KWF-T4X9P).
	Wasm WasmConfig `yaml:"wasm"`
	// Runtime holds multi-runtime coordination settings (KWN-MRCN).
	Runtime RuntimeConfig `yaml:"runtime"`
	// Dev holds development settings (e.g. dev server port).
	Dev DevConfig `yaml:"dev"`
	// Server holds server/service settings (e.g. host and port).
	Server ServerConfig `yaml:"server"`
	// Worker holds background worker settings (KWF-L5H2F).
	Worker *WorkerConfig `yaml:"worker"`
	// Service holds standalone microservice configuration (KWF-L5H2F).
	Service *ServiceConfig `yaml:"service"`
	// Services holds multiple declared services in the project (FRK-CONF-032).
	Services []ServiceConfig `yaml:"services"`
	// Infra holds cloud infrastructure configuration (KWF-B7N3D).
	Infra *InfraConfig `yaml:"infra"`
	// Gateway holds API gateway reverse proxy configuration (KWF-L5H2F).
	Gateway *GatewayConfig `yaml:"gateway"`
	// Scripts holds named shell commands runnable via `krew run <task>`.
	// Keys are task names, values are shell commands executed with `sh -c`
	// (Unix) or `cmd /C` (Windows). Optional, zero-value nil means no tasks.
	Scripts map[string]string `yaml:"scripts"`
	// I18n configures internationalization for the site/book build.
	// When set, krew build iterates over supported locales and generates
	// localized output trees (e.g. .krewire/build/en/docs/, .../id/docs/).
	I18n *ssg.I18nConfig `yaml:"i18n"`
	// API configures central API Gateway routing.
	API APIConfig `yaml:"api"`
	// Database tells `krew migrate` which database to migrate.
	Database DatabaseConfig `yaml:"database"`
}

// APIConfig holds settings for the unified API Gateway.
type APIConfig struct {
	// URL overrides the central API Gateway endpoint (default: "https://api.krewire.com").
	URL string `yaml:"url"`
}

// GatewayConfig holds API gateway configuration under the `gateway:` key (KWF-L5H2F).
type GatewayConfig struct {
	Port   int            `yaml:"port"`
	Host   string         `yaml:"host"`
	Routes []GatewayRoute `yaml:"routes"`
}

// GatewayRoute maps a path to a service or target URL.
type GatewayRoute struct {
	Path        string   `yaml:"path"`
	Method      string   `yaml:"method"`
	Service     string   `yaml:"service"`
	TargetURL   string   `yaml:"target_url"`
	Prefix      bool     `yaml:"prefix"`
	StripPrefix bool     `yaml:"strip_prefix"`
	Middleware  []string `yaml:"middleware"`
	RateLimit   int      `yaml:"rate_limit"`
	Timeout     string   `yaml:"timeout"`
}

// RuntimeConfig holds multi-runtime coordination settings under the `runtime:` key (KWN-MRCN).
type RuntimeConfig struct {
	// Namespace prefixes Krewire-specific environment variables (e.g. "go", "krewire").
	Namespace string `yaml:"namespace"`
	// EnvFile points to the .env file path (default: ".env").
	EnvFile string `yaml:"env_file"`
	// Port declares the default runtime port.
	Port int `yaml:"port"`
}

// DevConfig holds development server configuration under the `dev:` key.
type DevConfig struct {
	Port int `yaml:"port"`
}

// ServerConfig holds service/server settings under the `server:` key.
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// ServiceConfig holds service-specific settings under the `service:` or `services:` key (FRK-CONF-032).
type ServiceConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// WorkerConfig holds queue and background processing settings under the `worker:` key (KWF-L5H2F).
type WorkerConfig struct {
	Concurrency int      `yaml:"concurrency"`
	Queues      []string `yaml:"queues"`
}

// InfraConfig holds cloud infrastructure settings under the `infra:` key (KWF-B7N3D).
type InfraConfig struct {
	Provider  string              `yaml:"provider"`
	Region    string              `yaml:"region"`
	State     string              `yaml:"state"`
	Resources []provider.Resource `yaml:"resources"`
}

// BuildConfig tunes the build pipeline under the `build:` key.
type BuildConfig struct {
	// Base is the URL base the site is served under, e.g. "/docs/".
	// Legacy top-level `base:` stays supported; nested `build.base`
	// wins when both are set. krew --base wins over both.
	Base string `yaml:"base"`
	// Output is the output directory relative to the project root.
	Output string `yaml:"output"`
	// Include holds glob patterns selecting which content files build
	// (slash-separated, ** crosses segments). Unset (nil) uses the book
	// default "**/*.md"; an empty list disables the include filter.
	Include []string `yaml:"include"`
	// Exclude holds glob patterns removed after include filtering. Unset
	// (nil) uses the book defaults excluding README/readme developer
	// notes; an empty list excludes nothing.
	Exclude []string `yaml:"exclude"`
	// DependsOn lists shell commands to run in order before building
	// (e.g. ["npm run build", "composer install --no-dev"]).
	DependsOn []string `yaml:"depends_on"`
	// Outputs maps runtime names to output directories (e.g. nodejs: "dist", laravel: "public/build").
	Outputs map[string]string `yaml:"outputs"`
}

// BookConfig holds book-specific settings under the `book:` key.
type BookConfig struct {
	// Mount places the whole book under a sub-path of the output and URL
	// space, e.g. "/docs/" — letting a book co-exist with an ssg site
	// in one .krewire/build (progressive enhancement). Empty mounts at root.
	Mount string `yaml:"mount"`
	// TOC controls the generated root table-of-contents page. In a hybrid
	// project it defaults to false (suppressed so the ssg landing page owns
	// "/"); set `toc: true` to force it on. Standalone books always emit it.
	TOC *bool `yaml:"toc"`
}

// WasmConfig holds WASM build configuration under the `wasm:` key (KWF-T4X9P).
type WasmConfig struct {
	// Entry is the main package path to build (default "./wasm").
	Entry string `yaml:"entry"`
	// Name is the output module name (default "runtime").
	Name string `yaml:"name"`
}

// IsSSG reports whether this config targets the declarative SSG mode.
func (c *Config) IsSSG() bool {
	return c != nil && c.SSG != nil
}

// Kind returns the pinned project kind from config, or "" when unset.
func (c *Config) Kind() string {
	if c == nil {
		return ""
	}
	return c.Project.Kind
}

// APIURL returns the effective API Gateway endpoint, respecting KREWIRE_API_URL.
func (c *Config) APIURL() string {
	if env := os.Getenv("KREWIRE_API_URL"); env != "" {
		return strings.TrimRight(env, "/")
	}
	if c != nil && c.API.URL != "" {
		return strings.TrimRight(c.API.URL, "/")
	}
	return DefaultAPIURL
}

// AuthURL returns the central Auth service endpoint.
func (c *Config) AuthURL() string {
	return c.APIURL() + "/auth"
}

// CloudURL returns the Cloud engine endpoint.
func (c *Config) CloudURL() string {
	return c.APIURL() + "/cloud"
}

// DashURL returns the Dash workspace endpoint.
func (c *Config) DashURL() string {
	return c.APIURL() + "/dash"
}

// DeclaredPort returns the port declared in krewire.yaml across runtime.port,
// dev.port, server.port, or services (FRK-CONF-003, FRK-CONF-030).
func (c *Config) DeclaredPort() int {
	if c == nil {
		return 0
	}
	if c.Runtime.Port > 0 {
		return c.Runtime.Port
	}
	if c.Dev.Port > 0 {
		return c.Dev.Port
	}
	if c.Server.Port > 0 {
		return c.Server.Port
	}
	if len(c.Services) > 0 && c.Services[0].Port > 0 {
		return c.Services[0].Port
	}
	return 0
}

// EffectiveEnvFile returns the configured env file path or ".env" by default (FRK-CONF-012).
func (c *Config) EffectiveEnvFile() string {
	if c != nil && c.Runtime.EnvFile != "" {
		return c.Runtime.EnvFile
	}
	return ".env"
}

// Project holds per-project settings in krewire.yaml.
type Project struct {
	// Name is the project name.
	Name string `yaml:"name"`
	// Kind pins the project kind (see pkg/model/workload.go AllKinds:
	// "app", "cli", "site", "book", "worker", "service", "gateway", or
	// "infra"). Empty lets the CLI detect it from marker files.
	Kind string `yaml:"kind"`
	// Version is the product version declared next to the kind, the way a
	// project writes it:
	//
	//	project:
	//	  kind: site
	//	  version: v0.1.0
	//
	// Load folds it into Config.Version when no top-level `version:` is set,
	// so the version has one effective value regardless of where it is declared.
	Version string `yaml:"version"`
	// Author is the project author or team.
	Author string `yaml:"author"`
	// Description is a one-line project description.
	Description string `yaml:"description"`
	// Dirs overrides the canonical directory locations (FRK-STR-010).
	Dirs Dirs `yaml:"dirs"`
}

// Dirs holds custom directory locations for the project layout.
type Dirs struct {
	Web      string `yaml:"web"`
	Public   string `yaml:"public"`
	Internal string `yaml:"internal"`
	Cmd      string `yaml:"cmd"`
}

// Link is a named navigation target shown in the navbar.
type Link struct {
	Text string `yaml:"text"`
	URL  string `yaml:"url"`
}
