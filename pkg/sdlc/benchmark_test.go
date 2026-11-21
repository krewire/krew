package sdlc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/pkg/agent"
)

// Benchmark 1: Plan separation from execution.
// The workflow must expose its plan before modifications and fail closed if the plan is rejected.
func TestBenchmark_PlanBeforeImplement(t *testing.T) {
	tmpDir := t.TempDir()

	brief := TaskBrief{
		ID:                   "TASK-PLAN-1",
		Title:                "Refactor authentication middleware",
		Objective:            "Extract token validation into separate handler",
		Scope:                []string{"pkg/auth"},
		RequirePlan:          true,
		VerificationCommands: []string{"echo verify-ok"},
	}

	planSeen := false
	planApproved := false

	wf, err := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		PlanApprover: func(ctx context.Context, plan *Plan) (bool, string, error) {
			planSeen = true
			if len(plan.Steps) == 0 {
				t.Error("expected non-empty plan steps")
			}
			// Simulate plan rejection first
			return false, "security-reviewer", nil
		},
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			t.Fatal("ImplementFn must not be called when plan is rejected")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Rejection test
	res, err := wf.Run(context.Background())
	if err == nil || !errors.Is(err, ErrApprovalDenied) {
		t.Fatalf("expected ErrApprovalDenied on plan rejection, got %v", err)
	}
	if !planSeen {
		t.Fatal("plan was not exposed to reviewer before changes")
	}
	if res.Status != StatusFailed {
		t.Errorf("status = %v, want StatusFailed", res.Status)
	}

	// 2. Approval test
	wfApproved, _ := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		PlanApprover: func(ctx context.Context, plan *Plan) (bool, string, error) {
			planApproved = true
			return true, "lead-architect", nil
		},
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			if !p.Approved || p.ApprovedBy != "lead-architect" {
				t.Fatalf("implement called with unapproved plan: %+v", p)
			}
			return nil
		},
	})

	res, err = wfApproved.Run(context.Background())
	if err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if !planApproved || res.Plan == nil || !res.Plan.Approved {
		t.Errorf("plan was not properly marked as approved in result")
	}
}

// Benchmark 2: Scope restriction and permission denial.
// File writes and commands must be strictly restricted to granted scope.
func TestBenchmark_ScopeRestrictionAndPermissionDenial(t *testing.T) {
	tmpDir := t.TempDir()

	brief := TaskBrief{
		ID:                   "TASK-SCOPE-1",
		Title:                "Update billing service",
		Objective:            "Adjust invoice generator",
		Scope:                []string{"pkg/billing"},
		VerificationCommands: []string{"echo pass"},
	}

	scopeChecked := false
	wf, err := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			// Test permission on out-of-scope write
			outScopeArg, _ := json.Marshal(map[string]string{"path": "pkg/auth/secrets.go"})
			if app("write_file", outScopeArg) {
				return errors.New("security failure: out-of-scope write was approved")
			}

			// Test permission on in-scope write
			inScopeArg, _ := json.Marshal(map[string]string{"path": "pkg/billing/invoice.go"})
			if !app("write_file", inScopeArg) {
				return errors.New("in-scope write was denied")
			}

			scopeChecked = true
			// Create the file in scope
			_ = os.MkdirAll(filepath.Join(tmpDir, "pkg/billing"), 0o755)
			return os.WriteFile(filepath.Join(tmpDir, "pkg/billing/invoice.go"), []byte("// invoice\n"), 0o644)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := wf.Run(context.Background())
	if err != nil {
		t.Fatalf("workflow failed: %v", err)
	}
	if !scopeChecked {
		t.Fatal("scope was not verified during implementation")
	}
	if len(res.Diffs) == 0 {
		t.Fatal("expected diffs for in-scope file creation")
	}
}

