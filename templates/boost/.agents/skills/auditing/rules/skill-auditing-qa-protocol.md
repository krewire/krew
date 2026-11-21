# Skill Rule — Blind QA Protocol & Compliance Scorecard

Project-agnostic independent QA audit discipline: how a six-phase quality audit executes, the
independence ("blind") doctrine that keeps it honest against global standards, and the compliance
scorecard that makes the result measurable. Complements `skills/auditing/SKILL.md` and
`rules/global-standards.md` (§Quality Gates). Not a project-invariant check — that is arch-guard /
internal scanning territory, which this protocol sits *alongside*, never replacing.

## Metadata

- **ID:** `rule-skill-auditing-qa-protocol`
- **Scope:** `global` — governs any independent, benchmark-driven QA audit in any project
- **Severity:** `contextual` — binds while executing a QA audit
- **Enforcement:** `manual` (execution discipline) + `automated` (link/scan gates where the project has them)
- **Related:** `rules/global-standards.md`, `skills/auditing/rules/skill-auditing-security.md`, `rules/skill-planning-issues.md` (→ issues the findings), `skills/auditing/SKILL.md`

---

## Blind Execution — Independence Doctrine

The QA protocol is **explicitly blind**: it evaluates the codebase as if it were an unknown project
reviewed for the first time, against purely external benchmarks (OWASP, ISO 25010, CWE/SANS, WCAG,
PSR, framework community practice). Project docs must **not** be loaded before or during execution.
This protocol applies only when the audit caller selects blind mode. Contextual repository audits
use the normal auditing skill and may read project rules and runtime context.

### Rationale

The value of the audit collapses if it inherits the project's blind spots. An internal auditor
naturally excuses internal conventions and evaluates against the project's own rules; a blind
audit cannot — with no project context loaded, a finding is judged only against the external
standard it violates. The failure mode of a non-blind audit is a self-confirming pass: a codebase
can be internally consistent and still fail every global standard — and nobody sees it.

Contrast with internal invariant checks:

| Aspect | Internal scan (arch-guard) | Blind QA protocol |
|--------|--------------------------|-------------------|
| Rule base | Project-defined invariants | Global industry standards (OWASP, CWE, ISO, WCAG, PSR) |
| Project knowledge | Required — checks internal conventions | Forbidden — blind against external benchmarks |
| Output | Internal issues | Issues + compliance scorecard |

### Ten Key Rules

1. **Blind execution** — do NOT load project documentation before or during execution.
2. **External standards only** — every finding references an external standard (OWASP, CWE, PSR, ISO,
   WCAG, …), never project-specific invariants.
3. **Evidence-based** — every finding includes file path, line number, and concrete evidence.
4. **Severity follows CVSS** — use the Common Vulnerability Scoring System for security findings (see
   §Compliance Scorecard).
5. **No fixes during audit** — record findings; issue creation is opt-in and fixes happen downstream.
6. **Blocker exception** — if a finding actively prevents the audit from running (e.g. the app will
   not boot), fix minimally first, then continue.
7. **Comprehensive scope** — check every module, every route, every model — not just recently changed code.
8. **Independent of project rules** — project-specific conventions are NOT part of this audit; only
   global standards are.
9. **Report by default** — every audit ends with findings and a summary report. GitHub issue creation
   requires explicit authorization; audits never commit skill/rule files.
10. **Overlap transparency** — when a finding overlaps with an internal-scan finding, note the overlap
    in the issue body but still file independently.

Good vs not-blind finding:

```text
GOOD (blind):  "Stored XSS — user content rendered unescaped in views/foo/show.blade.php:41.
                Violates OWASP A03 (Injection) / CWE-79."
BAD (not):     "Violates a project strict_types convention in CreateFooAction."  ← internal rule
```

### Anti-Patterns

- Loading project docs "just for context" — breaks blindness; findings start rationalizing internal decisions.
- Scoring against project rules — defeats rule 8.
- Skipping the blind pass "because I know this codebase" — the discipline is the method.
- Fixing findings during the audit (except the minimal blocker intervention of rule 6).
- Reviewing only diffs — rule 7 demands full coverage.

---

## Six-Phase Protocol — Audit Execution Order and Blockers

A QA audit runs **6 phases in order**, each a major quality domain. Phases decide what to scan, in
what sequence, what blockers stop forward motion, and what each phase must output.

```
PHASE 1: Automated Scanning → PHASE 2: Security Audit → PHASE 3: Quality & Reliability
      → PHASE 4: Standards Compliance → PHASE 5: Performance & Efficiency
        → PHASE 6: Report, Issues & Commit
```

### Why order matters

Each phase depends on the previous: you cannot assess security meaningfully before the automated
tooling has surfaced the mechanical defects, and you cannot write a trustworthy report before every
phase produced its findings. *Blockers* — conditions that halt a phase — keep a phase that cannot run
from silently passing.

