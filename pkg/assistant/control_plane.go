package assistant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/sdlc"
	"github.com/krewire/krewire/pkg/sec"
)

var (
	ErrTaskNotFound = errors.New("control plane: task not found")
	ErrTaskHalted   = errors.New("control plane: task has already completed or halted")
)

// ActivityRecord represents an auditable execution, decision, or check event.
type ActivityRecord struct {
	ID         string         `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Type       string         `json:"type"` // "stage_transition", "plan_created", "plan_decided", "tool_call", "verification_check", "review_created", "review_decided", "task_status"
	Actor      string         `json:"actor"`
	IsVerified bool           `json:"is_verified"` // true for verified system checks and human approvals; false for unverified model suggestions
	Label      string         `json:"label"`       // "Verified System Evidence" vs "AI Model Suggestion"
	Summary    string         `json:"summary"`
	Details    map[string]any `json:"details,omitempty"`
}

// EngineeringTask is a supervised SDLC engineering workflow within a workspace.
type EngineeringTask struct {
	mu          sync.RWMutex
	ID          string                    `json:"id"`
	WorkspaceID string                    `json:"workspace_id"`
	Title       string                    `json:"title"`
	Brief       sdlc.TaskBrief            `json:"brief"`
	Stage       sdlc.Stage                `json:"stage"`
	Status      sdlc.Status               `json:"status"`
	Plan        *sdlc.Plan                `json:"plan,omitempty"`
	Diffs       []sdlc.FileDiff           `json:"diffs,omitempty"`
	Evidence    []sdlc.VerificationResult `json:"evidence,omitempty"`
	Review      *sdlc.ReviewSummary       `json:"review,omitempty"`
	Provenance  sdlc.Provenance           `json:"provenance"`
	Activities  []ActivityRecord          `json:"activities"`
	Error       string                    `json:"error,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`

	workflow       *sdlc.Workflow
	planDecision   chan decision
	reviewDecision chan decision
	listeners      map[chan ActivityRecord]struct{}
}

type decision struct {
	approve  bool
	reviewer string
}

// ControlPlane coordinates human-in-the-loop AI workflows across workspaces.
type ControlPlane struct {
	mu     sync.RWMutex
	tasks  map[string]*EngineeringTask // key: wsID + ":" + taskID
	runner CommandRunner
}

// NewControlPlane initializes a human-in-the-loop engineering control plane.
func NewControlPlane(runner CommandRunner) *ControlPlane {
	if runner == nil {
		runner = HostCommandRunner{}
	}
	return &ControlPlane{
		tasks:  make(map[string]*EngineeringTask),
		runner: runner,
	}
}

func taskKey(wsID, taskID string) string {
	return wsID + ":" + taskID
}

