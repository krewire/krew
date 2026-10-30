package agent

import (
	"encoding/json"

	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krewire/pkg/ai"
)

// Approver is a function called before invoking a tool to confirm permission.
type Approver func(toolName string, args json.RawMessage) bool

// Options configures an Agent instance.
type Options struct {
	Provider        ai.Provider
	Model           string
	SystemPrompt    string
	Persona         persona.Persona
	Registry        *tools.Registry
	Memory          memory.Memory
	MaxSteps        int
	Temperature     *float64
	Approver        Approver
	StrictApprovals bool // abort loop and return ErrPermissionDenied if approver rejects
	OnEvent         EventHandler
	Streaming       bool // stream LLM tokens as text_delta events

	// Subagent delegation plumbing.
	SubDepth int    // delegation levels below the root agent (0 = root)
	ThreadID string // nested thread id (parent_id:child_id) for events
}

// Option modifies Options.
type Option func(*Options)

// WithProvider sets the LLM provider.
func WithProvider(p ai.Provider) Option {
	return func(o *Options) { o.Provider = p }
}

// WithModel sets the model identifier.
func WithModel(m string) Option {
	return func(o *Options) { o.Model = m }
}

// WithPersona assigns a specialized persona and system prompt.
func WithPersona(p persona.Persona) Option {
	return func(o *Options) {
		o.Persona = p
		o.SystemPrompt = p.Instructions
	}
}

// WithSystemPrompt customizes the agent's base system prompt.
func WithSystemPrompt(sys string) Option {
	return func(o *Options) { o.SystemPrompt = sys }
}

// WithTools assigns a tools registry.
func WithTools(r *tools.Registry) Option {
	return func(o *Options) { o.Registry = r }
}

// WithMemory assigns conversation memory.
func WithMemory(m memory.Memory) Option {
	return func(o *Options) { o.Memory = m }
}

// WithMaxSteps sets the upper bound of iterative tool execution steps.
func WithMaxSteps(steps int) Option {
	return func(o *Options) { o.MaxSteps = steps }
}

// WithTemperature sets the generation temperature.
func WithTemperature(temp float64) Option {
	return func(o *Options) { o.Temperature = &temp }
}

// WithStreaming toggles token-level text_delta streaming via the provider's
// Stream endpoint when supported.
func WithStreaming(on bool) Option {
	return func(o *Options) { o.Streaming = on }
}

// WithApprover sets an approval callback for sensitive tool executions.
func WithApprover(appr Approver) Option {
	return func(o *Options) { o.Approver = appr }
}

// WithStrictApprovals controls whether tool denials abort execution with ErrPermissionDenied.
func WithStrictApprovals(strict bool) Option {
	return func(o *Options) { o.StrictApprovals = strict }
}

// WithEventHandler registers a streaming event listener.
func WithEventHandler(h EventHandler) Option {
	return func(o *Options) { o.OnEvent = h }
}
