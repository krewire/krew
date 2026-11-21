---
name: explore
description: Explore an unknown or large codebase into a navigable orientation. Use when find, locate, search, map, inventory, or orient is mentioned.
---

# Explore Command — Orient an Unknown Codebase

Generic exploration command for any domain. Turns "where is X / how does Y work" into a verified map: file inventory, module structure, key symbols, conventions, and integration points. Works with any AI agent. This command does **research only** — no code changes.

Follow `skills/exploring/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Verify every path and symbol against the actual filesystem — never hallucinate structure.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask what the exploration must answer (component location? data flow? convention? whole-map?) before proceeding.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Do not over-explore; target what is asked. Read `AGENTS.md`, RuntimeContextInterface, and tracked rules to reuse the existing map. Internal refs are unsupported; stop and report if required.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Pick grep/glob targets, key entry files, and call/import traces. Prefer project scanners (`tools/scan_*.py`) over manual greps where present (Automation-First). Decide depth: quick / medium / very thorough.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Run searches and reads; collect concrete `file:line` evidence for every claim. Produce a structured orientation: affected modules/layers, files touched, conventions to follow, patterns to re-use, risks/edge cases.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Check every cited path and symbol exists; re-read the question — did the exploration answer it or drift into a general survey?

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Report a concise map: what exists, where it is, how it connects, what to re-use, what to watch out for. Hand off to the execution phase with exact paths and line numbers.
## Anti-Patterns

- Quoting search results without verifying path/symbol in source; inventing structure; exhaustive survey when a targeted answer was asked.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | Yes | — | The question or area to explore |
| `{{depth}}` | enum | No | `medium` | `quick` / `medium` / `very-thorough` |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
## Handoff

Deliver exact file inventory, key symbols with `file:line`, conventions, risks, and a recommended entry path.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
