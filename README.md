# Krew (`github.com/krewire/krew`)

Unified developer toolchain, CLI driver, scaffolding engine, dev server, build orchestrator, and Model Context Protocol (MCP) server for the Krewire digital SDLC ecosystem.

The repository operates as a **Pure Multi-Module Monorepo** comprising **4 independent Go modules** linked via `go.work`, with zero `go.mod` at the repository root.

## Architecture & Modules

```text
krew/
├── go.work       # Go workspace linking internal modules and ecosystem checkouts
├── cli/          # Primary binary entrypoint (cmd/krew -> cli/krew), flag router
├── core/         # Core developer toolchain commands, scaffolding, dev server, builds, MCP
│   ├── commands/ # CLI subcommands (init, dev, build, test, migrate, ws, etc.)
│   ├── scaffold/ # Project templates, equipping variants (app, site, book, etc.)
│   ├── mcp/      # Model Context Protocol stdio devtools server
│   └── release/  # Qualification and release packaging
├── conformance/  # CLI command-contract and cross-platform conformance tests
└── version/      # Semantic versioning and toolchain watermark signature
```

## Features

- **Project Scaffolding**: `krew new` and `krew init` for 8 workload kinds (`app`, `site`, `book`, `service`, `worker`, `infra`, `gateway`, `template`).
- **Live Development**: `krew dev` with Hot Module Reloading (HMR) and in-memory assets.
- **Production Compilation**: `krew build` producing hermetic, CGO-free single static binaries.
- **Database Migrations**: `krew migrate` supporting Go-native migrations and SQL files with automatic WAL driver tuning via `github.com/krewire/hub/*`.
- **Model Context Protocol (MCP)**: `krew mcp` providing IDEs with unified tools for verification, scaffolding, and workspace navigation.

## Binary Identity & Compatibility Alias Policy

- **Canonical Executable**: The canonical CLI executable is named `krew` (built from `github.com/krewire/krew/cli`).
- **Compatibility Alias**: The command `kiw` is maintained as a permanent compatibility alias. In POSIX environments, `kiw` is symlinked to `krew`.
- **Release Artifacts**: All stable, prerelease, and nightly pipelines build and publish canonical `krew-<os>-<arch>` binaries alongside matching `kiw-<os>-<arch>` binaries in both release assets and checksum manifests (`checksums.txt`), ensuring backward compatibility with legacy installation scripts and downstream workflow automation.

## Quality Gates

Before submitting changes:
```bash
gofmt -l .
go vet ./...
go test ./...
CGO_ENABLED=0 go build ./...
```

## License

Private enterprise property of Krewire ID.
