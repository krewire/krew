# `assistant`

Import: `github.com/krewire/krew/pkg/assistant`

## Purpose

Coordinates autonomous embedded AI assistant pools, multi-workspace instances, human-in-the-loop approval decision gates, and supervised SDLC control planes.

## Main API

- `assistant.AssistantPool`: Multi-workspace assistant manager (`Chat`, `ChatStream`, `GetOrCreate`, `PendingApprovals`, `Decide`, `Remove`).
- `assistant.NewAssistantPool()` / `assistant.NewAssistantPoolWithOptions(opts PoolOptions)`: Factory functions.
- `assistant.PoolOptions`: Configures sandbox runner, approval policies, timeouts, and custom options.
- `assistant.ControlPlane`: Human-in-the-loop SDLC workflow orchestrator (`CreateTask`, `GetTask`, `ListTasks`, `DecidePlan`, `DecideReview`).
- `assistant.ApprovalPolicy`: Governs tool execution safety (`ApprovalRequired`, `ApprovalOff`).
- `assistant.CommandRunner`: Interface for executing commands in workspace sandboxes.
- `assistant.HostCommandRunner`: Default pure-Go command runner using `os/exec`.
- `assistant.CommandTool`: Governed command execution tool implementing `tools.Tool`.

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/krewire/krew/pkg/assistant"
)

func main() {
	pool := assistant.NewAssistantPool()
	reply, err := pool.Chat(context.Background(), "ws-1", ".", "Summarize project structure")
	if err != nil {
		panic(err)
	}
	fmt.Println(reply)
}
```
