# Contributing to Krew

This guide covers the workflow, conventions, and quality gates for contributing to **Krew** (`github.com/krewire/krew`).

See also [docs/contributing.md](docs/contributing.md) and [docs/architecture.md](docs/architecture.md).

## Language & Communication

- All code, documentation, comments, commit messages, and issues must be in **English**.
- If communicating in Indonesian or another language in chat, reply in that language, but keep all repository files in English.

## Branching Model

- Trunk-based development: feature/bugfix branches are cut from `main`.
- Branch naming: `feat/<name>`, `fix/<name>`, `perf/<improvement>`.
- One logical change per PR with atomic commits (≤~300 LOC).

## Conventional Commits

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):
```text
feat(agent): implement multi-step ReAct reasoning loop
fix(tools): guard against path traversal in read_file
perf(memory): optimize sliding window message buffer
```

## Quality Gates

Before pushing changes or submitting a PR:
```bash
gofmt -l .                          # must print nothing
go vet ./...                        # must pass
go test -v ./...                    # must pass
go build -o bin/krew ./cmd/krew     # must build
```
