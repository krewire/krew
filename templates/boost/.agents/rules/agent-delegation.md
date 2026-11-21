---
id: rule-agent-delegation
scope: global
severity: mandatory
enforcement: self-check
related: [AGENTS.md, rules/agent-ordering.md, rules/global-conventions.md, agents/manager.md, agents/leader.md, agents/orchestrator.md, agents/supervisor.md, skills/managing/SKILL.md, skills/orchestrating/SKILL.md]
---

# Agent Delegation — Bounded, Auditable Nested Execution

## Description

This rule governs delegation depth, fan-out, concurrency, retries, context transfer, and scope changes across the leader-first agent lifecycle. It permits one bounded nested level below a specialist to preserve useful decomposition without allowing unrestricted recursion or unbounded resource consumption.

## Trigger

Apply before creating or dispatching any child agent, when retrying a node, when passing a handoff, and whenever a child proposes work that changes scope, side effects, risk, or budget.

## Action

1. Preserve the explicit graph and ownership chain: `leader` defines goal, scope, and decision design; `manager` decomposes that design into the packet and graph input; `leader` authorizes the packet and owns outcome/retry decisions; `orchestrator` schedules the graph and handoffs; `supervisor` reports runtime health and applies only authorized controls.
2. Enforce these hard defaults for each graph:
    - Maximum depth is 4 edges from the top-level task: `leader → manager → orchestrator → specialist → one nested child`. A child at the ceiling cannot delegate.
   - A delegating node has at most 3 direct children.
   - At most 4 child agents are active concurrently. Reduce to 1 when tasks share mutable scope or runtime pressure is detected.
   - A node has at most 2 total attempts. Any retry requires leader authorization and must be idempotent or explicitly checkpointed.
3. Do not create duplicate child tasks, overlap mutable scopes, or propagate full history. Pass compact contracts, required inputs, and evidence summaries.
4. A nested request that changes scope, side effects, risk, or budget must pause and return to `leader`; it cannot proceed silently.
5. If CPU or memory telemetry is unavailable, enforce the structural limits above and report telemetry unavailable. Never invent measurements.

## Validation

Before dispatch, record each node's parent, depth, children, scope, attempt count, contract, and handoff evidence in the graph. Confirm the ceilings, non-overlapping mutable scopes, leader authorization for retries or changed scope, and telemetry availability status. Supervisor checkpoints must make these facts auditable to orchestrator and leader.

## Example — Correct

```text
leader → orchestrator → builder → tester
depth: 3 edges; direct children: 1; active children: 1; attempt: 1/2
handoff: compact contract + changed-file summary + test evidence
```

## Example — Incorrect

```text
builder → tester → auditor → researcher
and retry builder twice without leader authorization
```

This exceeds the depth ceiling, creates uncontrolled recursion, and exceeds the two-attempt node budget without approval.

## Exception

Runtime-specific limits may be lower than these defaults. A lower limit must be recorded by the runtime or supervisor. No exception permits unrestricted recursion, silent scope expansion, duplicate work, or fabricated telemetry.

## Quick References

- [Agent registry and lifecycle](../../AGENTS.md)
- [Ordering rule](agent-ordering.md)
- [Manager skill](../skills/managing/SKILL.md)
- [Orchestrating skill](../skills/orchestrating/SKILL.md)
