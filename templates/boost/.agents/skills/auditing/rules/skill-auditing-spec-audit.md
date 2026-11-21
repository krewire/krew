# Skill Rule — Bidirectional Specification–Implementation Audit

Project-agnostic audit procedure verifying that specs and their implementations are in sync — in
both directions. Complements `skills/auditing/SKILL.md` (security/quality) with a spec↔code
alignment pass. Fixes the lagging side: when the spec lags code, update the spec immediately; when
code lags the spec, file an issue (or fix small gaps in-run).

## Metadata

- **ID:** `rule-skill-auditing-spec-audit`
- **Scope:** `skill:auditing`
- **Severity:** `contextual` — applies when auditing spec/code alignment
- **Enforcement:** `manual` (review) + `automated` (where a spec-tests scanner exists)
- **Related:** `skills/auditing/SKILL.md`, `rules/skill-planning-feature-spec.md`,
  `rules/skill-planning-issues.md`, `rules/skill-testing-pest.md`

---

## Description

Detect three drift categories and decide which side is stale:

1. **Spec → Code:** spec promises something the implementation does not deliver (missing
   implementation).
2. **Code → Spec:** code does something the spec does not document (unspecified behavior).
3. **Both stale:** spec and code disagree on shared contracts (signatures, paths, names).

Coverage assertions (non-negotiable): every feature/API has a requirement ID in a spec (no
unregistered behavior — a **spec gap**); every testable requirement has an automated test (no
requirement without a test, no orphan test without a requirement).

## Trigger

- Post-feature or pre-release audit of spec↔code alignment.
- A spec change with no matching implementation, or code shipped without a spec update.
- Reviewing whether tests still map to their requirements.

## Action

1. **Scope** the audit: by spec, module, phase, area, or full. Size-triage before proceeding.
2. **Understand the drift categories** and which side is authoritative. **Trust nothing blindly** —
   verify via `git log --follow -- <file>` and `git blame`; if history is silent or contradictory,
   treat as a finding, do not silently pick a side.
3. **Audit both directions**: for each requirement check implementation presence; for each code
   unit check the spec documents it; for shared contracts compare signatures/paths/names.
4. **Run the tests** (targeted, not the whole suite) and **fill test gaps
   in-run** for spec'd components that lack tests.
5. **Triage findings** via a decision matrix:
   - Spec lags code → **fix the spec now** (amend then align).
   - Code lags spec / both stale on a shared contract → **file an issue** with severity + evidence
     (`rules/skill-planning-issues.md`); small gaps may be fixed in-run.
   - Agent docs & skills that reference the spec and are now stale → treat as Code → Spec drift
     (guide lagging); update them.
6. **Report**: per-finding verdict (spec-fixed / issue-filed / in-run-fixed), evidence
   (`file:line`), and remaining traceability gaps.

## Validation

- [ ] Every finding classified into one of the three drift categories
- [ ] Spec-lagging items fixed in-run (spec amended before code alignment)
- [ ] Code-lagging items filed as issues with severity + evidence (or fixed if small)
- [ ] No spec gap: every behavior/API traced to an FR/NFR/UC
- [ ] No orphan tests; every testable requirement has a test
- [ ] Decisions justified via git history — nothing assumed
- [ ] Report lists fixes, issues filed, and remaining gaps

## Example — Correct (spec lags code)

> Module `enrollment` implemented CSV export (FR-ENR-17 was in code + tests) but no spec recorded
> it. → Amend spec: add FR-ENR-17 with contract; align docs; leave code untouched.

## Example — Incorrect (silent accept)

> `sync-docs` updated the guide to match code while the spec still promises the old contract, and no
> requirement was recorded — behavior ships "unofficially", spec stays stale, nobody decides.

## Exception

Audit scope may be narrowed deliberately (single spec) — but within the chosen scope every finding
must be triaged, never dropped.