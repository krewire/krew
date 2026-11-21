# Skill Rule — Structured Issue Writing

Project-agnostic GitHub (or compatible) issue template so every issue is immediately actionable by
a human or an AI agent without extra context. Complements `skills/planning/SKILL.md` — issues are
how planned work gets tracked and delegated.

## Metadata

- **ID:** `rule-skill-planning-issues`
- **Scope:** `skill:planning`
- **Severity:** `contextual` — applies when writing any tracked-work issue (bug/feature/security/
  refactor/tech debt/docs)
- **Enforcement:** `manual` (review)
- **Related:** `skills/planning/SKILL.md`, `rules/skill-auditing-spec-audit.md`, `rules/global-standards.md`

---

## Description

Every issue follows one structure: title convention, description of the problem (not the solution),
scope & impact, reproduction for bugs, acceptance criteria, recommended approach (2+ when there are
trade-offs), design decisions, related links. A defect noticed is a defect tracked — pre-existing
issues found during other work must be filed, not silently tolerated.

## Trigger

- Filing a bug, feature, security finding, refactor, performance, or tech-debt issue.
- An audit (spec/security/quality) produces a finding that needs tracking.
- A defect is noticed during work that cannot be safely fixed in-session.

## Action

**Check for duplicates first** — search open issues for the same concern; link the existing one
instead of filing a twin.

Follow the unified template (remove irrelevant sections):

```markdown
## Title
{type}: {scope/area} — {short description}

## Description
{Problem, concretely, not the solution. Bugs: what happened vs what should have.
 Features: user story or problem statement.}

## Scope & Impact
| Field | Value |
|---|---|
| Module / Area | {area} |
| Files affected | `{file}`, `{file}` |
| Dependencies | {prerequisite task} |
| Severity | `critical` / `high` / `medium` / `low` |
| Priority | `P0` / `P1` / `P2` / `P3` |
**Impact:** {who/what/how many affected; why it matters}

## Reproduction (bug only)
### Steps to Reproduce: 1…2…3…
### Expected Behavior / Actual Behavior / Environment

## Acceptance Criteria
- [ ] Criterion 1 (checklist that must be satisfied to close)

## Recommended Approach
### Approach A: {name} (Recommended)   [pros + cons]
### Approach B: {alternative}          [pros + cons]

## Design Decisions
| Decision | Chosen | Rationale |

## Related
- {link to related issue / ADR / docs}
```

Key rules:

- **One issue = one concern.** Never combine a bug and a feature, or two unrelated bugs. If a second
  sentence is needed to describe a different concern, it is a second issue.
- Title `type: scope/area — desc` (`bug`, `feature`, `security`, `refactor`, `perf`, `tech-debt`,
  `docs`).
- **Scope must be specific** — name the module, submodule, and affected files; quantify the boundary.
  "Fix the module" is unknowable scope; a reader must be able to list the files they will touch.
- **Impact must be measurable** — numbers and concrete consequences, not feelings: "3s instead of
  200ms for 1000 students", not "slow". State population, current measurement, and expected target.
- Describe the **problem**, not the proposed solution — recommendation goes in its own section.
- 2+ approaches when there are nontrivial trade-offs; recommend one. Every material choice gets a
  Design Decisions row with Chosen + Rationale naming what was considered and rejected.
- **No sensitive information, ever** — no credentials, API keys, tokens, PII, or plaintext passwords.
  Relative project paths only (never `/home/name/...` or absolute machine paths).
- **Labels from the repo registry**: exactly 1 type + 1 severity + 1 priority (matching the body's
  Scope & Impact); never ad-hoc label names.
- **Trace to a spec** — a bug cites the `FR-*` / `NFR-*` / `UC-*` ID it contradicts; a feature/refactor
  names its governing spec or adds "Spec required — write it before implementation" as a blocker.
- **Duplicate check first** — search open issues; link the existing one instead of filing a twin.
- **Pre-existing defects are filed immediately** — a defect noticed during other work that cannot be
  safely fixed in-session becomes an issue in the same run, with the evidence gathered.
- L-size sessions: note the session-split plan in the body; do not attempt L-size in one pass.
- Severity + priority labeled on every issue (1 type + 1 severity + 1 priority minimum).

## Destructive Patterns — Never Write These

Recognized failure shapes; self-correct before submission:

- A bug + a feature combined in one issue
- A generic title ("Fix bugs" — which bugs?)
- A description that is just "please fix this"
- A technical issue with no recommended approach
- An issue containing credentials, API keys, or personal data
- A scope so wide it cannot be planned ("Make the app better")
- An issue without a referenced requirement or scope boundary (requirements-first applies to issues too)

## Validation

- [ ] Title follows `type: scope/area — desc`
- [ ] Problem described concretely, not the solution
- [ ] Files/dependencies named; severity + priority set; severity/priority match the GitHub labels
- [ ] Impact statements contain a number (population, latency, count) and a target
- [ ] Bugs include reproduction (or explicit "untestable/not reproduced")
- [ ] Acceptance criteria are a checklist
- [ ] 2+ approaches (or explicit single-obvious-approach note)
- [ ] No sensitive information; relative paths only
- [ ] Requirement ID cited, or "spec required" noted for features/refactors
- [ ] No duplicate of an existing open issue; related links resolve
- [ ] Matches none of the destructive patterns; a stranger to the module can start without questions

## Example — Correct

> `bug: enrollment/registration — duplicate entry on concurrent submit`
> When two students submit the registration form simultaneously, both requests pass the quota check
> before either transaction commits, resulting in over-quota placement (1 slot filled by 2).
> Module: Enrollment · Severity: high · Priority: P1
> **Impact:** affects all students during registration week; every over-quota placement needs manual
> cleanup.

## Example — Incorrect

> _Fix registration._ (No type, no scope, no repro, no acceptance criteria, no impact — a second
> agent cannot act on it.)

## Exception

Tiny, obvious chores (typo, one-line doc fix) may skip the template — but any behavior, security, or
cross-cutting change gets the full structure.