# Skill Rule — Security Audit Coverage & S-Rules

Project-agnostic security and privacy audit discipline: the full-sweep coverage umbrella plus the
S-rule baseline anti-patterns (OWASP + framework security practice). Complements
`skills/auditing/SKILL.md` (OWASP/CWE depth); arch-guard-class internal invariant scanning and the
blind QA protocol sit alongside, not instead of, this.

## Metadata

- **ID:** `rule-skill-auditing-security`
- **Scope:** `global` — governs any dedicated security/privacy audit in any project
- **Severity:** `contextual` — binds while auditing auth/input/output/PII/dependencies
- **Enforcement:** `manual` (audit discipline) + `automated` (security scanner where the project has one)
- **Related:** `rules/global-security.md` (OWASP/CWE/passwords/PII), `skills/auditing/rules/skill-auditing-qa-protocol.md`, `rules/skill-planning-issues.md` (→ files findings), `skills/auditing/SKILL.md`

---

## Audit Execution — Full-Sweep Coverage

A dedicated audit sweeps the codebase across the categories below and produces one issue per finding
with severity, impact, and fix recommendation.

### Why a dedicated audit

- **Vulnerabilities are cross-cutting** — a single XSS, a missing check on one mutation, or a leaked
  secret undermines the whole application regardless of how well the rest is built.
- **Automated scans produce false positives — and false negatives** — every finding must be verified
  manually against the code, or the audit outputs noise and misses real attacks.
- **Findings without structure are unactionable** — category, location, severity, reproduction, and
  fix can be triaged; a vague note cannot.

### Steps

1. **Run the scanners first (automation-first)** — any project security/invariant scanner (e.g.
   XSS, SQLi, mass assignment, auth patterns), then proceed.
2. **Audit each category in order** — authentication, authorization, XSS, SQL injection, mass
   assignment, file upload, PII, secrets & configuration, dependencies, then the cross-cutting
   controls (§Cross-Cutting Controls).
3. **Verify every scanner hit manually** — read the code, confirm or dismiss. Scans are triage, not truth.
4. **Record all findings, even out of scope** — prioritization happens downstream, not during the audit.
5. **Check existing issues before filing duplicates.**
6. **Output:** one issue per finding with the Issue Format below (via `skills/planning/SKILL.md` / the structured-issue rule).

### Issue Format

| Field | Content |
|-------|---------|
| **Category** | Which audit category (e.g. XSS Prevention) |
| **Location** | File path and line number |
| **Vulnerability** | OWASP category or specific risk |
| **Severity** | Critical / High / Medium / Low |
| **Reproduction** | Steps if applicable |
| **Fix** | Recommendation |

### Cross-Cutting Controls

Checked in **every** sweep, independent of category specifics:

- **CSRF** — all state-changing forms include a CSRF token (or use a framework mechanism that
  handles it); no state-changing route is exempt without a justification.
- **Rate limiting** on: login, password reset, account recovery, setup/claim tokens.
- **Session timeout** configured appropriately for the sensitivity of the application.
- **HTTPS** enforced in production — no cleartext on authenticated paths.

### Anti-Patterns

- Trusting scanner output verbatim — verify every hit; never file an unconfirmed false positive as
  Critical.
- Fixing findings mid-audit — scope creeps, the sweep stalls, and the fix is unverified against the
  remaining categories.
- Skipping out-of-scope findings — a leaked secret in docs or a dev-only route is still a finding.
- Filing duplicates — same issue, multiple categories; check existing issues and cross-reference.
- No reproduction steps — a finding without a way to confirm it cannot be triaged or verified.

---

## S-Rules — Baseline Anti-Pattern Checks

The scanner baseline checks the project against ten security rules (S1–S10) based on OWASP and
framework security best practices. These are the security dimension of the quality gate; deep OWASP
and PII handling beyond surface anti-patterns are covered by the security-audit skill.

