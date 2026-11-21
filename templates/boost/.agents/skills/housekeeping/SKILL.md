---
name: housekeeping
description: Use when maintaining hygiene of memory, context, rules, or knowledge stores. Triggers on cleanup, prune, organize, dedup, tidy, housekeep, maintain memory, hygiene, index sync.
---

# Housekeeping Skill

Comprehensive, project-agnostic hygiene for agent knowledge stores. Queries and prunes
RuntimeMemoryInterface, RuntimeContextInterface, and tracked `rules/`. No native runtime store
for internal `issues` or `refs` is verified; stop and report if either is requested. This skill
**maintains existing stores**; it does not fabricate new knowledge.

## When to Use

- Deduplicating repeated entries across RuntimeMemoryInterface, RuntimeContextInterface, and `rules/`
- Pruning ephemeral or expired runtime memory entries while preserving durable facts
- Redacting PII / secrets from store files (apply masking conventions)
- Keeping `index.md` files in sync with actual store contents
- Promoting signal seen ≥2 times from `memory` → `context` → `rules` (per Self-Improvement Loop)
- Detecting doc-link and reference drift across the store

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Query RuntimeContextInterface and RuntimeMemoryInterface to identify what is durable versus ephemeral.
- Identify which stores are in scope and what signals indicate stale/duplicate content.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Choose the hygiene targets (memory vs context vs rules) based on age, volume, and repetition.
- Prefer project scanners (`scan_broken_links.py`, `scan_dead_code.py`) where they cover store integrity.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- Prune/dedup/redact entries in place; update `index.md` rows; write descriptive commit messages.
- Never delete intentional states recorded in RuntimeContextInterface (e.g. deprecated, dummy-guard) without flagging them.
- One fact = one file; merge duplicates rather than duplicating.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Confirm index files match actual contents; run link-integrity and dead-code scans over the store.
- Ensure no durable decision or PII handling requirement was lost during pruning.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Report what was pruned/deduped/redacted, what was promoted, and any one-off signals left in RuntimeMemoryInterface. Keep PII redaction notes intentionally vague.

## Adaptive Modes

| Request Class | Housekeeping Response |
|---------------|----------------------|
| **Clear** | Prune, dedup, and redact the specified stores |
| **Materially ambiguous** | Clarify which stores to clean and whether to include history |
| **High-risk** | Preserve intentional states; flag durable decisions before pruning |
| **Blocked/unsupported** | Report the missing capability; do not fabricate store operations |

## Anti-Patterns

- Deleting durable context as transient; over-promoting one-offs to `rules/` (rule-bloat); leaving `index.md` out of sync; leaving PII intact.

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

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
