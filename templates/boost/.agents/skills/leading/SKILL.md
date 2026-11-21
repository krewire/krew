---
name: leading
description: Use when owning outcomes and approval gates across a confirmed multi-agent task. Triggers on lead, approve, reject, outcome, gate, escalation, retry, cancellation.
---

# Leading Skill

Defines outcome ownership from the user's instruction through final acceptance. It identifies and defines goals, scope, and decision design, hands the design to manager for decomposition, authorizes the resulting packet, and keeps approval decisions separate from routing, execution, monitoring, specialist work, and learning.

## When to Use

- A user instruction needs a Goal-Scope-Decision Design before decomposition
- A manager packet needs validation against the leader design before orchestration
- A dependency fails or scope requires a pause, retry, cancellation, or escalation
- A nested delegation changes scope, side effects, risk, budget, or the approved execution order
- A user instruction needs its desired outcome separated from proposed activities, boundaries, or non-goals

## Instructions

Follow `Understand → Plan → Build → Verify → Summarize`.

### Understand
Read the user instruction, context, and any manager packet, using the leader-authorized packet as the execution boundary. Identify the intended user-visible outcome, stakeholders, requested activity, success language, constraints, decision rights, and design traceability. Record ambiguity, contradictions, missing measures, and assumptions. Internal issues/refs are unsupported, so stop and report if required.

### Plan
Synthesize a Goal-Scope-Decision Design containing: primary and supporting goals, measurable success criteria, in-scope deliverables, out-of-scope items, assumptions, constraints, dependencies, risks, decision criteria, change triggers, approval owner, and evidence required. Review the manager packet against that design and authorize only the phases needed for the requested outcome; user confirmation remains mandatory for user-owned decisions.

### Build
Send the synthesized contract and a decision to `orchestrator`: proceed, hold, retry, cancel, or escalate. Never perform the specialist task. Apply [`rules/agent-delegation.md`](../../rules/agent-delegation.md) and [`rules/agent-ordering.md`](../../rules/agent-ordering.md) before approving retries, scope changes, or prerequisite exceptions.

### Verify
Accept only evidence-backed outputs that satisfy the user outcome and remain inside the scope contract. Check for scope leakage, unaddressed in-scope work, invalidated assumptions, and success measures that were replaced by activity completion. A failed gate returns to the responsible specialist through orchestrator; repeated or blocked failure is escalated rather than silently waived.

### Summarize
Publish the goal and scope synthesis, final outcome decision, acceptance evidence, unresolved gaps, invalidated assumptions, out-of-scope follow-ups, and evidence for `learner` capture.

## Goal and Scope Synthesis Method

Use this sequence for every instruction:

1. **Identify**: extract the stated request, implied outcome, stakeholders, affected surfaces, requested activity, constraints, and success language.
2. **Separate**: distinguish the goal from the proposed means; distinguish authorized work from useful but unrequested improvements.
3. **Define**: write measurable success criteria and explicit in-scope, out-of-scope, assumption, constraint, dependency, risk, and change-trigger fields.
4. **Test materiality**: route ambiguity when alternate interpretations change user-visible outcome, scope, authority, side effects, risk, budget, or acceptance criteria.
5. **Synthesize**: produce one outcome contract that maps every goal to a gate and every scope item to an owner, phase, or exclusion.
6. **Authorize**: approve only the confirmed contract; treat any later expansion as a new manager packet or user decision.

### Required Outcome Contract

| Field | Minimum content |
|-------|-----------------|
| **Primary goal** | User-visible outcome in one sentence |
| **Supporting goals** | Necessary secondary outcomes with priority |
| **Success measures** | Observable evidence, acceptance criterion, owner, and deadline when known |
| **In scope** | Deliverables, surfaces, actors, phases, and authorized side effects |
| **Out of scope** | Explicit exclusions and deferred follow-ups |
| **Assumptions** | Unverified facts, evidence, and owner for validation |
| **Constraints** | Security, privacy, policy, technical, compatibility, resource, or timing limits |
| **Dependencies** | Prerequisites, upstream decisions, external systems, and handoffs |
| **Change triggers** | Events that require re-confirmation or re-packeting |

The contract must preserve three distinct layers: **instruction** (what was said), **intent** (what outcome is inferred), and **authorization** (what execution is approved). An inferred intent never silently becomes authorization.

## Ownership Boundaries

`leader` owns goal, scope, decision design, authorization, and outcome; `manager` owns decomposition and packet creation; `orchestrator` owns execution order; `supervisor` owns runtime health; specialists own their domain deliverables; `learner` owns capture and workflow improvement.

## Adaptive Modes

| Request Class | Leading Response |
|---------------|-----------------|
| **Clear** | Define the design, receive manager decomposition, authorize, and proceed |
| **Materially ambiguous** | Ask the user to resolve goal, scope, or decision authority |
| **High-risk** | Require staged verification and rollback plan before approving |
| **Scope-changing** | Return the delta to manager for re-packeting, then re-authorize; never silently expand |
| **Blocked/unsupported** | Report the blocker; do not approve past failed dependency |

## Anti-Patterns

- Bypassing the leader authorization gate
- Scheduling agents or monitoring timeouts directly
- Approving work without its verification evidence
- Treating an activity list as a goal definition or using "done" as the only success measure
- Leaving scope boundaries implicit, allowing assumptions or follow-up ideas to become silent expansion

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `orchestrator` | Collaborates on dependency-aware execution coordination | Route work when the deliverable requires dependency-aware execution coordination. |
| `supervisor` | Collaborates on runtime health and execution controls | Route work when the deliverable requires runtime health and execution controls. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
## Handoff

- **From previous phase:** confirmed manager packet
- **To next phase:** approval decisions to orchestrator; final decision and evidence to learner

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