| Phase | Goal | Key checks (covered in depth by) |
|-------|------|----------------------------------|
| **1 — Automated Scanning** | Dependency audit, code-style tooling, project scanners, dead-code, build | Security scanners, style scan, build |
| **2 — Security Audit** | Evaluate against OWASP Top 10 and CWE/SANS Top 25 | `skills/auditing/rules/skill-auditing-security.md`, OWASP/CWE |
| **3 — Quality & Reliability** | Error handling, logging, input validation, auth/authz, test coverage, duplication | Error-handling, logging, validation, auth, test-coverage checks |
| **4 — Standards Compliance** | Framework best practices, WCAG, session, cryptography | PSR/WCAG, cryptography |
| **5 — Performance & Efficiency** | Database, caching, memory, frontend | Performance checks |
| **6 — Report & Optional Publication** | Consolidate, score, optionally publish issues, report | This rule §Scorecard + `rules/skill-planning-issues.md` |

### Blocker policy

A blocker prevents a phase from executing reliably. Fix it **minimally** (just enough to unblock),
or record it as a finding:

| Phase | Blocker | Action |
|-------|---------|--------|
| 1 | Dependency audit files a critical CVE | Record as finding; do NOT auto-update (out of scope) |
| 1 | Build fails | Record; attempt minimal fix if simple |
| 2 | Critical stored XSS | Record; do NOT fix (audit scope) |
| 2 | Critical SQL injection | Record; flag for immediate fix |
| 2 | Hardcoded credentials in source | Record; flag for immediate rotation |
| 3 | Debug output (`dd()`/`dump()`) in production code | Record; remove immediately |
| 3 | Test suite won't run | Record as blocker; fix if simple |
| 3 | Empty catch blocks swallowing errors | Record as finding |
| 4 | `declare(strict_types=1)` missing | Record; simple fix |
| 4 | No CSRF protection on forms | Record; may be handled by the framework |
| 6 | Issue CLI not authenticated | Report findings to the user directly; skip issue creation |
| 6 | No findings across all phases | Still create the report — a clean audit is valuable |

### Size-awareness

A full-project QA is **L** — inform the user, propose a session plan (e.g. by module), and run each
phase per session. Never attempt all 6 phases on the entire codebase in a single pass. Skipping a
phase is not allowed even when it "looks fine" — an empty phase is a valid, reportable result; a
skipped phase is not.

---

## Compliance Scorecard — Measuring Against Global Standards

The scorecard is the audit's quantitative output: one table, per global standard, of the score
achieved and the notes explaining it. It makes results comparable across audits and highlights which
standard is weakest, so triage goes to the biggest compliance gap. It also fixes severity assignment
for security findings: CVSS v3.1.

### Scorecard table

Delivered in the final report:

| Standard | Score | Notes |
|----------|-------|-------|
| OWASP Top 10 | X/10 categories clean | … |
| CWE/SANS Top 25 | X/25 CWEs absent | … |
| ISO 25010 | X/8 characteristics met | … |
| PSR-1/4/12 | Pass/Fail | … |
| WCAG 2.1 AA | X/11 criteria met | … |
| Framework best practices | X/Y checks pass | … |

ISO 25010, if scored per characteristic, uses 1–5 (5 = fully meets all sub-characteristics; 3 =
meets core requirements with notable gaps; 1 = fails to meet requirements).

### CVSS v3.1 severity classification

Severity is the authoritative scale, not an ad-hoc label:

| Rating | Score Range |
|--------|-------------|
| None | 0.0 |
| Low | 0.1 — 3.9 |
| Medium | 4.0 — 6.9 |
| High | 7.0 — 8.9 |
| Critical | 9.0 — 10.0 |

For web findings a simplified approximation is allowed *only* as a triage aid:
`Base ≈ Impact × Exploitability` (Impact: no access 2.6 · limited 5.0 · full data 7.5 · full system
10.0; Exploitability: requires auth 5.0 · special conditions 3.9 · no-auth trivial 8.0 · no-auth
automated 10.0). Official CVSS scoring wins for a contested finding.

### Scorecard anti-patterns

- Severity by gut instead of CVSS.
- Scoring rows the audit did not actually cover (mark untested standards as such, don't estimate).
- Overloading the scorecard with every finding — the scorecard is the summary; issues carry the detail.

---

## Verification

- Blind mode loads **no** project architecture/convention/module docs before or during the audit; contextual mode records which project rules were consulted.
- Every filed issue names an external standard (OWASP category, CWE-ID, PSR, ISO characteristic,
  WCAG criterion) plus file:line evidence.
- The report shows phases executed 6/6, no skipped phase; every blocker listed as resolved or
  recorded as a finding — never silently ignored.
- The report contains the full scorecard table, no standard missing; each score reconciles with the
  phase findings; every security severity maps to a CVSS v3.1 range.
- The audit run ends with findings and a user report. If issue publication was selected, issue URLs
  are included; no audit mode commits rule/skill files automatically.
