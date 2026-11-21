# `boost`

Import: `github.com/krewire/krew/pkg/boost`

## Purpose

Installer for the Krewire Boost AI-agent setup: `AGENTS.md`, `opencode.json`, and the `.agents/` preset (agents, commands, skills, rules).

## Main API

- `boost.Install(target string, opts ...Option) ([]string, error)`: Copies the Boost template into `target` and returns the written paths, sorted. Refuses to overwrite managed files unless forced.
- `boost.WithForce() Option`: Permits overwriting existing managed files.
- `boost.WithDryRun() Option`: Validates and reports the would-be writes without touching the target.
- `boost.Managed() []string`: Canonical top-level paths the installer manages.
- `boost.Template fs.FS`: The embedded template tree under `templates/boost`, readable directly.
- `boost.ErrTargetMissing` / `boost.ErrConflicts`: Typed errors for an unusable target and for existing managed files.

## Example

```go
package main

import (
	"log"

	"github.com/krewire/krew/pkg/boost"
)

func main() {
	written, err := boost.Install(".", boost.WithForce())
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range written {
		log.Println("wrote", path)
	}
}
```

## Design boundary

Driven by `krew boost install` during project initialization. The package reads the template tree from disk or environment.
