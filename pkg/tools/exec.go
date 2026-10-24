package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// RunCommandTool allows executing shell commands with strict timeouts.
type RunCommandTool struct {
	Workspace      string
	DefaultTimeout time.Duration
	AllowEnv       []string
}

func (t *RunCommandTool) Name() string { return "run_command" }
func (t *RunCommandTool) Description() string {
	return "Execute a shell command in the workspace directory with timeout."
}
func (t *RunCommandTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The exact shell command to run",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in seconds (default 30s)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *RunCommandTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Command        string `json:"command"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	timeout := t.DefaultTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if input.TimeoutSeconds > 0 {
		timeout = time.Duration(input.TimeoutSeconds) * time.Second
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// #nosec G204 — explicitly designed for AI Agent tool execution
	cmd := exec.CommandContext(cmdCtx, "sh", "-c", input.Command)
	if t.Workspace != "" {
		cmd.Dir = t.Workspace
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	outStr := stdout.String()
	errStr := stderr.String()

	resp := map[string]any{
		"stdout": outStr,
		"stderr": errStr,
	}

	if err != nil {
		resp["error"] = err.Error()
		if cmdCtx.Err() == context.DeadlineExceeded {
			resp["error"] = fmt.Sprintf("command timed out after %v", timeout)
		}
	} else {
		resp["exit_code"] = 0
	}

	return resp, nil
}
