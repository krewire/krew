---
id: rule-agent-conflict-resolution
scope: global
severity: mandatory
enforcement: self-check
related: [agents/thinker.md, agents/leader.md, agents/manager.md, agents/orchestrator.md, rules/agent-delegation.md, rules/agent-ordering.md, skills/thinking/SKILL.md, skills/managing/rules/skill-managing-confirmation.md]
---

# Agent Conflict Resolution — Win-Win Solution Score

## Description

This rule standardizes how any agent evaluates and resolves conflicting positions. The Win-Win
Solution Score measures the quality of the proposed outcome for the conflicting objectives; it is
not a confidence score, impact score, priority score, or permission to execute. The score supports
efficient internal decisions while preserving leader authorization and every
applicable authority, security, privacy, budget, deployment, and consent control.

## Trigger

Apply whenever an agent finds competing positions, interests, constraints, interpretations,
requirements, policies, recommendations, or execution options that require a choice. Apply before
resolving the conflict, proceeding, or claiming that a solution is win-win.

## Action

1. Record the conflicting positions, underlying interests, constraints, alternatives considered,
   evidence and assumptions, residual risks, reversibility, and the rationale for the selected
   score. Abstain and escalate if evidence is insufficient, an interest is misrepresented, an
   authority or policy conflicts, or the score is disputed.
2. Assign exactly one Win-Win Solution Score from 0 through 5 using these criteria:

   | Score | Required meaning |
   |---:|---|
   | 0 | No feasible common solution; objectives are mutually exclusive. |
   | 1 | One side wins and the other suffers material loss. |
   | 2 | Superficial compromise; one side still has a major unmet objective. |
   | 3 | Acceptable trade-off; both sides retain meaningful value, with known residual cost. |
   | 4 | Strong mutual benefit; both primary objectives are met with no material loss. |
   | 5 | Aligned solution; both objectives are met or improved, risks are reduced, and no material downside is identified. |

3. For scores 0–3, escalate to `leader` for goal, scope, decision-design, or outcome decisions,
   `manager` for decomposition gaps, or `thinker` for analysis. Obtain user confirmation when the
   decision is user-owned or changes the leader-authorized design.
4. For scores 4–5, no extra user confirmation is needed for the conflict itself when the decision
   is internal, reversible, low-risk, within the leader-authorized design, and does not change
   authorization. Report the decision, evidence, score, and rationale in the handoff and capture
   it through RuntimeContextInterface when appropriate.
5. A score of 4 or 5 never bypasses leader authorization,
   security or privacy controls, budget or resource limits, deployment or release approval, or
   consent for external or irreversible side effects. Pause and request the required approval
   whenever any of those boundaries is affected.

## Validation

Before proceeding, self-check that the conflict record contains every required field, the score
matches its table definition, and the score has not been substituted for confidence, impact,
priority, or authorization. Confirm packet scope, reversibility, risk, authority, resource limits,
and side effects. Review the handoff for the decision and score; use runtime context only through
the supported RuntimeContextInterface, never a local store.

## Example — Correct

```text
Conflict: implementation speed vs. testability.
Interests: deliver the confirmed change promptly; preserve regression protection.
Constraints: leader-authorized packet scope; no production release in this step.
Alternatives: skip tests (rejected); add a focused characterization test (selected).
Evidence/assumptions: the changed behavior has one isolated boundary; test command is available.
Residual risks: broader integration regressions remain. Reversibility: revert the isolated change.
Win-Win Solution Score: 4 — both primary objectives are met with no material loss.
Decision: proceed internally without an extra conflict confirmation; report the record in handoff.
```

## Example — Incorrect

```text
Score: 5 because the agent is highly confident and the task is high priority.
Decision: deploy immediately without leader, release, security, or user consent checks.
```

This confuses the score with confidence and priority and uses it as execution authorization.

## Exception

None for the score definitions or approval boundaries. A lower-level runtime or policy control may
require additional approval. Emergency authority, if any, comes from the applicable incident or
operational procedure and does not convert a score into permission; the resolver must still record
the conflict and escalate when its evidence or authority is insufficient.

## Quick References

- [Thinker scoring contract](../agents/thinker.md)
- [Leader authority boundary](../agents/leader.md)
- [Manager confirmation gate](../skills/managing/rules/skill-managing-confirmation.md)
- [Delegation boundaries](agent-delegation.md)
- [Execution ordering](agent-ordering.md)
- [Thinking skill](../skills/thinking/SKILL.md)
