---
name: write
description: Write and maintain project documentation and content/copy — all markdown files, project documentation, code docblocks, and copy. Use when document, write docs, update docs, sync docs, PHPDoc, doc drift, UI copy, copywriting, changelog, or docblock is mentioned.
---

# Write Command — Documentation, Content & Copy

Write, update, and keep all readable artifacts in sync with code and specs: markdown docs, project
documentation, code docblocks, and content/copy — following the documentation-first (SSOT) principle
and the two-tier model. Works with any AI agent and any project layout.

Follow `skills/writing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask which artifact to write/update and what changed (code, spec, or convention).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Identify affected docs (reference/conceptual, agent docs & skills, devtool catalogs), content/copy (UI, marketing, email, changelog), or docblocks; locate the governing spec; when code and docs disagree, check `git log --follow -- <file>` before deciding which is stale.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Choose conceptual vs reference tier per file; map every changed file to the doc that must change; decide additions vs cross-referenced updates (never duplicate a fact); confirm audience/tone for copy.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Apply the shapes in `skills/writing/SKILL.md` — metadata, structure, PHPDoc conventions, two-tier separation; keep edits surgical; ensure every `[label](path)` resolves; match existing voice for copy.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

`tools/scan_broken_links.py` for markdown-only changes; `pint` only if PHP touched; confirm links resolve, metadata current, no duplicated/stale facts, history via `git log` (no inline "Last updated").

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Files written/updated, tier decisions, drift resolved, gaps raised (one short paragraph).
## Anti-Patterns

- Full-file rewrites instead of surgical diffs; duplicating facts across files instead of cross-referencing; deciding code is stale without checking git history; inline "Last updated" banners; inventing data to fill a doc; inventing a new voice for copy where an established one exists.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
## Handoff

Docs and copy in sync with code and specs, links verified, quality gate passed, drift findings resolved or reported.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
