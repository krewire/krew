# Krewire Boost — Agent Constitution

This document is the primary operating ruleset for AI agents in projects
that install Krewire Boost (`kiw boost install`). It is tuned for the
**Krewire ecosystem** (Go projects driven by the `kiw` CLI) and provides a
comprehensive **Agent Triad** architecture covering all phases of the Software
Development Life Cycle (SDLC).

> [!NOTE]
> If the current project is not a Krewire project, follow the generic principles
> and skip Krewire-specific sections marked Krewire.

## Language

- **Write all stored artifacts in English** — code, docs, commit messages,
  specs, reports, logs.
- The human may write to you in Indonesian (or any language); **reply in the
  human's language in chat**, but keep all persisted artifacts English.

## Mission

Help developers build, verify, deploy, and maintain software correctly, quickly,
and safely — without breaking what already works. Never guess: understand and
verify before acting.

## Architecture: The Agent Triad Model

Every capability in Krewire Boost is built as an **Agent Triad** sharing a single
root verb:

| Component | Naming Pattern | Location | Role |
|---|---|---|---|
| **Agent** | Root + `-er` / `-or` | `.agents/agents/{name}.md` | Subagent persona, boundaries, and execution instructions |
| **Command** | Bare verb | `.agents/commands/{name}.md` | Slash command entry point and user arguments |
| **Skill** | Active verb (`-ing`) | `.agents/skills/{name}/SKILL.md` | Step-by-step procedures, domain knowledge, and tool rules |

### Available Agents (Full SDLC Triad Matrix)

| Agent (Subagent) | Command | Skill | SDLC Phase | Scope & Description |
|---|---|---|---|---|
| `planner` | `/plan` | `skills/planning/SKILL.md` | Planning | Requirements analysis, user stories, MoSCoW backlog (IEEE 29148) |
| `designer` | `/design` | `skills/designing/SKILL.md` | Architecture | System design, C4/UML models, OpenAPI specs, ADRs (ISO 42010) |
| `builder` | `/build` | `skills/building/SKILL.md` | Implementation | Feature development, Go/Krewire architecture, Clean Code, SOLID |
| `tester` | `/test` | `skills/testing/SKILL.md` | Testing | Test strategy, test pyramid, unit/integration/E2E QA (ISTQB) |
| `deployer` | `/deploy` | `skills/deploying/SKILL.md` | Deployment | CI/CD pipelines, release packaging, GitOps, environments (DORA) |
| `maintainer` | `/maintain` | `skills/maintaining/SKILL.md` | Maintenance | Observability, incident response, technical debt reduction (ITIL/SRE) |
| `auditor` | `/audit` | `skills/auditing/SKILL.md` | Cross-cutting | Security audits (OWASP Top 10), dependency hygiene, license review |
| `decomposer` | `/decompose` | `skills/decomposing/SKILL.md` | Cross-cutting | Single Responsibility Principle (SRP) decomposition & refactoring |
| `explorer` | `/explore` | `skills/exploring/SKILL.md` | Cross-cutting | Codebase search, orientation, symbol mapping, and inventory |
| `writer` | `/write` | `skills/writing/SKILL.md` | Cross-cutting | Technical docs, spec sync, documentation-first workflows |
| `automator` | `/automate` | `skills/automating/SKILL.md` | Cross-cutting | Tooling, batch automation, idempotent scripting |
| `tracker` | `/track` | `skills/tracking/SKILL.md` | Cross-cutting | GitHub Issues management and spec progress checklist SSOT |
| `customizer` | `/customize` | `skills/customizing/SKILL.md` | Governance | Agent triad scaffolding, renaming, and artifact standardization |
| `leader` | `/lead` | `skills/leading/SKILL.md` | Governance | Outcome authority, scope definition, and acceptance gates |
| `manager` | `/manage` | `skills/managing/SKILL.md` | Governance | Intent clarification, task-graph packaging, and delegation |
| `orchestrator` | `/orchestrate` | `skills/orchestrating/SKILL.md` | Governance | Dependency-aware execution graph coordination |
| `supervisor` | `/supervise` | `skills/supervising/SKILL.md` | Governance | Runtime monitoring, checkpoints, retries, and escalation |
| `housekeeper` | `/housekeep` | `skills/housekeeping/SKILL.md` | Hygiene | Rules, memory, context, and documentation hygiene |
| `learner` | `/learn` | `skills/learning/SKILL.md` | Improvement | Self-improvement loop: capture, consolidate, and tune workflow |
| `researcher` | `/research` | `skills/researching/SKILL.md` | Research | Real-world web research, external benchmarks, and landscape analysis |
| `thinker` | `/think` | `skills/thinking/SKILL.md` | Reasoning | Architectural decision design, trade-offs, and conflict resolution |

## Mandatory 5-Phase Execution Workflow

Every agent and command executes in five distinct phases:

1. **Understand** — Clarify requirements, gather facts, verify source code, and detect project shape. Cite real sources.
2. **Plan** — Formulate an approach with clear boundaries, steps, and test criteria before writing code.
3. **Build** — Implement minimal, high-cohesion changes following project conventions.
4. **Verify** — Run tests, linters, and builds. Never complete a task without positive verification.
5. **Summarize** — Concisely report changes, test evidence, and handoff instructions.

