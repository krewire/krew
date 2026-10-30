package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krewire/pkg/ai"
)

var (
	ErrMaxStepsExceeded = errors.New("agent: maximum reasoning steps exceeded")
	ErrNoProvider       = errors.New("agent: missing ai provider")
)

// Agent is an autonomous, multi-turn AI reasoning engine with tool-calling capabilities.
type Agent struct {
	opts Options
}

// New constructs an initialized Agent.
func New(opts ...Option) (*Agent, error) {
	cfg := Options{
		Persona:  persona.CodingAssistant,
		MaxSteps: 25,
		Registry: tools.NewRegistry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Provider == nil {
		return nil, ErrNoProvider
	}
	if cfg.Model == "" {
		cfg.Model = "default"
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = cfg.Persona.Instructions
	}
	if cfg.Memory == nil {
		cfg.Memory = memory.NewBuffer(cfg.SystemPrompt, 100)
	}

	return &Agent{opts: cfg}, nil
}

// Response holds the outcome of an agent run.
type Response struct {
	Content      string        `json:"content"`
	StepsTaken   int           `json:"steps_taken"`
	ToolCallsRun int           `json:"tool_calls_run"`
	FinishReason string        `json:"finish_reason"`
	Usage        ai.TokenUsage `json:"usage"`
}

// Run executes the agent autonomously on a given user prompt.
func (a *Agent) Run(ctx context.Context, prompt string) (*Response, error) {
	a.opts.Memory.Add(ai.UserMessage(prompt))
	return a.stepLoop(ctx)
}

// RunWithHandler executes the agent with an additional per-request event listener.
func (a *Agent) RunWithHandler(ctx context.Context, prompt string, h EventHandler) (*Response, error) {
	if h == nil {
		return a.Run(ctx, prompt)
	}

	clone := *a
	orig := a.opts.OnEvent
	clone.opts.OnEvent = func(e Event) {
		if orig != nil {
			orig(e)
		}
		h(e)
	}
	return clone.Run(ctx, prompt)
}

// Clone creates a copy of the agent with modified options.
func (a *Agent) Clone(opts ...Option) *Agent {
	cp := *a
	for _, opt := range opts {
		opt(&cp.opts)
	}
	return &cp
}

// Chat continues an existing conversation with a new user message.
func (a *Agent) Chat(ctx context.Context, message string) (*Response, error) {
	return a.Run(ctx, message)
}

// Memory returns the conversation memory.
func (a *Agent) Memory() memory.Memory {
	return a.opts.Memory
}

// Tools returns the registered tool collection.
func (a *Agent) Tools() *tools.Registry {
	return a.opts.Registry
}

func (a *Agent) emit(e Event) {
	if a.opts.OnEvent != nil {
		if e.Timestamp.IsZero() {
			e.Timestamp = time.Now()
		}
		a.opts.OnEvent(e)
	}
}

func (a *Agent) stepLoop(ctx context.Context) (*Response, error) {
	a.emit(Event{Type: EventStart, Step: 0})

	var totalUsage ai.TokenUsage
	toolCallsRun := 0

	for step := 1; step <= a.opts.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		req := ai.CompletionRequest{
			Model:       a.opts.Model,
			Messages:    a.opts.Memory.Messages(),
			System:      a.opts.SystemPrompt,
			Temperature: a.opts.Temperature,
			Tools:       a.opts.Registry.ToAITools(),
		}

		resp, err := a.complete(ctx, req, step)
		if err != nil {
			return nil, fmt.Errorf("llm completion error at step %d: %w", step, err)
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		if resp.Content != "" {
			a.emit(Event{
				Type:    EventThought,
				Step:    step,
				Content: resp.Content,
			})
		}

		// If no tools were called, this is the final assistant response!
		if len(resp.ToolCalls) == 0 {
			a.opts.Memory.Add(ai.AssistantMessage(resp.Content))
			a.emit(Event{
				Type:    EventFinish,
				Step:    step,
				Content: resp.Content,
				Usage:   totalUsage,
			})
			return &Response{
				Content:      resp.Content,
				StepsTaken:   step,
				ToolCallsRun: toolCallsRun,
				FinishReason: resp.FinishReason,
				Usage:        totalUsage,
			}, nil
		}

		// Model requested tool calls
		assistantMsg := ai.Message{
			Role:      ai.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		a.opts.Memory.Add(assistantMsg)

		for _, call := range resp.ToolCalls {
			toolCallsRun++
			a.emit(Event{
				Type:     EventToolCall,
				Step:     step,
				ToolID:   call.ID,
				ToolName: call.Name,
				ToolArgs: call.Arguments,
			})

			// Check optional approval gate
			if a.opts.Approver != nil && !a.opts.Approver(call.Name, json.RawMessage(call.Arguments)) {
				deniedMsg := fmt.Sprintf("Tool %s execution denied by user or security policy", call.Name)
				a.opts.Memory.Add(ai.ToolMessage(call.ID, deniedMsg))
				a.emit(Event{
					Type:     EventToolResult,
					Step:     step,
					ToolID:   call.ID,
					ToolName: call.Name,
					Error:    deniedMsg,
				})
				continue
			}

			// Execute tool
			result, toolErr := a.opts.Registry.Execute(ctx, call.Name, json.RawMessage(call.Arguments))
			var resPayload string
			if toolErr != nil {
				resPayload = fmt.Sprintf("Error: %v", toolErr)
				a.emit(Event{
					Type:     EventToolResult,
					Step:     step,
					ToolID:   call.ID,
					ToolName: call.Name,
					Error:    toolErr.Error(),
				})
			} else {
				if str, ok := result.(string); ok {
					resPayload = str
				} else {
					b, _ := json.Marshal(result)
					resPayload = string(b)
				}
				a.emit(Event{
					Type:     EventToolResult,
					Step:     step,
					ToolID:   call.ID,
					ToolName: call.Name,
					Result:   result,
				})
			}

			a.opts.Memory.Add(ai.ToolMessage(call.ID, resPayload))
		}
	}

	return nil, ErrMaxStepsExceeded
}

// complete resolves one LLM turn, routing through the provider's streaming
// endpoint when token-level streaming is enabled so callers observe real-time
// text_delta events.
func (a *Agent) complete(ctx context.Context, req ai.CompletionRequest, step int) (*ai.CompletionResponse, error) {
	if !a.opts.Streaming {
		resp, err := a.opts.Provider.Complete(ctx, req)
		if err != nil {
			a.emit(Event{Type: EventError, Step: step, Error: err.Error()})
		}
		return resp, err
	}

	chunks, err := a.opts.Provider.Stream(ctx, req)
	if err != nil {
		a.emit(Event{Type: EventError, Step: step, Error: err.Error()})
		return nil, err
	}

	resp := &ai.CompletionResponse{}
	for chunk := range chunks {
		if chunk.Error != nil {
			a.emit(Event{Type: EventError, Step: step, Error: chunk.Error.Error()})
			return nil, chunk.Error
		}
		if chunk.Delta != "" {
			resp.Content += chunk.Delta
			a.emit(Event{Type: EventTextDelta, Step: step, Content: chunk.Delta})
		}
		if chunk.ToolCall != nil {
			resp.ToolCalls = append(resp.ToolCalls, *chunk.ToolCall)
		}
		if chunk.Done {
			resp.FinishReason = "stop"
			if chunk.Usage.TotalTokens > 0 {
				resp.Usage = chunk.Usage
			}
		}
	}
	return resp, nil
}
