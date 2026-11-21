# Global Standards — SDLC Industry Alignment


Project-agnostic standards that this registry enforces by default; stack-specific extensions may be added (e.g. Laravel C1-C8/D1-D6).

## Metadata

- **ID:** `rule-global-standards`
- **Scope:** `global` (all agents/skills/phases)
- **Severity:** `mandatory`
- **Enforcement:** `manual` (review) + `self-check` (agent workflow)
- **Related:** `AGENTS.md §Global Standards per Phase`, `rules/global-conventions.md`, `rules/global-security.md`

---

## Planning — SRS & Backlog

**Standards:** IEEE 29148 (SRS), BABOK, PMBOK, Scrum Guide, ISO 25010

| Rule | Description | Validation |
|------|-------------|------------|
| **SRS-01** Requirements-driven | No behavior without a documented requirement or issue. | Issue checklist traces every change |
| **SRS-02** NFR quantification | Every NFR has an SLO and measurement point (e.g. "p95 < 200ms measured at `GET /api/x`"). | Plan gate checks NFR → SLO |
| **SRS-03** DoR/DoD | Stories meet Definition of Ready/Definition of Done before Build. | Checklist in planner |
| **SRS-04** 3S + ISO 25010 | Score functional suitability, performance, security, maintainability per ISO 25010. | Review gate |

**Wrong:** "We need faster login." (unmeasurable) **Right:** "FR-AUTH-03: p95 login < 800ms on 3G, measured via Pulse, 99.9% success SLO."

### ISO 25010 Quality Attributes

Source: ISO/IEC 25010:2023. Eight characteristics; score each per slice on the evaluation matrix below (not once per project). Priority weights are project-specific — the web-app baseline below is a starting point.

| Characteristic | Sub-characteristics (what to check) | Where it is gated |
|----------------|--------------------------------------|-------------------|
| Functional suitability | Completeness, correctness, appropriateness — edge cases (empty collections, nulls, zero-division); no stub/TODO features | Review gate |
| Performance efficiency | Time behaviour (N+1, unbounded queries, missing indexes), resource utilization (memory/file-handle leaks, huge collections), capacity (pagination, upload limits) | `### Performance & Efficiency` |
| Compatibility | Co-existence, interoperability (standard data formats, no vendor-specific SQL, browser fallbacks) | Review gate |
| Usability | Learnability, operability (labels, helpful errors, undo), user error protection (confirmation for destructive actions), inclusivity | `### WCAG 2.x AA — A11y Gates` |
| Reliability | Maturity (error handling), availability (health checks), fault tolerance (fallbacks, retries), recoverability (transactions, backups) | `### Error Handling` |
| Security | Confidentiality, integrity, non-repudiation, accountability, authenticity (audit trail, MFA, access control) | `### Security Standards` |
| Maintainability | Modularity, reusability, analysability (typed contracts, docs), modifiability (SOLID, DI), testability | `### Code Quality Metrics` |
| Portability | Adaptability (env-based config, no hardcoded values), installability (setup scripts, migrations), replaceability (standard interfaces, minimal lock-in) | Review gate |

**Evaluation matrix:** 5 = excellent, 4 = good, 3 = adequate, 2 = poor, 1 = very poor — a 1-2 on any security or reliability sub-characteristic blocks release.

**Priority weights (web-app baseline):** Security 25%, Reliability 20%, Functional Suitability 20%, Maintainability 15%, Usability 10%, Performance 5%, Compatibility 3%, Portability 2%.

---

## Design — Architecture

**Standards:** ISO 42010, C4 Model, UML, OpenAPI 3.1, 12-Factor, OWASP ASVS/STRIDE

| Rule | Description | Validation |
|------|-------------|------------|
| **ARC-01** C4 as SSOT | Level 1 (Context), Level 2 (Container), Level 3 (Component) live in `docs/architecture.md` + `refs/`. Code ≠ docs → git-history check first. | `scan_broken_links.py` + arch review |
| **ARC-02** ADR | Every structural decision gets `docs/adr/NNN-*.md` (link from AGENTS.md). | ADR staleness scan |
| **ARC-03** API-first | REST/GraphQL contracts in OpenAPI 3.1; breaking change = SemVer major + migration guide. | Spec lint |
| **ARC-04** Threat model | STRIDE + OWASP ASVS L1 before data/security design. | Design checklist |

