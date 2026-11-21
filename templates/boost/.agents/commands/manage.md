---
name: manage
description: Clarify vague requests, package precise instructions, and delegate to the fittest subagent. Use when request is vague, ambiguous, multi-intent, needs confirmation, needs disambiguation, needs routing, or needs orchestration.
---

# Manage Command — Instruction Decomposition & Task-Graph Packaging

Decomposition command for a leader-defined Goal-Scope-Decision Design. Converts the design into a precise Instruction Packet plus task-graph input for leader authorization and later orchestration. Manager does not redefine the outcome or execute specialist work.

Follow `skills/managing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for the leader's Goal-Scope-Decision Design, including goal, scope boundaries, decision authority, constraints, and acceptance gates. Do not invent user intent or scope.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Read the leader design plus matching RuntimeMemoryInterface/RuntimeContextInterface topic; extract goals, entities, constraints, decision criteria, implementation dependencies, and unknowns. Separate facts (sourced) from assumptions (labeled). Every gap becomes a design gap for leader review — never invented intent.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

When decomposition is ambiguous, draft 2–3 implementation options with task, risk, and dependency deltas for leader review. Pick exactly one specialist target per task from the `AGENTS.md` registry with rationale + why-not-others. Describe graph dependencies, sequential/parallel eligibility, per-task input/output, approval gates, and handoffs. Define testable acceptance criteria; state non-goals. The immediate handoff is to `leader` for authorization.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update a linked GitHub Issue with progress and resolving commit SHA; if internal issue tracking is required, stop and report the unsupported capability.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Produce the **Instruction Packet** in the response and STOP:

```markdown
## Instruction Packet
1. Leader Design Reference (version + traceability):
2. Restated Goal (primary + supporting outcomes):
3. Decision Design (question, options, criteria, authority, gates):
4. Scope In / Out:
5. Constraints and Dependencies:
6. Assumptions and Implementation Gaps (owner):
7. Risk and Side-Effect Controls:
8. Specialist Targets + Rationale + Why-not-others:
9. Task Graph (ordered nodes, edges, inputs, outputs, gates):
10. Acceptance Criteria and Evidence:
11. Leader Authorization Request (packet version, scope delta, unresolved items):
```

Do NOT delegate, execute, or write any file before explicit leader authorization. Return any goal, scope, authority, or decision-design delta to leader. Expand every vague adjective into a measurable criterion or an implementation gap.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the linked GitHub Issue with gate results when present; do not create a local issue or ref store.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Run the ambiguity check (no task line readable two ways), the design-integrity check (goals, scope, decision criteria, and authority preserved), and the authorization check (packet returned to leader; orchestration withheld).

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the linked GitHub Issue when present; do not create a local issue or ref store.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Report the packet version, leader design reference, specialist order, scope delta, and authorization state. Hand the packet to `leader`; after authorization, `orchestrator` must receive it verbatim.

## Anti-Patterns

- Delegating or executing before explicit leader authorization
- Inventing intent instead of asking
- Leaving vague adjectives unmeasured
- Routing to the wrong subagent or bundling multiple concerns into one task
- Dropping Open Questions without owners

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{request}}` | string | Yes | — | Raw user request to clarify, package, and route |
| `--target` | string | No | auto | Suggested subagent; manager still validates fit and may override with rationale |
| `--split` | flag | No | false | Force multi-task splitting even when the request looks singular |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `orchestrator` | Collaborates on dependency-aware execution coordination | Route work when the deliverable requires dependency-aware execution coordination. |
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |
## Handoff

The Instruction Packet and ordered delegation targets returned to leader for authorization. After authorization, orchestrator executes the packet verbatim; manager does not dispatch directly.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
