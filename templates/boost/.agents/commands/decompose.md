---
name: decompose
description: Decompose code by Single Responsibility Principle (SRP). Use when refactor, decompose, split, extract, simplify, SOLID, SRP, god class, god function, or long method is mentioned.
---

# Decompose Command — SRP-Driven Refactoring

Generic refactoring command for any domain or stack. Splits multi-responsibility code into single-purpose units following the **Single Responsibility Principle** (Clean Code / SOLID), keeping refactors behavior-preserving and bounded by a verification net. Works with any AI agent and any project.

Follow `skills/decomposing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for the file/class/function to decompose and confirm there is a test/typecheck baseline (SRP refactoring without a verification net is high-risk).

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Enumerate the distinct reasons-to-change the target currently owns. Splitting is by **cohesion** (what changes together), not length. Confirm a verification baseline (tests / typecheck / static analysis) exists before touching code.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Pick the extract boundary; each unit gets one responsibility and a name stating it. Preserve the public contract; list any API/naming breaks explicitly. Favor extract-and-delegate over move-and-rewrite.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Make small, behavior-preserving changes; no behavior rewrite in the same pass. Keep dependency direction acyclic and downward. Follow the host language's conventions; no magic numbers.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Run the baseline before and after; `git diff` each step for only intended structural changes. Confirm extracted units have non-overlapping responsibilities.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Report the extracted units (name + responsibility), any public API changes, and verification evidence. Call out any intended follow-up feature change separately.
## Core Concepts

- **SRP:** a unit should have exactly **one reason to change**. Split when a class/function owns multiple, unrelated reasons.
- **Cohesion over length:** a long function over one concern is fine; a short function doing two things is not. Judge by shared-vs-independent reasons to change.
- **Behavior-preserving:** a refactor must not change external behavior; feature changes are separate work, never folded in.
- **Verification net:** tests / typecheck / static analysis must cover the refactor, or the extract is deferred/flagged as risky.
- **Acyclic, downward dependencies:** extracted units depend on stable, single-purpose collaborators; never reintroduce a cycle.

## Examples

```bash
/decompose "app/Modules/Order/OrderService.php: 200-line God method doing validation, persistence, and email — has tests, split by responsibility"
/decompose "Order entity handles pricing + serialization + repo logic; extract pricing rule and serialization, keep API, run suite"
/decompose "checkout(): input parsing, discount calc, and DB commit all in one — extract discount calc with unit tests first"
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Suite fails after split | Behavior drifted or broken dependency | `git diff` step, realign extracted unit, re-extract single responsibility |
| New unit still multi-purpose | Split by size not cohesion | Re-map reasons-to-change; extract one responsibility at a time |
| Dependency cycle introduced | Units reference each other | Route through a stable single-purpose collaborator / invert dependency |
| Refactor tangled with a new feature | Two concerns in one pass | Separate the feature change; refactor stays behavior-preserving |

## Validation

- [ ] Verification net present and passing before/after (tests / typecheck / static analysis)
- [ ] Each extracted unit has exactly one responsibility (single reason to change)
- [ ] `git diff` across steps shows only intended structural changes — no behavior drift
- [ ] No dependency cycles; public API changes explicitly listed and documented
- [ ] Refactoring and feature work kept in separate steps/commits

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `automator` | Collaborates on automation and tooling | Route work when the deliverable requires automation and tooling. |
## Handoff

Delivers the refactored code with a summary of extracted units and their single responsibilities. `git diff` confirms only structural changes.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
