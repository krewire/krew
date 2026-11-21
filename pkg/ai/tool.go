package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// NewTool creates a Tool definition by reflecting on a Go function.
// Supported function signatures:
//   - func(ctx context.Context, args T) (Result, error)
//   - func(ctx context.Context, args *T) (Result, error)
//   - func(args T) (Result, error)
//   - func(args *T) (Result, error)
//   - func(ctx context.Context) (Result, error)
//   - func() (Result, error)
func NewTool(name, description string, fn any) (Tool, error) {
	if name == "" {
		return Tool{}, errors.New("ai: tool name is required")
	}
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		return Tool{}, fmt.Errorf("ai: expected func, got %T", fn)
	}

	typ := val.Type()
	numIn := typ.NumIn()
	numOut := typ.NumOut()

	if numOut < 1 || numOut > 2 {
		return Tool{}, fmt.Errorf("ai: tool func must return (result) or (result, error), got %d return values", numOut)
	}

	var hasCtx bool
	var argType reflect.Type

	for i := 0; i < numIn; i++ {
		inT := typ.In(i)
		if inT.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			if i != 0 {
				return Tool{}, errors.New("ai: context.Context must be the first parameter")
			}
			hasCtx = true
		} else {
			if argType != nil {
				return Tool{}, errors.New("ai: tool func accepts at most one argument struct besides context")
			}
			argType = inT
		}
	}

	var params map[string]any
	if argType != nil {
		params = GenerateSchema(argType)
	} else {
		params = map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}

	handler := func(ctx context.Context, args []byte) (any, error) {
		var inVals []reflect.Value
		if hasCtx {
			inVals = append(inVals, reflect.ValueOf(ctx))
		}

		if argType != nil {
			elemType := argType
			isPtr := false
			if elemType.Kind() == reflect.Pointer {
				elemType = elemType.Elem()
				isPtr = true
			}

			argVal := reflect.New(elemType)
			if len(args) > 0 && string(args) != "{}" {
				if err := json.Unmarshal(args, argVal.Interface()); err != nil {
					return nil, fmt.Errorf("ai: unmarshal tool arguments for %q: %w", name, err)
				}
			}

			if isPtr {
				inVals = append(inVals, argVal)
			} else {
				inVals = append(inVals, argVal.Elem())
			}
		}

		outVals := val.Call(inVals)
		if numOut == 2 {
			if !outVals[1].IsNil() {
				err, ok := outVals[1].Interface().(error)
				if ok {
					return nil, err
				}
				return nil, fmt.Errorf("%v", outVals[1].Interface())
			}
		}
		return outVals[0].Interface(), nil
	}

	return Tool{
		Name:        name,
		Description: description,
		Parameters:  params,
		Handler:     handler,
	}, nil
}

// ToolRegistry manages a collection of tools available for invocation.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewToolRegistry creates an empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

// RegisterFunc creates and registers a tool from a Go function.
func (r *ToolRegistry) RegisterFunc(name, description string, fn any) error {
	t, err := NewTool(name, description, fn)
	if err != nil {
		return err
	}
	r.Register(t)
	return nil
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Tools returns all registered tools.
func (r *ToolRegistry) Tools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// Execute invokes a registered tool with given JSON argument payload.
func (r *ToolRegistry) Execute(ctx context.Context, name string, args []byte) (any, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("ai: unknown tool %q", name)
	}
	if t.Handler == nil {
		return nil, fmt.Errorf("ai: tool %q has no execution handler", name)
	}
	return t.Handler(ctx, args)
}