---

## Implementation — Code

**Standards:** Clean Code, SOLID, DRY, Conventional Commits, SemVer, 12-Factor, PSR-1/4/12 (or language equivalent)

Covered in `rules/global-conventions.md`. Key gates: SOLID, DRY, strict types, no secrets in repo, 12-Factor env via `.env`, explicit parameter and return types on every signature, no debug output committed.

### PSR Standards — PSR-1 / PSR-4 / PSR-12

| Rule | Description | Validation |
|------|-------------|------------|
| **PSR-1** Files | UTF-8 without BOM; no closing `?>` tag; a file declares symbols **or** has side effects, never both. | `grep -rn "?>"` review |
| **PSR-1** Names | Classes PascalCase, methods camelCase, constants UPPER_SNAKE_CASE, boolean methods `is/has/can/should`-prefixed. | Lint / CS-fixer |
| **PSR-4** Autoload | Namespace maps 1:1 to directory; one class per file; class name = file name. | `composer dump-autoload --optimize` clean |
| **PSR-12** Style | `declare(strict_types=1)` in every non-framework file; explicit visibility; keyword spacing; opening brace on same line; one argument per line when a signature exceeds 80 chars. | `pint --test` (or language equivalent) |

### Code Quality Metrics

Maintainability floors per function/class (ISO 25010 Maintainability). A finding past the hard limits is a review blocker.

| Metric | Floor / Limit | Validation |
|--------|---------------|------------|
| Cyclomatic complexity | < 10 per method (alarm > 20: extract) | `phpmd` equivalent |
| Nesting depth | ≤ 3 levels | Review / linter |
| Method length | < 20 lines target, 50 max | Review / linter |
| Class length | < 300 lines target, 500 max (no god classes) | Component-count scan |
| Parameters | > 3 → data object / parameter object | Review |
| Duplication | No exact copy-paste; extract shared code for structural duplication; consolidate semantic duplication | `phpcpd` equivalent |
| Dead code | No unused classes/methods; TODO/FIXME older than 6 months → removed or filed | Dead-code scan |
| Magic numbers/strings | Named constants or enums, never literals in logic | Review |
| Dependencies | No circular imports between packages/modules | Dependency-graph scan |
| Return types | Consistent and explicit — nullable is `?Type`, never sometimes-null | Static analysis |

### Error Handling

| Rule | Description | Validation |
|------|-------------|------------|
| **ERR-01** No stack traces to users | User-facing error pages; raw exceptions stay server-side; no sensitive info (SQLSTATE, file paths) in messages. | Review; `APP_DEBUG=false` smoke test |
| **ERR-02** Never swallow | Empty catch blocks forbidden; at minimum log with context (message + relevant data). | Review / grep |
| **ERR-03** Specific exceptions | Domain/business-rule exception types distinct from generic infrastructure exceptions; no bare `catch (\Exception)`. | Review |
| **ERR-04** Fail fast | Validate input early, throw early; do not use exceptions for flow control. | Review |
| **ERR-05** Atomic mutations | Multi-step writes wrapped in a database transaction — all-or-nothing, no partial state. | Review |
| **ERR-06** Graceful degradation | Non-critical services (analytics, third parties) fail soft: catch, log warning, fallback value, page still renders. | Review |
| **ERR-07** User-friendly messages | Human-readable, localized messages for users; technical detail only in logs. | Review |
| **ERR-08** Null-safe rendering | Views use null-safe access with defaults (`a?->b ?? default`); no render-time exceptions for missing optional data. | Review |

**Checklist severity:** debug calls in production = High; empty catches = Medium; generic catches = Low; missing transaction on multi-step mutation = High; missing try/catch on external calls = Medium; styled, stack-trace-free error pages required.

### Security Standards — OWASP Top 10 2021 / SANS Top 25 / CWE

Baseline awareness = OWASP Top 10; weakness taxonomy = CWE; frequency × impact ranking = SANS Top 25. Fold into one triage: **fix by SANS rank, classify by CWE, verify by OWASP category.**