// CreateTask registers and initiates a new supervised SDLC task for a workspace.
func (cp *ControlPlane) CreateTask(ctx context.Context, wsID, wsPath string, brief sdlc.TaskBrief, actor string, configure ...func(*sdlc.WorkflowOptions)) (*EngineeringTask, error) {
	if err := brief.Validate(); err != nil {
		return nil, err
	}

	key := taskKey(wsID, brief.ID)
	cp.mu.Lock()
	if _, exists := cp.tasks[key]; exists {
		cp.mu.Unlock()
		return nil, fmt.Errorf("task %q already exists in workspace %q", brief.ID, wsID)
	}

	task := &EngineeringTask{
		ID:             brief.ID,
		WorkspaceID:    wsID,
		Title:          brief.Title,
		Brief:          brief,
		Stage:          sdlc.StageUnderstand,
		Status:         sdlc.StatusInProgress,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		planDecision:   make(chan decision, 1),
		reviewDecision: make(chan decision, 1),
		listeners:      make(map[chan ActivityRecord]struct{}),
		Activities:     make([]ActivityRecord, 0),
	}
	cp.tasks[key] = task
	cp.mu.Unlock()

	// Initial activity record
	task.recordActivity(ActivityRecord{
		ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Type:       "task_status",
		Actor:      actor,
		IsVerified: true,
		Label:      "Verified System Evidence",
		Summary:    fmt.Sprintf("Task %q initiated by %s", brief.ID, actor),
		Details:    map[string]any{"objective": brief.Objective, "scope": brief.Scope},
	})

	executor := func(c context.Context, dir, cmdStr string) (int, string, error) {
		if cp.runner != nil {
			res, err := cp.runner.RunCommand(c, wsID, wsPath, []string{"/bin/sh", "-c", cmdStr}, 60*time.Second)
			exitCode := 0
			output := ""
			if res != nil {
				exitCode = res.ExitCode
				output = res.Stdout
				if res.Stderr != "" {
					output += "\n" + res.Stderr
				}
			}
			if err != nil && exitCode == 0 {
				exitCode = 1
			}
			return exitCode, output, err
		}
		return sdlc.DefaultCommandExecutor(c, dir, cmdStr)
	}

	wfOpts := sdlc.WorkflowOptions{
		Workspace: wsPath,
		Brief:     brief,
		Executor:  executor,
		OnEvent: func(ev sdlc.WorkflowEvent) {
			task.mu.Lock()
			task.Stage = ev.Stage
			task.Status = ev.Status
			task.UpdatedAt = ev.Timestamp
			task.mu.Unlock()

			isVerified := true
			label := "Verified System Evidence"
			summary := fmt.Sprintf("Stage changed to %s (%s)", ev.Stage, ev.Status)
			if ev.Message != "" {
				summary = ev.Message
			}

			task.recordActivity(ActivityRecord{
				ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
				Timestamp:  ev.Timestamp,
				Type:       "stage_transition",
				Actor:      "sdlc_engine",
				IsVerified: isVerified,
				Label:      label,
				Summary:    summary,
				Details:    map[string]any{"stage": ev.Stage, "status": ev.Status},
			})
		},
		PlanApprover: func(ctx context.Context, plan *sdlc.Plan) (bool, string, error) {
			task.mu.Lock()
			task.Plan = plan
			task.Stage = sdlc.StagePlan
			task.Status = sdlc.StatusAwaitingPlanApproval
			task.UpdatedAt = time.Now()
			task.mu.Unlock()

			summary := fmt.Sprintf("Proposed execution plan with %d steps", len(plan.Steps))
			task.recordActivity(ActivityRecord{
				ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
				Timestamp:  time.Now(),
				Type:       "plan_created",
				Actor:      "planner_agent",
				IsVerified: false,
				Label:      "AI Model Suggestion",
				Summary:    summary,
				Details:    map[string]any{"steps": plan.Steps, "summary": plan.Summary},
			})

			select {
			case dec := <-task.planDecision:
				label := "Human Decision Gate"
				actionSummary := fmt.Sprintf("Plan approved by %s", dec.reviewer)
				if !dec.approve {
					actionSummary = fmt.Sprintf("Plan rejected by %s", dec.reviewer)
				}
				task.recordActivity(ActivityRecord{
					ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
					Timestamp:  time.Now(),
					Type:       "plan_decided",
					Actor:      dec.reviewer,
					IsVerified: true,
					Label:      label,
					Summary:    actionSummary,
					Details:    map[string]any{"approved": dec.approve},
				})
				return dec.approve, dec.reviewer, nil
			case <-ctx.Done():
				return false, "", ctx.Err()
			}
		},
		ReviewApprover: func(ctx context.Context, summary *sdlc.ReviewSummary) (bool, string, error) {
			task.mu.Lock()
			task.Review = summary
			task.Stage = sdlc.StageReview
			task.Status = sdlc.StatusAwaitingReview
			task.UpdatedAt = time.Now()
			if task.workflow != nil {
				res := task.workflow.Result()
				task.Diffs = res.Diffs
				task.Evidence = res.Evidence
			}
			task.mu.Unlock()

			task.recordActivity(ActivityRecord{
				ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
				Timestamp:  time.Now(),
				Type:       "review_created",
				Actor:      "sdlc_verifier",
				IsVerified: true,
				Label:      "Verified System Evidence",
				Summary:    fmt.Sprintf("Verification complete: %s", summary.ChangeSummary),
				Details:    map[string]any{"evidence": summary.VerificationEvidence, "risks": summary.Risks},
			})

			select {
			case dec := <-task.reviewDecision:
				label := "Human Decision Gate"
				actionSummary := fmt.Sprintf("Changes accepted by %s", dec.reviewer)
				if !dec.approve {
					actionSummary = fmt.Sprintf("Changes rejected by %s", dec.reviewer)
				}
				task.recordActivity(ActivityRecord{
					ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
					Timestamp:  time.Now(),
					Type:       "review_decided",
					Actor:      dec.reviewer,
					IsVerified: true,
					Label:      label,
					Summary:    actionSummary,
					Details:    map[string]any{"approved": dec.approve},
				})
				return dec.approve, dec.reviewer, nil
			case <-ctx.Done():
				return false, "", ctx.Err()
			}
		},
	}

	for _, cfg := range configure {
		cfg(&wfOpts)
	}

	wf, err := sdlc.NewWorkflow(wfOpts)
	if err != nil {
		cp.mu.Lock()
		delete(cp.tasks, key)
		cp.mu.Unlock()
		return nil, fmt.Errorf("failed to initialize workflow: %w", err)
	}

	task.workflow = wf

	// Run workflow asynchronously
	go func() {
		res, runErr := wf.Run(context.Background())
		task.mu.Lock()
		task.UpdatedAt = time.Now()
		if res != nil {
			task.Diffs = res.Diffs
			task.Evidence = res.Evidence
			task.Review = res.Review
			task.Provenance = res.Provenance
		}
		if runErr != nil {
			task.Status = sdlc.StatusFailed
			task.Error = runErr.Error()
		} else {
			task.Status = sdlc.StatusCompleted
		}
		task.mu.Unlock()

		statusSummary := fmt.Sprintf("Workflow ended with status: %s", task.Status)
		if task.Error != "" {
			statusSummary = fmt.Sprintf("Workflow failed: %s", task.Error)
		}
		task.recordActivity(ActivityRecord{
			ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
			Timestamp:  time.Now(),
			Type:       "task_status",
			Actor:      "sdlc_engine",
			IsVerified: true,
			Label:      "Verified System Evidence",
			Summary:    statusSummary,
			Details:    map[string]any{"status": task.Status, "error": task.Error},
		})
	}()

	return task, nil
}

