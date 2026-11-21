---
name: learn
description: Run a continuous-learning retrospective that mines the session and tunes agent workflow, rules, context, and memory. Use when learn, memory, retrospective, capture, consolidate, promote, workflow tuning, or self-improvement is mentioned.
---

# Learn Command — Continuous Self-Improvement Retrospective

Trigger an explicit **retrospective** that mines the current session (and recent history) for learning
signals and writes/tunes **workflow, rules, RuntimeContextInterface, and RuntimeMemoryInterface**. This is the manual form of the Learning
Loop (`rules/self-improvement.md`); the loop also runs automatically at session boundaries. Works with any
AI agent and any clone of this registry.

Follow `skills/learning/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask which stores to cover (workflow / rules / runtime context / runtime memory) and whether to include history (`--deep`).

## When to Invoke

- End of a complex or multi-step session.
- After a user correction or a failed verification gate.
- When you sense a repeated pattern but haven't promoted it yet, or a workflow asset looks outdated.
- On demand: `/learn` (optionally `/learn --deep` to also scan `git log`/`git diff`).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Review the session transcript and, with `--deep`, `git diff HEAD~1` + `git log -5 --oneline`. Classify each into: decision / correction / failure / pattern / constraint / gap, and route it to **workflow** (`agents/`/`commands/`/`skills/`), **rules**, **context**, or **memory**.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Decide: new RuntimeContextInterface topic, update existing, promote to `rules/`/`skills/`, or tune a workflow asset. Confirm no duplicate topic/rule/skill. Internal issue/ref stores are unsupported.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- **Workflow**: update `agents/{name}.md`, `commands/{name}.md`, or `skills/{name}/SKILL.md` minimally (preserve unrelated content); keep `AGENTS.md` rows accurate.
   - **Rules**: promote repeated signals → `rules/[scope]-[name].md`; reference from `AGENTS.md §Rules Index`.
   - **Context**: create/update a RuntimeContextInterface topic using the shape in `skills/customizing/SKILL.md`.
   - **Memory**: append/upsert/prune through RuntimeMemoryInterface when the operation is supported.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Index files match contents; no duplicate topics/rules; `tools/triad_check.py --root . --strict` clean after workflow edits; promoted rules referenced.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

One short paragraph: stores touched, files written/updated, rules promoted, and the one correction that must change future behavior. Keep narration discipline — surface decisions, not play-by-play.
## Output

- Tuned `agents/`, `commands/`, `skills/` (workflow) and/or updated RuntimeContextInterface, RuntimeMemoryInterface, or `rules/`.
- Updated runtime memory/context records when supported, plus any affected `AGENTS.md` registry rows.
- A one-line "Learning captured:" summary in the final report.

## Guardrails

- Fact-check every path/command before writing (no hallucinated learning).
- One fact = one file; update in place, never duplicate.
- Don't over-promote one-offs into mandatory rules; keep workflow edits surgical.
- Route to the right store: reusable generic asset → `customizer`; hygiene/pruning-only pass → `housekeeper`.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{store}}` | string | No | all | Stores to cover: `workflow` / `rules` / `runtime-context` / `runtime-memory` |
| `--deep` | flag | No | false | Also scan `git log`/`git diff` for historical signals |

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |
