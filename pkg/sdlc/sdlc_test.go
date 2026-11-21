package sdlc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/agent"
)

func TestTaskBrief_Validation(t *testing.T) {
	cases := []struct {
		name    string
		brief   TaskBrief
		wantErr bool
	}{
		{
			name: "missing id",
			brief: TaskBrief{
				Title:     "Add feature",
				Objective: "Do something",
				Scope:     []string{"pkg/auth"},
			},
			wantErr: true,
		},
		{
			name: "missing title",
			brief: TaskBrief{
				ID:        "TASK-1",
				Objective: "Do something",
				Scope:     []string{"pkg/auth"},
			},
			wantErr: true,
		},
		{
			name: "missing objective",
			brief: TaskBrief{
				ID:    "TASK-1",
				Title: "Add feature",
				Scope: []string{"pkg/auth"},
			},
			wantErr: true,
		},
		{
			name: "missing scope",
			brief: TaskBrief{
				ID:        "TASK-1",
				Title:     "Add feature",
				Objective: "Do something",
			},
			wantErr: true,
		},
		{
			name: "valid brief",
			brief: TaskBrief{
				ID:        "TASK-1",
				Title:     "Add feature",
				Objective: "Do something",
				Scope:     []string{"pkg/auth"},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.brief.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestTaskBrief_ParseFromMarkdown(t *testing.T) {
	md := `
## Objective
Implement TOTP code step deduplication to prevent replay attacks.

## Scope
- pkg/identity
- pkg/webauth

## Constraints
- Pure Go with CGO_ENABLED=0
- Must not leak secrets

## Acceptance Criteria
- [ ] TOTP replay is rejected
- [ ] Concurrency test passes

## Verification Commands
- go test -v ./pkg/identity/...
- go test -v ./pkg/webauth/...
`

	brief, err := ParseTaskBriefFromMarkdown("AUTH-12", "Prevent TOTP replay", md)
	if err != nil {
		t.Fatalf("ParseTaskBriefFromMarkdown failed: %v", err)
	}

	if brief.ID != "AUTH-12" {
		t.Errorf("ID = %q, want AUTH-12", brief.ID)
	}
	if brief.Title != "Prevent TOTP replay" {
		t.Errorf("Title = %q", brief.Title)
	}
	if !strings.Contains(brief.Objective, "TOTP code step deduplication") {
		t.Errorf("Objective = %q", brief.Objective)
	}
	if len(brief.Scope) != 2 || brief.Scope[0] != "pkg/identity" || brief.Scope[1] != "pkg/webauth" {
		t.Errorf("Scope = %v", brief.Scope)
	}
	if len(brief.Constraints) != 2 {
		t.Errorf("Constraints = %v", brief.Constraints)
	}
	if len(brief.AcceptanceCriteria) != 2 {
		t.Errorf("AcceptanceCriteria = %v", brief.AcceptanceCriteria)
	}
	if len(brief.VerificationCommands) != 2 || brief.VerificationCommands[0] != "go test -v ./pkg/identity/..." {
		t.Errorf("VerificationCommands = %v", brief.VerificationCommands)
	}
}

func TestPolicy_ScopeEnforcement(t *testing.T) {
	scopes := []string{"pkg/auth", "cmd/server"}

	if !CheckScope("pkg/auth/login.go", scopes) {
		t.Error("expected pkg/auth/login.go to be in scope")
	}
	if !CheckScope("cmd/server/main.go", scopes) {
		t.Error("expected cmd/server/main.go to be in scope")
	}
	if CheckScope("pkg/secret/key.go", scopes) {
		t.Error("expected pkg/secret/key.go to be outside scope")
	}
	if CheckScope("../outside.go", scopes) {
		t.Error("expected traversal to be outside scope")
	}

	// PolicyFilter tests
	brief := TaskBrief{
		ID:    "TEST-1",
		Scope: scopes,
	}

	var currentStage Stage = StageUnderstand
	stageGetter := func() Stage { return currentStage }

	filter := NewScopedApprover(brief, stageGetter, nil, "run_command")

	writeArgInScope, _ := json.Marshal(map[string]string{"path": "pkg/auth/login.go"})
	writeArgOutScope, _ := json.Marshal(map[string]string{"path": "pkg/secret/key.go"})

	// In StageUnderstand, write_file must be rejected
	if filter("write_file", writeArgInScope) {
		t.Error("write_file must be rejected during StageUnderstand")
	}

	// Switch to StageImplement
	currentStage = StageImplement
	if !filter("write_file", writeArgInScope) {
		t.Error("write_file within scope must be approved during StageImplement")
	}
	if filter("write_file", writeArgOutScope) {
		t.Error("write_file outside scope must be rejected during StageImplement")
	}

	// High risk tool without approver must fail closed
	if filter("run_command", nil) {
		t.Error("high risk tool without approver must fail closed")
	}
}

func TestDiff_SnapshotAndComputeDiff(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	_ = os.WriteFile(file1, []byte("line 1\nline 2\n"), 0o644)
	_ = os.WriteFile(file2, []byte("delete me\n"), 0o644)

	snap, err := TakeSnapshot(tmpDir, []string{"."})
	if err != nil {
		t.Fatalf("TakeSnapshot failed: %v", err)
	}

	// Modify file1
	_ = os.WriteFile(file1, []byte("line 1\nline 2 modified\nline 3\n"), 0o644)
	// Delete file2
	_ = os.Remove(file2)
	// Create file3
	file3 := filepath.Join(tmpDir, "file3.txt")
	_ = os.WriteFile(file3, []byte("new file\n"), 0o644)

	diffs, err := snap.ComputeDiff()
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}

	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d: %+v", len(diffs), diffs)
	}

	diffMap := make(map[string]FileDiff)
	for _, d := range diffs {
		diffMap[d.Path] = d
	}

	if d, ok := diffMap["file1.txt"]; !ok || d.Action != "modified" {
		t.Errorf("file1.txt diff expected modified, got %+v", d)
	}
	if d, ok := diffMap["file2.txt"]; !ok || d.Action != "deleted" {
		t.Errorf("file2.txt diff expected deleted, got %+v", d)
	}
	if d, ok := diffMap["file3.txt"]; !ok || d.Action != "created" {
		t.Errorf("file3.txt diff expected created, got %+v", d)
	}
}

func TestVerifier_Run(t *testing.T) {
	dir := t.TempDir()

	mockExec := func(ctx context.Context, d, command string) (int, string, error) {
		if strings.Contains(command, "fail") {
			return 1, "mock failure output", errors.New("exit 1")
		}
		return 0, "mock success output", nil
	}

	v := NewVerifier(dir, mockExec)

	// All passing
	results, passed := v.Run(context.Background(), []string{"echo pass", "test ok"})
	if !passed || len(results) != 2 {
		t.Fatalf("expected all checks to pass, got passed=%v results=%+v", passed, results)
	}

	// One failing
	results, passed = v.Run(context.Background(), []string{"echo pass", "run fail", "test ok"})
	if passed || len(results) != 3 {
		t.Fatalf("expected check to fail, got passed=%v results=%+v", passed, results)
	}
	if results[1].Passed || results[1].ExitCode != 1 || !strings.Contains(results[1].Output, "failure") {
		t.Errorf("failing result mismatch: %+v", results[1])
	}
}

func TestWorkflow_VerificationLedgerIntegrity(t *testing.T) {
	dir := t.TempDir()

	mockExec := func(ctx context.Context, d, command string) (int, string, error) {
		return 0, "gate ok", nil
	}

	brief := TaskBrief{
		ID:                   "TASK-LEDGER-01",
		Title:                "Test Ledger",
		Objective:            "Verify tamper-evident ledger chaining",
		Repository:           "test/repo",
		Scope:                []string{"."},
		VerificationCommands: []string{"go test ./..."},
	}

	wf, err := NewWorkflow(WorkflowOptions{
		Workspace:    dir,
		Brief:        brief,
		Executor:     mockExec,
		UnderstandFn: func(ctx context.Context, b TaskBrief) error { return nil },
		PlanFn: func(ctx context.Context, b TaskBrief) (*Plan, error) {
			return &Plan{Summary: "Plan ok", Approved: true}, nil
		},
		ImplementFn: func(ctx context.Context, b TaskBrief, p *Plan, app agent.Approver) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	result, err := wf.Run(context.Background())
	if err != nil {
		t.Fatalf("workflow execution failed: %v", err)
	}

	if result.Ledger == nil {
		t.Fatal("expected workflow result to have a non-nil verification ledger")
	}

	ok, err := result.Ledger.VerifyIntegrity()
	if !ok || err != nil {
		t.Fatalf("verification ledger integrity check failed: ok=%v, err=%v", ok, err)
	}

	latest, err := result.Ledger.LatestBlock()
	if err != nil {
		t.Fatalf("failed to get latest ledger block: %v", err)
	}
	if !latest.AllPassed {
		t.Errorf("expected all passed in latest block")
	}
}

func TestSDLC_ModelConversions(t *testing.T) {
	brief := TaskBrief{
		ID:                   "TASK-42",
		Title:                "Test Task",
		Objective:            "Objective text",
		Repository:           "krew",
		Scope:                []string{"pkg/sdlc"},
		Constraints:          []string{"pure-go"},
		AcceptanceCriteria:   []string{"tests pass"},
		VerificationCommands: []string{"go test ./..."},
		RequirePlan:          true,
		RequireReview:        true,
		Timeout:              5 * time.Minute,
		MaxRetries:           3,
	}

	modelBrief := brief.ToModel()
	if modelBrief.ID != brief.ID {
		t.Errorf("expected ID %s, got %s", brief.ID, modelBrief.ID)
	}
	if len(modelBrief.Scope) != len(brief.Scope) {
		t.Errorf("expected %d scope paths, got %d", len(brief.Scope), len(modelBrief.Scope))
	}

	roundtripBrief := FromModelBrief(modelBrief)
	if roundtripBrief.ID != brief.ID {
		t.Errorf("expected roundtrip ID %s, got %s", brief.ID, roundtripBrief.ID)
	}
	if roundtripBrief.Timeout != brief.Timeout {
		t.Errorf("expected roundtrip timeout %v, got %v", brief.Timeout, roundtripBrief.Timeout)
	}

	plan := Plan{
		Summary: "Test plan summary",
		Steps: []PlanStep{
			{
				ID:          "step-1",
				Description: "Create file",
				TargetFiles: []string{"pkg/test.go"},
				RiskLevel:   "low",
			},
		},
	}

	modelPlan := plan.ToModel()
	if len(modelPlan.Steps) != 1 {
		t.Fatalf("expected 1 step in model plan, got %d", len(modelPlan.Steps))
	}
	if modelPlan.Steps[0].ID != "step-1" {
		t.Errorf("expected step ID step-1, got %s", modelPlan.Steps[0].ID)
	}

	roundtripPlan := FromModelPlan(modelPlan)
	if len(roundtripPlan.Steps) != 1 {
		t.Fatalf("expected 1 step in roundtrip plan, got %d", len(roundtripPlan.Steps))
	}
	if roundtripPlan.Steps[0].ID != "step-1" {
		t.Errorf("expected roundtrip step ID step-1, got %s", roundtripPlan.Steps[0].ID)
	}
}
