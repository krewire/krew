# Getting Started with Krew

This guide explains how to install, build, and run **Krew** locally as a CLI tool or as an embedded Go package.

## Prerequisites

- **Go**: Version 1.27.1 or newer.
- **Git**: Configured for your GitHub account.
- **LLM API Key** (optional for live AI queries): An API key for an LLM provider (e.g., Anthropic, OpenAI, or local Ollama endpoint). Mock providers are built-in for local tests.

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
   # Output: Krew AI Agent v0.1.0
   ```

## Running the Interactive CLI

Start an interactive chat session:

```bash
./bin/krew chat
```

Run a one-shot instruction:

```bash
./bin/krew run "analyze git repository status and summarize recent commits"
```

## Running as an MCP Server

Start Krew in Model Context Protocol (MCP) server mode over standard input/output (stdio):

```bash
./bin/krew mcp
```

This command allows modern AI IDEs (Antigravity, Cursor, Claude Desktop) to connect to Krew and consume its tool capabilities.

## Embedding Krew in Your Go Application

Import `github.com/krewire/krew/pkg/embed` directly:

```go
package main

import (
	"context"
	"fmt"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
)

func main() {
	agent, err := embed.New(embed.Config{
		Persona: persona.CodingAssistant,
	})
	if err != nil {
		panic(err)
	}

	res, err := agent.Execute(context.Background(), "Summarize current directory")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
```

## Running Tests

Run the full automated test suite:

```bash
go test -v ./...
```
