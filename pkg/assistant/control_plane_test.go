package assistant_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/assistant"
	"github.com/krewire/krew/pkg/sdlc"
)

func TestControlPlane_LifecycleAndAudit(t *testing.T) {
	cp := assistant.NewControlPlane(assistant.HostCommandRunner{})

	tmpDir := t.TempDir()
	wsID := "ws-test-1"

	brief := sdlc.TaskBrief{
		ID:                   "TASK-CP-1",
		Title:                "Refactor tokenizer module",
		Objective:            "Improve token estimation accuracy",
		Scope:                []string{"pkg/token"},
		VerificationCommands: []string{"echo verify-ok"},
		RequirePlan:          true,
		RequireReview:        true,
	}

	_, err := cp.CreateTask(context.Background(), wsID, tmpDir, brief, "developer_alice", func(opts *sdlc.WorkflowOptions) {
		opts.PlanFn = func(ctx context.Context, b sdlc.TaskBrief) (*sdlc.Plan, error) {
			return &sdlc.Plan{
				Summary: "Add token count validation",
				Steps: []sdlc.PlanStep{
					{ID: "step-1", Description: "Add token count validation", TargetFiles: []string{"pkg/token/token.go"}},
				},
			}, nil
		}
		opts.ImplementFn = func(ctx context.Context, b sdlc.TaskBrief, plan *sdlc.Plan, approver agent.Approver) error {
			fpath := filepath.Join(tmpDir, "pkg", "token", "token.go")
			_ = os.MkdirAll(filepath.Dir(fpath), 0755)
			return os.WriteFile(fpath, []byte("package token\n\nfunc Count() int { return 42 }\n"), 0644)
		}
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 1. Wait for plan decision gate
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		taskObj, err := cp.GetTask(wsID, brief.ID)
		if err == nil {
			if taskObj.Status == sdlc.StatusAwaitingPlanApproval {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	taskObj, err := cp.GetTask(wsID, brief.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if taskObj.Status != sdlc.StatusAwaitingPlanApproval {
		t.Fatalf("expected status %s, got %s", sdlc.StatusAwaitingPlanApproval, taskObj.Status)
	}
	if taskObj.Plan == nil || len(taskObj.Plan.Steps) != 1 {
		t.Fatalf("expected plan with 1 step, got: %+v", taskObj.Plan)
	}

	// 2. Approve plan
	if err := cp.DecidePlan(wsID, brief.ID, true, "lead_bob"); err != nil {
		t.Fatalf("DecidePlan failed: %v", err)
	}

	// 3. Wait for review decision gate
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		taskObj, err = cp.GetTask(wsID, brief.ID)
		if err == nil && taskObj.Status == sdlc.StatusAwaitingReview {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	taskObj, err = cp.GetTask(wsID, brief.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if taskObj.Status != sdlc.StatusAwaitingReview {
		t.Fatalf("expected status %s, got %s", sdlc.StatusAwaitingReview, taskObj.Status)
	}
	if taskObj.Review == nil {
		t.Fatalf("expected review summary, got nil")
	}

	// 4. Accept review
	if err := cp.DecideReview(wsID, brief.ID, true, "lead_bob"); err != nil {
		t.Fatalf("DecideReview failed: %v", err)
	}

	// 5. Wait for workflow completion
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		taskObj, err = cp.GetTask(wsID, brief.ID)
		if err == nil && taskObj.Status == sdlc.StatusCompleted {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	taskObj, err = cp.GetTask(wsID, brief.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if taskObj.Status != sdlc.StatusCompleted {
		t.Fatalf("expected task completion, got: %s (err: %s)", taskObj.Status, taskObj.Error)
	}

	// 6. Audit activities verification
	if len(taskObj.Activities) < 5 {
		t.Fatalf("expected at least 5 audit activity records, got %d", len(taskObj.Activities))
	}

	foundEvidence := false
	foundPlanDecided := false
	for _, act := range taskObj.Activities {
		if act.Type == "plan_decided" && act.Actor == "lead_bob" {
			foundPlanDecided = true
		}
		if act.IsVerified && strings.Contains(act.Label, "Evidence") {
			foundEvidence = true
		}
	}
	if !foundPlanDecided {
		t.Errorf("expected plan_decided activity record by lead_bob")
	}
	if !foundEvidence {
		t.Errorf("expected at least one verified system evidence activity record")
	}

	// Verify task retrieval by workspace list
	list := cp.ListTasks(wsID)
	if len(list) != 1 || list[0].ID != brief.ID {
		t.Errorf("unexpected list tasks result: %+v", list)
	}
}

func TestControlPlane_RejectionGates(t *testing.T) {
	cp := assistant.NewControlPlane(assistant.HostCommandRunner{})
	tmpDir := t.TempDir()
	wsID := "ws-test-reject"

	brief := sdlc.TaskBrief{
		ID:          "TASK-REJECT-1",
		Title:       "Test rejection flow",
		Objective:   "Verify rejection halts task",
		Scope:       []string{"."},
		RequirePlan: true,
	}

	_, err := cp.CreateTask(context.Background(), wsID, tmpDir, brief, "dev", func(opts *sdlc.WorkflowOptions) {
		opts.PlanFn = func(ctx context.Context, b sdlc.TaskBrief) (*sdlc.Plan, error) {
			return &sdlc.Plan{
				Summary: "risky change",
				Steps:   []sdlc.PlanStep{{ID: "s1", Description: "risky step"}},
			}, nil
		}
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Wait for plan approval stage
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		task, err := cp.GetTask(wsID, brief.ID)
		if err == nil && task.Status == sdlc.StatusAwaitingPlanApproval {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Reject the plan
	if err := cp.DecidePlan(wsID, brief.ID, false, "security_officer"); err != nil {
		t.Fatalf("DecidePlan rejection failed: %v", err)
	}

	// Wait for workflow to halt
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		task, err := cp.GetTask(wsID, brief.ID)
		if err == nil && task.Status == sdlc.StatusFailed {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	task, err := cp.GetTask(wsID, brief.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if task.Status != sdlc.StatusFailed {
		t.Errorf("expected task status failed after rejection, got: %s", task.Status)
	}
}

func TestControlPlane_SubscribeActivities(t *testing.T) {
	cp := assistant.NewControlPlane(assistant.HostCommandRunner{})
	tmpDir := t.TempDir()
	wsID := "ws-sub-1"

	brief := sdlc.TaskBrief{
		ID:        "TASK-SUB-1",
		Title:     "Subscription test",
		Objective: "Test live activity stream",
		Scope:     []string{"."},
	}

	task, err := cp.CreateTask(context.Background(), wsID, tmpDir, brief, "operator", func(opts *sdlc.WorkflowOptions) {
		opts.PlanFn = func(ctx context.Context, b sdlc.TaskBrief) (*sdlc.Plan, error) {
			return nil, errors.New("abort early")
		}
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	ch, unsubscribe := task.SubscribeActivities()
	defer unsubscribe()

	select {
	case act := <-ch:
		if act.Type != "task_status" && act.Type != "stage_transition" {
			t.Errorf("unexpected initial activity: %+v", act)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for activity stream event")
	}
}
