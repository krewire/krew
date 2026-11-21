---
name: supervising
description: Use when monitoring agent runtime status and authorized timeout, retry, cancellation, or escalation controls. Triggers on supervise, checkpoint, timeout, retry, cancel, escalate, status.
---

# Supervising Skill

Provides operational checkpoints for orchestrated nodes. It observes and reports execution health; it does not route, implement, or approve.

## When to Use

- An orchestrator has dispatched a node that needs runtime monitoring
- A node times out, fails, is cancelled, or requires an escalation signal

## Instructions

Follow `Understand → Plan → Build → Verify → Summarize`.

### Understand
Require the node contract, graph identity, checkpoint policy, and leader-authorized controls. Do not invent runtime APIs.

### Plan
Observe start, progress, completion, failure, timeout, retry, cancellation, and escalation states.

### Build
Report structured status to orchestrator. Apply retry/cancellation only when explicitly authorized; otherwise signal leader escalation. Include delegation depth, fan-out, active concurrency, attempt count, mutable-scope conflicts, and telemetry availability in checkpoints; enforce structural limits when telemetry is unavailable.

### Verify
Each checkpoint includes node, state, evidence, and timestamp. Contradictory or missing status is escalated and never marked successful.

### Summarize
Return checkpoint evidence and unresolved signals to orchestrator; it forwards the graph result to leader.

## Ownership Boundaries

Supervisor owns runtime health signals. Orchestrator owns graph execution, leader owns outcome decisions, and specialists own deliverables.

## Adaptive Modes

| Request Class | Supervising Response |
|---------------|---------------------|
| **Clear** | Monitor and report node status with checkpoint evidence |
| **Materially ambiguous** | Return to orchestrator for node contract clarification |
| **High-risk** | Increase checkpoint frequency; require timeout and escalation policy |
| **Blocked/unsupported** | Report missing runtime telemetry; enforce structural limits |

## Anti-Patterns

- Treating monitoring as approval
- Retrying or cancelling without authorization
- Inventing unsupported runtime capabilities

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `orchestrator` | Collaborates on dependency-aware execution coordination | Route work when the deliverable requires dependency-aware execution coordination. |
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
## Handoff

- **From previous phase:** orchestrator node contract and checkpoint policy
- **To next phase:** checkpoint evidence and control signals

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
