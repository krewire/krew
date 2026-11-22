# Krew — Architecture & Subsystems Design

This document details the internal design and core packages within `github.com/krewire/krew`.

## System Overview

Krew is the unified developer toolchain and CLI driver for the Krewire digital SDLC ecosystem. Its architecture enforces private-by-default modularity with thin entrypoints and cohesive domain subsystems.

```text
                                cmd/krew (Composition Root)
                                             │
                                             ▼
                             internal/cli (App & Command Groups)
                                             │
       ┌────────────────────────┬────────────┴───────────┬────────────────────────┐
       ▼                        ▼                        ▼                        ▼
internal/commands      internal/dashboard         internal/scaffold       internal/config
- Command Handlers     - Dev Dashboard Server     - Code Generators       - Schema Validator
- Flags & Validation   - SSE Telemetry Stream     - Template Equipping    - krewire.yaml AST
- Execution Drivers    - Web Console UI (/mcp)    - Workload Shapes       - Diagnostics
       │                        │                        │                        │
       └────────────────────────┴────────────┬───────────┴────────────────────────┘
                                             ▼
                              Foundational Ecosystem Libraries
                              - github.com/krewire/tui (TUI primitives)
                              - github.com/krewire/krewire/pkg/* (Core domain & AI)
```

## Subsystem Domains

- **`cmd/krew`**:
  - Minimal composition root (~12 lines) invoking `cli.BuildApp("krew")`.

- **`internal/cli`**:
  - Root command builder and command-group definitions (`agent`, `project`, `build`, `develop`, `inspect`, `ship`).
  - Flag parsing and delegation to command handlers.

- **`internal/commands`**:
  - Command implementations categorized into logical functional domains:
    - **Project & Database**: `init`, `new`, `add`, `remove`, `update`, `make`, `migrate`, `seed`, `ui`, `pkg`, `auth`.
    - **Build & Artifacts**: `build`, `bundle`, `serve`.
    - **Develop & Runtime**: `dev`, `run`, `test`, `verify`, `vet`, `fmt`, `worker`, `dash`, `gateway`, `mcp`.
    - **Inspect & Diagnostics**: `version`, `info`, `compat`, `context`, `bench`, `tier`, `config`, `doctor`, `self-update`.
    - **Ship & Release**: `deploy`, `ws`, `boost`, `release`.

- **`internal/dashboard`**:
  - Standalone internal dashboard server subsystem.
  - Serves local development dashboard, SSE telemetry (`/api/events`), health checks (`/api/health`), project info (`/api/info`), and mounts the MCP server endpoint.

- **`internal/config` & `internal/shape`**:
  - Declarative configuration parser (`krewire.yaml`), JSON schema validation, and project workload shape detector (`app`, `site`, `book`, `cli`, `worker`, `service`, `gateway`, `infra`).

- **`internal/scaffold`**:
  - Deterministic project scaffolding, template extraction, and workload equipping.

- **`internal/mcp`**:
  - Model Context Protocol server exposing tool capabilities to IDE coding agents.

## Architectural Invariants

1. **Private by Default**: Krew provides developer toolchain binaries and contains no exported public Go SDK packages in `pkg/`. Reusable libraries belong strictly in `krewire/pkg/*`.
2. **Pure Go (`CGO_ENABLED=0`)**: All command implementations, asset bundling, and file operations compile cleanly without native C linkages.
3. **Downward-Only Dependencies**: Commands import internal domain libraries (`config`, `shape`, `scaffold`, `dashboard`), never the inverse.
