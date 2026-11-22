# AGENTS.md — Krew Developer Toolchain Monorepo

Agent guide for AI agents working inside the `krew` repository
(module = `github.com/krewire/krew`).

`krew` is the unified developer toolchain and CLI for the Krewire digital ecosystem:
providing project scaffolding (`new`, `init`, `make`), live dev server & reload (`dev`, `run`, `serve`),
production builds (`build`, `bundle`), database migrations and seeders (`migrate`, `seed`),
quality gates (`test`, `verify`, `vet`, `fmt`), component management (`ui`, `pkg`, `add`),
workspace orchestration (`dash`, `ws`, `gateway`, `worker`), and IDE integration (`mcp`).

All AI agent engine packages and runtime features reside in the core foundation at
`github.com/krewire/krewire/pkg/ai`. `krew` is strictly a developer toolchain.

## Language

- **Write all stored artifacts in English** — code, docs, commit messages, specs, reports.
- If the human writes in Indonesian or another language, reply in the human's language in chat, but keep all repository files in English.

## Architecture & Code Discipline

- **The Krewire Pyramid (Layer 3 Devtool)**: `krew` is the developer toolchain CLI. It depends strictly downward on Layer 2 libraries (`krewire/pkg/*`) and Layer 1 kernel primitives.
- **Pure Devtool Focus**: `krew` contains zero AI engine implementations; all AI capabilities are imported from `krewire/pkg/ai`.
- **Zero Peer Coupling**: Sibling platforms (`cloud`, `auth`, `dash`) must never be imported by `krew`.
- **Platform-Agnostic**: Keep `krew` free of platform-specific assumptions or hardcoded CGO dependencies (`CGO_ENABLED=0`).
- **Single Source of Truth (SSOT)**: Never duplicate tool schemas or version constants.

## Quality Gates

Before pushing changes:
```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build -o bin/krew ./cmd/krew
```

## Commit Discipline

- **Conventional Commits**: `feat|fix|docs|test|refactor|perf|chore(scope): desc`.
- Keep commit messages concise, precise, and in English.
