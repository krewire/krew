# Skill Rule — Writing & Documentation Sync

Discipline for every documentation/content change in any project: discover drift before editing,
write docs that match code and specs, keep links resolvable, and never duplicate content. Complements
`rules/global-conventions.md` (commit & style conventions) and `skills/writing/SKILL.md` (the writing
procedure).

## Metadata

- **ID:** `rule-skill-writing`
- **Scope:** `global` — governs markdown docs, project docs, content/copy, and code docblocks in any project
- **Severity:** `contextual` — binds for every doc/copy change and whenever docs must be synced to code/specs
- **Enforcement:** `manual` (review) + `automated` (link scanner where the project has one)
- **Related:** `rules/global-conventions.md`, `rules/global-standards.md`, `rules/skill-thinking.md` (drift = stale source conflict)

---

## Sync Workflow — Discover Drift Before Editing

### Intent

Before making any doc change, systematically discover what actually changed in the codebase and
specs in **at least the last 14 days** (minimum window), identify which files were added, deleted, or
modified, and map those changes to the docs that must be updated. Drift is found by inspection, not
by memory or guesswork. Widen the window (full log) whenever the change surface may predate it
(large refactor, dormant module, cross-module rename).

### Rationale

Docs go stale because agents update them "when they remember" — the reliable signal is git history:
it records exactly which code changed, so it tells you exactly which docs are now wrong.

1. **Updating the wrong docs** — guessing wastes effort and leaves the real drift in place.
2. **Re-verifying unchanged docs** — re-reading and "sync"-touching docs that have not drifted burns
   time, churns history, and hides future drift.

The 14-day window is the **floor, not a cap** — extend to full log whenever drift may predate it.

### Step 0 — Commit Uncommitted Changes Atomically First

```bash
git status --short
git diff --stat
git diff --cached --stat
```

If anything is modified (`M`), added (`A`), deleted (`D`), renamed (`R`), or untracked (`??`),
**commit it first** with an atomic, per-concern commit. Uncommitted code is invisible to `git log` —
syncing docs against `HEAD` while the tree has uncommitted business logic leaves the new docs stale
on arrival. Do not use `git add -A` nor `stash`; commit each logical change explicitly.

### Step 1 — Review Recent Git History

```bash
git log --since="14 days ago" --stat              # summary per commit (minimum window)
git log --since="14 days ago" --name-status       # consolidated file changes
git log --since="14 days ago" --format="%h %s"    # commit messages for context
# For suspected older drift: git log --stat
```

- Note which modules, layers, and files were affected.
- Skip commits that already updated their docs (`--stat` shows it).
- Focus on commits that introduced new code **without** doc updates.

### Step 2/3 — Identify What Changed and Which Docs Need Updates

Map changed artifacts to the docs that describe them:

| If you changed...         | Update these docs                                                     |
| ------------------------- | --------------------------------------------------------------------- |
| Module structure/products | Module conceptual + reference docs (file listing, actions/models/schema) |
| Business rules / features | Feature spec (requirements, user stories, data contracts)             |
| Architecture patterns     | Architecture guide / pattern docs                                     |
| Conventions               | Conventions doc                                                        |
| Database / schema         | Database/infra docs + spec data-model sections                        |
| Notable decisions         | Architecture decision records (ADRs) as applicable                    |
| Config                    | Configuration/infra docs                                              |
| Agent docs/skills         | `AGENTS.md`, agent skill docs, agent context, agent memory            |

This mapping **includes the agent layer** — a spec change (renamed default, new invariant, changed
path) must be mirrored in any guide or skill that documents it, not just the project docs.

### Ordered Sync Execution

Execute sync by impact-to-effort: dependency chains first (specs before refs that cite them), then
business urgency (core root-level before reference tiers), then quick wins. Commit per phase as a
checkpoint.

---

## Link Integrity — Resolvable References, No Duplicated Content

### Intent

Every cross-reference must resolve to a real target, and no content may be duplicated across docs.
When a doc needs a fact that lives elsewhere, it links to it — it never copies it.

### Rationale

1. **Broken links** — a reference to a renamed or deleted file, or an anchor that no longer matches a
   heading, sends the reader to a dead end and silently signals the docs are unmaintained.
2. **Duplicated content** — copies drift; one gets fixed, the other doesn't, producing two competing
   "truths". The fix is always the same: keep the fact in **one** authoritative location and
   cross-reference it with a relative link.

### How to Apply

- **Relative links**: all internal links are relative paths from the current file:
  `[Pattern](refs/arch/pattern-file.md)`.
- **Anchor links**: match the exact heading (lowercased, spaces → hyphens):
  `[Token Security](setup.md#token-security)` for `## Token Security`.
- **Before any doc commit, verify**: every `[text](path)` resolves; every `[text](path#anchor)`
  matches an existing heading; no content is duplicated — cross-reference instead.
- **Use the `edit` tool (not `write`) when updating docs** — rewrite only the changed sections to
  minimize risk of accidentally deleting content or breaking formatting. Never whole-file rewrite
  "while we're here".

### Anti-Patterns & Pitfalls

- **Copy-paste instead of link** — copies diverge on the next change.
- **Absolute paths** — resolve only from the repo root and break when docs move.
- **Stale anchors** — heading renamed but the anchor link not updated.
- **Deleting/renaming a file mid-sync** without fixing docs that reference it.
- **Bare URL links to internal files** — relative paths survive renames and forks.

---

## Sync Verification — The Final Quality Gate

### Intent

After docs are updated to match code and specs, run the final gate: verify every link resolves,
honor the sync key rules, and pass the verification checklist. This proves the sync actually
happened and is complete.

### Key Rules

1. Documentation is the SSOT — if code disagrees with docs, fix code (or fix docs if behavior changed intentionally).
2. Do NOT duplicate content — cross-reference with relative paths.
3. Conceptual docs contain **no implementation detail** (no paths, schemas); reference docs contain **no design rationale**.
4. **Always use the `edit` tool (not `write`)** for doc updates — preserve untouched content.

### Verification Checklist

- [ ] New/changed modules have their docs (conceptual + reference) updated; indexes list them
- [ ] Feature specs match implementation (requirements, data contracts)
- [ ] Agent docs/skills/context/memory match specs and code
- [ ] File paths, class names, and method signatures in docs verified against the actual codebase
- [ ] No broken relative links or stale anchors
- [ ] No stale content (removed features, renamed classes, changed signatures)
- [ ] No duplicated content — single authoritative location per fact

### Verification / Detection

- Link scanner (where the project has one, e.g. `tools/scan_broken_links.py`) for broken file/anchor
  links; freshness via `git log --follow -- <file>`.
- `git status` + `git diff` — exactly the intended docs changed, nothing dropped.
- Grep for tier violations: conceptual docs containing implementation detail; reference docs
  containing rationale.

---

## Quick References

- `rules/global-conventions.md` — commit atomicity, code style, metadata
- `rules/global-language.md` — English-only stored artifacts
- `skills/writing/SKILL.md` — writing procedure (doc-quality gates)
- `rules/skill-thinking.md` — stale-source conflict resolution (drift adjudication)