// GetTask retrieves an existing supervised task.
func (cp *ControlPlane) GetTask(wsID, taskID string) (*EngineeringTask, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	task, exists := cp.tasks[taskKey(wsID, taskID)]
	if !exists {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// ListTasks lists all tasks for a workspace.
func (cp *ControlPlane) ListTasks(wsID string) []*EngineeringTask {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	var list []*EngineeringTask
	for key, task := range cp.tasks {
		if strings.HasPrefix(key, wsID+":") {
			list = append(list, task)
		}
	}
	return list
}

// DecidePlan allows a human reviewer to approve or reject an execution plan.
func (cp *ControlPlane) DecidePlan(wsID, taskID string, approve bool, reviewer string) error {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return err
	}
	task.mu.RLock()
	status := task.Status
	task.mu.RUnlock()

	if status != sdlc.StatusAwaitingPlanApproval {
		return fmt.Errorf("task is not awaiting plan approval (status: %s)", status)
	}

	reviewer = strings.TrimSpace(reviewer)
	if reviewer == "" {
		reviewer = "human_operator"
	}

	select {
	case task.planDecision <- decision{approve: approve, reviewer: reviewer}:
		return nil
	default:
		return errors.New("plan decision already delivered or channel full")
	}
}

// DecideReview allows a human reviewer to accept or reject the final code review.
func (cp *ControlPlane) DecideReview(wsID, taskID string, approve bool, reviewer string) error {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return err
	}
	task.mu.RLock()
	status := task.Status
	task.mu.RUnlock()

	if status != sdlc.StatusAwaitingReview {
		return fmt.Errorf("task is not awaiting review approval (status: %s)", status)
	}

	reviewer = strings.TrimSpace(reviewer)
	if reviewer == "" {
		reviewer = "human_operator"
	}

	select {
	case task.reviewDecision <- decision{approve: approve, reviewer: reviewer}:
		return nil
	default:
		return errors.New("review decision already delivered or channel full")
	}
}

