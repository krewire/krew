package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
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
	opts   Options
	emitMu *sync.Mutex
}

// DefaultMaxToolResultTokens limits tool response sizes to avoid context explosion.
const (
	DefaultMaxToolResultTokens = 8_000
	DefaultMaxConcurrentTools  = 4
)

// New constructs an initialized Agent.
func New(opts ...Option) (*Agent, error) {
	cfg := Options{
		Persona:             persona.CodingAssistant,
		MaxSteps:            25,
		Registry:            tools.NewRegistry(),
		MaxToolResultTokens: DefaultMaxToolResultTokens,
		MaxConcurrentTools:  DefaultMaxConcurrentTools,
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
		cfg.Memory = memory.NewTokenBuffer(cfg.SystemPrompt)
	}
	if cfg.MaxToolResultTokens <= 0 {
		cfg.MaxToolResultTokens = DefaultMaxToolResultTokens
	}
	if cfg.MaxConcurrentTools <= 0 {
		cfg.MaxConcurrentTools = DefaultMaxConcurrentTools
	}

	return &Agent{opts: cfg, emitMu: &sync.Mutex{}}, nil
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

	clone := &Agent{opts: a.opts, emitMu: &sync.Mutex{}}
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
	cp := &Agent{opts: a.opts, emitMu: &sync.Mutex{}}
	for _, opt := range opts {
		opt(&cp.opts)
	}
	return cp
}

// Chat continues an existing conversation with a new user message.
func (a *Agent) Chat(ctx context.Context, message string) (*Response, error) {
	return a.Run(ctx, message)
}

// Memory returns the conversation memory.
func (a *Agent) Memory() memory.Memory {
	return a.opts.Memory
}

// RegisterTool adds a tool to the agent's registry.
func (a *Agent) RegisterTool(t tools.Tool) error {
	if a == nil || a.opts.Registry == nil {
		return fmt.Errorf("agent: registry is not available")
	}
	a.opts.Registry.Register(t)
	return nil
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
		if a.emitMu != nil {
			a.emitMu.Lock()
			defer a.emitMu.Unlock()
		}
		a.opts.OnEvent(e)
	}
}

func (a *Agent) isToolParallelSafe(name string) bool {
	if a.opts.Registry == nil {
		return false
	}
	t, ok := a.opts.Registry.Get(name)
	if !ok {
		return false
	}
	return tools.IsParallelSafe(t)
}

type toolExecOutcome struct {
	payload string
	denied  bool
	err     error
}

func (a *Agent) executeSingleTool(ctx context.Context, step int, call ai.ToolCall) toolExecOutcome {
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
		a.emit(Event{
			Type:     EventToolResult,
			Step:     step,
			ToolID:   call.ID,
			ToolName: call.Name,
			Error:    deniedMsg,
		})
		return toolExecOutcome{payload: deniedMsg, denied: true}
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

	if a.opts.MaxToolResultTokens > 0 && memory.EstimateTokens(resPayload) > a.opts.MaxToolResultTokens {
		maxRunes := a.opts.MaxToolResultTokens * 4
		runes := []rune(resPayload)
		if len(runes) > maxRunes {
			resPayload = string(runes[:maxRunes]) + "\n... [tool output truncated: exceeded token limit]"
		}
	}

	return toolExecOutcome{payload: resPayload, err: toolErr}
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

		results := make([]toolExecOutcome, len(resp.ToolCalls))

		// Partition tool calls into contiguous batches:
		// Sequences of parallel-safe tools can execute concurrently within a bounded worker pool.
		// Any non-parallel-safe tool acts as a single-item barrier executed sequentially.
		var batches [][]int
		var currentBatch []int

		concurrency := a.opts.MaxConcurrentTools
		if concurrency <= 0 {
			concurrency = DefaultMaxConcurrentTools
		}

		for i, call := range resp.ToolCalls {
			if concurrency > 1 && a.isToolParallelSafe(call.Name) {
				currentBatch = append(currentBatch, i)
			} else {
				if len(currentBatch) > 0 {
					batches = append(batches, currentBatch)
					currentBatch = nil
				}
				batches = append(batches, []int{i})
			}
		}
		if len(currentBatch) > 0 {
			batches = append(batches, currentBatch)
		}

		var strictDeniedTool string
		for _, batch := range batches {
			if len(batch) == 1 {
				idx := batch[0]
				call := resp.ToolCalls[idx]
				res := a.executeSingleTool(ctx, step, call)
				results[idx] = res
				if res.denied && a.opts.StrictApprovals {
					strictDeniedTool = call.Name
					break
				}
			} else {
				// Concurrent execution for parallel-safe batch
				sem := make(chan struct{}, concurrency)
				var wg sync.WaitGroup
				var mu sync.Mutex

				for _, idx := range batch {
					call := resp.ToolCalls[idx]
					wg.Add(1)
					go func(i int, c ai.ToolCall) {
						defer wg.Done()
						sem <- struct{}{}
						defer func() { <-sem }()

						res := a.executeSingleTool(ctx, step, c)
						results[i] = res
						if res.denied && a.opts.StrictApprovals {
							mu.Lock()
							if strictDeniedTool == "" {
								strictDeniedTool = c.Name
							}
							mu.Unlock()
						}
					}(idx, call)
				}
				wg.Wait()
				if strictDeniedTool != "" {
					break
				}
			}
		}

		// Preserve deterministic tool result order in Memory
		for i, call := range resp.ToolCalls {
			if results[i].payload != "" || results[i].denied || results[i].err != nil {
				toolCallsRun++
				a.opts.Memory.Add(ai.ToolMessage(call.ID, results[i].payload))
			}
		}

		if strictDeniedTool != "" {
			return nil, fmt.Errorf("%w: tool %s execution denied", ErrPermissionDenied, strictDeniedTool)
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
