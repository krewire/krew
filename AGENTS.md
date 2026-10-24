# AGENTS.md — Krew AI Agent Monorepo

Agent guide for AI agents working inside the `krew` repository
(module = `github.com/krewire/krew`).

`krew` is an autonomous, platform-agnostic AI agent engine designed to be embedded
anywhere: in-process inside Go applications (Krewire Boost), as a local/remote background
daemon over HTTP/SSE, as an interactive terminal CLI (`krew chat`, `krew run`), and inside
IDEs via Model Context Protocol (stdio MCP).

## Language

- **Write all stored artifacts in English** — code, docs, commit messages, specs, reports.
- If the human writes in Indonesian or another language, reply in the human's language in chat, but keep all repository files in English.

## Architecture & Code Discipline

- **Platform-Agnostic**: Keep `krew` free of platform-specific assumptions or hardcoded CGO dependencies.
- **Strict Separation of Concerns**:
  - `pkg/agent`: ReAct autonomous reasoning-action execution loop and streaming events.
  - `pkg/tools`: Sandboxed workspace tools (`read_file`, `write_file`, `edit_file`, `list_dir`, `run_command`, `fetch_url`).
  - `pkg/memory`: Sliding-window message history buffer and context compaction.
  - `pkg/persona`: Specialized role prompts (`CodingAssistant`, `DevOpsRunner`, `AutonomousWorker`, `GeneralAssistant`).
  - `pkg/embed`: High-level 3-line ergonomic embedding API.
  - `pkg/transport`: Transports (HTTP/SSE, Stdio, MCP).
  - `cmd/krew`: Unified CLI entry point.
- **Single Source of Truth (SSOT)**: Never duplicate tool schemas or version constants.

## Quality Gates

Before pushing changes:
```bash
gofmt -l .
go vet ./...
go test -v ./...
go build -o bin/krew ./cmd/krew
```

## Commit Discipline

- **Conventional Commits**: `feat|fix|docs|test|refactor|perf|chore(scope): desc`.
- Keep commit messages concise, precise, and in English.
