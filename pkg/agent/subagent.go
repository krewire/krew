package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/tools"
)

// MaxSubagentDepth bounds recursive delegation: a root agent may delegate to
// one level, which may delegate one further level only.
const MaxSubagentDepth = 2

// DefaultSubagentSteps is the step budget applied when an invocation does not
// supply max_steps.
const DefaultSubagentSteps = 15

// subagentToolName is the registry key for the delegation tool.
const subagentToolName = "invoke_subagent"

// SubAgentTool implements tools.Tool by delegating a scoped task to a fresh,
// isolated child Agent that inherits the parent's provider, model, tools, and
// workspace, but holds its own memory buffer.
type SubAgentTool struct {
	parent   *Agent
	depth    int           // delegation levels already below the root agent
	sequence *atomic.Int64 // unique child numbering per delegation root
}

// NewSubAgentTool builds the delegation tool for parent. depth counts how
// many delegation levels already separate the tool from the root agent.
func NewSubAgentTool(parent *Agent, depth int) *SubAgentTool {
	return &SubAgentTool{
		parent:   parent,
		depth:    depth,
		sequence: &atomic.Int64{},
	}
}

// EnableSubAgents registers the invoke_subagent tool on a's registry so the
// agent may delegate specialised work to nested child agents.
func EnableSubAgents(a *Agent) error {
	if a == nil || a.opts.Registry == nil {
		return fmt.Errorf("agent: cannot enable subagents without a registry")
	}
	return a.RegisterTool(NewSubAgentTool(a, 0))
}

// personaAliases maps short persona keys used in tool arguments to built-ins.
var personaAliases = map[string]persona.Persona{
	"coding":             persona.CodingAssistant,
	"devops":             persona.DevOpsRunner,
	"worker":             persona.AutonomousWorker,
	"research":           persona.ResearchAssistant,
	"general":            persona.GeneralAssistant,
	"coding-assistant":   persona.CodingAssistant,
	"devops-runner":      persona.DevOpsRunner,
	"autonomous-worker":  persona.AutonomousWorker,
	"research-assistant": persona.ResearchAssistant,
	"general-assistant":  persona.GeneralAssistant,
}

func (t *SubAgentTool) Name() string { return subagentToolName }

func (t *SubAgentTool) Description() string {
	return fmt.Sprintf("Delegate a self-contained task to a specialised sub-agent (persona: coding, devops, worker, research) that executes up to max_steps reasoning steps with its own context window and returns final content plus an execution summary. Nested delegation is capped at %d levels.", MaxSubagentDepth)
}

func (t *SubAgentTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "Self-contained task description the sub-agent executes independently",
			},
			"persona": map[string]any{
				"type":        "string",
				"description": "Sub-agent role: coding, devops, worker, or research (default coding)",
			},
			"max_steps": map[string]any{
				"type":        "integer",
				"description": fmt.Sprintf("Maximum reasoning steps for the child agent (default %d)", DefaultSubagentSteps),
			},
		},
		"required": []string{"prompt"},
	}
}

func (t *SubAgentTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Prompt   string `json:"prompt"`
		Persona  string `json:"persona"`
		MaxSteps int    `json:"max_steps"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Recursion guard: depth counts the levels between the root agent and this
	// delegation. At MaxSubagentDepth the task must execute directly.
	nextDepth := t.depth + 1
	if nextDepth > MaxSubagentDepth {
		return nil, fmt.Errorf("subagent depth limit reached (max %d levels); execute the task yourself instead of delegating", MaxSubagentDepth)
	}

	childNumber := t.sequence.Add(1)
	childThread := fmt.Sprintf("%s%d:", t.parent.opts.ThreadID, childNumber)

	child, err := t.spawnChild(input, nextDepth, childThread)
	if err != nil {
		return nil, err
	}

	resp, runErr := child.Run(ctx, input.Prompt)
	if runErr != nil {
		return nil, runErr
	}

	return map[string]any{
		"content":        resp.Content,
		"steps_taken":    resp.StepsTaken,
		"tool_calls_run": resp.ToolCallsRun,
		"usage":          resp.Usage,
		"thread_id":      strings.TrimSuffix(childThread, ":"),
	}, nil
}

// spawnChild builds an isolated child agent for one delegation: fresh memory
// buffer, requested persona, selected step budget, its own registry without
// the inherited delegation tool (replaced by a depth-aware copy), and event
// forwarding under nested thread ids.
func (t *SubAgentTool) spawnChild(input struct {
	Prompt   string `json:"prompt"`
	Persona  string `json:"persona"`
	MaxSteps int    `json:"max_steps"`
}, nextDepth int, childThread string) (*Agent, error) {
	child := t.parent.Clone(
		withPersonaResolved(input.Persona),
		withFreshMemory(),
		WithMaxSteps(stepBudget(input.MaxSteps)),
		withIsolatedRegistry(nextDepth),
		withThreadID(childThread),
	)
	if err := child.selfDelegate(); err != nil {
		return nil, err
	}

	// Forward child events to the parent's stream with nested thread ids
	// (parent_id:child_id), preserving real-time telemetry. Deeper lineage
	// already stamps its own thread id, so only unset ones are filled here.
	parent := t.parent
	child.opts.OnEvent = func(ev Event) {
		if ev.ThreadID == "" {
			ev.ThreadID = strings.TrimSuffix(childThread, ":")
		}
		parent.emit(ev)
	}
	return child, nil
}

// withIsolatedRegistry clones the parent registry, dropping the inherited
// delegation tool; subagent wiring is re-added with the elevated depth.
func withIsolatedRegistry(nextDepth int) Option {
	return func(o *Options) {
		fresh := tools.NewRegistry()
		for _, tool := range o.Registry.List() {
			if tool.Name() == subagentToolName {
				continue
			}
			fresh.Register(tool)
		}
		o.Registry = fresh
		o.SubDepth = nextDepth
	}
}

func withPersonaResolved(key string) Option {
	return func(o *Options) {
		trimmed := strings.ToLower(strings.TrimSpace(key))
		if trimmed == "" {
			return
		}
		if p, ok := personaAliases[trimmed]; ok {
			o.Persona = p
			o.SystemPrompt = p.Instructions
		}
	}
}

func withFreshMemory() Option {
	return func(o *Options) {
		o.Memory = memory.NewBuffer(o.SystemPrompt, 100)
	}
}

func withThreadID(id string) Option {
	return func(o *Options) { o.ThreadID = id }
}

// selfDelegate wires the child's registry with its own depth-aware delegation
// tool. Called after Clone so the binding references the child agent itself.
func (a *Agent) selfDelegate() error {
	a.opts.Registry.Register(NewSubAgentTool(a, a.opts.SubDepth))
	return nil
}

func stepBudget(requested int) int {
	if requested > 0 {
		return requested
	}
	return DefaultSubagentSteps
}
