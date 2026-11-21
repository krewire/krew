---
id: rule-agent-adaptability
scope: global
severity: mandatory
enforcement: self-check
related: [rules/agent-delegation.md, rules/agent-ordering.md, rules/agent-conflict-resolution.md, skills/managing/SKILL.md, agents/manager.md]
---

# Agent Adaptability — Request Classification and Adaptive Execution

## Description

This rule classifies incoming requests by ambiguity, risk, and complexity and defines the minimum
sufficient execution path for each class. It prevents both over-engineering simple requests and
under-handling complex or ambiguous ones.

## Trigger

Apply whenever an agent receives a user instruction, task handoff, or upstream artifact that
requires action.

## Action

### Request Classification

Classify every incoming request into exactly one primary class:

| Class | Criteria | Agent Response |
|-------|----------|----------------|
| **Clear** | Single interpretation, known scope, low risk, no conflicting requirements | Direct execution with standard workflow |
| **Locally incomplete** | Missing required inputs that the current agent can request without ambiguity | Ask for specific missing inputs; list exactly what is needed |
| **Materially ambiguous** | Two or more materially different readings exist; the wrong pick changes outcome, scope, risk, or acceptance criteria | Return to `leader` with the ambiguity record; do not guess |
| **Conflicting** | Two or more positions, interests, or constraints conflict | Apply `rules/agent-conflict-resolution.md`; escalate when score < 4 |
| **High-risk** | Change affects security, privacy, data integrity, external systems, budget, or is irreversible | Route through leader approval; apply higher verification gates |
| **Blocked/unsupported** | Required runtime capability or internal store is unavailable | Stop and report the gap; do not invent fallbacks |
| **Scope-changing** | Request expands or alters already-confirmed scope | Return to manager for re-packeting; do not silently expand |

### Materiality Test

A request is materially ambiguous when any of these is true:
- Two competent agents would produce different user-visible outcomes.
- The choice changes scope, authorization, side effects, security/privacy, budget, or acceptance criteria.
- The choice determines which agent or lifecycle phase owns the work.

### Specialist Autonomy

A specialist agent may resolve uncertainty that is ALL of:
- Local to its domain expertise.
- Reversible (can be reverted without external side effects).
- Low-risk (no security, privacy, data-integrity, budget, or deployment impact).
- Within the leader-authorized packet scope.

When resolving local uncertainty, the specialist must record the assumption, evidence, and rationale in its handoff.

### Simple Task Shortcut

When a request is Clear class and can be completed by one specialist without orchestration, dependency tracking, or cross-phase verification beyond that specialist's normal scope, the specialist may execute the full Understand → Plan → Build → Verify → Summarize workflow without routing through leader/orchestrator/supervisor. The specialist still follows all applicable delegation, ordering, and confirmation rules.

## Validation

Before execution, verify the request class, the materiality test result, and whether the executing agent has authority for that class. Confirm that material ambiguity was routed to leader, decomposition gaps were returned from manager to leader, high-risk work passed leader authorization, and unsupported capabilities were reported.

## Example — Correct

```text
User: "fix the typo in README"
Classification: Clear — single interpretation, low risk, no scope change.
Agent: direct execution; no orchestrator routing needed.
```

## Example — Incorrect

```text
User: "make the app faster"
Classification: Clear
Agent: starts optimizing random modules.
```

This is materially ambiguous — "faster" has no measurable definition and multiple valid interpretations exist.

## Exception

None for classification or materiality rules. Runtime-specific limits may tighten authority but never widen it.

## Quick References

- [Delegation](agent-delegation.md)
- [Ordering](agent-ordering.md)
- [Conflict resolution](agent-conflict-resolution.md)
- [Manager confirmation gate](../skills/managing/rules/skill-managing-confirmation.md)