| ID | Rule | Detection |
|----|------|-----------|
| **S1** | XSS: unescaped output only for trusted content, always escaped for user data | Unescaped (`{!! … !!}`-style) output of variables without an escaping wrapper |
| **S2** | SQL Injection: parameterized queries only | Raw SQL without bindings |
| **S3** | Mass Assignment: create/update with only allowed fields | Bulk `create/update` from the raw request |
| **S4** | CSRF: tokens on state-changing forms/AJAX | Missing CSRF tokens |
| **S5** | Authentication: `auth()->check()` or `@auth` gate before sensitive operations | Unprotected sensitive endpoints |
| **S6** | Authorization: explicit authorize/can check before actions | Missing authorization checks |
| **S7** | Rate limiting: limiter or throttle middleware on auth endpoints | No rate limiting on auth endpoints |
| **S8** | Secrets: no hardcoded passwords/tokens/keys | Hardcoded credentials |
| **S9** | File upload: validate type, size, content | Unrestricted uploads |
| **S10** | Dependency auditing: no known dependency vulnerabilities | Advisory findings from `composer audit` / `npm audit` |

### Per-Rule Rationale & How to Apply

- **S1 — XSS:** unescaped output of user-controlled data lets an attacker's `<script>` run in the
  victim's browser with the victim's session (stored XSS). Only render "trusted" raw HTML from a
  sanitizing renderer; everything else is auto-escaped by the template engine.
- **S2 — SQL Injection:** raw SQL with concatenated input is a direct query-injection vector. Use the
  query builder/ORM first; bind parameters if raw SQL is needed.
- **S3 — Mass Assignment:** bulk-passing the raw request passes every client-supplied key into the
  model; unexpected keys slip through and privilege-style fields get set. Slice to allowed fields or
  build from validated data.
- **S4 — CSRF:** a form missing its token (or an AJAX call missing the header) is a cross-site
  request-forgery window — a third-party site can post to logout/profile/admin endpoints while the
  victim is logged in.
- **S5 — Authentication:** sensitive operations must not be reachable anonymously; an
  unauthenticated request reaching an endpoint that echoes or mutates PII leaks data publicly.
- **S6 — Authorization:** authentication proves *who*; authorization proves *may*. A logged-in low
  user able to do a privileged action is a higher-severity bug than a public logout.
- **S7 — Rate Limiting:** auth endpoints (login, password reset, recovery) are brute-force targets;
  an attacker runs an offline dictionary at thousands of guesses per minute if unthrottled.
- **S8 — Secrets:** hardcoded passwords/tokens/keys survive repository history and every clone. Use
  env-backed config; never default secrets in committed config.
- **S9 — File upload:** unrestricted uploads accept executable content, oversized files (disk
  exhaustion), and malformed media (parser crashes). Validate type/mime/size; reject
  executable/polyglot content; store outside the webroot with validated collections.
- **S10 — Dependency auditing:** known vulnerabilities in dependencies (CVE/GHSA) are a direct remote
  execution / data-exfiltration channel. Run dependency audits; upgrade affected packages to patched
  versions.

### Severity Guidance

- **S1, S2, S3, S8, S9** — always **CRITICAL** when a live exploit path is confirmed (user input
  reaches an unsafe sink); **HIGH** when the sink exists but is unreachable from user input today.
- **S4, S6** — **HIGH** (authorization/CSRF bypass risk).
- **S5, S7, S10** — **MEDIUM** to **HIGH** depending on endpoint exposure (auth routes without rate
  limiting are **HIGH**).

When in doubt, err high: a security finding under-rated becomes a "known gap" nobody schedules.

---

## Verification

- The audit output covers all categories — no empty rows.
- Every scanner hit was either confirmed into an issue or explicitly dismissed with a reason.
- Each issue carries the full Issue Format — no missing severity, location, or fix.
- Pair every confirmed finding with the deep security-audit skill when the surface touches
  authentication or PII. Findings are filed via the structured-issue rule with a `security` scope.