// Benchmark 3: Verification failure fails closed.
// The agent must NOT claim completion when required checks fail.
func TestBenchmark_VerificationFailureFailsClosed(t *testing.T) {
	tmpDir := t.TempDir()

	brief := TaskBrief{
		ID:                   "TASK-VERIFY-1",
		Title:                "Add security check",
		Objective:            "Pass all test gates",
		Scope:                []string{"."},
		MaxRetries:           2,
		VerificationCommands: []string{"test-must-fail"},
	}

	attempts := 0
	wf, err := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		Executor: func(ctx context.Context, dir, command string) (int, string, error) {
			attempts++
			return 1, "FAIL: test assertion failed", errors.New("exit code 1")
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := wf.Run(context.Background())
	if err == nil || !errors.Is(err, ErrVerificationFailed) {
		t.Fatalf("expected ErrVerificationFailed, got %v", err)
	}
	if res.Status != StatusVerificationFailed {
		t.Errorf("status = %v, want StatusVerificationFailed", res.Status)
	}
	if attempts != 3 { // 1 initial attempt + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if len(res.Evidence) == 0 || res.Evidence[0].Passed {
		t.Errorf("expected structured failed evidence, got %+v", res.Evidence)
	}
}

// Benchmark 4: Human approval for high-risk actions.
func TestBenchmark_HumanApprovalForHighRiskAction(t *testing.T) {
	tmpDir := t.TempDir()

	brief := TaskBrief{
		ID:                   "TASK-APPROV-1",
		Title:                "Database schema migration",
		Objective:            "Run database drop and recreate",
		Scope:                []string{"pkg/db"},
		VerificationCommands: []string{"echo ok"},
	}

	approvalRequested := false
	humanApprover := func(toolName string, args json.RawMessage) bool {
		approvalRequested = true
		return false // Deny high-risk action
	}

	wf, err := NewWorkflow(WorkflowOptions{
		Workspace:     tmpDir,
		Brief:         brief,
		HumanApprover: humanApprover,
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			// run_command is configured as high-risk
			if !app("run_command", json.RawMessage(`{"command":"drop database"}`)) {
				return ErrApprovalDenied
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = wf.Run(context.Background())
	if err == nil || !errors.Is(err, ErrApprovalDenied) {
		t.Fatalf("expected ErrApprovalDenied, got %v", err)
	}
	if !approvalRequested {
		t.Fatal("human approval was not requested for high-risk tool")
	}
}

// Benchmark 5: Cancellation and Pause/Resume.
func TestBenchmark_CancellationAndPauseResume(t *testing.T) {
	tmpDir := t.TempDir()

	brief := TaskBrief{
		ID:                   "TASK-CANCEL-1",
		Title:                "Long running task",
		Objective:            "Execute long background task",
		Scope:                []string{"."},
		VerificationCommands: []string{"echo done"},
	}

	// 1. Cancellation test
	ctx, cancel := context.WithCancel(context.Background())
	wf, _ := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		UnderstandFn: func(c context.Context, b TaskBrief) error {
			cancel() // cancel context mid-flight
			return c.Err()
		},
	})

	_, err := wf.Run(ctx)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}

	// 2. Pause test
	var wfPause *Workflow
	wfPause, _ = NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		UnderstandFn: func(c context.Context, b TaskBrief) error {
			wfPause.Pause()
			return nil
		},
	})

	res, err := wfPause.Run(context.Background())
	if err == nil || !errors.Is(err, ErrWorkflowPaused) {
		t.Fatalf("expected ErrWorkflowPaused, got %v", err)
	}
	if res.Status != StatusPaused {
		t.Errorf("status = %v, want StatusPaused", res.Status)
	}
}

// Benchmark 6: End-to-end Issue Lifecycle Qualification.
// Validates issue -> brief -> understand -> plan -> implement -> verify -> review -> report.
func TestBenchmark_FullIssueLifecycle(t *testing.T) {
	tmpDir := t.TempDir()

	issueMarkdown := `
## Objective
Implement resilient database health check with timeout.

## Scope
- pkg/db

## Constraints
- SafePath traversal guard
- CGO_ENABLED=0 pure Go

## Acceptance Criteria
- [ ] HealthCheck handles unreachable database gracefully
- [ ] Timeout is enforced

## Verification Commands
- echo "DB verification passed"
`

	brief, err := ParseTaskBriefFromMarkdown("CORE-42", "Database health check", issueMarkdown)
	if err != nil {
		t.Fatalf("ParseTaskBriefFromMarkdown failed: %v", err)
	}

	var events []WorkflowEvent

	wf, err := NewWorkflow(WorkflowOptions{
		Workspace: tmpDir,
		Brief:     brief,
		OnEvent: func(ev WorkflowEvent) {
			events = append(events, ev)
		},
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			dbDir := filepath.Join(tmpDir, "pkg/db")
			_ = os.MkdirAll(dbDir, 0o755)
			return os.WriteFile(filepath.Join(dbDir, "health.go"), []byte("package db\n\nfunc Health() bool { return true }\n"), 0o644)
		},
		ReviewApprover: func(ctx context.Context, rev *ReviewSummary) (bool, string, error) {
			if len(rev.VerificationEvidence) == 0 {
				t.Error("expected verification evidence in review summary")
			}
			return true, "security-lead", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := wf.Run(context.Background())
	if err != nil {
		t.Fatalf("workflow failed: %v", err)
	}

	if res.Status != StatusCompleted {
		t.Fatalf("expected StatusCompleted, got %v", res.Status)
	}
	if res.Stage != StageReport {
		t.Fatalf("expected StageReport, got %v", res.Stage)
	}
	if res.Review == nil || !res.Review.Approved || res.Review.ReviewedBy != "security-lead" {
		t.Fatalf("review not recorded properly: %+v", res.Review)
	}
	if len(res.Evidence) != 1 || !res.Evidence[0].Passed {
		t.Fatalf("evidence missing or not passed: %+v", res.Evidence)
	}
	if len(res.Diffs) != 1 || res.Diffs[0].Path != "pkg/db/health.go" || res.Diffs[0].Action != "created" {
		t.Fatalf("diffs incorrect: %+v", res.Diffs)
	}
	if res.Provenance.TaskID != "CORE-42" || len(res.Provenance.CreatedFiles) != 1 {
		t.Fatalf("provenance incorrect: %+v", res.Provenance)
	}

	// Verify all lifecycle stages emitted events
	seenStages := make(map[Stage]bool)
	for _, ev := range events {
		seenStages[ev.Stage] = true
	}

	expectedStages := []Stage{StageUnderstand, StagePlan, StageImplement, StageVerify, StageReview, StageReport}
	for _, st := range expectedStages {
		if !seenStages[st] {
			t.Errorf("missing lifecycle event for stage: %s", st)
		}
	}
}
