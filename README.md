# Krew — Unified Developer Toolchain & CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/krewire/krew.svg)](https://pkg.go.dev/github.com/krewire/krew)
[![License: Proprietary](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

**Krew** is the unified developer toolchain and command-line driver for the entire Krewire digital SDLC ecosystem. It orchestrates project scaffolding, live reloading, production compilation, asset bundling, database migrations, quality gate verifications, and IDE Model Context Protocol (MCP) integrations.

> **Note**: All autonomous AI agent packages, reasoning loops, and provider adapters have moved to the foundational core monorepo: [`github.com/krewire/krewire/pkg/ai`](https://github.com/krewire/krewire/tree/main/pkg/ai).

---

## Core Workflow Commands

- `krew new <name> [--site|--book|--cli|--app]` — Scaffold a new project from authoritative templates.
- `krew init` — Equip Krewire into an existing directory.
- `krew dev` — Live development server with hot reload and file watching.
- `krew build` — Static site generation, application binary compilation, and infrastructure bundles.
- `krew bundle` — CSS and JS static asset bundling with minification.
- `krew serve` — Preview production output from `.krewire/build`.
- `krew run` — Execute tasks or arbitrary Go source files with runtime context.
- `krew test` — Run unit and integration tests.
- `krew verify` — Run reproducible quality gates and emit verifiable task records.
- `krew vet` — Static analysis, schema linting, and vulnerability scanning.
- `krew fmt` — Enforce Go formatting standards.
- `krew make <kind> <Name>` — Generate components, pages, models, migrations, or seeders.
- `krew migrate [up|down|status]` — Database schema migration management.
- `krew seed` — Database seeders and factory runners.
- `krew ui` — Browse, install, and diff `.kiw` Pop-Brutalist UI components.
- `krew pkg` — Package management for standalone `.kiw` component bundles.
- `krew dash` — Launch local development dashboard and sandbox environments.
- `krew gateway` — Local edge API gateway reverse proxy.
- `krew mcp` — Native Model Context Protocol (MCP) server for IDE coding assistants.
- `krew ws` — Monorepo and multi-repo workspace management (`go.work`).
- `krew boost` — Scaffold Boost AI agent configuration (`AGENTS.md`, `.cursor/rules`, etc.).

---

## AI Agent SDK

For the autonomous AI agent engine, ReAct reasoning loops, memory compactor, and LLM providers, import directly from the core library:

```go
import (
    "github.com/krewire/krewire/pkg/ai"
    "github.com/krewire/krewire/pkg/ai/agent"
    "github.com/krewire/krewire/pkg/ai/assistant"
    "github.com/krewire/krewire/pkg/ai/embed"
    "github.com/krewire/krewire/pkg/ai/providers"
)
```

CLI commands for interactive AI sessions (`chat`, `agent`, `daemon`, `sessions`, `export`) are provided via `krewire/cmd/ai`.

---

## Quality Gates

Before pushing changes to `krew`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build -o bin/krew ./cmd/krew
```