// PauseTask pauses active execution of an engineering task.
func (cp *ControlPlane) PauseTask(wsID, taskID, actor string) error {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return err
	}
	task.mu.Lock()
	if task.workflow == nil {
		task.mu.Unlock()
		return ErrTaskHalted
	}
	task.workflow.Pause()
	task.Status = sdlc.StatusPaused
	task.UpdatedAt = time.Now()
	task.mu.Unlock()

	task.recordActivity(ActivityRecord{
		ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Type:       "task_status",
		Actor:      actor,
		IsVerified: true,
		Label:      "Verified System Evidence",
		Summary:    fmt.Sprintf("Task paused by %s", actor),
	})
	return nil
}

// CancelTask cancels an active engineering task.
func (cp *ControlPlane) CancelTask(wsID, taskID, actor string) error {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return err
	}
	task.mu.Lock()
	if task.workflow == nil {
		task.mu.Unlock()
		return ErrTaskHalted
	}
	task.workflow.Cancel()
	task.Status = sdlc.StatusCancelled
	task.UpdatedAt = time.Now()
	task.mu.Unlock()

	task.recordActivity(ActivityRecord{
		ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Type:       "task_status",
		Actor:      actor,
		IsVerified: true,
		Label:      "Verified System Evidence",
		Summary:    fmt.Sprintf("Task cancelled by %s", actor),
	})
	return nil
}

// DecideDiff sets approval status ("accepted" or "rejected") on a specific file diff.
func (cp *ControlPlane) DecideDiff(wsID, taskID, filePath string, accept bool, reviewer string) error {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return err
	}

	task.mu.Lock()
	found := false
	status := "rejected"
	if accept {
		status = "accepted"
	}
	for i := range task.Diffs {
		if task.Diffs[i].Path == filePath {
			task.Diffs[i].Status = status
			found = true
			break
		}
	}
	task.mu.Unlock()

	if !found {
		return fmt.Errorf("diff not found for file %q", filePath)
	}

	task.recordActivity(ActivityRecord{
		ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Type:       "diff_decided",
		Actor:      reviewer,
		IsVerified: true,
		Label:      "Verified System Decision",
		Summary:    fmt.Sprintf("File diff %q %s by %s", filePath, status, reviewer),
		Details:    map[string]any{"path": filePath, "status": status, "reviewer": reviewer},
	})
	return nil
}

// GenerateApprovalSignal issues a cryptographically authenticated approval signal
// for a privileged task operation within a workspace.
func (cp *ControlPlane) GenerateApprovalSignal(wsID, taskID, approverID, secretKey string, ttl time.Duration) (*sec.ApprovalSignal, error) {
	task, err := cp.GetTask(wsID, taskID)
	if err != nil {
		return nil, err
	}

	scope := fmt.Sprintf("dash:task:%s:%s", wsID, taskID)
	sig, err := sec.NewApprovalSignal(approverID, scope, secretKey, ttl)
	if err != nil {
		return nil, err
	}

	task.recordActivity(ActivityRecord{
		ID:         fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Type:       "approval_signal_issued",
		Actor:      approverID,
		IsVerified: true,
		Label:      "Verified System Decision",
		Summary:    fmt.Sprintf("Cryptographic approval signal issued by %s for scope %q", approverID, scope),
		Details:    map[string]any{"scope": scope, "valid_until": sig.ValidUntil},
	})

	return sig, nil
}

// Subscribe opens a real-time event channel for task execution events.
func (task *EngineeringTask) Subscribe() (<-chan ActivityRecord, func()) {
	return task.SubscribeActivities()
}

// SubscribeActivities registers a channel to receive live activity updates for a task.
func (task *EngineeringTask) SubscribeActivities() (<-chan ActivityRecord, func()) {
	task.mu.Lock()
	defer task.mu.Unlock()
	ch := make(chan ActivityRecord, 32)
	task.listeners[ch] = struct{}{}

	// Replay existing activities
	for _, act := range task.Activities {
		ch <- act
	}

	unsubscribe := func() {
		task.mu.Lock()
		defer task.mu.Unlock()
		delete(task.listeners, ch)
		close(ch)
	}
	return ch, unsubscribe
}

func (task *EngineeringTask) recordActivity(act ActivityRecord) {
	task.mu.Lock()
	task.Activities = append(task.Activities, act)
	listeners := make([]chan ActivityRecord, 0, len(task.listeners))
	for ch := range task.listeners {
		listeners = append(listeners, ch)
	}
	task.mu.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- act:
		default:
			// Non-blocking write to avoid stalling the control plane
		}
	}
}
