---
name: orchestrating
description: Use when coordinating a confirmed multi-agent execution graph with dependencies, parallelism, checkpoints, retries, and handoffs. Triggers on orchestrate, graph, phase, dependency, parallel, sequential.
---

# Orchestrating Skill

Coordinates execution without absorbing specialist responsibilities. It requires a manager-produced packet and explicit leader authorization against the Goal-Scope-Decision Design before dispatch.

## When to Use

- A confirmed request spans multiple specialist phases
- Tasks have dependencies, eligible parallel work, checkpoints, or authorized retries

## Instructions

Follow `Understand → Plan → Build → Verify → Summarize`.

### Understand
Require the verbatim manager packet, its Goal-Scope-Decision trace, leader authorization, and approved graph. Internal issues/refs are unsupported; stop and report if any required capability is unavailable.

### Plan
- **Skip:** omit a phase only when its output is not needed; record the reason.
- **Sequential:** run dependent phases after their upstream artifact and approval gate pass.
- **Parallel:** run independent phases concurrently only when inputs and mutable scope do not conflict; join before dependent work.
- Every edge names input, output, acceptance criteria, and approval/checkpoint gate.
- Apply [`rules/agent-ordering.md`](../../rules/agent-ordering.md): topologically sort first, then schedule ready nodes by band, ratio, and deterministic tie-breaks; log every evidence-based reorder.
- Apply [`rules/agent-delegation.md`](../../rules/agent-delegation.md): allow only one bounded nested level below a specialist, cap depth/fan-out/concurrency/attempts, and pass compact contracts rather than full history.

### Build
Dispatch specialists by role: planner owns requirements, builder implementation, tester tests/QA, auditor independent audit, deployer release/infra, maintainer incidents/debt, and writer documentation. Pass only contract-required upstream inputs and compact evidence summaries; send status events to supervisor.

### Verify
Supervisor checkpoints cover start, completion, timeout, failure, retry, cancellation, and escalation. A failed node blocks dependents; orchestrator asks leader whether to retry, re-plan, or stop.

### Summarize
Return graph status and evidence to leader. Leader decides the outcome; learner captures the final decision through RuntimeMemoryInterface and RuntimeContextInterface when supported.

## Lifecycle Contract

`user instruction → leader goal/scope/decision design → manager instruction packet → leader authorization → orchestrator → specialist agents → supervisor checkpoints → leader outcome decision → learner capture`.

The leader produces the goal, scope, and decision design. The manager produces the precise task graph input and never redefines the design or executes specialist work. No orchestrator action may bypass the leader authorization gate.

## Practical Examples

- **Simple:** “Explain this function” → leader defines goal/scope/decision design → manager packages one explorer task → leader authorizes → orchestrator runs explorer sequentially → supervisor checkpoint → leader accepts → learner captures the outcome.
- **Multi-phase:** “Add an invoicing feature” → leader design → manager packet → leader authorization → planner → designer → builder → tester; independent auditor may run parallel with tester after build, then deployer runs only after all gates pass.
- **Failed/retried:** builder times out → supervisor reports timeout → orchestrator blocks tester and informs leader → leader authorizes one retry → supervisor reports success or escalation → leader decides; learner captures the failure and decision.

## Ownership Boundaries

| Role | Owns | Does not own |
|---|---|---|
| manager | decomposition, packet, graph input | redefining goal/scope or authorization |
| leader | goal, scope, decision design, authorization, outcome | specialist execution or runtime monitoring |
| orchestrator | graph, order, handoffs | outcome approval or specialist deliverables |
| supervisor | status, timeout, retry, cancellation, escalation signals | outcome acceptance or task implementation |
| planner/builder/tester/auditor/deployer/maintainer | respective specialist deliverable | graph-wide coordination |
| learner | capture and workflow improvement | live execution decisions |

## Adaptive Modes

| Request Class | Orchestrating Response |
|---------------|----------------------|
| **Clear** | Dispatch the graph and monitor execution |
| **Materially ambiguous** | Return to leader for clarification; do not invent graph order |
| **High-risk** | Reduce concurrency, add checkpoints, require leader approval at each stage |
| **Blocked/unsupported** | Report the blocker to leader; do not proceed past failed dependency |

## Anti-Patterns

- Dispatching before leader authorization of the manager packet
- Calling unrelated phases "parallel" when dependencies conflict
- Treating a supervisor status as an approval or a specialist output

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `supervisor` | Collaborates on runtime health and execution controls | Route work when the deliverable requires runtime health and execution controls. |
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
## Handoff

- **From previous phase:** leader-authorized manager packet, Goal-Scope-Decision trace, and approved graph
- **To next phase:** checkpointed graph result to leader; final outcome evidence to learner

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
