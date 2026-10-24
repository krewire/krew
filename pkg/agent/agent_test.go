package agent

import (
	"context"
	"encoding/json"
	"testing"

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
