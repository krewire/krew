# Global Conventions — Naming, Commits, Style, Docs


Applies to **every** repo that uses this agent registry.

## Metadata

- **ID:** `rule-global-conventions`
- **Scope:** `global`
- **Severity:** `mandatory`
- **Enforcement:** `automated` (lint/CI where possible) + `manual` (review)
- **Related:** `rules/global-standards.md` §Implementation, `rules/global-security.md`

---

## Commits & Versioning

> **NOTICE — Atomic commit & push (mandatory):** every change in this repo **MUST be committed and pushed immediately** after it is verified. **FORBIDDEN to batch multiple concerns into one large commit** (`one big commit`). One commit = one concern, one `type(scope)`, ≤300 LOC, push right after `git commit`. Never leave verified work uncommitted at session end.

- **Conventional Commits** — `type(scope): desc` where `type = feat|fix|docs|chore|test|refactor|perf|security`, `scope = module/area` (kebab-case). One concern per commit. Example: `feat(auth): add throttled login (FR-AUTH-03)`.
- **SemVer** — breaking API/contract change = major bump + migration guide.
- **Atomic + immediate push** — commit at every verified milestone and `git push` right after. No single big commit: split `memory` + `context` + `rules` changes into separate commits (see AGENTS.md §Conventions). CI checks commit granularity via `git log --stat`.
- **No secrets** — fail build on `gitleaks`/secret scan. Use env vars.

## Code Style & Quality

- **SOLID, DRY, Clean Code** — single-responsibility, no copy-paste; extract when 3+ repetitions or a concept has a name; deduplicate and align by default, even when the instruction did not ask (see below).
- **Strictness** — `declare(strict_types=1)`-equivalent everywhere (PHP strict types, TS strict, Python type hints). No untyped public APIs.
- **No debug artefacts** — `dd`/`dump`/`console.log`/`print_r` forbidden in committed code. CI fails on lint-equivalent checks.
- **Formatting** — project formatter is law (`pint`/`prettier`/`black`). Run before Verify; CI enforces.
- **No service locator** — constructor injection only (no `app()->make` / locator).

### Clean Code, DRY, Dedup-Align

- **DRY is the default bias** — extract repeated logic into shared, named units (helper, trait, function, module, doc cross-reference). Never introduce a second copy of something that already exists; reuse or extract.
- **More, smaller, well-named modules** over one dense blob — modularization is the prescribed direction, not over-engineering.
- **Align the outlier, don't tolerate drift** — when code, docs, tests, or requirements disagree, align the outlier to the SSOT instead of leaving the contradiction standing.
- **Surface structural decisions only** — extraction, dedup, refactor, or re-scoping that changes scope, structure, or behavior is stated briefly and confirmed when it does; routine dedup runs silently (narration discipline).
- **Record decisions** — any dedup/alignment that affects a spec or an invariant is recorded (ADR or spec amendment), never left implicit.
- **Validation:** `git diff` review shows only intended consolidation; no new duplication introduced by the change.

### File Header

- **Ordering** — for source files with a declaration header (PHP/TS/Java-style): strict-type/attribute declarations first, `namespace` matching the directory path, `use`/`import` statements one per line sorted alphabetically, then the type (class/interface) declaration.
- **Dependency declaration** — inject dependencies via promoted, read-only constructor parameters so the signature documents the dependency graph at a glance; avoid non-promoted assignments that hide injected dependencies.
- **Command entry point** — command-style classes expose exactly one public entry method (`execute`/`run`); the rest stays private/protected.
- **Why it matters** — every file parses the same way: strictness is enforced before anything is declared, autoloading matches namespace→path, imports stay greppable and diff-stable.
- **Pitfalls** — `use`/`namespace` placed above the strict-type line; imports grouped by category instead of alphabetically; header fields that duplicate git metadata (dates, authors, versions are forbidden — see Documentation).
- **Validation:** formatter (`pint` or equivalent) enforces header style; review of each new file.

## Workflow

### Impact-to-Effort Ordering

All work — batched instructions, multi-stage plans, backlog picks, issue triage — is **ordered before execution**, never taken verbatim. Three passes run in fixed order.

1. **Dependency chains (topological)** — if B needs A, A executes first even when B scores higher (a high-ratio item behind an unmet prerequisite is not actually available). When candidates are equal, take the one others depend on; never start blocked work "anyway" unless reordering the chain costs more than waiting.
2. **Priority bands** — band every candidate before comparing ratios; urgency never outranks importance alone:

   | Band | Importance | Urgency | Handling |
   |------|-----------|---------|----------|
   | **P0 — Critical** | Core flow broken / data loss / security | Now | Supersedes ratio; prerequisites still apply unless waiting increases harm |
   | **Important-first** | Mission-relevant | Low–medium | Scheduled deliberately after P0s and quick wins |
   | **Urgent-only** | Low mission impact | Deadline-driven | Batch tightly; challenge whether the deadline is real |
   | **Neither** | Cosmetic / hygiene | None | Fill-ins, deferrals, or filed issues |

   Tie-break inside a band: higher importance beats higher urgency.

