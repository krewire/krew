# Contributing to Krew

Thank you for contributing to **Krew** (`github.com/krewire/krew`), the unified developer toolchain, CLI driver, and autonomous agent runtime for the Krewire digital SDLC ecosystem!

This guide covers the development workflow, coding standards, design principles, and quality gates for contributors.

---

## Language & Communication Standards

- All code, documentation, comments, commit messages, and issues must be in **English**.
- If communicating in Indonesian or another language in chat, reply in that language, but keep all repository files and code artifacts strictly in English.

---

## Development Workflow

1. **Issue Single Source of Truth**:
   - Check open issues on `krewire/krew`.
   - All non-trivial PRs should reference an issue (`ref #<IssueNumber>`).

2. **Branching Model**:
   - Trunk-based development: feature and bugfix branches are cut directly from `main`.
   - Branch naming convention: `feat/<name>`, `fix/<name>`, `docs/<name>`, `perf/<improvement>`.
   - One logical change per PR with atomic commits ($\le$ ~300 LOC per commit).

3. **Conventional Commits**:
   Commit messages strictly follow [Conventional Commits](https://www.conventionalcommits.org/):
   ```text
   feat(mcp): add json-rpc ping and notification handling (ref #14)
   fix(tools): guard against path traversal in file operations
   perf(memory): optimize sliding window context buffer
   test(agent): add test cases for multi-step ReAct reasoning loop
   ```

---

## Design Principles

- **Zero Monolithic Framework Lock-In**: Krew is a modular library first, a CLI toolchain second. Core agent functionality must remain usable in-process via `pkg/embed` without requiring CLI flags or external daemon processes.
- **Transport Independence**: Business logic belongs in `pkg/agent` and `pkg/tools`, never inside transport handlers (MCP, HTTP, CLI).
- **Safety by Default**: Tool executions that modify files or execute shell commands must follow explicit user authorization or sandbox policies.
- **Pure Go (`CGO_ENABLED=0`)**: The CLI and agent libraries compile cleanly without CGO, ensuring instant cross-platform distribution.

---

## Quality Gates

Before pushing changes or submitting a pull request, run all quality gates locally:

```bash
# 1. Format check (must output nothing)
gofmt -l .

# 2. Static analysis
go vet ./...

# 3. Unit and integration tests
go test -v ./...

# 4. Binary build verification
CGO_ENABLED=0 go build -o bin/krew ./cmd/krew
```
