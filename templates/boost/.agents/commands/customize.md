---
name: customize
description: Customize any agent, command, skill, or agent-store artifact. Use when creating, renaming, standardizing, scaffolding, or generating any agent need — triad, memory, context, rule, ref, skill, command, script, or config wiring.
---

# Customize Command — Agent-Asset Creation & Shaping

Generic customization command for this registry's asset library. Builds and shapes agent assets from the
canonical shapes in `skills/customizing/SKILL.md` — triads (agent ↔ command ↔ skill), memory/context/
rules/refs files, and config wiring — while keeping the `AGENTS.md` registry accurate and the
`triad_check.py` scan clean. Works with any AI agent. This command **creates and rewires assets**; it
does not prune existing stores (that is `/housekeep`).

Follow `skills/customizing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask what asset to create/rename (the triad `root`, or the store + artifact name).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Confirm the root word; check for collisions across `agents/`, `commands/`, `skills/`.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Derive `root-er` / `root` / `root-ing`; select the artifact shape from `skills/customizing/SKILL.md`; reuse an existing matching asset as a richer pattern.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Create the files from the shape (`agents/`, `commands/`, `skills/`, or the store file); register in the `AGENTS.md`; wire `opencode.json`/`kilo.json` `skills.paths` when outside the root.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

`tools/triad_check.py` returns zero findings; the `AGENTS.md` rows point to existing files; no stale references left behind.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

What was created/renamed, files touched, registration, and any opencode restart needed.
## Anti-Patterns

- Mismatched triad names (agent `fooer`, skill `bar`); leaving an agent without its command or skill; forgetting to register in `AGENTS.md`; inventing structure instead of loading it from `skills/customizing/SKILL.md`.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{target}}` | string | Yes | — | Asset to create/rename: triad root (e.g. `auditor`), or supported tracked artifact |
| `--rename` | flag | No | false | Rename/standardize an existing asset to the triad convention |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `automator` | Collaborates on automation and tooling | Route work when the deliverable requires automation and tooling. |
## Handoff

A complete, conflict-free, registered asset (triad or store file) plus any restarts or follow-ups required.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
