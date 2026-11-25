# `cli` (`github.com/krewire/krew/cli`)

## Purpose & Design Role

`cli` is the thin `package main` composition root for the `krew` binary. Keeping `cli/main.go` intentionally minimal (12 lines) decouples the executable entrypoint from the command implementations, scaffolding engines, and servers housed in `github.com/krewire/krew/core`, making the toolchain testable as regular Go packages and fast to cross-compile.

## Architectural Position

- **Layer**: Layer 3 (Universal Developer Driver & CLI).
- **Module**: `github.com/krewire/krew/cli` (`cli/go.mod`).
- **Dependencies**: Imports `os` from the Go standard library and `github.com/krewire/krew/core/cli` (resolved via `go.work`).

## Core Concepts & APIs

- `main()` (`cli/main.go`):
  1. Calls `cli.BuildApp("krew")` to assemble the `*tui.App` command tree, flags, hints, and registered database drivers.
  2. Invokes `app.Run(os.Args[1:])` and passes the integer exit code to `os.Exit`.

## Realistic Usage Examples

Building and running the `krew` binary locally with `CGO_ENABLED=0`:

```bash
cd /home/krewire/workspace/krewire/krew
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/krew ./cli
./bin/krew version
./bin/krew doctor
```

## Verification

Run from `/home/krewire/workspace/krewire/krew/cli`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build ./...
```
