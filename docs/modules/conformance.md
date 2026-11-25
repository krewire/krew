# `conformance` (`github.com/krewire/krew/conformance`)

## Purpose & Design Role

`conformance` is the automated platform-integrity and build-reproducibility verification suite for `krew`. Rather than testing individual command logic (which lives in `*_test.go` alongside each `core` subpackage), `conformance` validates ecosystem-wide non-functional invariants on the compiled `./cli` binary: CGO-free compilation, multi-OS/multi-arch cross-compilation, and deterministic bitwise reproducibility.

## Architectural Position

- **Layer**: Layer 3 (Quality Gate & Conformance Module for `krew`).
- **Module**: `github.com/krewire/krew/conformance` (`conformance/go.mod`).
- **Dependencies**: Go standard library only (`bytes`, `crypto/sha256`, `os`, `os/exec`, `path/filepath`, `testing`). Shells out to `go build` against `./cli` from the repository root.

## Core Concepts & Tests (`platform_test.go`)

- `TestPlatform_PureGo(t *testing.T)`: Runs `CGO_ENABLED=0 go build ./cli` into a temporary directory to guarantee zero C/CGO dependencies. (Skipped when `-short` is passed.)
- `TestPlatform_CrossCompileCLI(t *testing.T)`: Cross-compiles `./cli` with `CGO_ENABLED=0` across:
  - `linux/amd64`
  - `darwin/arm64`
  - `windows/amd64`
- `TestPlatform_BitwiseReproducibility(t *testing.T)`: Performs two consecutive clean builds with `-trimpath -ldflags="-s -w"` and `CGO_ENABLED=0`, hashing both binaries with SHA-256 and failing if the digests differ.

## Realistic Usage Examples

Running the full platform conformance suite (including cross-compilation and reproducibility checks):

```bash
cd /home/krewire/workspace/krewire/krew/conformance
go test -v ./...
```

Running in fast mode (skips full binary builds):

```bash
cd /home/krewire/workspace/krewire/krew/conformance
go test -short -v ./...
```

## Verification

Run from `/home/krewire/workspace/krewire/krew/conformance`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build ./...
```
