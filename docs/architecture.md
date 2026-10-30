# Krew — Architecture & Subsystems Design

This document details the internal design and core packages within `github.com/krewire/krew`.

## System Overview

```text
                                  Krew Agent Architecture
                                             │
         ┌───────────────────┬───────────────┴───────────────┬───────────────────┐
         ▼                   ▼                               ▼                   ▼
    [pkg/persona]       [pkg/memory]                    [pkg/tools]        [pkg/transport]
    - System Prompts    - Transcript Log                - Tool Registry    - Stdio (MCP)
    - Tool Policies     - Context Compactor             - Schema Validation- HTTP / SSE
    - Role Attributes   - Multi-Turn History            - Execution Guards - CLI (pty)
         │                   │                               │                   │
         └───────────────────┼───────────────────────────────┼───────────────────┘
                             ▼                               ▼
                      [pkg/agent]                     [pkg/embed]
                      - Autonomous Loop               - In-Process Embed API
                      - LLM Provider Interface        - Go Native Client
```

## Package Overview

- **`pkg/agent`**:
  - Contains the core iterative execution loop: `Prompt` $\rightarrow$ `LLM Call` $\rightarrow$ `Tool Resolution` $\rightarrow$ `Execution` $\rightarrow$ `Observation` $\rightarrow$ `Completion`.
  - Implements step limits, timeout controls, and graceful cancellation.

- **`pkg/memory`**:
  - Manages conversation turns, token consumption tracking, and context window sliding.
  - Implements deterministic compaction strategies to avoid exceeding model context limits.

- **`pkg/tools`**:
  - Defines the `Tool` interface (`Name()`, `Description()`, `Parameters()`, `Execute()`).
  - Implements safety policies, sandboxing constraints, and parameter validation against JSON Schemas.

- **`pkg/persona`**:
  - Declarative personas (`CodingAssistant`, `Researcher`, `Reviewer`, `Orchestrator`).
  - Sets tone, operational constraints, and allowed tool subsets for each persona.

- **`pkg/transport`**:
  - Provides transport adapters:
    - **MCP (Model Context Protocol)**: Exposes tool declarations and execution over JSON-RPC 2.0 stdio.
    - **HTTP/SSE**: Streams tokens, reasoning blocks, and tool results over HTTP Server-Sent Events.
    - **Terminal**: Interactive CLI driver for human user communication.

- **`pkg/embed`**:
  - The primary public Go API for embedding Krew directly in external Go applications with zero external dependencies.
