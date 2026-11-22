# Getting Started with Krew

This guide explains how to install, build, and run **Krew** locally as the developer toolchain CLI.

## Prerequisites

- **Go**: Version 1.27.1 or newer.
- **Git**: Configured for your GitHub account.
- **Docker** (optional): For running local containerized services (e.g. `krewire-db`).

## Local Setup

1. **Clone the repository** (or navigate to `krew/` in the workspace root):
   ```bash
   git clone https://github.com/krewire/krew.git
   cd krew
   ```

2. **Synchronize Go dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the `krew` CLI**:
   ```bash
   go build -o bin/krew ./cmd/krew
   ```

4. **Verify the installation**:
   ```bash
   ./bin/krew version
   ```

## Common Workflows

### 1. Scaffold a New Project

```bash
# Create a new web application
./bin/krew new my-app --app

# Create a documentation site
./bin/krew new my-docs --site
```

### 2. Live Development Server

```bash
cd my-app
krew dev
```

### 3. Database Migrations (Laravel-Style in Go)

Generate a new typed Go migration:

```bash
krew make migration create_users_table
```

This creates `database/migrations/<timestamp>_create_users_table.go`:

```go
package migrations

import (
	"context"

	"github.com/krewire/krewire/pkg/orm/migrate"
	"github.com/krewire/krewire/pkg/orm/schema"
)

func init() {
	migrate.Register(&migrate.Migration{
		Name: "create_users_table",
		Up: func(ctx context.Context, s *schema.Builder) error {
			return s.Create("users", func(t *schema.Blueprint) {
				t.Increments("id")
				t.String("name", 255)
				t.String("email", 255).Unique()
				t.Timestamps()
			})
		},
		Down: func(ctx context.Context, s *schema.Builder) error {
			return s.DropIfExists("users")
		},
	})
}
```

Apply migrations:

```bash
krew migrate up
```

### 4. Developer Authentication

Authenticate your developer session with Krewire Auth:

```bash
krew auth login --token <token>
krew auth status
krew auth logout
```

### 5. Running as an MCP Server

Start Krew in Model Context Protocol (MCP) server mode over standard input/output (stdio):

```bash
./bin/krew mcp --stdio
```

This command allows modern AI IDEs (Antigravity, Cursor, Claude Desktop) to connect to Krew and consume its tool capabilities.
