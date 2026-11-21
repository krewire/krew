---
name: building
description: Use when building features, fixing bugs, or refactoring. Triggers on implement, code, refactor, build, clean code, SOLID, commit, branch, code review, lint, SAST.
---

# Building Skill

Comprehensive, project-agnostic code development for any language, framework, or repo shape. Adapts branching, review, and quality gates to team and constraints. Works with any AI agent. This skill is for **building** — it does not elicit requirements or define architecture beyond what is needed to slice and implement safely.

## When to Use

- Building increments from a design, spec, or user story; fixing defects; or improving existing code with safety
- Need branching, commit hygiene, review discipline, or security gates
- Implementing with quality gates (lint, typecheck, test, SAST) and need to keep changes reviewable and revertible
- Refactoring legacy code or paying down tech debt without breaking behavior
- Preparing a change for testing and deployment handoff

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Favor small, verifiable increments over large synthetic dumps. Every line should be reviewable by a human in under an hour.

### Understand — Ground in Design and Code
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Review design pack and coding conventions from `refs/` when present.** Inspect codebase patterns, dependency graph, test strategy, and conventions without assuming a specific language. Read at least 3-5 existing files in the area you will change — note naming, error handling, test style, and config handling. Ground understanding in files you have read, not invented structure. Capture exact file paths and line numbers when referencing code.
2. **Identify contracts that must not break.** Interface schemas (OpenAPI, protobuf, schema), data migrations (forward and rollback), feature flags, and integration points. List at least one consumer per contract and the compatibility window. Check if the change is breaking, non-breaking, or expanding.
3. **Note NFRs and security requirements that affect code.** Performance budget, error handling, observability (logs, metrics, traces), and secrets handling. Check the rule shape in `skills/customizing/SKILL.md` when present for phase gates, otherwise use team standard. Surface hidden coupling: which tests will need to change if you change this code? Which data will be migrated?
4. **Clarify what is out of scope for this increment.** If the design left a spike, time-box it and produce a decision, not production code.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Slice for Review and Rollback
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Slice into small vertical increments (<1 day, TDD-ready) that can be reviewed and reverted safely.** Each increment should be demoable or at least testable in isolation — e.g., "add column nullable, backfill, then add not-null" is two slices, not one. Avoid generating large blocks of unreviewed code at once — that is where repetitive structures, copy-paste, and hallucinations hide. If a slice cannot be described in one sentence, split it.
2. **Choose branching fitting team:** trunk-based for CI/CD with short-lived branches and continuous integration, Gitflow for release trains with hotfixes, or simple feature branches for small teams. Define PR size (<300 LOC as guideline) to keep review human-scale. Choose merge strategy (squash, rebase, merge) per team. Do not mandate a single model — justify by release needs.
3. **Plan error handling, idempotency, retries, and observability** proportional to risk. For a payment flow, idempotency and retries with jitter are critical; for an internal report, simple error propagation may suffice. Define what "done" means for this slice: tests, docs, observability, and rollback plan.
4. **Select tooling available to project:** linter, formatter, type checker, SAST per language — do not mandate a specific tool. Use what the repo already uses; introduce a root-level `tools/*.py` helper only with a reason, and keep one-offs in `/tmp`.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Write Small, Tested, Secure Code
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Write code with intent, not hype:**
   - Clean Code, SOLID, DRY, KISS as applicable; small functions (ideally <20 lines), meaningful names, intention-revealing APIs, explicit dependencies. Replace vague comments like "delve into the realm" or "in today's fast-paced world" with precise intent: "retry with jitter for idempotency, see ADR-003, backoff 2^attempt * 100ms."
   - Prefer concrete, project-specific names over generic "Manager" and "Handler" repeated everywhere. One `UserManager` is unclear; `UserRegistrationService` with `register(email, password)` is specific and testable.
   - 12-Factor config: env-based, no secrets in repo, explicit dependencies, disposable processes. Check for hardcoded secrets with a pre-commit hook — do not commit and then fix. Document env vars in `.env.example`.