## Identify the Project Kind (Krewire)

Krewire projects use a single config file (`krewire.yaml`) and eight workload
shapes. Detect the kind before executing workflows:

| Kind | Detection Marker | Driven by | Needs `main.go` |
|---|---|---|---|
| `app` | `project.kind: app`, root `main.go`, or `cmd/*` | `kiw run`, `kiw dev` | yes |
| `cli` | `project.kind: cli` | `kiw run <args>` | yes |
| `site` | `project.kind: site`, `ssg:` key, or `pages/` | `kiw build`, `kiw serve` | no |
| `book` | `project.kind: book` or `content/` dir | `kiw build`, `kiw serve` | no |
| `worker` | `project.kind: worker`, `worker:` key | `kiw run`, `kiw worker`, `kiw dev` | yes |
| `service` | `project.kind: service`, `service:` key | `kiw run`, `kiw dev` | yes |
| `gateway` | `project.kind: gateway`, `gateway:` key | `kiw run`, `kiw gateway`, `kiw dev` | yes |
| `infra` | `project.kind: infra`, `infra:` key | `kiw deploy`, `kiw build --plan` | no |

Validate with `kiw info`. Never guess kind: read `krewire.yaml` or inspect file markers.

## The `kiw` Command Matrix (Krewire)

The `kiw` CLI is the unified entry point:

| Command | Purpose | Target Kinds |
|---|---|---|
| `kiw new <name>` | Scaffold a new project (fullstack `app` by default) | any |
| `kiw init [<variant>]` | Equip a pre-init bootstrap project with layout (see `kiw init --help`) | any |
| `kiw build` | Compile binary or build static output (`.krewire/build`) | all |
| `kiw serve` | Start local server or preview output | all |
| `kiw run [args...]` | Build and run application binary | app, cli, worker, service |
| `kiw dev` | Live rebuild and reload development server | app, cli, worker, service |
| `kiw worker` | Run background worker queues | worker |
| `kiw deploy` | Provision and deploy infrastructure | app, site, book, worker, service, infra |
| `kiw test` | Run `go test ./...` | all Go projects |
| `kiw vet` | Run `go vet ./...` | all Go projects |
| `kiw fmt` | Run `gofmt -l .` / format code | all Go projects |
| `kiw make <kind> <name>` | Scaffold component, page, handler, migration, model, middleware, service | all |
| `kiw dash` | Interactive dev dashboard, metrics, virtual workspaces & SSE telemetry (`:4000`) | all |
| `kiw mcp [--stdio]` | Native Model Context Protocol (MCP) server for AI coding agents | all |
| `kiw info` | Display project environment and kind | all |
| `kiw version` | Display CLI and ecosystem version | all |
| `kiw boost install` | Install Krewire Boost AI templates | all |

## Native Agentic Integration: Model Context Protocol (MCP)

Krewire projects feature native bidirectional Model Context Protocol (MCP) support for AI coding assistants (Antigravity, Cursor, Claude Desktop, OpenCode):

- **stdio transport**: Connect local IDEs and desktop assistants directly with `kiw mcp --stdio`. Configured out of the box in `opencode.json` and `.cursor/mcp.json`.
- **SSE / HTTP transport**: Access telemetry and tools over network streams via `kiw dash` at `http://127.0.0.1:4000/mcp/sse`.

### MCP Resources
- `krewire://project/info` — Project metadata, workload kind, Go runtime, and active services.
- `krewire://project/routes` — Discovered `.kiw` pages, Markdown articles, and Go HTTP handlers.
- `krewire://runtime/metrics` — Live memory allocation, system memory, goroutines, and GC stats.
- `krewire://runtime/health` — Continuous uptime and health status probe.

### MCP Tools
- `kiw_inspect_project` — Read project configuration and ecosystem topology.
- `kiw_list_routes` — Inspect full routing table.
- `kiw_make_scaffold` — Execute verified Krewire generators for components, pages, handlers, models, middleware, and services.
- `kiw_get_metrics` — Retrieve runtime memory telemetry.
- `kiw_run_tests` — Run test suite (`go test ./...`) with structured outcome.

## Quality Gates

Verify changes before completing any task:

```bash
gofmt -l .        # must print nothing
go vet ./...      # must report zero defects
go test ./...     # must pass all tests
```

- For `site` and `book` kinds, run `kiw build` and verify `.krewire/build` generation.
- Never declare success without running and checking actual command output.

## Commit & Progress Tracking Discipline

- **Conventional Commits**: `feat|fix|docs|test|refactor|perf|security|chore(scope): desc`.
- Atomic commits ($\le$ ~300 LOC per commit).
- **GitHub Issues SSOT**:
  - Tasks and features are tracked as GitHub Issues with checklist items (`- [ ]` / `- [x]`).
  - Commits reference issue numbers: `feat(scope): implement feature (ref #<IssueNumber>)`.