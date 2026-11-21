package sdlc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krewire/pkg/ai"
	"github.com/krewire/krewire/pkg/model"
)

// WorkflowOptions configures an SDLC workflow execution.
type WorkflowOptions struct {
	Workspace      string
	Brief          TaskBrief
	Provider       ai.Provider
	Model          string
	HumanApprover  agent.Approver
	PlanApprover   func(ctx context.Context, plan *Plan) (bool, string, error)
	ReviewApprover func(ctx context.Context, review *ReviewSummary) (bool, string, error)
	OnEvent        func(event WorkflowEvent)
	Executor       CommandExecutor

	// Pluggable stage handlers (optional overrides for testing, subagent delegation, or specialized personas)
	UnderstandFn func(ctx context.Context, brief TaskBrief) error
	PlanFn       func(ctx context.Context, brief TaskBrief) (*Plan, error)
	ImplementFn  func(ctx context.Context, brief TaskBrief, plan *Plan, approver agent.Approver) error
	ReviewFn     func(ctx context.Context, brief TaskBrief, diffs []FileDiff, evidence []VerificationResult) (*ReviewSummary, error)
}

// Workflow coordinates the end-to-end SDLC lifecycle with explicit stages and verification.
type Workflow struct {
	opts   WorkflowOptions
	mu     sync.Mutex
	status Status
	stage  Stage
	state  WorkflowResult

	snapshot *WorkspaceSnapshot
	cancelFn context.CancelFunc
	paused   bool
}

// NewWorkflow initializes a new governed SDLC workflow.
func NewWorkflow(opts WorkflowOptions) (*Workflow, error) {
	if err := opts.Brief.Validate(); err != nil {
		return nil, err
	}
	if opts.Workspace == "" {
		opts.Workspace = "."
	}
	if opts.Brief.MaxRetries <= 0 {
		opts.Brief.MaxRetries = 2
	}

	w := &Workflow{
		opts:   opts,
		status: StatusPending,
		stage:  StageUnderstand,
		state: WorkflowResult{
			TaskID:    opts.Brief.ID,
			Status:    StatusPending,
			Stage:     StageUnderstand,
			StartTime: time.Now(),
			Provenance: Provenance{
				TaskID:    opts.Brief.ID,
				AgentID:   "krew-sdlc-agent",
				Model:     opts.Model,
				Timestamp: time.Now(),
			},
			Ledger: model.NewVerificationLedger(opts.Brief.ID),
		},
	}

	if opts.Provider != nil {
		w.state.Provenance.Provider = opts.Provider.Name()
	}

	return w, nil
}

// emit sends a structured event to the registered event listener.
func (w *Workflow) emit(stage Stage, status Status, msg string, payload any) {
	w.mu.Lock()
	w.stage = stage
	w.status = status
	w.state.Stage = stage
	w.state.Status = status
	onEvent := w.opts.OnEvent
	w.mu.Unlock()

	if onEvent != nil {
		onEvent(WorkflowEvent{
			Stage:     stage,
			Status:    status,
			Message:   msg,
			Timestamp: time.Now(),
			Payload:   payload,
		})
	}
}

// Pause pauses the workflow execution safely between stages.
func (w *Workflow) Pause() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = true
	w.status = StatusPaused
	w.state.Status = StatusPaused
	if w.cancelFn != nil {
		w.cancelFn()
	}
}

// Cancel safely terminates active workflow execution.
func (w *Workflow) Cancel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = StatusCancelled
	w.state.Status = StatusCancelled
	w.state.Error = ErrWorkflowCancelled.Error()
	if w.cancelFn != nil {
		w.cancelFn()
	}
}

// Status returns current workflow status and stage.
func (w *Workflow) Status() (Status, Stage) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status, w.stage
}

// Result returns the current workflow state snapshot.
func (w *Workflow) Result() WorkflowResult {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state
}

// Run executes the complete governed SDLC workflow pipeline.
func (w *Workflow) Run(parentCtx context.Context) (*WorkflowResult, error) {
	timeout := w.opts.Brief.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	w.mu.Lock()
	w.cancelFn = cancel
	w.paused = false
	w.mu.Unlock()
	defer cancel()

	// --- 1. Stage: Understand ---
	if err := w.runStageUnderstand(ctx); err != nil {
		if w.checkHalted() {
			return &w.state, w.haltError()
		}
		return w.fail(err)
	}
	if w.checkHalted() {
		return &w.state, w.haltError()
	}

	// --- 2. Stage: Plan ---
	if err := w.runStagePlan(ctx); err != nil {
		if w.checkHalted() {
			return &w.state, w.haltError()
		}
		return w.fail(err)
	}
	if w.checkHalted() {
		return &w.state, w.haltError()
	}

	// --- 3. Stage: Implement & 4. Stage: Verify (bounded retry loop) ---
	if err := w.runImplementAndVerify(ctx); err != nil {
		if w.checkHalted() {
			return &w.state, w.haltError()
		}
		return w.fail(err)
	}
	if w.checkHalted() {
		return &w.state, w.haltError()
	}

	// --- 5. Stage: Review ---
	if err := w.runStageReview(ctx); err != nil {
		if w.checkHalted() {
			return &w.state, w.haltError()
		}
		return w.fail(err)
	}
	if w.checkHalted() {
		return &w.state, w.haltError()
	}

	// --- 6. Stage: Report ---
	return w.runStageReport(ctx)
}