2. **Commit with hygiene:** Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `BREAKING CHANGE:`) and SemVer. Branch as `feat/`, `fix/`, `chore/` with descriptive names that include the story ID. Keep commits atomic and messages specific: not "fix bug" but "fix: handle null avatar in profile render, add test for guest user, fixes #123."
3. **Add tests alongside code.** Produce at least one test per slice that would have caught the bug or proven the feature. Document complex logic and update `refs/` when a decision changes. Provide unique, project-specific examples in docs, not zero-example placeholders. In comments, explain *why*, not *what* the code does: not "loop over users" but "batch by 100 to avoid OOM on 50k users, see load test #45."
4. **Handle config and migrations safely.** Env vars documented with safe defaults, migrations forward and backward tested on a copy of prod data, feature flag new behavior where rollback must be instant. Keep root-level tools only when helper automation adds clear value.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Make Quality Verifiable
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Local gates must pass from real runs:** lint, typecheck, test with coverage gate (e.g., ≥80% as team standard); SAST/SCA (semgrep, govulncheck, npm audit, trivy or equivalent per language). Do not invent coverage numbers — report the tool's actual output (e.g., "87.3% line, 72% branch from c8"). Fail the slice if any gate is red; do not adjust thresholds to pass.
2. **Self-review before PR:** check for repetitive sentence structures in comments, generic hedging, and meaningless words (cutting-edge, state-of-the-art, quietly in the background, secret sauce). Rewrite into specific, verifiable statements: not "state-of-the-art implementation" but "adds index on users.email, reduces lookup from 120ms to 8ms at 1k RPS (measured in staging)."
3. **PR review discipline:** at least one approval, CI green, no untracked TODOs, no secrets, no hallucinated data or links, no `// TODO` without an issue ID. Keep PR description specific: what changed, why, how to verify (curl, test command, or demo), and rollback plan ("revert commit abc123, flag `new_checkout` off"). Validate against acceptance criteria and interface contracts with a concrete test run, not a claim.
4. **Summarize AI's logic back to yourself** to catch nonsensical cause-effect before merging. If AI helped draft, use it for ideation — rewrite the bulk in your own precise language after a break. Take a break before final review — fresh eyes catch slop the first pass missed.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Close the Loop
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

1. **Update changelog (Keep a Changelog) and docs;** close linked issues with the commit that resolved them. Keep AI-generated prose to a small fraction — rewrite the bulk in your own precise language. Hand off to testing with a clear handoff: coverage report from the tool, migration notes (forward/backward commands), feature flag name, and how to demo the slice.
2. **Log tech debt and follow-ups for maintenance** in `refs/` with owner and date — not vague "tech debt." Example: "debt: extract payment retry into service, owner @jane, due 2026-09-10, cost if not done: 3 days of flake triage per month."
3. **Define what is done for this increment** and what is explicitly deferred to the next slice.

## Core Concepts

- **Vertical slice vs horizontal layer:** A vertical slice delivers a user-visible increment (API + logic + test + migration) that can be demoed; a horizontal layer delivers a technical layer that cannot be demoed — prefer vertical for flow.
- **Trunk-based vs Gitflow:** Trunk-based optimizes for CI/CD and small batches; Gitflow optimizes for release trains and hotfixes — choose by release needs and team size.
- **Conventional Commits vs freeform:** Conventional Commits enables automated changelogs and SemVer bump; freeform hides breaking changes — use the former when releases are frequent.
- **Refactor vs rewrite:** Refactor preserves behavior with tests as safety net; rewrite discards behavior and risks — prefer refactor unless tests are absent and behavior is unknown.

## Best Practices

1. **Small PRs, frequent merges.** A 100-line PR reviewed in 30 minutes beats a 1k-line PR reviewed in 3 days.
2. **Tests as safety net, not afterthought.** Write at least one test that would have caught the bug you fixed, before you fix it.
3. **Secrets never in repo.** Use env and secret manager; scan pre-commit; rotate if leaked; document rotation steps.
4. **Comment the why, not the what.** The code shows what; the comment shows why the alternative was rejected and which ADR decided it.

## Common Patterns

