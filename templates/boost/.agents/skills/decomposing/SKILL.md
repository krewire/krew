---
name: decomposing
description: Use when refactoring code to satisfy the Single Responsibility Principle (SRP). Triggers on refactor, decompose, split, extract, simplify, SOLID, SRP, god class, god function, long method, cohesion, responsibility.
---

# Decomposing Skill

Standards and conventions for decomposing multi-responsibility code into single-purpose units guided by the **Single Responsibility Principle** from Clean Code / SOLID. Applies to any domain, stack, and methodology. Works with any AI agent. This skill is about **structural refactoring only** — it splits and extracts code, never changes externally observable behavior.

## When to Use

- A class/function/module owns two or more unrelated responsibilities (mixed persistence+logic, data+view, input+processing+output)
- You need to split a god class or long method while preserving behavior and the public contract
- You are about to refactor and need the SRP yardstick, split strategy, and verification discipline
- You must decide whether a split is safe given the coverage in place

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`; this skill adds the SRP decomposition discipline below.

### Understand — map responsibilities (SRP = one reason to change)
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Enumerate the target's distinct **reasons to change**. Each independent reason is a candidate unit; if two things change for different reasons, they belong apart.
- Judge by **cohesion** (what changes together), not by size: a long function over one concern is fine; a short function doing two things is not.
- Confirm a **verification net** (tests / typecheck / static analysis) covers the target before touching it. Without one, flag the refactor as high-risk and defer or add tests first.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — choose split sites
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Pick extract boundaries where responsibilities are cohesive; each extracted unit gets **one** responsibility and a name stating it.
- **Extract-and-delegate** over move-and-rewrite; prefer small, reviewable steps over one large diff.
- Preserve the public contract and external behavior; list any internal API/naming breaks explicitly.
- Keep dependency direction acyclic and downward — extracted units depend on stable, single-purpose collaborators.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — extract one responsibility per step
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- One responsibility per step, one compose-able commit per extract. Never rewrite behavior in the same pass.
- Name units by the single reason to change they own (e.g. `calculateDiscount`, `OrderPersistence`).
- Follow host-language conventions (naming, file layout, PSR-12 / SOLID equivalents). No magic numbers; clear names over cleverness.
- Never bundle a refactor with a feature change or unrelated fix in the same step.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — prove behavior preserved
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Run the verification net before and after; a behavior-preserving split must pass identically (or by equivalent tests).
- `git diff` each step to confirm only intended structural changes — no hidden behavior drift.
- Confirm extracted units have non-overlapping responsibilities (each has exactly one reason to change).

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — hand off
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Report the extract list (unit name + responsibility), any public API changes, and verification evidence.
- If a follow-up feature change is intended, state it as next work — never fold it into the refactor.

## Decomposition Checklist (before/after each extract)

- [ ] Verification net present and passing before the change
- [ ] Extracted unit has exactly one reason to change
- [ ] Public contract / external behavior unchanged
- [ ] Dependency direction acyclic, downward; no god dependency
- [ ] No feature change or unrelated fix bundled in
- [ ] `git diff` for the step shows only intended structural change

## Adaptive Modes

| Request Class | Decomposing Response |
|---------------|---------------------|
| **Clear** | Extract one responsibility per step with verification net |
| **Materially ambiguous** | Clarify the target file/class/function and SRP boundary before splitting |
| **High-risk** | Require verification net before touching; add characterization tests |
| **Blocked/unsupported** | Report the missing capability; do not refactor without a test baseline |

## Anti-Slop

1. Never split by line count or "feels long" — split by cohesion (independent reasons to change).
2. Never extract a unit that still has overlapping responsibilities (spreads the god class across files).
3. Never hide a behavior/feature change inside a refactor — refactors are behavior-preserving by definition.
4. Never create dependency cycles while "simplifying".
5. Never refactor safety-critical or uncovered code without adding a verification net first.
6. Never hallucinate test results or `git diff` outcomes — verify against the actual run.

## Validation

- [ ] Verification net passing before/after; no behavior drift
- [ ] Each extracted unit single-responsibility (one reason to change)
- [ ] Public API changes listed; no dependency cycles
- [ ] Refactor and feature work separated into distinct steps/commits

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `automator` | Collaborates on automation and tooling | Route work when the deliverable requires automation and tooling. |

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
