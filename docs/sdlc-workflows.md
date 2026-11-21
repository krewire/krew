# Governed SDLC Workflows (`pkg/sdlc`)

Krew provides governed, repeatable software development lifecycle (SDLC) workflows that transition autonomous agents from free-form chat assistants into accountable engineering participants.

---

## 1. Core Principles

- **Separation of Planning and Execution**: Agents must produce and expose a structured plan (`Plan`) before mutating files.
- **Strict Scope Boundaries**: Modifying files or executing commands outside the declared `TaskBrief.Scope` is rejected (`ErrScopeViolation`).
- **Deterministic Verification Evidence**: Changes are verified with reproducible test and build commands. The agent cannot claim completion if any check fails.
- **Interactive Human Approval Gates**: High-risk actions and plan approval can be gated on human reviewer authorization.
- **Traceable Provenance**: Generated changes record task, agent, model/provider, tools used, and verification evidence.

---

## 2. SDLC Workflow Stages

```text
Understand ──► Plan ──► Implement ──► Verify ──► Review ──► Report
                             ▲            │
                             └──(retry)───┘
```

1. **`understand`**: Examines repository context, constraints, and dependencies using read-only tools.
2. **`plan`**: Outlines atomic implementation steps, target files, and risk levels. Requires human approval when configured.
3. **`implement`**: Restricts tool permissions strictly to granted scope; applies atomic code edits.
4. **`verify`**: Executes deterministic verification commands. If checks fail, triggers bounded retries; fails closed if checks cannot be satisfied.
5. **`review`**: Computes unified diffs, summarizes changes, and surfaces architectural or security risks for human review.
6. **`report`**: Produces a provenance-linked audit record of the completed workflow.

---

## 3. Usage Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/krewire/krew/pkg/sdlc"
)

func main() {
	brief, err := sdlc.ParseTaskBriefFromMarkdown(
		"ISSUE-42",
		"Resilient database health check",
		`## Objective
Implement resilient database health check with timeout.

## Scope
- pkg/db

## Verification Commands
- go test -v ./pkg/db/...
`,
	)
	if err != nil {
		log.Fatal(err)
	}

	wf, err := sdlc.NewWorkflow(sdlc.WorkflowOptions{
		Workspace: ".",
		Brief:     brief,
		PlanApprover: func(ctx context.Context, plan *sdlc.Plan) (bool, string, error) {
			fmt.Printf("Plan summary: %s (Steps: %d)\n", plan.Summary, len(plan.Steps))
			return true, "lead-engineer", nil
		},
		ReviewApprover: func(ctx context.Context, rev *sdlc.ReviewSummary) (bool, string, error) {
			fmt.Printf("Change summary: %s\n", rev.ChangeSummary)
			return true, "security-reviewer", nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	res, err := wf.Run(context.Background())
	if err != nil {
		log.Fatalf("workflow failed: %v", err)
	}

	fmt.Printf("Status: %s, Stage: %s, Diffs: %d\n", res.Status, res.Stage, len(res.Diffs))
}
```

---

## 4. Specialized SDLC Personas

- `persona.SDLCPlanner`: Architectural analysis, dependency mapping, and risk estimation.
- `persona.SDLCImplementer`: Test-driven, scoped atomic code changes.
- `persona.SDLCVerifier`: Deterministic verification, test error root-cause diagnostics.
- `persona.SDLCReviewer`: Diff review, security risk identification, and human review preparation.
