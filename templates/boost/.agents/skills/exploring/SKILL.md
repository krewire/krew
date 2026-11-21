---
name: exploring
description: Use when exploring, locating, or inventorying an unknown or large codebase. Triggers on explore, find, locate, search, map, orient, inventory, codebase, grep, discover.
---

# Exploring Skill

Comprehensive, project-agnostic codebase exploration and orientation. Turns an unknown or large codebase into a navigable map: file inventory, module structure, key symbols, conventions, and integration points. Works with any AI agent and any project. This skill is for **research only** — it produces findings and a map, not code changes.

## When to Use

- Orienting in a new or unfamiliar codebase before making any change
- Locating a component, class, function, route, or data flow by name or behavior
- Mapping module boundaries, layered structure, or dependency graph
- Discovering conventions, patterns, and integration seams to follow
- Inventorying dead code, drift between docs and implementation, or risky areas
- Producing a concise orientation handoff for the execution phase

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`; for exploration this maps to orient → plan searches → gather evidence → verify findings → report.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Confirm what question must be answered (component location? data flow? convention?). Do not over-explore.
- Read the project's `AGENTS.md`, `context/`, and `refs/` orientation docs if present — reuse the map instead of re-deriving it.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Choose the search strategy: grep/glob across probable locations, read key entry files, trace imports/calls.
- Prefer project scanners (`tools/scan_*.py`) over manual greps where present (Automation-First).
- Decide depth: quick (basic search), medium (moderate), or very thorough (comprehensive across naming variants).

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- Run searches and reads; collect concrete `file:line` evidence for every claim. Never hallucinate paths or contents — verify each against the actual filesystem.
- Produce a structured orientation: affected modules/layers, files touched, conventions to follow, patterns to re-use, risks/edge cases.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Confirm every cited path and symbol exists; run lint/build/scanners for the explored area only where feasible.
- Re-read the question — did the exploration actually answer it, or drift into general survey?

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Report a concise map: what exists, where it is, how it connects, what to re-use, what to watch out for.
- Hand off to the execution phase with exact file paths and line numbers.

## Adaptive Modes

| Request Class | Exploring Response |
|---------------|-------------------|
| **Clear** | Execute the search strategy and produce a structured orientation |
| **Materially ambiguous** | Clarify the exploration question before searching; do not over-explore |
| **High-risk** | Prefer read-only exploration; flag any proposed change as requiring separate approval |
| **Blocked/unsupported** | Report the missing capability; do not fabricate paths or contents |

## Anti-Patterns

- Quoting search results without verifying path/symbol in source; inventing structure; gold-plating an exhaustive survey when a targeted answer was asked.

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

Exact file inventory, key symbols with `file:line`, conventions, risks, and a recommended entry path.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
