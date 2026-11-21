---
name: automating
description: Use when automating, scripting, or building devtools. Triggers on automate, automation, script, batch, tool, devtool, dry-run, idempotent, repetitive work, bulk edit, scanner, generator.
---

# Automating Skill

Standards and conventions for detecting repetitive work, writing scripts and devtools, and choosing the right language and placement. Applies to any domain, stack, and methodology. Works with any AI agent. This skill is about **automation and tooling only** — it does not do business logic; it makes recurring work repeatable, safe, and reusable.

## When to Use

- The same operation runs on 3+ items (files, lines, records, renames, translations) or is scan/verify/batch-shaped
- A recurring pattern has no existing tool and needs one (Automation-First)
- You are about to build, extend, or place a devtool and need interface/placement/verification standards
- You must decide between a durable tool vs a one-off throwaway script

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`; this skill adds the tooling decision discipline below.

### Understand — find the repetition
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Count the items/operations that repeat. If fewer than ~3, do it directly — do not script.
- Classify the job: **durable** (will recur, reusable value) vs **throwaway** (one-off / few-off).
- Rule: one-off / few-off scripts go to `/tmp`, run, then discard — `/tmp/opencode` for sandboxed scratch. `tools/` is exclusively for durable, reusable devtools.
- Reuse before create: survey the project's `tools/` + skill `Automation Scripts` tables before writing a new tool.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — approach & language
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Choose language by **fit and performance**, not habit:
  - Python — JSON/text/batch transforms, stdlib scanners, most devtools
  - Shell — one-liners, piping, filesystem automation
  - node.js — JS/TS-adjacent automation
  - PHP — Laravel-adjacent automation
- Decide the contract up front: arguments, input/output format, exit codes, idempotency, and a `--dry-run` guard for destructive ops.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — write & place
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- Header docstring, argument parsing, deterministic output, error handling (exit non-zero on failure).
- Durable tools → `tools/` (this registry); throwaways → `/tmp`.
- Batch your own edits into few passes; never run per-edit what a tool can do once.
- Destructive operations require a dry-run or explicit confirmation; preserve data.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — prove it works
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Run and confirm output matches expected shape; check idempotency (safe to re-run).
- For renames/batch: `git status` + `git diff` confirm only intended changes.
- For a durable tool, sanity-check dry-run before applying to real data.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — hand off
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Report script path, invocation, and outcome. For throwaways, note they were discarded.
- For durable tools, add an integration note so future agents reuse them.

## Tool Placement

| Need | Location | Lifecycle |
|------|----------|-----------|
| Durable, reusable devtool | `tools/` (this registry) | Committed, reused |
| One-off / few-off script | `/tmp` (e.g. `/tmp/opencode`) | Run, then discard — never commit |
| Ad-hoc verification | Inline / shell | — |

## Adaptive Modes

| Request Class | Automating Response |
|---------------|---------------------|
| **Clear** | Build the script/tool with idempotency and dry-run guard |
| **Materially ambiguous** | Clarify durable vs throwaway, item count, and target outcome |
| **High-risk** | Add dry-run guard, confirmation step, and rollback plan |
| **Blocked/unsupported** | Report the missing capability; do not fabricate tool paths |

## Anti-Slop

1. Never build a durable tool for a one-off job — use `/tmp`, discard after.
2. Never redo by hand what a script already does (Automation-First).
3. Never run destructive batch ops without a dry-run or confirmation.
4. Never choose a language by habit — choose by fit and performance.
5. Never hallucinate a command or path — verify against the actual filesystem before running.

## Validation

- [ ] Placement correct (durable → `tools/`; throwaway → `/tmp` + discarded)
- [ ] Idempotent, safe to re-run; dry-run guard for destructive ops
- [ ] Language chosen by fit; reuse-before-create honored
- [ ] No committed throwaway scripts

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `decomposer` | Collaborates on SRP-based decomposition and refactoring | Route work when the deliverable requires SRP-based decomposition and refactoring. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |

- **Bundle contents:** `SKILL.md`; `rules/` (1 item); `tools/` (none); `resources/` (none); `refs/` (none).
