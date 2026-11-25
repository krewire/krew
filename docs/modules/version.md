# `version` (`github.com/krewire/krew/version`)

## Purpose & Design Role

`version` is the standalone semantic version and ecosystem compatibility declaration module for `krew`. Separating version metadata into its own leaf module allows external tools, release automation, and compatibility checkers (`krew compat`, `core/release`) to inspect `krew`'s release version and minimum `krewire` foundation requirements without pulling in the entire `core` command tree.

## Architectural Position

- **Layer**: Layer 3 (Metadata Leaf Module within `github.com/krewire/krew`).
- **Module**: `github.com/krewire/krew/version` (`version/go.mod`).
- **Dependencies**: Imports `github.com/krewire/krewire/version` (Layer 1 kernel primitive).

## Core Concepts & APIs

Defined in `version/version.go`:

- `Version` (`version.Version`): Semantic release version of `krew`, currently `0.2.0` (`version.MustParseVersion("0.2.0")`).
- `EcosystemRequires` (`map[string]version.Version`): Declares the minimum compatible version of each ecosystem dependency (`"krewire": version.MustParseVersion("0.2.0")`).
- `VersionString() string`: Returns `Version.String()` (e.g., `"0.2.0"`).

## Realistic Usage Examples

```go
package main

import (
	"fmt"

	krewver "github.com/krewire/krew/version"
)

func main() {
	fmt.Println("krew version:", krewver.VersionString())
	for mod, minVer := range krewver.EcosystemRequires {
		fmt.Printf("requires %s >= %s\n", mod, minVer.String())
	}
}
```

## Verification

Run from `/home/krewire/workspace/krewire/krew/version`:

```bash
gofmt -l .
go vet ./...
go test -v ./...
CGO_ENABLED=0 go build ./...
```