### Pattern: Atomic Commit
```bash
feat(auth): add login with RBAC check
- handle null token as 401, not 500 (test: guest login, see #123)
- add index on users.email for lookup (120ms → 8ms at 1k RPS, measured in staging)
```

### Pattern: Safe Migration (Two-Phase)
```sql
-- Phase 1 (deploy 1): add column nullable, backfill in batches, keep old code reading both
ALTER TABLE users ADD COLUMN avatar_url TEXT;
-- Phase 2 (deploy 2): add NOT NULL after backfill verified, drop old column
ALTER TABLE users ALTER COLUMN avatar_url SET NOT NULL;
-- Rollback: drop column if no data yet, or keep nullable
```

### Pattern: Idempotent Handler
```python
def handle(event):
    if is_processed(event.id):
        return ok(event.id)  # already handled, safe to retry
    process(event)
    mark_processed(event.id)
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| PR stuck in review for days | Too large, mixed refactor and feature, no Tests | Split into 2 PRs: one refactor with characterization tests (≤200 LOC), one feature with tests; keep each <300 LOC |
| Flaky tests after change | Shared mutable fixture, random data | Isolate fixtures per test, use deterministic factories, reset state via transaction rollback or truncate |
| Secret committed | No pre-commit hook, env not documented | Add detect-secrets hook, move to env, rotate secret, add to `.env.example` and README, revoke old token |
| Coverage drops, but CI still green | Gate not enforced, new code without tests | Add at least one test per slice; gate in CI (e.g., `coverage <80% → fail`), not locally only |
| Merge conflicts constantly | Long-lived branch, infrequent rebase | Rebase daily, keep branches <1 day, use trunk-based with feature flags |

## Standards

- **Clean Code / SOLID / DRY/KISS**: design and naming — adapt to language idioms, not slogans; measure via lint and review.
- **Conventional Commits 1.0 / SemVer 2.0**: history and versioning — enable automated releases and changelogs.
- **Trunk-Based / Gitflow**: branching — choose by release needs and team size; document choice in `refs/`.
- **OWASP Top 10 / ASVS**: secure coding — apply controls by risk, validate via SAST and manual review.
- **12-Factor App**: config and services — adapt to hosting model; env-based, disposable, explicit deps.

## Checklists

### Pre-Execution
- [ ] Design and conventions available or explicitly deferred — not hallucinated
- [ ] Branch from main, up to date, with descriptive name and story link (e.g., `feat/PROJ-123-login`)
- [ ] Toolchain available for project's language (lint, formatter, type checker, SAST) and pre-commit configured

### Post-Execution
- [ ] Formatted, linted, typechecked, tests pass — reports are from real runs, not invented numbers
- [ ] Coverage gate met, no flaky tests, no invented coverage, no hallucinated links
- [ ] No secrets committed, env vars documented with safe defaults and `.env.example` updated
- [ ] Changelog and docs updated with specific, verifiable changes — no generic intro/conclusion, no empty buzzwords
- [ ] PR reviewed and CI green, no repetitive phrasing or hypophoras, PR description has what/why/how to verify/rollback

## Adaptive Modes

| Request Class | Building Response |
|---------------|------------------|
| **Clear** | Implement the increment with tests and quality gates |
| **Materially ambiguous** | Clarify spec, branch, and quality gates before coding |
| **High-risk** | Add characterization tests, staged verification, and rollback plan |
| **Blocked/unsupported** | Report the missing capability; do not proceed without verification net |

## Anti-Patterns to Avoid

- Large, unreviewable PRs; mixing refactor with feature in one commit; cleverness over clarity; deep nesting and hidden side effects; `// TODO` without issue; mass-produced boilerplate that repeats the same three verbs ("automate, optimize, streamline") with no specific artifact.

## Example

- "Implement {feature} per {spec} with tests, migration (forward/backward), feature flag {flag}, and access control."
- "Refactor {module} for {goal} with safety via incremental PRs and characterization tests."
- "Fix {bug} with reproduction test, minimal code change, and regression coverage — PR <100 LOC."

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

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

- **Bundle contents:** `SKILL.md`; `rules/` (10 items); `tools/` (none); `resources/` (none); `refs/` (none).