| OWASP A0x | Name | CWE anchor | What to check |
|-----------|------|------------|---------------|
| A01 | Broken Access Control | CWE-862/863 | IDOR, missing function-level access control, least privilege, CORS `*` on authenticated endpoints, open redirect on user-controlled URLs |
| A02 | Cryptographic Failures | CWE-327/328/798 | Strong credential hashing, cryptographically strong randomness for security tokens, no secrets in source/config, TLS everywhere, no sensitive data in URLs/query strings |
| A03 | Injection | CWE-89/79/78/94 | Parameterized queries; escaped output; no shell/exec with user input; no eval/assert; no template injection |
| A04 | Insecure Design | — | Business-logic bypasses, rate limiting on auth flows, resource limits (uploads/body/query), anti-automation |
| A05 | Security Misconfiguration | CWE-732 | Debug off in production, security headers (CSP, `X-Frame-Options` DENY, HSTS, `X-Content-Type-Options` nosniff), no default credentials, no directory listing |
| A06 | Vulnerable Components | — | `composer audit` / `npm audit` per release; no `*` constraints; no abandoned dependencies |
| A07 | Authentication Failures | CWE-287/306/521 | Rate-limited login, account lockout, session regeneration after login, session expiry enforced, password policy ≥ 8 chars with complexity, MFA for sensitive operations |
| A08 | Software/Data Integrity | CWE-502/347 | No deserialization of untrusted data; signed updates/cookies; CI/CD secrets hygiene; lockfile integrity |
| A09 | Logging & Monitoring Failures | CWE-778/532 | Security events logged, no PII in logs, alerting on repeated failures, audit trail for mutations — see `### Logging` |
| A10 | SSRF | CWE-918 | Validate/allowlist user-controlled URLs; block loopback and cloud-metadata (169.254.169.254); DNS-rebinding aware |

**SANS priority tiers:** Tier 1 (must-check) — XSS (CWE-79), SQL injection (CWE-89), path traversal (CWE-22), CSRF (CWE-352), unsafe upload (CWE-434), missing authorization (CWE-862), hardcoded credentials (CWE-798), missing authentication (CWE-306). Tier 2 (common) — input validation (CWE-20), deserialization (CWE-502), file permissions (CWE-732), code injection (CWE-94), privilege management (CWE-269), SSRF (CWE-918). Tier 3 — XXE (CWE-611), command injection (CWE-77).

**Severity triage (CVSS v3.1):** None 0.0 / Low 0.1-3.9 / Medium 4.0-6.9 / High 7.0-8.9 / Critical 9.0-10.0. Any finding in CWE-79/89/22 auto-escalates to High/Critical.