func (w *Workflow) checkHalted() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.paused || w.status == StatusCancelled
}

func (w *Workflow) haltError() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.paused {
		return ErrWorkflowPaused
	}
	return ErrWorkflowCancelled
}

func (w *Workflow) fail(err error) (*WorkflowResult, error) {
	w.mu.Lock()
	if errors.Is(err, ErrVerificationFailed) {
		w.status = StatusVerificationFailed
		w.state.Status = StatusVerificationFailed
	} else if w.paused {
		w.status = StatusPaused
		w.state.Status = StatusPaused
		err = ErrWorkflowPaused
	} else if errors.Is(err, context.Canceled) && w.status == StatusCancelled {
		err = ErrWorkflowCancelled
	} else {
		w.status = StatusFailed
		w.state.Status = StatusFailed
	}
	w.state.Error = err.Error()
	w.state.EndTime = time.Now()
	res := w.state
	st := w.status
	w.mu.Unlock()
	w.emit(w.stage, st, err.Error(), nil)
	return &res, err
}

// 1. Understand
func (w *Workflow) runStageUnderstand(ctx context.Context) error {
	if w.checkHalted() {
		return w.haltError()
	}
	select {
	case <-ctx.Done():
		if w.checkHalted() {
			return w.haltError()
		}
		return ctx.Err()
	default:
	}
	w.emit(StageUnderstand, StatusInProgress, "Analyzing repository context and scope", nil)

	snap, err := TakeSnapshot(w.opts.Workspace, w.opts.Brief.Scope)
	if err != nil {
		return fmt.Errorf("failed taking workspace snapshot: %w", err)
	}
	w.mu.Lock()
	w.snapshot = snap
	w.mu.Unlock()

	if w.opts.UnderstandFn != nil {
		return w.opts.UnderstandFn(ctx, w.opts.Brief)
	}
	return nil
}

// 2. Plan
func (w *Workflow) runStagePlan(ctx context.Context) error {
	if w.checkHalted() {
		return w.haltError()
	}
	select {
	case <-ctx.Done():
		if w.checkHalted() {
			return w.haltError()
		}
		return ctx.Err()
	default:
	}
	w.emit(StagePlan, StatusInProgress, "Generating execution plan", nil)

	var plan *Plan
	var err error

	if w.opts.PlanFn != nil {
		plan, err = w.opts.PlanFn(ctx, w.opts.Brief)
	} else {
		// Default plan generation
		plan = &Plan{
			Summary: fmt.Sprintf("Implementation plan for: %s", w.opts.Brief.Title),
			Steps: []PlanStep{
				{
					ID:          "step-1",
					Description: w.opts.Brief.Objective,
					TargetFiles: w.opts.Brief.Scope,
					RiskLevel:   "medium",
				},
			},
		}
	}

	if err != nil {
		return fmt.Errorf("planning stage failed: %w", err)
	}

	w.mu.Lock()
	w.state.Plan = plan
	w.mu.Unlock()

	w.emit(StagePlan, StatusInProgress, "Plan generated", plan)

	// Approval gate for plan
	if w.opts.Brief.RequirePlan {
		w.emit(StagePlan, StatusAwaitingPlanApproval, "Waiting for human plan approval", plan)
		if w.opts.PlanApprover != nil {
			approved, reviewer, appErr := w.opts.PlanApprover(ctx, plan)
			if appErr != nil {
				return appErr
			}
			if !approved {
				return fmt.Errorf("%w: plan was rejected by %s", ErrApprovalDenied, reviewer)
			}
			plan.Approved = true
			plan.ApprovedBy = reviewer
			plan.ApprovedAt = time.Now()
		} else {
			// Auto-approve if no interactive approver hook is supplied
			plan.Approved = true
			plan.ApprovedBy = "system"
			plan.ApprovedAt = time.Now()
		}
	} else {
		plan.Approved = true
	}

	return nil
}

