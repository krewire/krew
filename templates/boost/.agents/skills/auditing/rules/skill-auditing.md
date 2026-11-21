# Skill Rule — Auditing Baselines & Intentional States

An audit is only meaningful against a known, current baseline. This rule defines how to treat
**intentional or deferred states** that look like defects, how to read scanner baselines so a
"regression" means a delta above pre-existing counts (not a match), and when to fix vs file.

## Metadata

- **ID:** `rule-skill-auditing`
- **Scope:** `global` — governs auditing, arch-guard scans, and any "is this a defect?" triage
- **Severity:** `mandatory` — applies to every audit and scan readout
- **Enforcement:** `manual` (review) + `automated` (scanner output interpretation)
- **Related:** `skills/auditing/rules/skill-auditing-qa-protocol.md`, `skills/auditing/rules/skill-auditing-security.md`, `skills/auditing/rules/skill-auditing-spec-audit.md`, `skills/building/rules/skill-building-refactoring.md`, `rules/self-improvement.md` (fix-or-file)

---

## Baselines — Regression = Delta Above Pre-existing Counts

**What it enforces:** never treat a pre-existing scan finding as a regression caused by current work.

**Why it matters:** scanners are upgraded and improve heuristics. Old counts are **not comparable**
to new ones; treating a match against a post-upgrade baseline as a regression produces false alarms
and drive-by "fixes" that churn unrelated files.

**How to apply:**

- Record refreshed counts per scanner in a project context file (e.g. `codebase-intentional-states.md`).
- A **regression** is a delta above the current baseline, never a match to it.
- A finding that predates current work is **not** a regression; fix it in a dedicated cleanup session
  or file an issue — never as a drive-by edit.
- When scanner heuristics change, refresh the baseline in the same run before interpreting output.

**Pitfalls to avoid:**

- Quoting old baseline figures after a scanner upgrade ("32/11/232") — refresh first.
- Folding a pre-existing finding into the current change's commit history without recording it as baseline.
- "Fixing" intentional states because they look like defects (see next section).

**Verification:** scan readouts cite the baseline file + its refresh date; any delta is explicitly
attributed to current work.

---

## Intentional / Tracked States — Do Not Casually "Fix"

**What it enforces:** some codebase states are deliberate or tracked as duplication/debt. Recognize
them before touching exception behavior, scan output, or specs.

**Why it matters:** a tracked duplication (e.g. two specs describing the same contract) or a deferred
state may look like an obvious defect. Resolving it casually destroys a registered decision and
breaks test mappings.

**How to apply:** a state is _intentional/tracked_ when a context file or ADR records it — e.g.
"exception hierarchy defined twice — resolve the duplication in a spec pass _before_ touching
exception behavior", or "specs use 5-char alphanumeric IDs — do not reintroduce sequential numbering".
When a task touches such a state:

1. Check the project context/ADR for a recorded resolution rule.
2. If one exists, follow the recorded order of operations (e.g. reconcile the spec duplication first).
3. If none exists, treat it as a finding: record a decision (context file or ADR) before acting.

**Pitfalls to avoid:**

- "Fixing" what is recorded as intentional — the change itself becomes the defect.
- Skipping the recorded prerequisite step (e.g. updating `ExceptionsTest` after reconciling specs).

**Verification:** touched area's recorded state/rule was consulted first (cited in the report); no
recorded decision was silently reversed.

---

## Fix or File — Never Silently Tolerate

**What it enforces:** a finding you choose not to fix now must be recorded as a next step, not dropped.

**Why it matters:** silently tolerating a pre-existing defect forces the next agent to rediscover it
and ships the risk to the user-under-test.

**How to apply:** for any finding out of current scope, either fix the safe adjacent issue or file a
tracked issue (see `skills/planning/rules/skill-planning-issues.md`); name the issue in the report. Baselines are the
recording vehicle for deferred debt.

**Pitfalls to avoid:** reading a scan, deciding findings are "pre-existing", and ending the session
without filing or refreshing the baseline.

**Verification:** every finding is either fixed, filed, or absorbed into the recorded baseline.

---

## Skill Auditing References

- `skills/auditing/rules/skill-auditing-spec-audit.md` — bidirectional spec↔implementation audit
- `skills/auditing/rules/skill-auditing-qa-protocol.md` — blind QA protocol vs external benchmarks
- `skills/auditing/rules/skill-auditing-security.md` — security & privacy audit (OWASP)
- Workspace `tools/scan_*.py` — project scanners (see `AGENTS.md §Automation Scripts`)
