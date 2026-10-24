package embed

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krewire/pkg/ai"
	"github.com/krewire/krewire/pkg/ai/providers"
)

// Options holds configuration for embedding a Krew agent.
type Options struct {
	// Workspace is the working root path for file tools and command execution.
	// Defaults to current working directory if empty.
	Workspace string

	// Persona provides system instructions and role definition.
	// Defaults to persona.CodingAssistant if nil.
	Persona *persona.Persona

	// SystemPrompt optionally overrides the persona instructions.
	SystemPrompt string

	// Provider is the LLM inference provider (Gemini, OpenAI, Anthropic, Ollama, etc.).
	// If nil, New() attempts auto-detection via environment variables.
	Provider ai.Provider

	// Model specifies the target model name (e.g. "gemini-2.5-flash", "gpt-4o", "claude-3-7-sonnet").
	Model string

	// MaxSteps is the safety limit on reasoning-action loops per turn. Default: 25.
	MaxSteps int

	// CustomTools allows adding extra tools beyond the standard workspace tool suite.
	CustomTools []tools.Tool

	// DisableDefaultTools prevents loading standard filesystem, execution, and web tools.
	DisableDefaultTools bool

	// Memory optionally overrides the default sliding-window memory buffer.
	Memory memory.Memory

	// Approver handles interactive approval gates for destructive operations.
	Approver agent.Approver

	// OnEvent receives streaming lifecycle events (thoughts, tool calls, results).
	OnEvent agent.EventHandler
}

// Instance is an embedded Krew agent ready to be invoked in-process.
type Instance struct {
	underlying *agent.Agent
	registry   *tools.Registry
	memory     memory.Memory
	persona    *persona.Persona
}

// New creates and configures an embedded Krew AI agent instance.
func New(opts Options) (*Instance, error) {
	ws := opts.Workspace
	if ws == "" {
		cwd, err := os.Getwd()
		if err == nil {
			ws = cwd
		} else {
			ws = "."
		}
	}

	p := opts.Persona
	if p == nil {
		coding := persona.CodingAssistant
		p = &coding
	}

	var reg *tools.Registry
	if !opts.DisableDefaultTools {
		reg = tools.DefaultWorkspaceTools(ws)
	} else {
		reg = tools.NewRegistry()
	}

	for _, t := range opts.CustomTools {
		reg.Register(t)
	}

	prov := opts.Provider
	if prov == nil {
		detected, err := detectProvider()
		if err != nil {
			return nil, fmt.Errorf("no AI provider supplied and auto-detection failed: %w", err)
		}
		prov = detected
	}

	sysPrompt := opts.SystemPrompt
	if sysPrompt == "" {
		sysPrompt = p.Instructions
	}

	mem := opts.Memory
	if mem == nil {
		mem = memory.NewBuffer(sysPrompt, 40)
	}

	agentOpts := []agent.Option{
		agent.WithProvider(prov),
		agent.WithPersona(*p),
		agent.WithSystemPrompt(sysPrompt),
		agent.WithMemory(mem),
		agent.WithTools(reg),
	}

	if opts.Model != "" {
		agentOpts = append(agentOpts, agent.WithModel(opts.Model))
	}
	if opts.MaxSteps > 0 {
		agentOpts = append(agentOpts, agent.WithMaxSteps(opts.MaxSteps))
	}
	if opts.Approver != nil {
		agentOpts = append(agentOpts, agent.WithApprover(opts.Approver))
	}
	if opts.OnEvent != nil {
		agentOpts = append(agentOpts, agent.WithEventHandler(opts.OnEvent))
	}

	ag, err := agent.New(agentOpts...)
	if err != nil {
		return nil, fmt.Errorf("init agent: %w", err)
	}

	return &Instance{
		underlying: ag,
		registry:   reg,
		memory:     mem,
		persona:    p,
	}, nil
}

// Run executes a single prompt to completion through the multi-turn reasoning loop.
func (i *Instance) Run(ctx context.Context, prompt string) (*agent.Response, error) {
	return i.underlying.Run(ctx, prompt)
}

// Chat continues an existing conversation with a new user message.
func (i *Instance) Chat(ctx context.Context, message string) (*agent.Response, error) {
	return i.underlying.Chat(ctx, message)
}

// Agent returns the underlying core agent engine.
func (i *Instance) Agent() *agent.Agent {
	return i.underlying
}

// Registry returns the agent's tool registry.
func (i *Instance) Registry() *tools.Registry {
	return i.registry
}

// Memory returns the agent's memory buffer.
func (i *Instance) Memory() memory.Memory {
	return i.memory
}

// detectProvider checks environment variables for known AI provider keys.
func detectProvider() (ai.Provider, error) {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return providers.NewGemini(key), nil
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return providers.NewOpenAI(key), nil
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return providers.NewAnthropic(key), nil
	}
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		return providers.NewOllama(func(o *providers.OllamaProvider) {
			o.BaseURL = host
		}), nil
	}

	return nil, errors.New("please set GEMINI_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, or OLLAMA_HOST")
}
