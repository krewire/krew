---
id: rule-agent-ordering
scope: global
severity: mandatory
enforcement: self-check
related: [AGENTS.md, rules/agent-delegation.md, rules/global-conventions.md, agents/manager.md, agents/leader.md, agents/orchestrator.md, agents/supervisor.md, skills/managing/SKILL.md, skills/orchestrating/SKILL.md]
---

# Agent Ordering — Deterministic Dependency-Aware Scheduling

## Description

This rule governs the order in which leader-authorized graph nodes become eligible for execution. It makes dependencies and critical work authoritative, then uses an evidence-backed impact-to-effort ratio and deterministic tie-breakers so scheduling is repeatable and auditable without bypassing leader authorization.

## Trigger

Apply before execution, when selecting among dependency-ready nodes, when deciding whether to parallelize, and whenever new evidence, a failed gate, a dependency change, or a scope change may alter the order.

## Action

1. Build the dependency DAG and topologically sort it before execution. Every dependency precedes its dependents.
2. Prioritize dependency-ready nodes by bands: **P0 Critical**, **Important-first**, **Urgent-only**, then **Neither**. P0 covers core-flow breakage, data loss, and security; it supersedes ratio but still respects prerequisites unless waiting would increase harm.
3. Within a band, calculate the canonical ratio as `impact ÷ effort`, with both scores from 1 to 5:
   - Impact: `1` cosmetic or negligible, `2` limited inconvenience, `3` material user or operational effect, `4` major mission or reliability effect, `5` core-flow failure, data loss, or security harm.
   - Effort: `1` isolated, low-risk change, `2` small bounded change, `3` multi-file or integration work, `4` substantial cross-component work, `5` broad or high-risk work requiring staged coordination.
   Record evidence or explicit assumptions for every score.
4. Break equal ratios in this order: (a) task that unblocks more downstream nodes, (b) higher impact, (c) lower effort, (d) earlier creation or stable graph order.
5. Re-score only after new evidence, a failed gate, a dependency change, or a scope change. Record the reason for every reorder.
6. Never parallelize tasks that share mutable scope. Run them sequentially with a checkpoint. Ordering is scheduling only; it does not grant permission to bypass leader authorization or user-owned approval.

## Validation

Inspect the recorded DAG, topological order, band, scores, evidence or assumptions, tie-break results, parallel-scope decision, and any reorder reason before dispatch. Confirm that blocked nodes were not scheduled and that P0 handling did not bypass prerequisites without a documented harm rationale.

## Example — Correct

```text
DAG: A → B; A → C
Ready: A (Important-first, impact 4, effort 2, ratio 2.0)
Order: A, then B/C after A's gate; B and C sequential if they share mutable scope.
```

## Example — Incorrect

```text
Run B first because its ratio is 5.0, although B depends on A.
```

This violates the topological prerequisite and treats a blocked task as ready.

## Exception

A P0 task may precede a prerequisite only when waiting would increase core-flow, data-loss, or security harm; leader must record the harm rationale and the resulting checkpoint. Runtime pressure may reduce parallelism to one active child, but it does not alter dependency order or approval gates.

## Quick References

- [Agent registry and lifecycle](../../AGENTS.md)
- [Delegation rule](agent-delegation.md)
- [Global conventions](global-conventions.md)
- [Manager skill](../skills/managing/SKILL.md)
- [Orchestrating skill](../skills/orchestrating/SKILL.md)
