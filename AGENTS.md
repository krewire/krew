# AGENTS.md — Krew Developer Toolchain Monorepo

Agent guide for AI agents working inside the `krew` developer toolchain repository.

`krew` is the universal developer CLI and build toolchain for the Krewire ecosystem.

## Monorepo Layout (Pure Multi-Module Monorepo)

Coordinated via `go.work` across 4 independent Go modules:
- `cli`: Binary entrypoint (`main.go`).
- `core`: Toolchain commands, scaffolding, dev server, builds, MCP devtools server.
- `conformance`: End-to-end command contract and platform conformance tests.
- `version`: Version metadata.

## Invariants

- **The Krewire Pyramid (Layer 3 Devtool)**: `krew` depends strictly downward on Layer 2 libraries (`krewire/*`) and Layer 1 kernel primitives.
- **Port & Adapter Integration**: External database drivers and frontend compilers are integrated through `krewire/hub/*`.
- **Zero Peer Coupling**: Sibling platforms (`cloud`, `auth`, `dash`) must never be imported by `krew`.
- **Pure Go & CGO-Free**: All code must compile cleanly with `CGO_ENABLED=0`.

## Quality Gates

```bash
gofmt -l .
go vet ./...
go test ./...
CGO_ENABLED=0 go build ./...
```
