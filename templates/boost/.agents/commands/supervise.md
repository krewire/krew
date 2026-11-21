---
name: supervise
description: Monitor an orchestrated agent node and report status, timeout, retry, cancellation, or escalation checkpoints.
---

# Supervise Command — Runtime Monitoring

Follow `skills/supervising/SKILL.md` and `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for the orchestrator node contract and checkpoint policy. Do not invent runtime controls.

### Understand
Load the node identity, authorized controls, and checkpoint policy. Classify the request per `rules/agent-adaptability.md`.

### Plan
Set start, progress, completion, and failure observations without changing scope.

### Build
Monitor and report status; apply retry or cancellation only when authorized.

### Verify
Check status evidence, timeout handling, and escalation completeness.

### Summarize
Return checkpoint history and unresolved signals to orchestrator.

## Adaptive Behavior

| Request Class | Supervise Response |
|---------------|-------------------|
| **Clear** | Monitor and report node status with checkpoint evidence |
| **Materially ambiguous** | Return to orchestrator for node contract clarification |
| **High-risk** | Increase checkpoint frequency; require timeout and escalation policy |
| **Blocked/unsupported** | Report missing runtime telemetry; enforce structural limits |

## Examples

- "Monitor builder node with 5-min timeout and canary checkpoint"
- "Escalate — tester exceeded timeout, retry not authorized"

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

The orchestrator receives runtime evidence; leader remains the outcome authority.
