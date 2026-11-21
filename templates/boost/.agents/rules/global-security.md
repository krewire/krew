# Global Security Baseline — OWASP, CWE, Secrets, Privacy


Applies to **every** repo. Fails the Build/Verify gates if violated. Stack-specific supplements (e.g. a framework security scanner, Dependabot rules) extend, not replace, this baseline.

## Metadata

- **ID:** `rule-global-security`
- **Scope:** `global`
- **Severity:** `mandatory`
- **Enforcement:** `automated` (SAST/scan) + `manual` (review)
- **Related:** `rules/global-standards.md` §Implementation/Deployment, `rules/global-conventions.md`

---

## OWASP Top 10 (2021) + ASVS L1 — Minimum Bar

| # | Risk | Must-Do (ASVS L1) | Check |
|---|------|-------------------|-------|
| **A01** | Broken Access Control | Authorize **every** layer (route → policy → action). Default deny; tests cover `allow` + `deny` per role. | SAST auth check, policy tests |
| **A02** | Cryptographic Failures | Secrets in env only; app/DB creds never in repo; TLS 1.2+; encrypt PII at rest. | `gitleaks`, `.env` gitignore |
| **A03** | Injection (SQL/XSS) | No raw SQL without bindings; no unescaped user content; CSP header. | SAST injection scan |
| **A04** | Insecure Design | STRIDE + ASVS threat model before design; security requirements are FR/NFR, not afterthought. | Design checklist |
| **A05** | Security Misconfiguration | Hardening checklist (health/config audit), least-privilege, remove demo/default creds before prod. | Setup audit |
| **A07** | Identification & Auth Failures | Throttle/rate-limit auth, lockout after N attempts, hashed passwords (argon2), 2FA where required. | SAST + auth tests |
| **A08** | Software & Data Integrity | SLSA L2+ provenance + CycloneDX SBOM per release; pin deps; verify signatures. | SLSA scan |
| **A09** | Logging & Monitoring Failures | Structured audit log: `who/when/where/what` per mutation; alert on 5xx / auth anomaly; retain per policy. | Arch-guard + CSIRT |
| **A10** | SSRF | Allow-list outbound fetch; no user-controlled URL without validation. | Code review |

A06 (Vulnerable Components) → ### Dependency Management below (Dependabot/`npm audit`/`composer audit` + CVSS triage). A05-A07 are verified in Verify.

## CWE Top 25 Awareness (Representative)

- **CWE-79** XSS, **CWE-89** SQLi, **CWE-306** Missing Auth, **CWE-287** Improper Auth, **CWE-798** Hardcoded Creds, **CWE-862** Missing Authorization. Patterns blocked by the OWASP rules above.
- Also covered by the checklists below: **CWE-20** Input Validation, **CWE-327** Broken/Risky Crypto, **CWE-338** Predictable PRNG, **CWE-352** CSRF, **CWE-384** Session Fixation, **CWE-614** Insecure Cookie Flags, **CWE-915** Mass Assignment.

## Control Checklists

### Authentication & Session Management

| Rule | Validation |
|------|------------|
| Rate-limit login, password reset, and registration; lock out after N failed attempts; throttle by user/IP when anonymous. | Throttle/rate-limit config; auth tests |
| Password policy: minimum length 8+ with complexity (mixed case, numbers, symbols); prefer long passphrases; reset tokens single-use, short-lived, invalidated on new request; success/error messages never reveal whether an account exists. | Password rules; reset-flow tests |
| Hash credentials with a strong adaptive algorithm — never plaintext (see Cryptography for algorithm rules). | SAST scan; code review |
| Regenerate the session ID after **any** auth state change (login, privilege change, logout); invalidate session + CSRF token on logout; prevent session fixation. | Session tests |
| Session cookies: `HttpOnly`, `Secure` in production, `SameSite=Lax/Strict`; idle + absolute timeout; domain restricted to the app. | Session config review |
| Session storage driver appropriate for deployment (shared/multi-server needs a shared store, not local files) with garbage collection configured; enforce a max concurrent-sessions policy or invalidate older sessions on new login when single-session is required. | Session config review |
| MFA for sensitive operations where required; recovery codes provided; no bypass path. | Auth tests |
| Authorize every mutation at each layer (route → policy → action); default deny; policy per resource with correct context-aware logic. | Policy tests: `allow` + `deny` per role |
| Prevent IDOR: ownership checks via policy or queries scoped to the authenticated owner; never trust a client-supplied ID alone. | Policy/query review; tests |
| RBAC via the framework role system; no hardcoded role strings in controllers; role hierarchy documented and changeable without code edits. | Code review |
| Middleware order: authenticate first, then authorize. | Route review |