// 3. Implement & 4. Verify loop
func (w *Workflow) runImplementAndVerify(ctx context.Context) error {
	maxRetries := w.opts.Brief.MaxRetries
	verifier := NewVerifier(w.opts.Workspace, w.opts.Executor)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			if w.checkHalted() {
				return w.haltError()
			}
			return ctx.Err()
		default:
		}

		if w.checkHalted() {
			return w.haltError()
		}

		// Stage: Implement
		w.emit(StageImplement, StatusInProgress, fmt.Sprintf("Implementing changes (attempt %d/%d)", attempt+1, maxRetries+1), nil)

		approver := NewScopedApprover(w.opts.Brief, func() Stage { return StageImplement }, w.opts.HumanApprover)

		if w.opts.ImplementFn != nil {
			if err := w.opts.ImplementFn(ctx, w.opts.Brief, w.state.Plan, approver); err != nil {
				return fmt.Errorf("implementation error: %w", err)
			}
		}

		// Stage: Verify
		w.emit(StageVerify, StatusInProgress, "Running deterministic verification commands", nil)

		evidence, passed := verifier.Run(ctx, w.opts.Brief.VerificationCommands)
		w.mu.Lock()
		w.state.Evidence = evidence
		w.state.Provenance.VerificationEvidence = evidence

		if w.state.Ledger != nil {
			var modelGates []model.VerificationEvidence
			for _, res := range evidence {
				status := model.GatePassed
				if !res.Passed {
					status = model.GateFailed
				}
				ev := model.VerificationEvidence{
					GateID:     res.Command,
					Type:       model.GateTypeTest,
					Command:    res.Command,
					Required:   true,
					Status:     status,
					ExitCode:   res.ExitCode,
					Output:     res.Output,
					DurationMs: res.Duration.Milliseconds(),
				}
				ev.EvidenceHash = ev.ComputeHash(w.opts.Brief.ID)
				modelGates = append(modelGates, ev)
			}
			_, _ = w.state.Ledger.AppendBlock(w.opts.Brief.ID, modelGates)
		}
		w.mu.Unlock()

		if passed {
			w.emit(StageVerify, StatusInProgress, "All verification checks passed", evidence)
			return nil
		}

		// Verification failed
		w.emit(StageVerify, StatusVerificationFailed, fmt.Sprintf("Verification failed on attempt %d", attempt+1), evidence)

		if attempt >= maxRetries {
			w.mu.Lock()
			w.status = StatusVerificationFailed
			w.state.Status = StatusVerificationFailed
			w.state.Error = "verification checks failed: maximum retries reached"
			w.mu.Unlock()
			return fmt.Errorf("%w: checks failed after %d retries", ErrVerificationFailed, maxRetries)
		}
	}

	return ErrVerificationFailed
}

// 5. Review
func (w *Workflow) runStageReview(ctx context.Context) error {
	if w.checkHalted() {
		return w.haltError()
	}
	select {
	case <-ctx.Done():
		if w.checkHalted() {
			return w.haltError()
		}
		return ctx.Err()
	default:
	}
	w.emit(StageReview, StatusInProgress, "Preparing change review and diff summary", nil)

	// Compute workspace diffs
	var diffs []FileDiff
	if w.snapshot != nil {
		var err error
		diffs, err = w.snapshot.ComputeDiff()
		if err != nil {
			return fmt.Errorf("failed computing workspace diffs: %w", err)
		}
	}

	w.mu.Lock()
	w.state.Diffs = diffs
	w.mu.Unlock()

	var review *ReviewSummary
	if w.opts.ReviewFn != nil {
		var err error
		review, err = w.opts.ReviewFn(ctx, w.opts.Brief, diffs, w.state.Evidence)
		if err != nil {
			return fmt.Errorf("review generation failed: %w", err)
		}
	} else {
		review = &ReviewSummary{
			ChangeSummary:        fmt.Sprintf("%d file(s) modified across scope", len(diffs)),
			VerificationEvidence: w.state.Evidence,
			Risks:                []string{},
			UnresolvedQuestions:  []string{},
		}
	}

	w.mu.Lock()
	w.state.Review = review
	w.mu.Unlock()

	// Review Approval Gate
	if w.opts.Brief.RequireReview {
		w.emit(StageReview, StatusAwaitingReview, "Waiting for human change review", review)
		if w.opts.ReviewApprover != nil {
			approved, reviewer, err := w.opts.ReviewApprover(ctx, review)
			if err != nil {
				return err
			}
			if !approved {
				return fmt.Errorf("%w: changes rejected by %s", ErrApprovalDenied, reviewer)
			}
			review.Approved = true
			review.ReviewedBy = reviewer
			review.ReviewedAt = time.Now()
		} else {
			review.Approved = true
			review.ReviewedBy = "system"
			review.ReviewedAt = time.Now()
		}
	} else {
		review.Approved = true
	}

	return nil
}

// 6. Report
func (w *Workflow) runStageReport(ctx context.Context) (*WorkflowResult, error) {
	if w.checkHalted() {
		return &w.state, w.haltError()
	}
	select {
	case <-ctx.Done():
		if w.checkHalted() {
			return &w.state, w.haltError()
		}
		return nil, ctx.Err()
	default:
	}
	w.mu.Lock()
	w.stage = StageReport
	w.status = StatusCompleted
	w.state.Stage = StageReport
	w.state.Status = StatusCompleted
	w.state.EndTime = time.Now()

	// Populate file provenance lists
	var created, modified []string
	for _, d := range w.state.Diffs {
		if d.Action == "created" {
			created = append(created, d.Path)
		} else {
			modified = append(modified, d.Path)
		}
	}
	w.state.Provenance.CreatedFiles = created
	w.state.Provenance.ModifiedFiles = modified

	result := w.state
	w.mu.Unlock()

	w.emit(StageReport, StatusCompleted, "Workflow completed successfully", result)
	return &result, nil
}
