package commands

import "github.com/krewire/krewire/pkg/model"

// CanonicalPackageCatalog returns the objective metadata inventory for foundational
// and modular domain packages in the Krewire monorepo (github.com/krewire/krewire).
func CanonicalPackageCatalog() []model.PackageMetadata {
	return []model.PackageMetadata{
		// ── Layer 1: Core Foundation & Kernel ─────────────────────────────────
		{
			Name:                 "model",
			ImportPath:           "github.com/krewire/krewire/pkg/model",
			Layer:                1,
			Purpose:              "Pure domain models, workload definitions, scopes, tiers, and security baselines",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: zero breaking changes within minor releases. Depends only on stdlib.",
		},
		{
			Name:                 "errs",
			ImportPath:           "github.com/krewire/krewire/pkg/errs",
			Layer:                1,
			Purpose:              "POSIX exit codes, structured diagnostic attributes, and error chain helpers",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: immutable exit code values and error constructors.",
		},
		{
			Name:                 "lifecycle",
			ImportPath:           "github.com/krewire/krewire/pkg/lifecycle",
			Layer:                1,
			Purpose:              "Workload supervisor, graceful shutdown, module registry, and execution kernel",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: Startable interface and supervisor contracts are backward-compatible.",
		},
		{
			Name:                 "version",
			ImportPath:           "github.com/krewire/krewire/pkg/version",
			Layer:                1,
			Purpose:              "SemVer parsing, build watermark stamps, and cross-repository version constraints",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: immutable semver comparison rules.",
		},

		// ── Layer 2: Modular Domain Libraries ─────────────────────────────────
		{
			Name:                 "sec",
			ImportPath:           "github.com/krewire/krewire/pkg/sec",
			Layer:                2,
			Purpose:              "Defensive web headers, CSRF guards, SSRF protection, SafePath, and PBKDF2 hashing",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       []string{"golang.org/x/net/html (pinned sanitizer)"},
			BreakingChangePolicy: "Strict SemVer: fail-closed security defaults remain permanent.",
		},
		{
			Name:                 "auth",
			ImportPath:           "github.com/krewire/krewire/pkg/auth",
			Layer:                2,
			Purpose:              "Identity propagation, Basic auth, JWT verification, and structured HTTP errors",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: stdlib-only crypto and context identity propagation.",
		},
		{
			Name:                 "web",
			ImportPath:           "github.com/krewire/krewire/pkg/web",
			Layer:                2,
			Purpose:              "High-performance HTTP router, middleware chain, SSG/SSR pipelines, and gateway reverse proxy",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: standard net/http handler interface compatibility.",
		},
		{
			Name:                 "log",
			ImportPath:           "github.com/krewire/krewire/pkg/log",
			Layer:                2,
			Purpose:              "Structured slog logger, JSON production formatting, and request correlation IDs",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: standard log/slog handler extension.",
		},
		{
			Name:                 "config",
			ImportPath:           "github.com/krewire/krewire/pkg/config",
			Layer:                2,
			Purpose:              "Declarative krewire.yaml parser, .env loader, and environment variable overlays",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: backward-compatible schema field additions.",
		},
		{
			Name:                 "orm",
			ImportPath:           "github.com/krewire/krewire/pkg/orm",
			Layer:                2,
			Purpose:              "Pure-Go multi-driver ORM, SQL query builder, and migration runner",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       []string{"modernc.org/sqlite (opt-in pure-Go)"},
			BreakingChangePolicy: "Strict SemVer: query builder expressions remain additive.",
		},
		{
			Name:                 "cache",
			ImportPath:           "github.com/krewire/krewire/pkg/cache",
			Layer:                2,
			Purpose:              "Multi-driver caching interface (Memory, File, Redis) with TTL eviction",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: cache.Driver interface compatibility.",
		},
		{
			Name:                 "resilience",
			ImportPath:           "github.com/krewire/krewire/pkg/resilience",
			Layer:                2,
			Purpose:              "Retry policies with exponential backoff and circuit breakers",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: policy execution signatures remain stable.",
		},
		{
			Name:                 "storage",
			ImportPath:           "github.com/krewire/krewire/pkg/storage",
			Layer:                2,
			Purpose:              "Backend-independent object storage abstraction (disk, memory)",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: storage.Provider interface compatibility.",
		},
		{
			Name:                 "kiw",
			ImportPath:           "github.com/krewire/ui/kiw",
			Layer:                2,
			Purpose:              "Unified .kiw DSL parser, AST desugaring, and HTML component compiler",
			License:              "MIT",
			Stability:            model.StabilityStable,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Strict SemVer: compiler produces deterministic Go html/template bytecode.",
		},
		{
			Name:                 "runtime",
			ImportPath:           "github.com/krewire/krewire/pkg/runtime",
			Layer:                2,
			Purpose:              "Client WebAssembly VDOM engine, reactive components, and client-side hydration",
			License:              "MIT",
			Stability:            model.StabilityBeta,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Beta: VDOM diffing API stabilizing towards v1.0.",
		},
		{
			Name:                 "ai",
			ImportPath:           "github.com/krewire/krewire/pkg/ai",
			Layer:                2,
			Purpose:              "Native Go AI SDK: provider streaming, function calling, and structured JSON schemas",
			License:              "MIT",
			Stability:            model.StabilityBeta,
			MinGoVersion:         "1.22+",
			ThirdPartyDeps:       nil,
			BreakingChangePolicy: "Beta: LLM provider adapters may add new capabilities additively.",
		},
	}
}

// GetPackageMetadata looks up a package by name or import path in the canonical catalog.
func GetPackageMetadata(name string) (model.PackageMetadata, bool) {
	for _, pkg := range CanonicalPackageCatalog() {
		if pkg.Name == name || pkg.ImportPath == name {
			return pkg, true
		}
	}
	return model.PackageMetadata{}, false
}
