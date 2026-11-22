# Krew — Developer & Contributor Documentation

Welcome to the **Krew** repository documentation. This repository (`github.com/krewire/krew`) implements the unified developer toolchain, CLI driver, and Model Context Protocol (MCP) server for the Krewire digital SDLC ecosystem.

## Documentation Index

- [Architecture & Design](architecture.md): CLI composition root (`cmd/krew`), internal dashboard subsystem (`internal/dashboard`), command execution domains, and private-by-default modularity.
- [Getting Started](getting-started.md): Installation, local setup, building the `krew` CLI, project scaffolding, database migrations, and developer authentication.
- [Contributing Guide](contributing.md): Contribution guidelines, testing requirements, quality gates, and code standards.

## Core Capabilities

- **Unified Developer CLI**: Standard entrypoint across all 8 software workload kinds (`app`, `cli`, `site`, `book`, `worker`, `service`, `gateway`, `infra`).
- **Typed Database Migrations**: Laravel-style Go migrations generated with `krew make migration <name>` and managed via `krew migrate`.
- **Developer Authentication**: Seamless identity session management (`krew auth login`, `krew auth logout`, `krew auth status`).
- **Internal Dev Dashboard**: Standalone dashboard HTTP server (`internal/dashboard`) providing real-time telemetry over Server-Sent Events (SSE).
- **Model Context Protocol (MCP)**: Native stdio and HTTP MCP server exposing tool capabilities to IDE coding assistants (Cursor, Antigravity, Claude Desktop).
- **Hermetic Scaffolding**: Declarative project bootstrapping and template equipping via `internal/scaffold`.
