package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/krewire/krewire/pkg/ai"
)

// Registry manages a thread-safe collection of invokable tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry initializes an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds one or more tools to the registry.
func (r *Registry) Register(tools ...Tool) *Registry {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range tools {
		if t != nil && t.Name() != "" {
			r.tools[t.Name()] = t
		}
	}
	return r
}

// Get returns a tool by name if present.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// ToAITools transforms all registered tools into []ai.Tool for LLM requests.
func (r *Registry) ToAITools() []ai.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ai.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, ToAITool(t))
	}
	return out
}

// Execute invokes a registered tool with JSON arguments.
func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) (any, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return t.Execute(ctx, args)
}
