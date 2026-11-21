# Skill Rule — Generic Testing Discipline

Project-agnostic testing discipline for any language or framework: what to test, where to fake,
how to diagnose failures, and how to run verification cheaply. Complements the framework-specific
`rules/skill-testing-pest.md` and the standards in `rules/global-standards.md` §Testing; does not
replace either.

## Metadata

- **ID:** `rule-skill-testing`
- **Scope:** `global` — governs all test-writing, reviewing, and diagnosis in any project
- **Severity:** `contextual` — binds when writing, reviewing, or diagnosing tests
- **Enforcement:** `manual` (review) + `automated` (grep/linter checks where available)
- **Related:** `rules/global-standards.md` §Testing, `rules/skill-testing-pest.md`, `skills/testing/SKILL.md`

---

## Test Writing

Tests verify the requirement — nothing more — one behavior at a time, in the project's established
shape. Never invent a new test style per file, and never pad beyond the named scenarios.

| Rule | Description | Validation |
|------|-------------|------------|
| **TW-01** Behavior-focused | Test what the code *does* (result + side effects), not internal calls or implementation details. | Review; a refactor that preserves behavior does not break the test |
| **TW-02** One behavior per test | AAA structure (Arrange → Act → Assert) with a single clear intent per test; the name reveals intent. | Review |
| **TW-03** Follow existing patterns | Read a sibling test of the same type before writing; the new file is structurally indistinguishable from its siblings. | Diff new test vs sibling |
| **TW-04** No framework testing | Assert on your logic's outcomes, not framework mechanics the vendor already tests upstream. | Review |
| **TW-05** Named scenarios only | Cover the happy path plus explicitly named rejections/alternatives; skip edge-case matrices, markup/CSS detail, and internals the requirement never names. | Scenario checklist against the requirement |
| **TW-06** Extract repetitive setup | Shared helpers for repeated fixture chains; never paste the same chain into N tests. | Grep for duplicated setup |

- A test that maps to no current requirement or acceptance criterion is **orphan noise**: flag it
  for removal, do not "fix" it to pass.
- Coverage means **requirements covered**, not lines of code. Line-coverage floors are diagnostic
  only (see Quality & Diagnosis).

## Mocking & Boundaries

Fake at the edge, never in the core. A fake exists to isolate the code under test from side effects
that are slow, external, or non-deterministic — not to avoid running the code being tested.

| Rule | Description | Validation |
|------|-------------|------------|
| **MB-01** Mock only externals | Real objects for own code; fake only the boundaries: external HTTP, file system, queue, notifications, events, cache, auth. | Review |
| **MB-02** Never mock the data layer | Real persistence (factories + real database with transactional refresh per test) — a mocked model/ORM proves nothing: green tests pass while the real data layer is broken. | Grep for mock-of-model |
| **MB-03** Prefer built-in fakes | The framework's native fake/stub drivers over hand-rolled mock-object stubs — scoped to the test, leak-free, less code. | Review; grep stub calls |
| **MB-04** Fake narrowly, after arranging | Fake only what the test needs, positioned *after* fixture setup so real setup events still fire and listeners register. | Review |
| **MB-05** No real side effects | A test that hits real HTTP, writes to disk, or enqueues real jobs is slow, flaky, and environment-dependent. | Review |

Boundary reference:

| Boundary | Approach |
|----------|----------|
| External HTTP | Fake driver, no real network |
| File system | Fake storage driver |
| Queue / async jobs | Fake bus/driver |
| Notifications / mail | Fake delivery |
| Events | Fake only the specific event under test |
| Cache | In-memory fake store |
| Auth / identity | Act as an authenticated actor, never stub the guard |
| Persistence | Real database: factories + transaction-per-test |

Concrete fake names and helpers are framework-specific — see `rules/skill-testing-pest.md`.

## Quality & Diagnosis

A failing test is information. Read the output before re-running: parse the assertion mismatch and
the file:line, classify the failure, then choose the smallest remedy.

| Rule | Description | Validation |
|------|-------------|------------|
| **QD-01** Read before re-running | Diagnose from output: assertion mismatch → test/source logic; missing class → autoload/namespace; database error → migration/connection; timeout → un-drained async work. | Fix-workflow review |
| **QD-02** Classify the failure | Test defect (fix the test), source defect (fix the code), or environment/pre-existing (flag it — do not silently fix). | Root-cause note on the fix |
| **QD-03** Isolated tests | Every test creates its own fixtures; no test reads state produced by another. | Run alone vs in suite |
| **QD-04** Shared state is a defect | Passes alone / fails in suite → shared mutable state or ordering — fix the test, never reorder to hide it. | Suite isolation check |
| **QD-05** Flaky means race | Intermittent failures point to race conditions or missing cleanup, not assertion tweaks. | Observe across runs |
| **QD-06** No debug calls | No dump/print-style debug output in test files; use the test harness's inspection methods instead. | Grep debug calls |
| **QD-07** Floors are informational | Reported line-coverage floors (business logic ≥ 90%, domain rules 100%, handlers/data ≥ 80%) inform, never replace requirement coverage. | Report floors; do not gate on them |
| **QD-08** Smallest confirming run | Failing test → siblings → affected area suite → full suite only for core infrastructure. | Run sequence review |

Health indicators:

| Symptom | Diagnosis |
|---------|-----------|
| Fails in isolation, passes in suite | Shared state or ordering issue |
| Passes in isolation, fails in suite | Shared mutable fixture or ordering defect |
| Flaky (sometimes passes) | Race condition or missing cleanup |
| High coverage but bugs persist | Testing the wrong thing (internals, not behavior) |
| Breaks on unrelated refactor | Test couples to implementation details |
| Fails on the clean baseline too | Pre-existing issue — flag it, don't fix it silently |

## Efficiency

Run tests in the smallest scope that gives confidence, batch executions instead of running after
every change, and budget memory. The full suite is the exception, not the default.

| Rule | Description | Validation |
|------|-------------|------------|
| **EF-01** Lightest verification first | Pick the smallest run that answers the change: targeted filter → area suite → full suite. | Change-type matrix |
| **EF-02** Batch verification | Make ALL changes first, then verify once — never run tests after every individual edit. | Session workflow |
| **EF-03** Scope matches risk | Single-method change → targeted; cross-area refactor → area suite; core infrastructure changed → full suite ONCE. | Review |
| **EF-04** Respect the memory budget | Run suites sequentially on memory-limited machines; chaining the full suite twice for one change is not batching. | Time/memory smoke test |
| **EF-05** Flag scope creep | A run slower than its scope implies broad filters, full-suite includes, or an unintended database trait on pure-logic tests. | Time smoke test |

Typical cost ladder (scales with project size): single test/class in seconds; area suite in
minutes; full suite an order of magnitude more. When the full suite must run, run it once after all
changes are batched — never per edit.

## Validation

- [ ] Tests trace to a requirement ID / acceptance criterion; no orphans, no padding
- [ ] AAA, one behavior per test, names reveal intent
- [ ] Mocks only at the boundary; the real data layer is under test
- [ ] No debug calls; no real side effects (HTTP/disk/queue/mail)
- [ ] Run alone vs in suite; flaky runs investigated as races, not tweaked
- [ ] Verification ran at the lightest scope that fits the change