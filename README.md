# Krew — Unified Developer Toolchain & CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/krewire/krew.svg)](https://pkg.go.dev/github.com/krewire/krew)
[![License: Proprietary](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

**Krew** (`github.com/krewire/krew`) is the unified developer toolchain and command-line driver for the entire Krewire digital SDLC ecosystem. It orchestrates project scaffolding, live reloading, production compilation, asset bundling, typed database migrations, quality gate verification, developer authentication, and Model Context Protocol (MCP) integrations.

> **Architecture Notice**: All autonomous AI agent packages, reasoning loops, and provider adapters are located in the foundational core monorepo: [`github.com/krewire/krewire/pkg/ai`](https://github.com/krewire/krewire/tree/main/pkg/ai).

---

## Documentation Index

- [Architecture & Design](docs/architecture.md): CLI composition root (`cmd/krew`), command domains, internal dashboard subsystem (`internal/dashboard`), and private-by-default modularity.
- [Getting Started](docs/getting-started.md): Installation, project creation, database migrations, developer authentication, and local development.
- [Contributing Guide](CONTRIBUTING.md): Quality gates, pure-Go invariant, and command contribution guidelines.

---

## Core Workflow Commands

### 1. Project & Workspace Creation
- `krew new <name> [--site|--book|--cli|--app]` — Scaffold a new project from authoritative templates.
- `krew init [--site|--book|--app]` — Equip Krewire into an existing directory.
- `krew add <package>` — Add dependencies, plugins, or UI components (`ui/<component>`, `ui:<pack>`).
- `krew remove <package>` — Remove a project dependency or plugin.
- `krew update` — Update project dependencies across the workspace.
- `krew ws` — Monorepo and multi-repo workspace management (`go.work`).

### 2. Development, Build & Servicing
- `krew dev` — Live development server with hot reload and file watching.
- `krew build` — Static site generation, application binary compilation, and infrastructure bundles.
- `krew bundle` — CSS and JS static asset bundling with minification.
- `krew serve` — Preview production output from `.krewire/build`.
- `krew run [task|file.go]` — Execute runnable tasks or arbitrary Go source files with runtime context.
- `krew dash` — Launch local development dashboard, metrics, and virtual workspace sandboxes.

### 3. Database & Migrations
- `krew make migration <name>` — Generate a new typed Laravel-style Go migration in `database/migrations/<timestamp>_<name>.go` with programmatic `schema.Blueprint`.
- `krew migrate [up|down|redo|status]` — Apply, roll back, or inspect database schema migrations.
- `krew seed [--class=<Seeder>]` — Seed the database with records via Go seeders and factories (`db:seed`).

### 4. Quality Gates & Diagnostics
- `krew test` — Run unit and integration tests.
- `krew verify` — Run reproducible quality gates and emit verifiable task records (`--json`).
- `krew vet` — Static analysis, schema linting, and vulnerability scanning (`--sec`, `--vuln`).
- `krew fmt` — Enforce Go formatting standards (`--write`).
- `krew doctor` — Diagnose developer environment, toolchain prerequisites, and workspace.
- `krew compat` — Check version compatibility across all ecosystem modules.
- `krew config` — Validate `krewire.yaml` against its canonical JSON Schema.

### 5. Developer Authentication & Identity
- `krew auth login [--token=<token>]` — Authenticate developer credentials against Krewire Auth.
- `krew auth logout` — Remove active developer credentials.
- `krew auth status` — Display current authentication identity and active session token.

### 6. AI Agent & MCP Integration
- `krew chat` — Start an interactive AI coding session in the terminal.
- `krew agent <prompt> [--brief=<path>] [--sdlc]` — Execute an autonomous task or prompt headless to completion.
- `krew mcp [--stdio]` — Native Model Context Protocol (MCP) server for IDE coding assistants.
- `krew boost` — Scaffold Boost AI agent configuration (`AGENTS.md`, `.cursor/rules`, etc.).

---

## Local Development Database Container

For local development with PostgreSQL, start the official `krewire-db` container:

```bash
docker compose -f ../cloud/docker-compose.yml up -d krewire-db
```

This binds PostgreSQL 17 to `127.0.0.1:5432` with credentials defined in `cloud/.env`.

---

## Quality Gates

Before pushing changes to `krew`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build -o bin/krew ./cmd/krew
```
