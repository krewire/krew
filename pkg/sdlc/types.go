package sdlc

import (
	"errors"
	"time"

	"github.com/krewire/krewire/pkg/model"
)

// Standard SDLC error definitions.
var (
	ErrScopeViolation     = errors.New("sdlc: file or command execution is outside granted scope")
	ErrPlanRequired       = errors.New("sdlc: implementation requires an approved plan")
	ErrVerificationFailed = errors.New("sdlc: required verification checks failed")
	ErrWorkflowCancelled  = errors.New("sdlc: workflow execution was cancelled")
	ErrWorkflowPaused     = errors.New("sdlc: workflow execution is paused")
	ErrApprovalDenied     = errors.New("sdlc: human approval was denied")
	ErrInvalidBrief       = errors.New("sdlc: task brief is invalid or missing required fields")
)

// Stage represents a governed SDLC workflow stage.
type Stage string

const (
	StageUnderstand Stage = "understand"
	StagePlan       Stage = "plan"
	StageImplement  Stage = "implement"
	StageVerify     Stage = "verify"
	StageReview     Stage = "review"
	StageReport     Stage = "report"
)

// Status represents the operational state of an SDLC workflow.
type Status string

const (
	StatusPending              Status = "pending"
	StatusInProgress           Status = "in_progress"
	StatusAwaitingPlanApproval Status = "awaiting_plan_approval"
	StatusAwaitingReview       Status = "awaiting_review"
	StatusAwaitingApproval     Status = "awaiting_approval"
	StatusCompleted            Status = "completed"
	StatusVerificationFailed   Status = "verification_failed"
	StatusFailed               Status = "failed"
	StatusCancelled            Status = "cancelled"
	StatusPaused               Status = "paused"
)

// TaskBrief specifies the task requirements, bounds, constraints, and qualification criteria.
type TaskBrief struct {
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	Objective            string        `json:"objective"`
	Repository           string        `json:"repository"`
	Scope                []string      `json:"scope"`
	Constraints          []string      `json:"constraints"`
	AcceptanceCriteria   []string      `json:"acceptance_criteria"`
	VerificationCommands []string      `json:"verification_commands"`
	RequirePlan          bool          `json:"require_plan"`
	RequireReview        bool          `json:"require_review"`
	Timeout              time.Duration `json:"timeout,omitempty"`
	MaxRetries           int           `json:"max_retries,omitempty"`
}

// PlanStep defines an atomic action in an SDLC execution plan.
type PlanStep struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	TargetFiles []string `json:"target_files"`
	RiskLevel   string   `json:"risk_level"` // "low", "medium", "high"
}

// Plan is the structured plan generated in StagePlan prior to file modifications.
type Plan struct {
	Summary    string     `json:"summary"`
	Steps      []PlanStep `json:"steps"`
	Approved   bool       `json:"approved"`
	ApprovedBy string     `json:"approved_by,omitempty"`
	ApprovedAt time.Time  `json:"approved_at,omitempty"`
}

// VerificationResult captures deterministic test or check execution outcomes.
type VerificationResult struct {
	Command   string        `json:"command"`
	ExitCode  int           `json:"exit_code"`
	Passed    bool          `json:"passed"`
	Output    string        `json:"output"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// FileDiff describes file changes produced by the implementation stage.
type FileDiff struct {
	Path         string `json:"path"`
	Action       string `json:"action"` // "created", "modified", "deleted"
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
	DiffText     string `json:"diff_text,omitempty"`
	Status       string `json:"status,omitempty"` // "pending", "accepted", "rejected"
}

// Provenance tracks the complete provenance of generated changes.
type Provenance struct {
	TaskID               string               `json:"task_id"`
	AgentID              string               `json:"agent_id"`
	Model                string               `json:"model"`
	Provider             string               `json:"provider"`
	ToolsUsed            []string             `json:"tools_used"`
	VerificationEvidence []VerificationResult `json:"verification_evidence"`
	CreatedFiles         []string             `json:"created_files"`
	ModifiedFiles        []string             `json:"modified_files"`
	Timestamp            time.Time            `json:"timestamp"`
}

// ReviewSummary encapsulates human review data, diffs, risks, and questions.
type ReviewSummary struct {
	ChangeSummary        string               `json:"change_summary"`
	VerificationEvidence []VerificationResult `json:"verification_evidence"`
	Risks                []string             `json:"risks"`
	UnresolvedQuestions  []string             `json:"unresolved_questions"`
	Approved             bool                 `json:"approved"`
	ReviewedBy           string               `json:"reviewed_by,omitempty"`
	ReviewedAt           time.Time            `json:"reviewed_at,omitempty"`
}

// WorkflowEvent is emitted during stage transitions and major milestone events.
type WorkflowEvent struct {
	Stage     Stage     `json:"stage"`
	Status    Status    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload,omitempty"`
}

// WorkflowResult is the final output of a completed or halted SDLC workflow.
type WorkflowResult struct {
	TaskID     string                    `json:"task_id"`
	Status     Status                    `json:"status"`
	Stage      Stage                     `json:"stage"`
	Plan       *Plan                     `json:"plan,omitempty"`
	Diffs      []FileDiff                `json:"diffs,omitempty"`
	Evidence   []VerificationResult      `json:"evidence,omitempty"`
	Review     *ReviewSummary            `json:"review,omitempty"`
	Provenance Provenance                `json:"provenance"`
	Ledger     *model.VerificationLedger `json:"ledger,omitempty"`
	Error      string                    `json:"error,omitempty"`
	StartTime  time.Time                 `json:"start_time"`
	EndTime    time.Time                 `json:"end_time"`
}