### Input Validation

| Rule | Validation |
|------|------------|
| All user input is untrusted until validated **server-side** — client-side validation is UX, not security. | Code review; tests |
| Whitelist known-good, reject everything else; validate type, length, format, and range — not just presence; reject invalid input early, before processing. | Validation-rule review |
| Typed fields: integers (avoid string comparison traps), booleans (reject truthy `"false"`), dates (format + ordering), arrays (validate every element). | Tests with malformed/mixed types |
| Email with RFC/DNS checks; URLs with scheme allow-list; no raw input passed to pagination/query params without `(int)`/validation. | Validation-rule review |
| Rich-text/HTML input is sanitized on the way in (HTMLPurifier or sanitized markdown) — never trust the client's markup. | SAST scan; review |
| File uploads: MIME + size + sanitized filename — see Mass Assignment & File Uploads below. | See that section |

### XSS & Injection

| Rule | Validation |
|------|------------|
| Escape all dynamic output by default; every unescaped output carries an inline justification naming the sanitizer — no justification, no unescaped output. | SAST regex on unescaped output |
| Never bind raw user input into client-side HTML (framework equivalent of `x-html`/`innerHTML`); no inline `<script>` tags — use the framework's data bindings. | SAST scan; review |
| CSP: `default-src 'self'`; scope `script-src`/`style-src`/`img-src`; no wildcard origins, no `unsafe-inline`/`unsafe-eval`; never disable CSP (not even "during development"). | Header test; review |
| All queries parameterized or built through the query builder; raw SQL fragments only with bound parameters; no string-concatenated values into queries. | SAST injection scan |
| Output escaping (XSS) and query bindings (SQLi) are distinct defenses — one never substitutes for the other. | Code review |

### Mass Assignment & File Uploads

| Rule | Validation |
|------|------------|
| Every model declares an explicit fillable allow-list (framework attribute or default-deny `$guarded`); never `$fillable = ['*']`. | Arch-guard scan |
| No raw request array (`$request->all()` / framework state `->all()`) reaches `create()`/`update()`; writes take validated, explicitly listed keys or a DTO. | SAST scan: `create($request->all())` regex |
| Extra untrusted params (e.g. `role=admin`) must never be persisted. | Test with forged request body |
| Uploads validated server-side by **MIME/content**, not just the client extension; per-collection/per-type size limits; filenames sanitized (strip path separators, control chars) or replaced with random names; stored outside the web root / in private storage. | Upload tests; config review |
| User uploads go through the framework's managed media/storage layer with validation rules — not raw filesystem writes on user-controlled paths. | SAST scan |

### Cryptography

| Rule | Validation |
|------|------------|
| Never roll your own crypto — use established libraries or the framework's crypto primitives. | Code review |
| Strong algorithms only: bcrypt/argon2id for passwords; AES-256 (GCM) for encryption; no MD5/SHA1 in security contexts; no ECB mode; no `base64_encode` masquerading as encryption. | SAST scan; review |
| Cryptographically secure randomness for tokens/OTPs/nonces (`random_bytes` or framework equivalent); never `rand()`/`mt_rand()`/`uniqid(md5(...))` for secrets. | SAST scan |
| Keys/secrets via env or secret manager, never hardcoded; the framework's encryption key ≥ 32 bytes, generated by the framework tool, unique per environment. | `gitleaks`; env review |
| Encrypt sensitive data at rest (PII, tokens) with the framework's encryption layer — see PII & Privacy. | Code review; tests |

### API Security

