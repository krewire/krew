# Krew — Platform-Agnostic Autonomous AI Agent

[![Go Reference](https://pkg.go.dev/badge/github.com/krewire/krew.svg)](https://pkg.go.dev/github.com/krewire/krew)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Krew** is an autonomous, platform-agnostic AI agent engine designed to be embedded **literally anywhere**:
- In-process inside **Krewire Boost** or custom Go microservices.
- As a local/remote headless runner in **CI/CD pipelines** or cron jobs.
- As a background daemon via **HTTP REST + Server-Sent Events (SSE)** in Docker or Kubernetes.
- Inside IDEs (Cursor, Windsurf, Claude Desktop, Antigravity) via **Model Context Protocol (MCP)**.
- Directly in your terminal via an interactive CLI (`krew chat`).

---

## Architecture Overview

```
                      ┌────────────────────────────────────────┐
                      │            Target Application          │
                      │  (Krewire Boost, Go App, Server, IDE)  │
                      └───────────────────┬────────────────────┘
                                          │
                   ┌──────────────────────┴──────────────────────┐
                   │                 Transports                  │
                   │   Embed API  │  CLI  │  HTTP/SSE  │   MCP   │
                   └──────────────────────┬──────────────────────┘
                                          │
                               ┌──────────┴──────────┐
                               │     Krew Agent      │
                               │  (Autonomous ReAct) │
                               └──────────┬──────────┘
                                          │
             ┌────────────────────────────┼────────────────────────────┐
             │                            │                            │
   ┌─────────┴─────────┐        ┌─────────┴─────────┐        ┌─────────┴─────────┐
   │    Tool Suite     │        │   Memory Buffer   │        │   LLM Providers   │
   │  read/write/edit  │        │   Sliding Window  │        │  Gemini, OpenAI,  │
   │  exec, fetch_url  │        │     Compaction    │        │ Anthropic, Ollama │
   └───────────────────┘        └───────────────────┘        └───────────────────┘
```

---

## 1. Embed in Go (3 Lines)

You can embed Krew into any Go service or toolchain:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
)

func main() {
	// 1. Initialize embedded agent (auto-detects GEMINI_API_KEY, OPENAI_API_KEY, etc.)
	ag, err := embed.New(embed.Options{
		Workspace: ".",
		Persona:   &persona.CodingAssistant,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Execute an autonomous task
	res, err := ag.Run(context.Background(), "Inspect the codebase and add tests for pkg/utils")
	if err != nil {
		log.Fatal(err)
	}

	// 3. Print result
	fmt.Println(res.Content)
}
```

---

## 2. CLI Usage

### Interactive Terminal Session
```bash
# Launch interactive coding session
krew chat --persona=coding --workspace=.
```

### Headless Single-Shot Execution
```bash
# Run headless automation (ideal for CI/CD or background scripts)
krew run "Check all Go tests and fix any failing assertions"
```

### Remote Server Daemon
```bash
# Start HTTP/REST + Server-Sent Events daemon
krew serve --port=8080 --host=0.0.0.0
```

Available HTTP endpoints:
- `GET /health`: Health check status.
- `POST /api/agent/run`: Single-turn autonomous execution (supports `{"stream": true}` for SSE).
- `POST /api/agent/chat`: Multi-turn conversational execution.

### Model Context Protocol (MCP)
Run Krew as a stdio MCP server for Cursor, Windsurf, or Claude Desktop:
```bash
krew mcp --workspace=/path/to/project
```

#### Claude Desktop Configuration (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "krew": {
      "command": "krew",
      "args": ["mcp", "--workspace", "/path/to/workspace"],
      "env": {
        "GEMINI_API_KEY": "your-key-here"
      }
    }
  }
}
```

---

## 3. Built-In Personas

Krew provides battle-tested system personas tailored for specific operational environments:

| Persona | Role | Key Capabilities |
|---|---|---|
| `persona.CodingAssistant` | Senior software engineer | Code exploration, refactoring, atomic diffs, test execution. |
| `persona.DevOpsRunner` | Site reliability & DevOps | Server diagnostics, Docker/systemd automation, network debugging. |
| `persona.AutonomousWorker` | Headless worker | Step-by-step task queue resolution with structured summaries. |
| `persona.GeneralAssistant` | General companion | Explanations, Q&A, and workflow assistance. |

---

## 4. Built-In Sandboxed Tool Suite

All workspace tools respect path sandboxing to protect against directory traversal attacks (`SafePath`):

- `read_file(path)`: Inspect file contents up to size limits.
- `write_file(path, content, overwrite)`: Atomically create or overwrite files.
- `edit_file(path, target, replacement)`: Precise exact-match string replacement.
- `list_dir(path, recursive)`: File tree inventory.
- `run_command(command, timeout_seconds)`: Execute commands with timeout enforcement.
- `fetch_url(url, method, timeout_seconds)`: Query external APIs or scrape documentation.

---

## 5. Integrating with Krewire Boost

Krewire Boost leverages `krew` as its default coding assistant runtime:

```go
import (
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
)

func RegisterBoostAssistant(workspace string) (*embed.Instance, error) {
	return embed.New(embed.Options{
		Workspace: workspace,
		Persona:   &persona.CodingAssistant,
		MaxSteps:  30,
	})
}
```

---

## 6. Supported LLM Providers

Krew uses `github.com/krewire/krewire/pkg/ai` as its unified inference provider abstraction:
- **Google Gemini**: set `GEMINI_API_KEY` (defaults to Gemini 2.5 Flash / Pro).
- **OpenAI**: set `OPENAI_API_KEY` (GPT-4o, o3-mini).
- **Anthropic**: set `ANTHROPIC_API_KEY` (Claude 3.7 Sonnet).
- **Ollama**: set `OLLAMA_HOST` (e.g. `http://localhost:11434`) for 100% offline local inference.

---

## 7. Strategic Roadmap

Krew follows a multi-year engineering roadmap (2026–2028):
- **2026**: Core ReAct Engine, Multi-Protocol Transports (Embed, CLI, SSE Daemon, MCP), Sandboxed Tools, and in-CLI companion (`kiw krew`). *(Delivered)*
- **2027**: 3-Tier Hierarchical Memory (Working, Episodic, Semantic Vector RAG), OpenTelemetry step tracing, cloud-hosted autonomous runners, and cryptographic agent identity (A-Identity) with RBAC.
- **2028**: Agent-to-Agent (A2A) peer-to-peer swarm mesh protocol, autonomous SRE remediation, sovereign air-gapped local inference engine (Ollama/GGUF/Metal/CUDA), and v1.0.0 LTS.

For detailed milestone specifications, see [Krew Strategic Roadmap](../internal/docs/roadmaps/krew-roadmap.md).

---

## License

Krew is open source software licensed under the [MIT License](LICENSE).
Part of the [Krewire](https://krewire.com) ecosystem.
