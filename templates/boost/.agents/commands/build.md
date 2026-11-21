---
name: build
description: Implement features with quality gates — Clean Code/SOLID/Conventional Commits. Use when implement, code, refactor, build, or feature is mentioned.
---

# Implement Command — Spec to Tested Code

Generic implementation command for any language and framework. Turns a design/spec into tested, reviewed code with security and config discipline. Adapts branching, commit, and quality gates to team. Works with any AI agent.

Follow `skills/building/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Favor small, verifiable increments over large synthetic dumps.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for spec or feature description, target branch, and quality gates before proceeding.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Review design pack and coding conventions from tracked rules or RuntimeContextInterface when present. An internal ref store is unsupported; stop and report if one is required. Inspect codebase patterns, dependency graph, test strategy, and conventions without assuming a specific language. Read at least 3-5 existing files in the area you will change — note naming, error handling, and test style. Ground understanding in files you have read, not invented structure. Capture exact file paths and line numbers when referencing code. List at least one consumer per contract and the compatibility window. Note NFRs and security requirements.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Slice into small vertical increments (<1 day, TDD-ready) that can be reviewed and reverted safely. Each increment should be demoable or testable in isolation. Avoid generating large blocks of unreviewed code at once — that is where repetitive structures and hallucinations hide. If a slice cannot be described in one sentence, split it. Choose branching fitting team: trunk-based for CI/CD, Gitflow for release trains, or simple feature branches. Define PR size (<300 LOC guideline) and merge strategy. Plan error handling, idempotency, retries, and observability proportional to risk. Select tooling available to project: linter, formatter, type checker, SAST per language — do not mandate a specific tool.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update a linked GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Write with intent, not hype: Clean Code, SOLID, DRY, KISS; small functions (<20 lines), meaningful names, explicit dependencies. Replace vague comments like "delve into the realm" with precise intent: "retry with jitter for idempotency, see ADR-003." Prefer concrete names over generic "Manager" and "Handler". Handle config via env (12-Factor), no secrets in repo, explicit dependencies. Commit with Conventional Commits and SemVer; branch as `feat/`, `fix/`, `chore/` with story link. Add tests alongside code. Create a root-level `tools/*.py` helper only if automation adds clear value; keep one-offs in `/tmp`.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Local gates must pass from real runs: lint, typecheck, test with coverage gate (e.g., ≥80%); SAST/SCA per language. Do not invent coverage numbers — report tool output. Self-review for repetitive structures, generic hedging, and meaningless words (cutting-edge, quietly in the background). Rewrite into specific, verifiable statements. PR review: at least one approval, CI green, no untracked TODOs, no secrets, no hallucinated links. Verify every URL resolves and every stat has a source. Validate against acceptance criteria and interface contracts with a concrete test run.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Update changelog (Keep a Changelog) and docs; close linked issues with the commit that resolved them. Keep AI-generated prose to a small fraction — rewrite the bulk after a break. Hand off to testing with coverage report from the tool, migration notes (forward/backward), feature flag name, and how to demo the slice. Log tech debt with owner and date.
## Core Concepts

- **Vertical slice vs horizontal layer:** Vertical slice delivers user-visible increment; horizontal layer delivers technical layer that cannot be demoed — prefer vertical.
- **Trunk-based vs Gitflow:** Trunk-based for CI/CD and small batches; Gitflow for release trains — choose by release needs.
- **Refactor vs rewrite:** Refactor preserves behavior with tests; rewrite discards behavior — prefer refactor.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{spec}}` | string | Yes | — | RuntimeContextInterface topic, tracked rule, or feature description |
| `{{branch}}` | string | No | `feat/auto` | Branch name (`feat/`, `fix/`, `chore/`) with story link |
| `--test` | flag | No | true | Generate tests alongside code |
| `--lint` | flag | No | true | Run lint/typecheck before commit |
| `--dry-run` | flag | No | false | Scaffold without committing — show files |

## Examples

```bash
/implement "Per {spec} implement {feature} with tests, migration, and access control" "feat/feature"
/implement "Add {capability} to {component}" "feat/capability" --test
/implement --dry-run "RuntimeContextInterface:payment-api" "feat/payment"
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| PR stuck in review for days | Too large, mixed refactor and feature | Split into 2 PRs: one refactor with tests (≤200 LOC), one feature; keep each <300 LOC |
| Flaky tests after change | Shared mutable fixture | Isolate per test, use deterministic factories, reset via transaction |
| Secret committed | No pre-commit hook | Add detect-secrets hook, move to env, rotate, add to `.env.example` |

## Standards

Clean Code, SOLID, DRY/KISS, Conventional Commits 1.0, SemVer 2.0, Trunk-Based/Gitflow, OWASP Top 10, 12-Factor — adapt to language idioms.

## Validation

- [ ] Lint, typecheck, tests pass with gate (tool output, not invented), Conventional Commits, no secrets, SAST clean, no repetitive phrasing or hallucinated links, PR has what/why/how to verify/rollback

## Core Concepts — Phase-Specific

- **Decision vs artifact:** This phase produces decisions and contracts, not just documents — each artifact must have an owner and a verification.
- **Incremental vs big-bang:** Prefer small, reversible increments that can be reviewed in one sitting over large, unreviewable dumps.
- **Quantified vs vague:** Every NFR/SLO/gate must have a number, a measurement point, and an owner — otherwise it is wishful thinking.

## Best Practices — Anti-Slop

1. One artifact, one interpretation — if two people can read it differently, split or rewrite.
2. Replace hedging ("in today's fast-paced world") with quantified context and cited source.
3. Provide a unique, project-specific example per item — no zero-example lists.
4. State a point of view when choosing — do not hedge to the center when a trade-off is needed.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `designer` | Collaborates on architecture and API design | Route work when the deliverable requires architecture and API design. |
| `decomposer` | Collaborates on SRP-based decomposition and refactoring | Route work when the deliverable requires SRP-based decomposition and refactoring. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
## Handoff

- **From previous phase:** Requires versioned input with traceability and quantified NFRs/SLOs — do not start without it.
- **To next phase:** Delivers versioned artifacts with explicit non-goals, open spikes, owners, and due dates. No TBD without an owner.

## Metrics

- Track phase-specific metrics with tool output, not invented numbers: scope stability, estimation accuracy, design review throughput, build lead time, test flake rate, deployment frequency, MTTR.


## Anti-Patterns — Phase-Specific

- Generic hedging, repetitive "It's not X; it's Y" rhetoric, and formulaic three-item lists that obscure the trade-off.
- Mass-produced boilerplate with no unique example per item; zero-example artifacts.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
