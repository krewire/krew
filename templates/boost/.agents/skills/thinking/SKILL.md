---
name: thinking
description: Use when resolving decisions or conflicts through structured reasoning. Triggers on conflict, decision, trade-off, alternatives, pros/cons, rationale, ADR, deep thinking, choose between. Read-only — never writes files.
---

# Thinking Skill

Standards and conventions for resolving architectural decisions and conflicts through structured, logical reasoning that follows ADR logic and produces **several alternatives with pros/cons** and a reasoned recommendation. Applies to any domain, stack, and methodology. Works with any AI agent. This skill is **reasoning-only and ephemeral**: it must **never create, edit, or write any file** — all output is delivered in the conversation.

## When to Use

- Two or more positions/options conflict and a decision is required
- An architectural choice needs structured alternatives, trade-offs, and a recommendation
- A reviewer/user asks "why this way?" or "which should we choose and why?"
- Deep, multi-step reasoning over options is needed before any implementation or documentation step

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`, adapted so that **Build** outputs reasoning to the user rather than to a file.

### Understand — frame the decision
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Read the existing context (code, ADRs, `context/`, `refs/`) to ground the conflict in first-party facts. Never invent stakeholders, numbers, or constraints.
- Restate the decision precisely and the **criteria** for a good outcome (cost, risk, fit, maintenance, reversibility, …).
- Separate **facts** (grounded) from **assumptions** (explicitly labeled, with an owner to confirm).
- Apply the SSOT [`rules/agent-conflict-resolution.md`](../../rules/agent-conflict-resolution.md): record positions, interests, constraints, alternatives, evidence/assumptions, residual risks, reversibility, and a 0–5 Win-Win Solution Score with rationale.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — generate alternatives
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Produce **several** distinct candidates (3+ when the space allows), each centered on a different core trade-off — not minor variations of one idea.
- Define the evaluation criteria so alternatives are compared on a level field.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — reason, never write
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- For each alternative, list concise **pros** and **cons**, each tied to evidence or a labeled assumption.
- Compare alternatives against the criteria directly; reason through the trade-offs.
- State a **recommendation** with the reasoning that leads to it, plus the **conditions** under which a different alternative would win. Scores 0–3 require the rule's escalation path; no extra user confirmation is needed for scores 4–5 only for internal, reversible, low-risk decisions inside the leader-authorized design, and they never bypass approval or safety controls.
- **Hard rule: deliver everything in the response. Do NOT create, edit, or write any file** — no scratch, no `refs/`, no ADR file, no `/tmp` artifact. This skill is ephemeral by design.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — check the logic
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Does the recommendation follow from the listed pros/cons and criteria? Would an omitted con flip it?
- Any assumption presented as a fact? Surface and label it.
- Confirm no file was created during the reasoning.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — decide-ready answer
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Present the conflict record, Win-Win Solution Score and rationale, alternatives with pros/cons, recommendation, and the approvals or user confirmation required by [`rules/agent-conflict-resolution.md`](../../rules/agent-conflict-resolution.md) — as an in-chat deliverable.

## Adaptive Modes

| Request Class | Thinking Response |
|---------------|------------------|
| **Clear** | Produce alternatives, pros/cons, and a recommendation with Win-Win Score |
| **Materially ambiguous** | Restate the decision with explicit criteria before reasoning |
| **Conflicting** | Apply conflict resolution rule; surface positions, interests, and a score |
| **High-risk** | Require structural limits and leader checkpoint in the recommendation |

## Anti-Slop

1. **Never** create, edit, or write any file — this skill is reasoning-only; persisting a decision is a separate explicit step.
2. Never give only two forced options when the space supports more.
3. Never invent pros/cons, benchmarks, or constraints — ground them or label them as assumptions.
4. Never recommend without stating the decision criteria and the conditions favoring each alternative.
5. Never hide the reasoning — the user must see the step-by-step logic, not just a verdict.
6. Never present an assumption as a verified fact.

## Validation

- [ ] Several distinct alternatives produced, each with a different core trade-off
- [ ] Pros/cons per alternative grounded or labeled as assumptions — no invented data
- [ ] Recommendation stated with reasoning and the conditions favoring each other option
- [ ] **No file created, edited, or written** (verify `git status` unchanged if inside a repo)
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

- **Bundle contents:** `SKILL.md`; `rules/` (1 item); `tools/` (none); `resources/` (none); `refs/` (none).
