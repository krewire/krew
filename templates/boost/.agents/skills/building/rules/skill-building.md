# Skill Rule — Defense-in-Depth Guards

Production-destructive capabilities (seeding demo data, enabling debug, exposing admin surfaces)
are protected by **layered guards**, not a single check. When a guard looks like an unnecessary
check, it is usually the visible layer of a deliberate defense-in-depth; weakening it breaks a
recorded product decision. This rule defines how to recognize, preserve, and safely operate around
such guards.

## Metadata

- **ID:** `rule-skill-building`
- **Scope:** `global` — governs building, extending, or touching guarded capabilities
- **Severity:** `contextual` — binds while working on a capability protected by one or more guards
- **Enforcement:** `manual` (review) + `automated` (tests asserting guard behavior)
- **Related:** `rules/skill-building-refactoring.md`, `rules/skill-auditing.md`, `rules/global-security.md`, `rules/skill-planning-feature-spec.md` (spec amendment before guard change)

---

## Recognize the Guard Layers

**What it enforces:** before "simplifying"/"fixing" what looks like a redundant production check,
identify whether it is one layer of a deliberate defense-in-depth stack.

**Why it matters:** demo data is deliberately blocked in production at **multiple independent
layers** (seeder guard + action guard + CLI guard). Removing any one layer based on "this is
redundant, the other check already handles it" silently weakens the product's protection against
accidentally polluting a live database, and violates the recorded spec/ADR decision.

**How to apply:** when touching a guarded capability, enumerate the guard layers (in the project
context file, e.g. `production-dummy-guard.md`). Treat the whole stack as intentional. A change to
any layer **requires a spec amendment first** — reference the spec requirement ID (e.g. `NFR-S1`,
`NFR-S13`, defense-in-depth `DD-10`) in the change.

**Pitfalls to avoid:**

- Weakening a single guard because "the other one covers it".
- "Simplifying" a guard to reduce code — the layer count is the product decision, not redundancy.
- Relaxing a production check to run a one-off operation (see next section).

**Verification:** every guard layer is still present and asserted in tests; no guard was altered
without a recorded spec amendment.

---

## Operate Around Guards — Ephemeral One-Off Pattern

**What it enforces:** when a legitimately guarded operation must run on a live/staging instance
(e.g. demo seeding), the guard is **not** relaxed; instead a throwaway container satisfies it.

**Why it matters:** production-quality instances need populated data for demos (never for real
schools), while the running stack and its guards must stay untouched.

**How to apply:**

1. Run a **one-off container** from the deployed image, attached to the compose network.
2. Override the env **for that container only** (e.g. `APP_ENV=local`) to satisfy the guard condition
   — without changing the deployment.
3. Run the operation with `--rm`; container-layer changes (vendor, autoloader) are discarded;
   **only the DB rows persist** (the operation wraps work in a transaction).
4. Verify counts afterward; confirm no running container changed.

**Pitfalls to avoid:**

- Seed/operate directly on the running stack — the running container keeps its guard env.
- Committing ops tooling that bypasses the guard into the repo (repo spec forbids it).

**Verification:** after the one-off completes, no service container was restarted or re-deployed; the
guarded capability's tests still pass.

---

## Where to Record

- Guard definition & layers → project context file (`context/{project}-{guard}.md`) with spec/ADR
  references and the "safe procedure" section.
- Spec amendment required for any guard change → `rules/skill-planning-feature-spec.md` (requirements-first).
