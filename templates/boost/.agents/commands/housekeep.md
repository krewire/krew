---
name: housekeep
description: Maintain hygiene of memory, context, rules, and knowledge-base stores. Use when cleanup, prune, organize, dedup, tidy, maintain memory, or housekeeping is mentioned.
---

# Housekeep Command — Knowledge-Store Hygiene

Generic hygiene command for any agent's knowledge stores. Queries and prunes RuntimeMemoryInterface,
RuntimeContextInterface, and tracked `rules/`. Internal issues and refs have no verified runtime
store; stop and report if either is requested. This command **maintains existing stores**; it does not fabricate new knowledge.

Follow `skills/housekeeping/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask which stores to clean and whether to include history (`--deep`).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Query RuntimeContextInterface and RuntimeMemoryInterface to identify durable versus ephemeral records and stale/duplicate indicators.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Select which stores to clean (memory vs context vs rules) based on age, volume, and repetition.         Prefer project scanners (`scan_broken_links.py`, `scan_doc_template.py`, `triad_check.py`) where they cover store integrity.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Prune/dedup/redact runtime entries when supported and write descriptive commit messages. Never delete intentional states recorded in RuntimeContextInterface. One fact = one runtime topic; merge duplicates rather than duplicating.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Index files match actual contents; link/dead-code scans over the store pass; no durable decision or PII requirement lost.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

What was pruned/deduped/redacted, what was promoted, and any one-off signals left in RuntimeMemoryInterface. Keep PII redaction notes intentionally vague.
## Anti-Patterns

- Deleting durable context as transient; over-promoting one-offs to `rules/` (rule-bloat); leaving `index.md` out of sync; leaving PII intact.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | No | all | Which store(s) to clean |
| `--deep` | flag | No | false | Also scan git history and runtime memory for stale & PII |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
## Handoff

A clean, deduplicated runtime store plus a short list of remaining single-sighting signals for RuntimeMemoryInterface.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
