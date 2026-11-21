---
name: think
description: Structured reasoning for decision and conflict resolution. Use when conflict, decision, trade-off, choose between, alternatives, pros, cons, rationale, or deep thinking is mentioned. Read-only — never creates files.
---

# Think Command — Structured Decision Reasoning

Reasoning command for any domain or stack. Resolves a conflict or architectural decision through structured, logical analysis grounded in decision records, producing **several alternatives with pros/cons** and a reasoned recommendation — all communicated **directly to the user**. Follows ADR-style logic. **This command is read-only and never creates any file.**

Follow `skills/thinking/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask what needs to be decided or which conflict to resolve, and the criteria for a good outcome.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Read existing context (code, ADRs, RuntimeContextInterface, and tracked rules) to ground the conflict in facts. Internal refs are unsupported; stop and report if required. Restate what must be decided and the evaluation criteria. Mark assumptions separately from facts.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Produce **several** distinct candidate solutions (3+ when the space allows), each centered on a different core trade-off. No file is written at any point.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

For each alternative, list its **pros** and **cons** grounded in evidence or a labeled assumption. Compare alternatives against the criteria, then give a **recommendation** with the reasoning and the conditions that would favor another option. Deliver all of this **in the response** — do not create, edit, or write any file.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Confirm the recommendation follows from the listed pros/cons and criteria; no omitted con would flip it. Surface any assumption presented as fact.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Give a tight summary: the conflict, alternatives with pros/cons, the recommendation, and the open assumptions for the user to confirm.
## Core Concepts

- **Alternatives, not binaries:** produce multiple distinct options with different trade-offs, not two forced sides.
- **Pros/cons grounded:** every point is tied to evidence or an explicitly labeled assumption — never invented.
- **Criteria-led:** alternatives are compared on a shared, stated set of criteria, not each in isolation.
- **Reasoning visible:** the user sees the step-by-step logic, not just a verdict.
- **No files:** this command reasons in the conversation only. Persisting a decision (ADR, rule, context) is a separate, explicit step for `designer`/`builder`/user — never done here.

## Examples

```bash
/think "Split OrderService into two services vs keep monolith — decide, with alternatives and pros/cons, considering cost, ops burden, and reversibility"
/think "In-memory cache vs Redis for session store — reason through latency, consistency, and operational cost, recommend one"
/think "Should we adopt primary-replica now or defer? Walk through 3 alternatives with trade-offs and what would flip each"
```

## Validation

- [ ] Several distinct alternatives produced (not just two), each with a different core trade-off
- [ ] Per-alternative pros/cons grounded in evidence or labeled assumptions — no invented data
- [ ] Recommendation stated with the reasoning and the conditions favoring each other option
- [ ] **No file created, edited, or written** during the command (verify `git status` unchanged if inside a repo)
- [ ] Open assumptions surfaced for the user to confirm

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `designer` | Collaborates on architecture and API design | Route work when the deliverable requires architecture and API design. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
