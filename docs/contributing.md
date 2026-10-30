# Contributing to Krew

Thank you for contributing to **Krew**! This document explains our standards, development flow, and quality gates.

## Development Workflow

1. **Issue Tracking**:
   - Check open issues on `krewire/krew`.
   - All non-trivial PRs should reference an issue.

2. **Branching & Commits**:
   - Branch from `main`: `feat/<name>`, `fix/<name>`, `docs/<name>`, `test/<name>`.
   - Conventional Commits:
     ```text
     feat(mcp): add json-rpc ping and notification handling (ref #14)
     fix(memory): prevent context loss during aggressive compaction
     test(tools): add boundary tests for file editing parameters
     ```
   - Keep commits atomic ($\le$ ~300 LOC per commit).

3. **Quality Gates**:
   Before submitting changes, run quality gates locally:
   ```bash
   gofmt -l .        # must print nothing
   go vet ./...      # must pass
   go test -v ./...  # must pass
   ```

## Design Principles

- **Zero Monolithic Framework Lock-In**: Krew is a modular library first, a CLI second. Core functionality must remain usable in-process via `pkg/embed` without requiring CLI flags or external daemon processes.
- **Transport Independence**: Business logic belongs in `pkg/agent` and `pkg/tools`, never inside transport handlers (MCP, HTTP, CLI).
- **Safety by Default**: Tool executions that modify files or execute shell commands must follow explicit user authorization or sandbox policies.
- **English Language**: All code, documentation, and commits must be in English.
