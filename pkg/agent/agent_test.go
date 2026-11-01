package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krewire/pkg/ai"
)

// multiStepMock is a sequential response mock provider.
type multiStepMock struct {
	responses []*ai.CompletionResponse
	step      int
}

func (m *multiStepMock) Name() string { return "multi-mock" }
func (m *multiStepMock) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if m.step >= len(m.responses) {
		return &ai.CompletionResponse{Content: "Done.", FinishReason: "stop"}, nil
	}
	resp := m.responses[m.step]
	m.step++
	return resp, nil
}
func (m *multiStepMock) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestAgent_DirectResponse(t *testing.T) {
	mock := &multiStepMock{
		responses: []*ai.CompletionResponse{
			{Content: "Hello! How can I help you?", FinishReason: "stop"},
		},
	}

	var events []Event
	ag, err := New(
		WithProvider(mock),
		WithEventHandler(func(e Event) {
			events = append(events, e)
		}),
	)
	if err != nil {
		t.Fatalf("agent init failed: %v", err)
	}

	resp, err := ag.Run(context.Background(), "Hi")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
	if len(events) < 2 {
		t.Errorf("expected at least start and done events, got %d", len(events))
	}
}

func TestAgent_ToolExecutionLoop(t *testing.T) {
	tmp := t.TempDir()
	registry := tools.DefaultWorkspaceTools(tmp)

	// Step 1: Model decides to call write_file
	// Step 2: Model sees file written and gives final answer
	mock := &multiStepMock{
		responses: []*ai.CompletionResponse{
			{
				Content: "Let me create the greeting file.",
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_1",
						Name:      "write_file",
						Arguments: `{"path": "hello.txt", "content": "Krew Agent is alive"}`,
					},
				},
			},
			{
				Content:      "I have successfully created hello.txt.",
				FinishReason: "stop",
			},
		},
	}

	var toolCallsSeen []string
	ag, err := New(
		WithProvider(mock),
		WithTools(registry),
		WithEventHandler(func(e Event) {
			if e.Type == EventToolCall {
				toolCallsSeen = append(toolCallsSeen, e.ToolName)
			}
		}),
	)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	resp, err := ag.Run(context.Background(), "Create hello.txt")
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if resp.ToolCallsRun != 1 {
		t.Errorf("expected 1 tool call run, got %d", resp.ToolCallsRun)
	}
	if len(toolCallsSeen) != 1 || toolCallsSeen[0] != "write_file" {
		t.Errorf("expected write_file tool call event, got %v", toolCallsSeen)
	}

	// Verify file was physically written to disk
	readRes, err := registry.Execute(context.Background(), "read_file", json.RawMessage(`{"path": "hello.txt"}`))
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}
	if readRes.(string) != "Krew Agent is alive" {
		t.Errorf("unexpected file content: %s", readRes)
	}
}

func TestAgent_ApprovalGate(t *testing.T) {
	mock := &multiStepMock{
		responses: []*ai.CompletionResponse{
			{
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_deny",
						Name:      "run_command",
						Arguments: `{"command": "rm -rf /"}`,
					},
				},
			},
			{
				Content:      "Action was rejected as expected.",
				FinishReason: "stop",
			},
		},
	}

	ag, err := New(
		WithProvider(mock),
		WithTools(tools.NewRegistry().Register(&tools.RunCommandTool{})),
		WithApprover(func(name string, _ json.RawMessage) bool {
			if name == "run_command" {
				return false // deny
			}
			return true
		}),
	)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	resp, err := ag.Run(context.Background(), "Run command")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if resp.Content != "Action was rejected as expected." {
		t.Errorf("unexpected content: %s", resp.Content)
	}
}

func TestAgent_TokenBufferDefaultAndCompaction(t *testing.T) {
	mock := &multiStepMock{
		responses: []*ai.CompletionResponse{
			{Content: "turn response", FinishReason: "stop"},
		},
	}

	ag, err := New(WithProvider(mock))
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	// 1. Check default memory backend is TokenBuffer
	tb, ok := ag.Memory().(*memory.TokenBuffer)
	if !ok || tb == nil {
		t.Fatalf("expected default Memory to be *memory.TokenBuffer, got %T", ag.Memory())
	}

	// 2. Verify compaction bounds context
	smallBuffer := memory.NewTokenBuffer("system", memory.WithMaxTokens(50), memory.WithKeepTurns(2))
	agCompact, err := New(WithProvider(mock), WithMemory(smallBuffer))
	if err != nil {
		t.Fatal(err)
	}

	// Add several large messages
	for i := 0; i < 10; i++ {
		smallBuffer.Add(ai.UserMessage(fmt.Sprintf("very long message turn %d with lots of words and padding text here", i)))
		smallBuffer.Add(ai.AssistantMessage(fmt.Sprintf("assistant long response turn %d with lots of words and text", i)))
	}

	// Confirm messages are compacted and within bound
	msgs := agCompact.Memory().Messages()
	if len(msgs) > 6 {
		t.Errorf("expected compaction to reduce message count, got %d", len(msgs))
	}
}

func TestAgent_ToolResultTruncation(t *testing.T) {
	hugeToolOutput := strings.Repeat("huge output line\n", 5000)

	reg := tools.NewRegistry()
	reg.Register(&mockHugeTool{output: hugeToolOutput})

	mock := &multiStepMock{
		responses: []*ai.CompletionResponse{
			{
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_huge",
						Name:      "huge_tool",
						Arguments: `{}`,
					},
				},
			},
			{
				Content:      "Processed huge tool output",
				FinishReason: "stop",
			},
		},
	}

	ag, err := New(
		WithProvider(mock),
		WithTools(reg),
		WithMaxToolResultTokens(100), // restrict to 100 tokens (~400 runes)
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ag.Run(context.Background(), "Run huge tool")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	msgs := ag.Memory().Messages()
	var toolMsg *ai.Message
	for _, m := range msgs {
		if m.Role == ai.RoleTool {
			toolMsg = &m
			break
		}
	}
	if toolMsg == nil {
		t.Fatal("expected tool message in memory")
	}
	if !strings.Contains(toolMsg.Content, "[tool output truncated: exceeded token limit]") {
		t.Errorf("expected truncated tool message, got length %d", len(toolMsg.Content))
	}
	if len(toolMsg.Content) > 1000 {
		t.Errorf("expected tool message length <= 1000, got %d", len(toolMsg.Content))
	}
}

type mockHugeTool struct {
	output string
}

func (m *mockHugeTool) Name() string        { return "huge_tool" }
func (m *mockHugeTool) Description() string { return "returns huge output" }
func (m *mockHugeTool) Parameters() map[string]any {
	return map[string]any{"type": "object"}
}
func (m *mockHugeTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	return m.output, nil
}
