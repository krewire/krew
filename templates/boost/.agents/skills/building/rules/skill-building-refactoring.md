# Skill Rule — Refactoring Safety

Refactoring changes structure without changing behavior; structural edits must never silently alter
observable behavior. A refactor that changes behavior is not a refactor — it is a bug shipping under
a refactor label. This rule sets the safety gates, the during-refactor discipline, the
after-refactor verification, and the destructive patterns to avoid. Complements the project-agnostic
`skills/building/SKILL.md` and `skills/decomposing/SKILL.md`.

## Metadata

- **ID:** `rule-skill-building-refactoring`
- **Scope:** `global` — governs any structural refactor in any project
- **Severity:** `contextual` — binds while restructuring code
- **Enforcement:** `manual` (review) + `automated` (test suite, style/lint scanner where the project has one)
- **Related:** `rules/global-conventions.md` (atomic commits, edit policy), `rules/skill-testing.md`, `rules/skill-building-laravel.md` (project-level specifics)

---

## Safety Gates — Run Before Any Refactoring

These gates establish a trusted baseline. Skipping any of them means you cannot detect damage
because you cannot prove what "before" looked like.

### Test suite passes

**Intent:** a green baseline lets you attribute later failures to your refactor rather than to a
pre-existing fault.

**How to apply:** run the full suite (or the affected module/area suite at minimum) at session
start; record which suites were green. If the suite is already red, flag the pre-existing failure and
do not treat your refactor's red results as clean.

**Pitfall:** starting a refactor with a red suite and "hoping" the target area still works.

### `git status` clean (start from a clean slate)

**Intent:** the working tree holds only your intended changes, so `git diff` stays a reliable signal
of what you altered.

**How to apply:** check `git status`; commit or stash unrelated work before starting. During
refactoring use `git diff` after every step to prove only the target lines changed.

**Pitfall:** refactoring while carrying a half-finished feature in the same working tree.

### No debug output in the codebase

**Intent:** no `dd`/`dump`/`ray`/`var_dump`-style calls that could fire during refactored paths.

**How to apply:** grep the touched files for debug calls before and after; never introduce them to
inspect refactored logic — use tests instead.

**Pitfall:** using a debug call inside a temporary extraction "just to check" and forgetting it.

### If no tests exist for target code, write characterization tests first

**Intent:** lock in current behavior with tests before restructuring, so equivalence is verifiable.

**How to apply:** for any class you will restructure that lacks coverage, write tests that assert
the current observable behavior (inputs → outputs, exceptions, side effects), run them green, then
refactor, then re-run. These become permanent regression tests.

**Pitfall:** refactoring untested legacy code and confirming only "it still compiles".

---

## During Refactoring

- **One concern per commit** — never mix a refactor with a feature fix. A mixed commit defeats code
  review and makes `git bisect` point at the wrong change. If a behavior bug is discovered mid-
  refactor, fix it in a separate commit with its own message. "While I was in there" is scope creep.
- **Short feedback loops** — verify after every extraction, rename, or move (run the targeted test or
  the affected area's suite). Never batch structural steps before verifying; a failure after many
  moves could be anywhere.
- **Strangler pattern** — introduce the new module/class next to the old; route callers to the new
  one; verify behavior matches (ideally via characterization tests); delete the old once unused. A
  full rewrite-in-place leaves no fallback if the new structure misbehaves.
- **Do not change public API signatures without updating all callers** — gather callers first
  (`grep`/`rg` for the class and method across code and tests). Prefer extracting/new overloads over
  modifying signatures. If a signature must change, update all callers and any documented contract in
  the same commit, then run the affected suite.

---

## After Refactoring — Verification

- **Test suite passes** — structure changed, behavior must not have; the suite is the objective
  equivalence measure. Compare pass/fail counts to the safety-gate baseline.
- **Style/lint clean** — refactored code meets the same style bar as new code; run the project's
  style/lint tool on the dirty files and fix findings before commit.
- **No new unanchored `TODO`/`FIXME`** — debt annotations carry an owner and a date
  (`TODO(name, YYYY-MM-DD): message`); an undated note is un-actionable. Keep them in moved code if
  the debt survives the refactor.
- **New files honor the project's file conventions** — strict types, sorted imports, file-header
  order where the project mandates them.
- **Dead code cleaned up** — extractions often leave unused imports, dead branches, or now-unused
  private methods; grep the old file for removed symbols and delete leftovers.

---

## Destructive Patterns to Avoid

Banned outright — each has produced real breakage:

- **Changing a shared base class without updating all subclasses and callers** — a structural switch
  that shifts base-class method signatures breaks wiring that grep cannot easily discover.
- **Moving domain methods into the persistence layer** — a business rule relocated to a model ends
  up on the wrong layer; domain questions stay in the domain layer.
- **Removing an event/side-effect dispatch without checking for listeners** — an orphaned dispatch
  may leave documented side effects (notifications, projections) permanently silent. Grep for
  listeners before removing any dispatch.
- **Refactor and feature fix in the same commit** — see "One concern per commit".
- **Altering behavior that is already tested** — if a test asserts the old behavior and you change it
  "because the refactor needs it", you are changing behavior, not structure. Address the root cause
  in a separate concern, or document the new requirement first.

---

## Health-Based Fix Order — Debt Before New Logic

**What it enforces:** when touching a module/area that has known defects, fix the existing debt in
priority order **before** adding new logic; when the area is healthy, use it as the template.

**Why it matters:** a module can be structurally complete yet crash at runtime (schema mismatches,
broken views, missing layers). Adding a feature on top of that ships a broken slice and forces the
next agent to untangle debt + feature together. The standard fix order prevents compounding.

**How to apply (when a task touches a module with known issues):**

1. `git log --oneline -p` for that module's recent fixes — see what P0s are already filed.
2. Run the project's scan tooling scoped to the module (`scan_violations --module {Name}`) to
   baseline the debt; treat the result as the pre-existing baseline (see
   `rules/skill-auditing.md`).
3. Fix P0 items in this order before adding new logic:
   - **schema mismatches** (migration columns ≠ code refs → SQL errors);
   - **contract gaps** (actions returning `Model`/`void` instead of the expected response wrapper);
   - **hard-coded user-facing strings** (blocks localization);
   - **missing domain layer** (business rules living in models/actions instead of entities);
   - **side-effect dispatch inside transactions** (race conditions — dispatch after commit);
   - **dead code** (unused DTOs, observers, events without listeners).
4. Add comprehensive automated tests for the new logic — no orphan tests.

**Healthy-area pattern:** if the task is a new feature in a healthy module, use that module as the
template; if it is in a debt-laden module, budget extra time for the debt up front.

**Pitfalls to avoid:**

- Building a feature over a known crash (schema/Blade) and calling it complete.
- Mixing debt fixes into the feature's "one concern" commit — debt fixes are separate concerns.
- Treating baseline findings as regressions caused by current work.

**Verification:** P0 fixes landed before the feature in the diff order; the module's schema/Blade
no longer crashes; new tests trace to a spec requirement ID; debt fixes are in their own commits.