3. **Impact-to-effort ratio** — within the same band, sort by ratio: quick wins first, heavy lifts scheduled, low-value work batched or challenged. Use the canonical [`rules/agent-ordering.md`](agent-ordering.md) scoring scale — do not re-invent scales locally.

**Applies to** message batches of 2+ instructions (full sort before executing anything), multi-stage plans / big-session splits (stage order follows the passes; each stage ends in a checkpoint commit), backlog or issue triage (P0 floats to top; questionable items challenged or filed low), and mid-task discovery of extra work (same passes decide *now vs file-as-issue*).

**Surfacing** — score silently; surface the resulting order only when it differs from the given order, when a dependency forces resequencing, or when an item is deferred/challenged.

**Validation:** execution order is deterministic per the passes; a one-line reorder note reaches the user whenever their sequence changes.

### Edit Policy

- **Read before edit** — read the full current content of every file you may touch before changing anything.
- **Edit, don't rewrite** — change only what the instruction requires; preserve unrelated code, comments, formatting, and context. A full rewrite is justified only for small files where the rewrite IS the intent.
- **Verify with git** — compare `git diff` before/after each change to prove nothing unintended was altered or dropped; `git status` shows only intended files.
- **Scope smallest** — keep the change surface minimal; fewer touched files = fewer places for errors to hide.
- **Validation:** `git diff` review after every edit (lossless-edit proof); no silent reordering or deletion of unrelated lines.

## Naming

| Artifact | Convention | Example |
|----------|------------|---------|
| Files / dirs | `kebab-case` for docs/scripts, `PascalCase` for `app/Module/Submodule/` | `scan_naming.py`, `app/Auth/Services/` |
| Classes | `PascalCase`, one class per file | `SendWelcomeEmail` |
| Methods / vars | `camelCase` (typed), booleans `is/has/can` | `canActivate()` |
| Tests | `ClassNameTest` or `FR-XXX - description` | `FR-AUTH-03: throttles after 5 attempts` |
| Commits/branches | `type(scope): desc` / `feat/kebab` | `fix(billing): clamp slot overflow` |

## Documentation

- **History via git (mandatory):** no `> **Last updated:**` banner — track changes via `git log`, `git diff`, and `git blame`. Commit messages carry context (`type(scope): desc`). The same rule applies to docblocks: no authorship/version tags (see `### PHPDoc / TSDoc Conventions`).
- **Language (mandatory):** all stored artifacts are **English-only** per `rules/global-language.md`; direct chat follows user's language.
- **Single source of truth:** never duplicate a fact — cross-reference (`[label](path#anchor)`). Every `[label](path)` must resolve (`scan_broken_links.py`). A requirement duplicated across sources is a defect (dedup/align doctrine — see `### Clean Code, DRY, Dedup-Align`).
- **Tiering:** human-facing docs in `docs/` (SSOT), agent knowledge in RuntimeMemoryInterface/RuntimeContextInterface, and durable rules in `rules/`. Direction: RuntimeMemoryInterface → RuntimeContextInterface → `rules/` (promote), never `rules/` → runtime context.

### PHPDoc / TSDoc Conventions

- **Supplement, never duplicate** — docblocks carry what the signature cannot: business meaning, side effects, failure modes. Never restate native type hints (`@param string $name` beside `string $name` is dead weight that rots and misleads).
- **Forbidden tags** — `@author`, `@version`, `@created`, `@package` are never allowed; authorship and versioning belong to git and these tags go stale immediately.
- **`@throws` lists concrete exception types** — never generic `Exception`; name the failure modes callers must handle. `@see` links related handlers for non-obvious side effects.
- **When to write** — complex algorithms (multi-line approach description), business-rule methods (one-line description of the business question), and non-obvious side effects. Simple getters/property access need no docblock.
- **Type crutch = bug** — annotating `@return ?User` instead of typing `?User`; fix the signature, don't document it.
- **Validation:** formatter enforces docblock style; grep for forbidden tags (`@author|@version|@created|@package`) and for `@param`/`@return` lines that merely restate the signature.

### Project Layout

```
./
  agents/ skills/ rules/ context/ refs/ tools/ issues/ memory/ AGENTS.md
```

## Exceptions

Stack-specific rules may be added (e.g. language/framework-specific invariant sets) but may not relax a `mandatory` global rule without an ADR (`docs/adr/NNN-*.md`) and reviewer sign-off.
