package agent

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/krewire/krewire/pkg/ai"
)

// scriptedProvider yields queued responses across every agent in the
// delegation chain, sharing one call counter like a real shared provider.
type scriptedProvider struct {
	mu        sync.Mutex
	responses []*ai.CompletionResponse
	calls     int
}

func (p *scriptedProvider) Name() string { return "scripted" }

func (p *scriptedProvider) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	if p.calls > len(p.responses) {
		return &ai.CompletionResponse{Content: "fallback final", FinishReason: "stop"}, nil
	}
	return p.responses[p.calls-1], nil
}

func (p *scriptedProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func subCall(id, prompt, persona string, maxSteps int) ai.ToolCall {
	args, _ := json.Marshal(map[string]any{
		"prompt":    prompt,
		"persona":   persona,
		"max_steps": maxSteps,
	})
	return ai.ToolCall{ID: id, Name: subagentToolName, Arguments: string(args)}
}

func agentWithScript(t *testing.T, provider *scriptedProvider, events *[]Event) *Agent {
	t.Helper()
	ag, err := New(WithProvider(provider), WithEventHandler(func(e Event) {
		if events != nil {
			*events = append(*events, e)
		}
	}))
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	if err := EnableSubAgents(ag); err != nil {
		t.Fatalf("EnableSubAgents: %v", err)
	}
	return ag
}

func TestSubAgent_DelegationRoundTrip(t *testing.T) {
	provider := &scriptedProvider{responses: []*ai.CompletionResponse{
		{ToolCalls: []ai.ToolCall{subCall("call_1", "investigate the caching layer", "research", 5)}},
		{Content: "research: cache invalidation happens in refresh()", FinishReason: "stop"},
		{Content: "Synthesis: caching invalidation lives in refresh()", FinishReason: "stop"},
	}}

	var events []Event
	ag := agentWithScript(t, provider, &events)

	resp, err := ag.Run(context.Background(), "Where does cache invalidation live?")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.Content != "Synthesis: caching invalidation lives in refresh()" {
		t.Errorf("expected synthesized final answer, got %q", resp.Content)
	}
	if resp.StepsTaken != 2 {
		t.Errorf("expected 2 parent steps, got %d", resp.StepsTaken)
	}
	if resp.ToolCallsRun != 1 {
		t.Errorf("expected 1 delegated tool call, got %d", resp.ToolCallsRun)
	}

	// Child events must arrive under a nested thread id.
	sawChildThought := false
	for _, ev := range events {
		if ev.Type == EventThought && ev.ThreadID == "1" && strings.Contains(ev.Content, "research:") {
			sawChildThought = true
		}
	}
	if !sawChildThought {
		t.Errorf("expected child thought event with thread id 1, got %+v", events)
	}

	// Closed-out delegation result delivered back to the parent.
	var delegation map[string]any
	for _, ev := range events {
		if ev.Type == EventToolResult && ev.ToolID == "call_1" {
			delegation, _ = ev.Result.(map[string]any)
		}
	}
	if delegation == nil {
		t.Fatalf("expected delegation tool_result, got %+v", events)
	}
	if delegation["thread_id"] != "1" || asFloat(delegation["steps_taken"]) != 1 {
		t.Errorf("unexpected delegation payload: %+v", delegation)
	}
}

// asFloat normalises int/float64 JSON values for assertions.
func asFloat(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	default:
		return -1
	}
}

func TestSubAgent_DelegationChained(t *testing.T) {
	provider := &scriptedProvider{responses: []*ai.CompletionResponse{
		{ToolCalls: []ai.ToolCall{subCall("call_1", "sub task", "worker", 5)}},
		{ToolCalls: []ai.ToolCall{subCall("call_2", "sub sub task", "coding", 5)}},
		{Content: "grandchild direct result", FinishReason: "stop"},
		{Content: "child synthesis", FinishReason: "stop"},
		{Content: "root synthesis", FinishReason: "stop"},
	}}

	var events []Event
	ag := agentWithScript(t, provider, &events)

	resp, err := ag.Run(context.Background(), "do nested work")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if resp.Content != "root synthesis" {
		t.Errorf("unexpected final response: %q", resp.Content)
	}

	// Nested lineage: root event has empty thread id, first child is "1",
	// deepest grandchild is "1:1".
	threads := map[string]bool{}
	for _, ev := range events {
		threads[ev.ThreadID] = true
	}
	if !threads[""] || !threads["1"] || !threads["1:1"] {
		t.Errorf("expected nested thread lineage root/1/1:1, saw %v", threads)
	}
}

func TestSubAgent_RecursionGuard(t *testing.T) {
	// The deepest agent (depth 2) attempts one further delegation and must
	// receive a clear error, not a crash.
	provider := &scriptedProvider{responses: []*ai.CompletionResponse{
		{ToolCalls: []ai.ToolCall{subCall("call_1", "level 1", "worker", 5)}},
		{ToolCalls: []ai.ToolCall{subCall("call_2", "level 2", "worker", 5)}},
		{ToolCalls: []ai.ToolCall{subCall("call_3", "level 3", "worker", 5)}},
		{Content: "grandchild direct result", FinishReason: "stop"},
		{Content: "child synthesis", FinishReason: "stop"},
		{Content: "root synthesis", FinishReason: "stop"},
	}}

	var events []Event
	ag := agentWithScript(t, provider, &events)

	resp, err := ag.Run(context.Background(), "recurse too far")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if resp.Content != "root synthesis" {
		t.Errorf("unexpected final response: %q", resp.Content)
	}

	sawGuardError := false
	for _, ev := range events {
		if ev.Type == EventError && strings.Contains(ev.Error, "depth limit") {
			sawGuardError = true
		}
	}
	// The guard surfaces as the denial text in the tool result stream.
	sawDenial := false
	for _, ev := range events {
		if ev.Type == EventToolResult && strings.Contains(ev.Error, "depth limit") {
			sawDenial = true
		}
	}
	if !sawGuardError && !sawDenial {
		t.Errorf("expected a depth-limit signal in events, got %+v", events)
	}
}

func TestSubAgent_DirectDepthGuardUnit(t *testing.T) {
	ag, err := New(WithProvider(&scriptedProvider{}))
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	tool := NewSubAgentTool(ag, MaxSubagentDepth)
	_, got := tool.Execute(context.Background(), json.RawMessage(`{"prompt":"x"}`))
	if got == nil || !strings.Contains(got.Error(), "depth limit") {
		t.Fatalf("expected depth limit error, got %v", got)
	}
}

func TestSubAgent_ArgValidation(t *testing.T) {
	ag, err := New(WithProvider(&scriptedProvider{}))
	if err != nil {
		t.Fatalf("agent.New: %v", err)
	}
	if err := EnableSubAgents(ag); err != nil {
		t.Fatalf("EnableSubAgents: %v", err)
	}

	if _, got := ag.Tools().Execute(context.Background(), subagentToolName, json.RawMessage(`{}`)); got == nil || !strings.Contains(got.Error(), "prompt is required") {
		t.Errorf("expected missing prompt error, got %v", got)
	}
	// Unknown persona keys fall back to the parent persona silently.
	provider := &scriptedProvider{responses: []*ai.CompletionResponse{
		{ToolCalls: []ai.ToolCall{subCall("call_1", "task", "unknown-persona", 2)}},
		{Content: "child did the task", FinishReason: "stop"},
		{Content: "done", FinishReason: "stop"},
	}}
	ag2 := agentWithScript(t, provider, nil)
	if _, err := ag2.Run(context.Background(), "go"); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}
