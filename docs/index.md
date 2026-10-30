# Krew — Developer & Contributor Documentation

Welcome to the **Krew** repository documentation. This repository (`github.com/krewire/krew`) implements the platform-agnostic, embeddable autonomous AI agent engine for the Krewire ecosystem.

## Documentation Index

- [Getting Started](getting-started.md): Installation, local setup, building the `krew` CLI, running in embedded Go programs, and local testing.
- [Architecture & Design](architecture.md): Core agent execution loop (`pkg/agent`), memory management (`pkg/memory`), tool registry (`pkg/tools`), personas (`pkg/persona`), and transport adapters (`pkg/transport`).
- [Contributing Guide](contributing.md): Contribution guidelines, testing requirements, quality gates, and code standards.

## Core Features & Interfaces

- **Platform-Agnostic Embeddability**: Embed directly into any Go microservice with zero framework lock-in via `pkg/embed`.
- **Stand-Alone CLI**: Interactive terminal assistant (`krew chat`) and headless command-line execution (`krew run`).
- **Model Context Protocol (MCP)**: Native support for Anthropic's Model Context Protocol, enabling Krew to connect seamlessly to IDEs (Cursor, Claude Desktop, Antigravity) as an MCP server.
- **HTTP / Server-Sent Events (SSE)**: Microservice mode streaming real-time agent thoughts and tool executions over HTTP.
- **Pluggable Personas & Tooling**: Declarative persona system with structured tool registries and safety policy gating.
