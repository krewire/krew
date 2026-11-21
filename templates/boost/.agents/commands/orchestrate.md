---
name: orchestrate
description: Execute a leader-approved dependency graph across specialist agent phases with supervisor checkpoints.
---

# Orchestrate Command — Execution Graph

Follow `skills/orchestrating/SKILL.md` and `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for the leader-authorized manager packet, its Goal-Scope-Decision trace, and approved graph. Do not start work without all three.

### Understand
Load graph nodes, dependencies, handoff contracts, approval gates, leader authorization, and Goal-Scope-Decision trace. Classify the request per `rules/agent-adaptability.md`.

### Plan
Mark nodes skipped, sequential, or parallel using dependency and shared-scope checks. Apply `rules/agent-ordering.md`.

### Build
Dispatch eligible specialists and request supervisor checkpoints before and after each stage.

### Verify
Stop on failed gates, preserve evidence, and request leader authorization before retrying or changing scope. Apply `rules/agent-delegation.md`.

### Summarize
Return node statuses, artifacts, checkpoint evidence, and blockers to leader.

## Adaptive Behavior

| Request Class | Orchestrate Response |
|---------------|---------------------|
| **Clear** | Dispatch the graph and monitor execution |
| **Materially ambiguous** | Return to leader for design clarification; do not invent graph order |
| **High-risk** | Reduce concurrency, add checkpoints, require leader approval at each stage |
| **Blocked/unsupported** | Report the blocker to leader; do not proceed past failed dependency |

## Examples

- "Execute graph: planner→designer→builder with tester gate after builder"
- "Hold at B — dependency A not satisfied, waiting for leader authorization"

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

Leader receives the complete execution graph result; no leader authorization or user-owned approval is bypassed.
