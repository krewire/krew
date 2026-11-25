# `core` (`github.com/krewire/krew/core`)

## Purpose & Design Role

`core` houses the entire implementation of the `krew` developer toolchain across 15 cohesive subpackages. It wires the TUI command tree, executes project lifecycle commands (`new`, `init`, `dev`, `build`, `test`, `verify`, `deploy`), parses and validates `krewire.yaml`, scaffolds all 8 workload shapes, serves the local developer dashboard, runs the native MCP server, and manages `.kiw` UI components and releases.

## Architectural Position

- **Layer**: Layer 3 (Universal Developer Driver & CLI).
- **Module**: `github.com/krewire/krew/core` (`core/go.mod`).
- **Dependencies**:
  - Layer 0–2 ecosystem packages: `github.com/krewire/krewire/*` (`model`, `sec`, `version`, etc.), `github.com/krewire/tui/core`, `github.com/krewire/ui/*`, and `github.com/krewire/hub/{mysql,postgres,sqlite}` (blank-imported in `core/cli/app.go` for migration/database commands).
  - Never imports Layer 4–5 peer platforms (`cloud`, `auth`, `dash`).

## Core Subpackages & APIs

### CLI Assembly & Command Execution (`cli`, `commands`)
- `core/cli` (`BuildApp(name string) *tui.App`): Registers all command groups (`agent`, `inspect`, `project`, `develop`, `build`, `ship`), attaches `commands.Hints`, and links database drivers.
- `core/commands`: Implements handlers and flag registrars (`Register*` / `Run*`) for `new`, `init`, `add`, `remove`, `update`, `make`, `migrate`, `seed`, `ui`, `pkg`, `auth`, `build`, `bundle`, `serve`, `dev`, `run`, `test`, `verify`, `vet`, `fmt`, `worker`, `dash`, `gateway`, `mcp`, `version`, `info`, `compat`, `context`, `bench`, `tier`, `config`, `doctor`, `self-update`, `deploy`, `ws`, `boost`, and `release`.

### Configuration, Shape & Context (`config`, `shape`, `context`, `gomod`, `ecosystem`)
- `core/config`: Defines `Config` (`krewire.yaml`), `APIConfig`, `GatewayConfig`, `GatewayRoute`, `RuntimeConfig`, `DevConfig`, `ServerConfig`, `ServiceConfig`, and canonical defaults (`DefaultOutput = ".krewire/build"`, `DefaultInput = "content"`, `DefaultAPIURL`).
- `core/shape`: Inspects project directories via `Detect(dir string, ...string) (Result, error)` and `IsApp(dir string) bool` to report `Capabilities` (`HasApp`, `HasPages`, `HasSSG`, `HasBook`, `HasWorker`, `HasService`, `HasGateway`, `HasInfra`) or `MarkerPreInit`.
- `core/context`: `Load(targetDir string) (*model.ProjectContext, error)` and `Lint(ctx *model.ProjectContext, targetDir string) []model.ContextDiagnostic` (`SchemaVersion = "1.0.0"`) for AI-agent and human project context validation.
- `core/gomod`: `Read`, `Find`, `(*Module).UsesKrewire()`, and `CheckGoVersion` for `go.mod` inspection.
- `core/ecosystem`: `DetectAllHosts`, `DetectHost`, and `FindColocated` for discovering colocated Node/PHP/Go host projects.

### Scaffolding & UI Component Management (`scaffold`, `kui`)
- `core/scaffold`: Generates projects and equips `Variant`s (`VariantApp`, `VariantStatic`, `VariantBook`, `VariantCLI`, `VariantWorker`, `VariantService`, `VariantGateway`, `VariantInfra`, `VariantTemplate`) with sentinel guards (`ErrInvalidName`, `ErrProjectExists`, `ErrNotEmpty`, `ErrConflict`).
- `core/kui`: Manages vendored `.kiw` components in `components/` (`Dir = "components"`) tracked by `.kui-lock.json` (`LockFile`), `ReadLock`, `State`, and `ErrModified`.

### Developer Server, MCP & Diagnostics (`dashboard`, `mcp`, `bench`)
- `core/dashboard`: `NewServer(root, cfg, env, host, port)` (`DefaultDashboardHost = "127.0.0.1"`, `DefaultDashboardPort = "4000"`). Exposes `Handler()` with `/`, `/api/info`, `/api/metrics`, `/api/health`, SSE telemetry (`DefaultDashboardPulseInterval = 2 * time.Second`), and mounted MCP endpoints.
- `core/mcp`: Native JSON-RPC 2.0 Model Context Protocol server (`ProtocolVersion = "2024-11-05"`, `NewServer(root, config)`, `HandleRequest(ctx, req)`) over stdio and HTTP.
- `core/bench`: `CanonicalInventory`, `CurrentEnvironment`, `FormatReportSummary`, `DefaultPolicy()`, and `CompareReports()` for benchmark regression tracking.

### Build Provenance, Versioning & Release (`buildinfo`, `release`, `version`)
- `core/buildinfo`: `ModuleVersion`, `KnownVersion`, `ResolveVersion`, `ReadProvenance`, and `LDFlags(strip bool)`.
- `core/release`: `ParseTrack`, `ParseBump`, `Bump`, `BumpTrack`, `Modules`, `DefaultRequirements`, and `FindWorkspaceRoot`.
- `core/version`: Exposes `Version` (`0.2.0`), `VersionString()`, and `EcosystemRequires`.

## Realistic Usage Examples

Inspecting a project's shape and launching an MCP server programmatically:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/mcp"
	"github.com/krewire/krew/core/shape"
)

func main() {
	res, err := shape.Detect(".")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Detected marker=%s app=%v ssg=%v\n",
		res.String(), res.Capabilities.HasApp, res.Capabilities.HasSSG)

	srv := mcp.NewServer(".", &config.Config{})
	resp := srv.HandleRequest(context.Background(), mcp.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	})
	fmt.Printf("MCP response error=%v\n", resp.Error)
}
```

## Verification

Run from `/home/krewire/workspace/krewire/krew/core`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build ./...
```
