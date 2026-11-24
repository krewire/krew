# Contributing to Krew

Guidelines for contributing to **Krew** (`github.com/krewire/krew`).

> All code, comments, docstrings, commit messages, and documentation are strictly in **English**.

## Invariants

1. **Pure Multi-Module Monorepo**: Each module has its own `go.mod`. Never add a `go.mod` to the root of the repository.
2. **Pure Go**: Always compile with `CGO_ENABLED=0`.
3. **Hub Integration**: Third-party database drivers and frontend compilers are consumed via `krewire/hub/*`.

## Quality Gates

```bash
gofmt -l .
go vet ./...
go test ./...
CGO_ENABLED=0 go build ./...
```
