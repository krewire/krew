package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/tools"
)

// PoolOptions configures an AssistantPool. Zero values select safe defaults.
type PoolOptions struct {
	// Runner executes the agent's run_command tool. Defaults to HostCommandRunner.
	Runner CommandRunner
	// Approval selects whether mutating tool calls (everything except
	// read_file, list_dir and ast_search) wait for a human decision.
	// Defaults to ApprovalRequired.
	Approval ApprovalPolicy
	// ApprovalTimeout bounds the wait for a decision; expiry denies the call.
	// Defaults to DefaultApprovalTimeout.
	ApprovalTimeout time.Duration
	// Configure, when set, may adjust the embed options before an agent is
	// created (for example to inject an LLM provider in tests).
	Configure func(*embed.Options)
}

type runState struct {
	ctx  context.Context
	emit func(agent.Event)
}

// AssistantPool coordinates embedded Krew AI assistants across virtual workspaces.
type AssistantPool struct {
	opts         PoolOptions
	approvals    *approvals
	controlPlane *ControlPlane

	mu        sync.RWMutex
	instances map[string]*embed.Instance
	runLocks  map[string]*sync.Mutex
	runs      map[string]*runState
}

// NewAssistantPool creates an assistant pool with default options:
// host runner and approvals required.
func NewAssistantPool() *AssistantPool {
	return NewAssistantPoolWithOptions(PoolOptions{})
}

// NewAssistantPoolWithOptions creates an assistant pool.
func NewAssistantPoolWithOptions(opts PoolOptions) *AssistantPool {
	if opts.Runner == nil {
		opts.Runner = HostCommandRunner{}
	}
	if opts.Approval == "" {
		opts.Approval = ApprovalRequired
	}
	return &AssistantPool{
		opts:         opts,
		approvals:    newApprovals(opts.ApprovalTimeout),
		controlPlane: NewControlPlane(opts.Runner),
		instances:    make(map[string]*embed.Instance),
		runLocks:     make(map[string]*sync.Mutex),
		runs:         make(map[string]*runState),
	}
}

// ControlPlane returns the human-in-the-loop engineering control plane.
func (p *AssistantPool) ControlPlane() *ControlPlane {
	return p.controlPlane
}

// ApprovalPolicy returns the pool's approval policy.
func (p *AssistantPool) ApprovalPolicy() ApprovalPolicy {
	return p.opts.Approval
}

// GetOrCreate returns an embedded Krew agent instance bound to the workspace root.
func (p *AssistantPool) GetOrCreate(workspaceID, wsPath string) (*embed.Instance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if inst, ok := p.instances[workspaceID]; ok {
		return inst, nil
	}

	astTool := tools.NewASTSearchTool(wsPath)
	cmdTool := NewCommandTool(workspaceID, wsPath, p.opts.Runner)

	opts := embed.Options{
		Workspace:   wsPath,
		Persona:     &persona.CodingAssistant,
		MaxSteps:    30,
		CustomTools: []tools.Tool{astTool, cmdTool},
		Approver:    p.approver(workspaceID),
	}
	if p.opts.Configure != nil {
		p.opts.Configure(&opts)
	}
	inst, err := embed.New(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI assistant for workspace %s: %w", workspaceID, err)
	}

	p.instances[workspaceID] = inst
	return inst, nil
}

// approver returns the Krew approval callback for a workspace.
func (p *AssistantPool) approver(workspaceID string) agent.Approver {
	return func(toolName string, args json.RawMessage) bool {
		if p.opts.Approval == ApprovalOff || !NeedsApproval(toolName) {
			return true
		}
		ctx := context.Background()
		var emit func(agent.Event)
		p.mu.RLock()
		if run := p.runs[workspaceID]; run != nil {
			ctx, emit = run.ctx, run.emit
		}
		p.mu.RUnlock()
		return p.approvals.wait(ctx, workspaceID, toolName, args, emit)
	}
}

// beginRun serialises agent runs per workspace and records the context and
// event sink used by the approval gate. The returned func ends the run.
func (p *AssistantPool) beginRun(ctx context.Context, workspaceID string, emit func(agent.Event)) func() {
	p.mu.Lock()
	lock, ok := p.runLocks[workspaceID]
	if !ok {
		lock = &sync.Mutex{}
		p.runLocks[workspaceID] = lock
	}
	p.mu.Unlock()

	lock.Lock()
	p.mu.Lock()
	p.runs[workspaceID] = &runState{ctx: ctx, emit: emit}
	p.mu.Unlock()
	return func() {
		p.mu.Lock()
		delete(p.runs, workspaceID)
		p.mu.Unlock()
		lock.Unlock()
	}
}

// Chat executes a synchronous conversational turn with the AI assistant in the workspace.
// Tool calls needing approval block until decided through Decide.
func (p *AssistantPool) Chat(ctx context.Context, workspaceID, wsPath, prompt string) (string, error) {
	inst, err := p.GetOrCreate(workspaceID, wsPath)
	if err != nil {
		return "", err
	}

	end := p.beginRun(ctx, workspaceID, nil)
	defer end()
	res, err := inst.Chat(ctx, prompt)
	if err != nil {
		return "", err
	}
	return res.Content, nil
}

// ChatStream executes a streaming conversational turn, broadcasting thinking,
// tool calls, approval requests and text deltas.
func (p *AssistantPool) ChatStream(ctx context.Context, workspaceID, wsPath, prompt string, onEvent agent.EventHandler) (*agent.Response, error) {
	inst, err := p.GetOrCreate(workspaceID, wsPath)
	if err != nil {
		return nil, err
	}

	end := p.beginRun(ctx, workspaceID, onEvent)
	defer end()
	return inst.RunWithHandler(ctx, prompt, onEvent)
}

// PendingApprovals lists the tool calls waiting for a decision in a workspace.
func (p *AssistantPool) PendingApprovals(workspaceID string) []PendingApproval {
	return p.approvals.list(workspaceID)
}

// Decide approves or denies a pending tool call.
func (p *AssistantPool) Decide(workspaceID, approvalID string, approve bool) error {
	return p.approvals.decide(workspaceID, approvalID, approve)
}

// SetInstance registers a pre-configured or mock instance for a workspace.
// The caller owns that instance's tools and approval policy.
func (p *AssistantPool) SetInstance(workspaceID string, inst *embed.Instance) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.instances[workspaceID] = inst
}

// Remove releases the agent instance when a workspace is stopped or destroyed
// and denies its pending approvals.
func (p *AssistantPool) Remove(workspaceID string) {
	p.approvals.denyAll(workspaceID)
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.instances, workspaceID)
}
