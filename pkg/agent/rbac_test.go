package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/krewire/auth/pkg/audit"
	"github.com/krewire/auth/pkg/rbac"
	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/tools"
	"github.com/krewire/krewire/pkg/ai"
)

type rbacMockProvider struct {
	responses []*ai.CompletionResponse
	step      int
}

func (m *rbacMockProvider) Name() string { return "rbac-mock" }
func (m *rbacMockProvider) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if m.step >= len(m.responses) {
		return &ai.CompletionResponse{Content: "Completed", FinishReason: "stop"}, nil
	}
	resp := m.responses[m.step]
	m.step++
	return resp, nil
}
func (m *rbacMockProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestAgent_RBAC_ViewerDeniedRunCommand(t *testing.T) {
	eval := rbac.NewEvaluator()
	ledger := audit.NewMemoryLedger()

	approver := agent.NewAgentApprover(eval, []string{"viewer"}, "tenant_test", "user_viewer", ledger)

	mock := &rbacMockProvider{
		responses: []*ai.CompletionResponse{
			{
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_exec",
						Name:      "run_command",
						Arguments: `{"command": "uptime"}`,
					},
				},
			},
		},
	}

	ag, err := agent.New(
		agent.WithProvider(mock),
		agent.WithTools(tools.NewRegistry().Register(&tools.RunCommandTool{})),
		agent.WithApprover(approver),
		agent.WithStrictApprovals(true),
	)
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	_, err = ag.Run(context.Background(), "Check system uptime")
	if err == nil {
		t.Fatal("expected run_command to fail with permission denied for viewer")
	}
	if !errors.Is(err, agent.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got: %v", err)
	}

	// Verify audit ledger captured the violation
	events, err := ledger.Query(context.Background(), audit.Filter{
		TenantID: "tenant_test",
		Action:   "agent.tool_denied",
	})
	if err != nil {
		t.Fatalf("ledger query failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].TargetID != "run_command" {
		t.Errorf("expected target run_command, got %q", events[0].TargetID)
	}
	if events[0].ActorID != "user_viewer" {
		t.Errorf("expected actor user_viewer, got %q", events[0].ActorID)
	}
}

func TestAgent_RBAC_DeveloperAllowedFileTools(t *testing.T) {
	tmp := t.TempDir()
	reg := tools.DefaultWorkspaceTools(tmp)
	eval := rbac.NewEvaluator()
	ledger := audit.NewMemoryLedger()

	approver := agent.NewAgentApprover(eval, []string{"developer"}, "tenant_test", "user_dev", ledger)

	mock := &rbacMockProvider{
		responses: []*ai.CompletionResponse{
			{
				Content: "Creating file",
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_w",
						Name:      "write_file",
						Arguments: `{"path": "test.txt", "content": "developer access verified"}`,
					},
				},
			},
			{
				Content: "Reading file",
				ToolCalls: []ai.ToolCall{
					{
						ID:        "call_r",
						Name:      "read_file",
						Arguments: `{"path": "test.txt"}`,
					},
				},
			},
			{
				Content:      "All file operations succeeded.",
				FinishReason: "stop",
			},
		},
	}

	ag, err := agent.New(
		agent.WithProvider(mock),
		agent.WithTools(reg),
		agent.WithApprover(approver),
		agent.WithStrictApprovals(true),
	)
	if err != nil {
		t.Fatalf("agent.New failed: %v", err)
	}

	resp, err := ag.Run(context.Background(), "Write and read test.txt")
	if err != nil {
		t.Fatalf("developer execution failed unexpectedly: %v", err)
	}
	if resp.ToolCallsRun != 2 {
		t.Errorf("expected 2 tool calls run, got %d", resp.ToolCallsRun)
	}

	// Verify no permission denials were logged
	events, _ := ledger.Query(context.Background(), audit.Filter{})
	if len(events) != 0 {
		t.Errorf("expected 0 denial events for developer, got %d", len(events))
	}
}

func TestAgent_RBAC_PersonaPolicyDefaults(t *testing.T) {
	roleCoding, ok := agent.DefaultPersonaRole("CodingAssistant")
	if !ok || roleCoding.Name != "CodingAssistant" {
		t.Errorf("expected CodingAssistant persona role, got %+v", roleCoding)
	}

	roleDevOps, ok := agent.DefaultPersonaRole("DevOpsRunner")
	if !ok || roleDevOps.Name != "DevOpsRunner" {
		t.Errorf("expected DevOpsRunner persona role, got %+v", roleDevOps)
	}

	roleViewer, ok := agent.DefaultPersonaRole("ViewerAssistant")
	if !ok || roleViewer.Name != "ViewerAssistant" {
		t.Errorf("expected ViewerAssistant persona role, got %+v", roleViewer)
	}

	_, ok = agent.DefaultPersonaRole("unknown")
	if ok {
		t.Error("expected unknown persona role to return false")
	}
}
