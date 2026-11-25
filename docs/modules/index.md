# Krew (`github.com/krewire/krew`) — Module Reference

Krew is the Layer 3 unified developer toolchain, CLI driver, and Model Context Protocol (MCP) server for the Krewire digital SDLC ecosystem. Following the pure multi-module monorepo invariant, there is no `go.mod` at the repository root; `go.work` coordinates 4 independent Go modules:

## Module Catalog

- [`cli`](cli.md) (`github.com/krewire/krew/cli`) — Minimal `main` composition root that builds and executes the `krew` CLI binary via `core/cli.BuildApp("krew")`.
- [`core`](core.md) (`github.com/krewire/krew/core`) — Toolchain implementation containing 15 cohesive subpackages: CLI wiring, command handlers, `krewire.yaml` config, project context linting, scaffolding, shape detection, dev dashboard, MCP server, benchmarks, release tooling, and `kui` component management.
- [`conformance`](conformance.md) (`github.com/krewire/krew/conformance`) — Platform and build conformance test suite verifying `CGO_ENABLED=0` pure-Go builds, cross-compilation (`linux/amd64`, `darwin/arm64`, `windows/amd64`), and bitwise binary reproducibility.
- [`version`](version.md) (`github.com/krewire/krew/version`) — Semantic release version (`0.2.0`) and `EcosystemRequires` compatibility declarations against `github.com/krewire/krewire/version`.

## Architectural Invariants

1. **Layer 3 Tooling Isolation**: `krew` sits at Layer 3 of the 6-Layer Pyramid Architecture. It depends strictly downward on Layer 1–2 libraries (`krewire/*`, `tui/*`, `ui/*`, `hub/*`), and foundational domain modules (`krewire/*`) must never import `krew`.
2. **Zero Peer Coupling**: `krew` never imports lateral peer platforms (`cloud`, `auth`, `dash`).
3. **Pure Go (`CGO_ENABLED=0`)**: All 4 modules compile without CGO, verified by `conformance/platform_test.go`.
