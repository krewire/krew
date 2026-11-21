---
name: automate
description: Automate repetitive or batch work into scripts and devtools. Use when automate, script, batch, tool, devtool, or repetitive multi-item work is mentioned.
---

# Automate Command — Turn Repetitive Work into Tools

Generic automation command for any domain. Detects repetitive, batch, or pattern-based work and scripts it (Automation-First): if the same operation runs on 3+ items or is scan/verify-shaped, write a tool instead of doing it by hand. Works with any AI agent and any project.

Follow `skills/automating/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask what repeats: how many items, how often, and the target outcome. A one-off / few-off job goes to `/tmp` (run then discard); only durable, reusable value goes to `tools/` (this registry).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Identify the items/operations that repeat (files, lines, records, renames, scans). If fewer than ~3 items, do it directly — do not script. Confirm whether it will recur (durable) or is one-off (throwaway).

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Choose by need and performance, not habit: Python (`tools/`) for JSON/text/batch + stdlib scanners; shell for one-liners; node.js for JS-adjacent; PHP for Laravel-adjacent. Survey existing devtools first — reuse before create. Define the contract: args, input/output format, exit codes.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

For durable value, write to `tools/` (this registry) with a header docstring, arg handling, deterministic output, and error handling. For one-offs, write to `/tmp`, run, and discard. Batch your edits into few passes. Include a dry-run mode for destructive operations.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Run, confirm output shape, and check idempotency (safe to re-run). For renames/batch: `git status` + `git diff` show only intended changes. Never apply destructive ops without a dry-run first.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Report the script path, invocation, and outcome. If throwaway, note it was discarded. If durable, add an integration note so future agents reuse it.
## Core Concepts

- **Automation-First:** 3+ repetitive items or a scan/verify shape → script it; never redo by hand what a tool does.
- **Durable vs throwaway:** `tools/` (this registry) for lasting value; `/tmp` for one-off/few-off — committing throwaways pollutes the toolchain.
- **Reuse before create:** check project `tools/` and existing skills before writing a new script.
- **Idempotency:** automation must be safe to re-run; destructive ops need a dry-run guard.
- **Fit over habit:** choose language by need and performance.

## Examples

```bash
/automate "Bulk-rename tracked documentation with a dry-run first"
/automate "Write a scanner that flags inline cache keys across app/Modules"
/automate "One-off: convert 5 hand-written tables in docs/ to markdown"  # -> /tmp, run, discard
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Re-running a script corrupts data | Not idempotent | Make it safe to re-run; add a dry-run/guard |
| `tools/` full of junk | Throwaway scripts committed | Keep one-offs in `/tmp`; durable tools only in `tools/` |
| Script is slow/brittle | Wrong language or no error handling | Choose by fit; add exit codes + error handling |
| Presenting renamed a wrong file | No dry-run before destructive op | Always dry-run + `git diff` before applying |

## Validation

- [ ] Script exists at the right place (`tools/` in this registry, or `/tmp` + discarded)
- [ ] Idempotent and safe to re-run; `git status`/`diff` shows only intended changes
- [ ] No committed throwaway; language chosen by fit; reuse-before-create honored

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `decomposer` | Collaborates on SRP-based decomposition and refactoring | Route work when the deliverable requires SRP-based decomposition and refactoring. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
## Handoff

Delivers the script path, idempotency guarantee, and run instructions. No file created if script lives in `/tmp`.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