| Rule | Validation |
|------|------------|
| Rate-limit API routes (per-user, or per-IP when anonymous); auth endpoints additionally per the Authentication checklist. | Rate-limiter config |
| CSRF token on every state-changing request; exemptions rare, require an inline justification, and stay narrow (e.g. a verified webhook). | SAST scan; review |
| CORS: allow-list origins; never `*` together with credentials; restrict methods/headers. | CORS config review |
| Security headers on all responses: CSP, `X-Content-Type-Options: nosniff`, `X-Frame-Options`/`frame-ancestors`, `Strict-Transport-Security`, `Referrer-Policy`, `Permissions-Policy`. | Header test |
| Correct HTTP methods: POST/PUT/PATCH/DELETE for mutations — never GET for state change. | Route review |
| Error responses minimal: no stack traces, query strings, or SQL leaked to the client. | Code review |
| Open redirects: allow-list redirect targets; never redirect to raw user input. | Code review |

### Dependency Management

| Rule | Validation |
|------|------------|
| SCA on every change/PR: `composer audit` / `npm audit` / Dependabot. Triage by CVSS: ≥ 9.0 fix immediately, 7.0–8.9 within 1 week, 4.0–6.9 next update cycle, < 4.0 accept/defer. | CI gate green |
| Flag abandoned packages (no maintainer, no recent release, unresolved advisories); record replacement evaluation. | Review |
| Pin dependencies: caret (`^`) ranges for most packages, exact pins for critical infrastructure; never unbounded (`*` or no upper bound). | Manifest review |
| License review: permissive (MIT/BSD/Apache-2.0) low risk; LGPL medium; GPL/AGPL high (copyleft) — record usage and compatibility. | Review |
| Automated dependency scanning configured and enabled (e.g. `dependabot.yml` present); suppressing an audit finding requires a justification. | CI file check |

## Secrets & Config

- Never commit `.env`, `*.key`, `credentials`, `auth.json`. `.gitignore` must cover `/vendor`, `/node_modules`, `.env*`, `/storage/*.key`, `*.sqlite`.
- Env via 12-Factor: app URL, app key, `DB_*` from env — never hardcoded; database creds live in `.env` only, never defaulted in `config/*.php`.
- A `.env` previously committed is a finding even after removal — rotate everything it contained; check `git ls-files | rg '^\.env'` returns nothing.
- Framework app key unique per installation; sharing a key across environments enables cross-install decryption of cookies/encrypted data. Generate a fresh key per deployment.
- No secrets in example files, seeders, or config defaults "for convenience" (`.env.example` may contain placeholder values only).
- Recovery keys printed once, stored file-scoped with `storage/*.key` gitignored.
- Secret scan hook before commit (pre-commit `gitleaks`/detect-secrets).

## PII & Privacy (GDPR + ISO 27001)

- Minimize: collect only what a requirement (FR) requires; mask PII in logs/traces; never log raw request payloads.
- PII inventory: record PII fields at design time (spec/ADR) with a masking plan; no new PII column without one.
- Separate credentials from profile data: the auth table holds what *authenticates*, not full identity — a credentials leak must not become a full-PII leak.
- Masked logging: every log/audit write that could carry PII passes through the masking layer; audit/activity logs never capture raw PII verbatim.
- Retain: define retention per entity; purge on request. GDPR erasure must leave a traceable record (deletion log trail) and cover backups.
- Field-level access control: gate sensitive columns in the policy/domain layer, pass a boolean to the view; exports masked by default unless explicitly authorized.
- Transit/rest: TLS + encrypted at-rest for sensitive fields (framework encryption casts) — see Cryptography.
- Docs: `docs/` never contains real user PII in examples.

## Verification

- **SAST:** framework security scanner + arch-guard scan (injection/unescaped output/mass-assignment regexes), linter/formatter.
- **SCA:** `composer audit` / `npm audit` / Dependabot with CVSS triage per Dependency Management.
- **Secrets:** `gitleaks detect --no-git` (or detect-secrets) — clean on the full tree, including git history check for `.env`.
- **Dast/manual:** auth/authorization matrix test per route and action (allow + deny per role); session/cookie flag checks; security-header check.

Any blocker from the above fails Verify — fix or file an issue with `CWE-` / `OWASP-A##` label.