**Runtime probes:** authentication bypass (protected routes unauthenticated), authorization bypass (other users' resources), rate limiting (rapid-fire requests), error-message leakage, upload type/size/content validation.

### Performance & Efficiency

Query and data-shape anti-patterns (P1-P5) keep the query count, memory footprint, and latency bounded as data grows:

| ID | Rule | Detection |
|----|------|-----------|
| **P1** | No N+1 — eager load relationships read in loops (`with()`) | Relationship access inside `foreach`/`map` triggers a query per iteration |
| **P2** | Select only needed columns (`select('id', 'name')`), never `*` | `Model::all()` / `->get()` without column selection on wide tables |
| **P3** | `chunk()`/`cursor()` for large datasets (exports, batch jobs) | `->get()` on potentially huge collections materializes all rows in memory |
| **P4** | Cache expensive/hot reads with a TTL — keys from a registered registry | Repeated heavy aggregates recomputed per request |
| **P5** | `exists()` instead of `count() > 0` for presence checks | `count() > 0` scans whole index for a boolean answer |

**Additional checks:** index columns used in WHERE/JOIN/ORDER BY; filter at the database level, not in memory; frontend — non-critical scripts `defer`, bundle size watched at build.

**Budgets (web baseline):** TTFB < 200ms (warn > 500ms), queries per page < 15 (warn > 30), LCP < 2.5s (warn > 4s), total page weight < 500KB (warn > 2MB), JS bundle < 200KB (warn > 500KB), N+1 = 0.

**Enforcement:** profile hot paths with APM (Pulse-equivalent); Lighthouse CI for frontend budgets; review gate for N+1/pagination/eager-loading.

---

## Testing — Verification

**Standards:** ISTQB, ISO 29119, Test Pyramid (unit 70% / integration 20% / e2e 10%), TDD/BDD, OWASP Testing Guide, WCAG 2.2

| Rule | Description | Validation |
|------|-------------|------------|
| **TST-01** Traceable tests | Every test has clear descriptive naming. No orphan tests. Coverage = behavior covered, not just lines. | Test suite execution |
| **TST-02** Pyramid | Prefer fast isolated tests; e2e only for critical UC paths. | Test-plan review |
| **TST-03** Risk-based | High-impact × high-likelihood paths first (OWASP, payment, auth). | Planning impact matrix |
| **TST-04** A11y | UI tests include WCAG 2.2 AA checks (axe/contrast/keyboard). | Verify gate |

### Test Coverage Policy

Behavioral coverage governs (TST-01); line coverage is a **diagnostic floor**, never the definition of done — a requirement with no test is a gap.

| Rule | Description | Validation |
|------|-------------|------------|
| **TC-01** Cover behavior | Test what the code *does* (result + side effects), not internal calls or implementation details. | Review |
| **TC-02** Edge + failure paths | Explicit rejections, invalid formats, empty/null inputs, timeouts — not just the happy path. | Targeted test runs |
| **TC-03** AAA structure | Arrange → Act → Assert; one clear behavior per test. | Review |
| **TC-04** No shared state | Independent tests via factories/transactions; never cross-test dependencies. | Suite isolation check |
| **TC-05** Mock only externals | Real objects for own code; fake payment gateways, email, external APIs. | Review |
| **TC-06** Don't test framework | Assert on your logic's outcomes, not framework mechanics. | Review |

**Diagnostic floors (per-layer, informational):** business logic ≥ 90%, domain rules 100%, HTTP handlers ≥ 80%, data layer ≥ 80%, utilities ≥ 90%, views ≥ 50%.

**Health indicators:** passes alone / fails in suite → shared state or ordering; flaky → race condition or missing DB cleanup; high coverage but bugs persist → testing the wrong thing; breaks on refactor → testing implementation details.

### WCAG 2.x AA — A11y Gates

Target WCAG 2.2 AA (a superset of the 2.1 AA baseline). Four principles: Perceivable, Operable, Understandable, Robust.

| Rule | Description | Validation |
|------|-------------|------------|
| **A11Y-01** Contrast | Normal text ≥ 4.5:1, large text (≥ 18pt or ≥ 14pt bold) ≥ 3:1; status never color-only (WCAG 1.4.1, 1.4.3). | axe / Lighthouse / contrast analyser |
| **A11Y-02** Focus | Visible focus on every interactive element; never `outline: none` without a visible replacement (2.4.7). | Keyboard walkthrough |
| **A11Y-03** Keyboard | All functionality keyboard-reachable; natural Tab order (no positive `tabindex`); Enter/Space opens, Escape closes; no keyboard traps; modal focus managed (2.1.1). | Keyboard walkthrough |
| **A11Y-04** Reflow | No horizontal scroll at 320px; text resizes to 200% without loss; tables reflow or scroll with visible indicators; content not clipped/overlapping (1.4.4, 1.4.10). | Viewport check |
| **A11Y-05** Text alternatives | `alt` on images (empty for decorative); `aria-label` on icon-only controls; icons paired with text do not duplicate that text in `alt`; captions/transcripts for media (1.1.1). | axe |
| **A11Y-06** Semantics | Semantic landmarks, logical heading hierarchy, real list/table elements (`th`, `caption`, `thead`), labels on all inputs (1.3.1). | axe |
| **A11Y-07** Navigation | Skip link, descriptive link text (not "click here"), unique page titles, consistent navigation, no unexpected context change on focus (2.4.x, 3.2.x). | Review |
| **A11Y-08** Dynamic content | `aria-live` regions for async/loading updates and status messages (4.1.x). | Screen-reader check |
| **A11Y-09** Input assistance | Errors identified in text and associated with the field; instructions provided; destructive submissions confirmed or reversible (3.3.x). | Review |
| **A11Y-10** Robust HTML | Valid HTML, no duplicate IDs, correct ARIA roles, page language declared (4.1.1-4.1.3). | HTML validator |
| **A11Y-11** Motion & time | No flashing more than 3×/second; animations pausable; session timeouts warned in advance (2.2.1, 2.3.1). | Review |

**Tools:** axe DevTools / Lighthouse / WAVE / pa11y (CI); colour contrast analyser for palette validation.

---

## Deployment — Release & Supply Chain

**Standards:** DORA, SRE, GitOps, SLSA L3 + SBOM (CycloneDX/SPDX), SOC 2

| Rule | Description | Validation |
|------|-------------|------------|
| **DEP-01** Version-tag deploy | `v*.*.*` tag drives `build-and-deploy.yml` → `deploy.sh`. No manual SCP. | Release checklist |
| **DEP-02** Provenance & SBOM | SLSA attestation + CycloneDX SBOM per release. | SLSA scanner |
| **DEP-03** Rollback | Every release has a 1-command rollback + <15min RTO. | Runbook |
| **DEP-04** Audit trail | `activitylog`-equivalent per mutation (who/when/where). | Arch-guard |
| **DEP-05** Dependency audit | `composer audit` / `npm audit` clean; lockfile committed and verified; no unconstrained (`*`) or abandoned dependencies. | Release gate |

---

## Maintenance — Ops & Privacy

**Standards:** ITIL 4, SRE (SLO/Error Budget), ISO 27001, GDPR, OpenTelemetry (logs/metrics/traces), NIST CSF

| Rule | Description | Validation |
|------|-------------|------------|
| **OPS-01** Observability | OTel logs/metrics/traces; SLO + error budget per SLI. | Pulse/dashboard |
| **OPS-02** PII | Minimize, encrypt at rest, mask in logs, GDPR retention. | `rules/global-security.md` |
| **OPS-03** Incident | Runbook + blameless postmortem (`docs/adr/` or `context/`). | Maintenance-agent |

### Logging

| Rule | Description | Validation |
|------|-------------|------------|
| **LOG-01** No sensitive data | Never log passwords, tokens, PII, payment data; log references (user/order IDs) and masked values. | Grep log calls for PII fields |
| **LOG-02** Security events | Log auth failures, access denials, privilege changes, data exports, admin actions — with actor, target, outcome. | Audit-trail review |
| **LOG-03** Structured | Consistent, parseable format (key=value / JSON); context via keys, not string concatenation. | Review |
| **LOG-04** Levels | Use the full ladder (debug → emergency); not everything is `error`; debug detail off in production. | Review |
| **LOG-05** Injection-safe | User input never concatenated raw into log lines (newline/CR forgery); log as structured data or sanitized. | Review |
| **LOG-06** Retention/rotation | Rotated (e.g. daily, ~30 days), separate security/audit channel, log directory not web-accessible. | Config review |

**Security events that MUST be logged:** successful/failed login, password change/reset, account lockout, unauthorized access attempt, admin action, data export, role/permission change.

---

## Quality Gates — Lightest by Change Type

| Change | Gate (see also `AGENTS.md §Quality Gates`) |
|--------|---------------------------------------------|
| Docs/config | Visual + `scan_broken_links.py` |
| Style | Lint (`pint`/`prettier`) |
| Single-module | `--filter` targeted test |
| Cross-module | Module suite + `scan_violations`/`scan_security` |
| New feature | Full suite + static analysis once (PHPStan level 8 equiv.) + coverage floor reported |
| Release | Full suite + SLSA + staging smoke + dependency audit (`composer audit`/`npm audit`) |

No gate is skipped; gates are **batched once** (Verify), never per-edit.

**Deliverable completeness (DoD):** a slice is done only when tests (happy path + explicit failure path rejections), localization of user-facing strings, API/route registration, authorization, cache invalidation, and docs are all complete. A slice that fails any item is not ready for review or merge — shipping it forces the next agent to inherit the defect. Pre-existing gate failures are **fixed or filed, never silently tolerated**.