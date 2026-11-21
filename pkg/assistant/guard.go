package assistant

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/agent"
)

// ApprovalPolicy decides whether agent tool calls need a human decision.
type ApprovalPolicy string

const (
	// ApprovalRequired (the default) pauses every tool call that can change
	// the workspace or reach the network until a user approves it.
	ApprovalRequired ApprovalPolicy = "required"
	// ApprovalOff lets the agent run every tool without review.
	ApprovalOff ApprovalPolicy = "off"
)

// EventApprovalRequired is streamed when a tool call waits for approval.
// ToolID carries the approval ID to pass to Decide.
const EventApprovalRequired agent.EventType = "approval_required"

// ErrApprovalNotFound is returned by Decide for unknown or expired approvals.
var ErrApprovalNotFound = errors.New("approval not found")

// DefaultApprovalTimeout is how long a tool call waits before it is denied.
const DefaultApprovalTimeout = 5 * time.Minute

// readOnlyTools only read inside the workspace and never need approval.
var readOnlyTools = map[string]bool{
	"read_file":  true,
	"list_dir":   true,
	"ast_search": true,
}

// NeedsApproval reports whether a tool call must be approved under
// ApprovalRequired. Unknown tools need approval.
func NeedsApproval(toolName string) bool {
	return !readOnlyTools[toolName]
}

// ParseApprovalPolicy maps a configuration value to a policy. The empty
// string selects ApprovalRequired.
func ParseApprovalPolicy(v string) (ApprovalPolicy, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", string(ApprovalRequired):
		return ApprovalRequired, nil
	case string(ApprovalOff):
		return ApprovalOff, nil
	}
	return "", fmt.Errorf("unknown agent approval policy %q (want %q or %q)", v, ApprovalRequired, ApprovalOff)
}

// PendingApproval is a tool call waiting for a decision.
type PendingApproval struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	ToolName    string    `json:"tool_name"`
	ToolArgs    string    `json:"tool_args"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`

	decision chan bool
}

type approvals struct {
	mu      sync.Mutex
	pending map[string]*PendingApproval
	timeout time.Duration
}

func newApprovals(timeout time.Duration) *approvals {
	if timeout <= 0 {
		timeout = DefaultApprovalTimeout
	}
	return &approvals{pending: make(map[string]*PendingApproval), timeout: timeout}
}

func newApprovalID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "apr-" + hex.EncodeToString(b)
}

// wait registers a pending approval, announces it through emit and blocks
// until it is decided, ctx ends or the timeout expires (both deny).
func (a *approvals) wait(ctx context.Context, workspaceID, tool string, args json.RawMessage, emit func(agent.Event)) bool {
	now := time.Now().UTC()
	p := &PendingApproval{
		ID:          newApprovalID(),
		WorkspaceID: workspaceID,
		ToolName:    tool,
		ToolArgs:    string(args),
		CreatedAt:   now,
		ExpiresAt:   now.Add(a.timeout),
		decision:    make(chan bool, 1),
	}
	a.mu.Lock()
	a.pending[p.ID] = p
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.pending, p.ID)
		a.mu.Unlock()
	}()

	if emit != nil {
		emit(agent.Event{
			Type:      EventApprovalRequired,
			ToolID:    p.ID,
			ToolName:  tool,
			ToolArgs:  string(args),
			Content:   fmt.Sprintf("Tool %s needs approval (expires %s)", tool, p.ExpiresAt.Format(time.RFC3339)),
			Timestamp: now,
		})
	}

	timer := time.NewTimer(a.timeout)
	defer timer.Stop()
	select {
	case ok := <-p.decision:
		return ok
	case <-ctx.Done():
		return false
	case <-timer.C:
		return false
	}
}

func (a *approvals) decide(workspaceID, id string, approve bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	p, ok := a.pending[id]
	if !ok || p.WorkspaceID != workspaceID {
		return ErrApprovalNotFound
	}
	delete(a.pending, id)
	p.decision <- approve
	return nil
}

func (a *approvals) list(workspaceID string) []PendingApproval {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []PendingApproval{}
	for _, p := range a.pending {
		if p.WorkspaceID == workspaceID {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}

func (a *approvals) denyAll(workspaceID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, p := range a.pending {
		if p.WorkspaceID == workspaceID {
			delete(a.pending, id)
			p.decision <- false
		}
	}
}

// ExecutionResult holds stdout, stderr, and exit code from running a command.
type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CommandRunner executes a command within an isolated or specified workspace.
type CommandRunner interface {
	RunCommand(ctx context.Context, workspaceID, workspacePath string, command []string, timeout time.Duration) (*ExecutionResult, error)
}

// HostCommandRunner executes commands directly on the host using os/exec.
type HostCommandRunner struct{}

func (h HostCommandRunner) RunCommand(ctx context.Context, workspaceID, workspacePath string, command []string, timeout time.Duration) (*ExecutionResult, error) {
	if len(command) == 0 {
		return nil, errors.New("empty command")
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(tctx, command[0], command[1:]...)
	if workspacePath != "" {
		cmd.Dir = workspacePath
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return &ExecutionResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, err
}

// DefaultCommandTimeout and MaxCommandTimeout bound the agent's run_command.
const (
	DefaultCommandTimeout = 30 * time.Second
	MaxCommandTimeout     = 2 * time.Minute
)

// CommandTool replaces default run_command to route execution through CommandRunner.
type CommandTool struct {
	WorkspaceID   string
	WorkspacePath string
	Runner        CommandRunner
}

// NewCommandTool creates a new CommandTool.
func NewCommandTool(workspaceID, workspacePath string, runner CommandRunner) *CommandTool {
	if runner == nil {
		runner = HostCommandRunner{}
	}
	return &CommandTool{
		WorkspaceID:   workspaceID,
		WorkspacePath: workspacePath,
		Runner:        runner,
	}
}

func (t *CommandTool) Name() string { return "run_command" }

func (t *CommandTool) Description() string {
	return "Execute a shell command in the workspace directory (sandboxed, with timeout)."
}

func (t *CommandTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The exact shell command to run",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in seconds (default 30, max 120)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *CommandTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Command        string `json:"command"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if strings.TrimSpace(input.Command) == "" {
		return nil, errors.New("command is required")
	}
	timeout := DefaultCommandTimeout
	if input.TimeoutSeconds > 0 {
		timeout = time.Duration(input.TimeoutSeconds) * time.Second
	}
	timeout = min(timeout, MaxCommandTimeout)

	runner := t.Runner
	if runner == nil {
		runner = HostCommandRunner{}
	}

	res, err := runner.RunCommand(ctx, t.WorkspaceID, t.WorkspacePath, []string{"/bin/sh", "-c", input.Command}, timeout)
	resp := map[string]any{"stdout": "", "stderr": ""}
	if res != nil {
		resp["stdout"] = res.Stdout
		resp["stderr"] = res.Stderr
		resp["exit_code"] = res.ExitCode
	}
	if err != nil {
		resp["error"] = err.Error()
	}
	return resp, nil
}
