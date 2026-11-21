# Skill Rule — Thinking & Conflict Resolution

Durable reasoning discipline for any agent and any project: how to break down work, and how to
resolve disagreement between governing sources without guessing. Complements `rules/global-standards.md`
(precedence the higher-tier specs operate within) and `skills/thinking/SKILL.md` (the read-only
reasoning procedure).

## Metadata

- **ID:** `rule-skill-thinking`
- **Scope:** `global` — governs decision-making and source-conflict resolution in any project
- **Severity:** `contextual` — binds whenever sources disagree or a plan must be decomposed
- **Enforcement:** `manual` (review) + `automated` (arch-guard scans where the project has them)
- **Related:** `rules/global-conventions.md` (edit/surgical discipline), `rules/self-improvement.md`

---

## Computational Thinking

Four pillars applied to every instruction to stay autonomous, anticipate the next step, and avoid
blind execution.

| Pillar | How the agent applies it |
|--------|--------------------------|
| **Decomposition** | Break the instruction into smaller sub-problems (files, layers, concerns). Solve each independently, then integrate. Never try to hold the whole problem at once. |
| **Pattern recognition** | Classify the instruction (bug? feature? refactor? docs? audit?) and reuse known patterns: relevant skills, existing code, docs, past conventions. A known pattern is a solved problem. |
| **Abstraction** | Filter out irrelevant detail (versions, formatting, noise) and focus on the essential structure — entities, flows, contracts, invariants. See the forest before the trees. |
| **Algorithm design** | Plan ordered steps with clear inputs/outputs. Before acting, ask: what is the expected outcome, what could go wrong, what does the next step depend on? |

**Decision loop** — before each action, run: *predict outcome → act → verify → adjust*. After every
step, anticipate the next: what must follow, what can break, what to verify. When the instruction is
ambiguous, resolve it yourself when the cost is low (look up the answer in code or docs); escalate to
the user only when the decision changes scope or architecture.

---

## Conflict Resolution — Precedence Hierarchy

Defines the absolute precedence when governing sources disagree. This rule is the tie-breaker for
requirements vs architecture-pattern conflicts.

### Precedence (highest → lowest)

1. **Architecture patterns / invariants — HIGHEST, NON-NEGOTIABLE**
   - The projects' arch patterns (where they exist) are the ultimate SSOT for structural invariants.
   - Covers: module boundaries, Entity/DTO/Model/Action/presentation contracts, non-negotiable
     conventions.
   - Encodes global industry standards; protects security, modularity, scalability across the project.
   - **No spec, code, test, doc, or user instruction may violate an arch pattern.**

2. **Feature specs — SECOND**
   - SSOT for feature behavior (requirement IDs) *within* arch constraints.
   - If a spec requirement contradicts an arch pattern, **the spec is wrong and must be amended** —
     arch pattern wins.

3. **Code — LOWEST**
   - Must align to both arch patterns and specs. If code disagrees with either, fix code.

### SSoT Priority Framework

When **any** two sources conflict — not just arch vs spec — resolve by priority, but never trust
blindly: `adr > specs > guides > code > refs` (higher wins). Even the highest source must be verified
via git history (`git log --follow`, `git blame`) and intent before acting. If history is silent or
contradictory, treat it as a finding and justify the decision comprehensively.

### Operational Rules

- **Load the arch pattern before a spec change** that touches an arch concern (purity of Entities,
  thin presentation, Action triad, DTO shape, cache keys): read the governing pattern's
  non-negotiable section first.
- **Spec violates arch?** Amend spec *immediately* with a recorded decision (commit with
  `type(scope): desc` citing the arch pattern as higher authority), then align code/tests. Do not
  implement violating code.
- **Code violates arch/spec?** Fix code. Never amend an arch pattern to accommodate code.
- **Tests:** Must trace to spec requirement IDs, but the spec itself must already be arch-compliant.
  No orphan tests that would require an arch violation.

### Why Arch > Spec

Arch patterns protect system-wide invariants across a large codebase. A single feature spec cannot
override them without systemic risk (security, modularity, testability).

---

## Verification / Detection

- When sources conflict, the transcript shows which side was found stale and why (git history, not
  assumption).
- No spec amendment landed without a recorded decision citing the higher authority.

## Quick References

- `rules/global-standards.md` — SDLC-standard precedence layer
- `skills/thinking/SKILL.md` — read-only structured reasoning procedure
- `rules/self-improvement.md` — capturing decisions as durable learning