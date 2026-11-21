# Self-Improvement (Continuous Learning) Rules


Rules for scope: `global`. Every agent and skill MUST participate in the loop below. This is the agent's "deep learning": experience → extracted pattern → behavior change.

## Metadata

- **ID**: `rule-self-improvement`
- **Scope**: `global`
- **Severity**: `mandatory`
- **Enforcement**: `self-check` (agent at session boundaries) + `manual` (retro via `learner`)
- **Related Refs**: RuntimeMemoryInterface, `skills/customizing/SKILL.md` (artifact shapes), `rules/terminology.md`, `skills/learning/SKILL.md`, `commands/learn.md`

## The Learning Loop (3 stages, closed — with memory)

```
RuntimeMemoryInterface storage (raw, autonomous; verified runtime interface) ──▶ RuntimeContextInterface (curated, human-approved) ──▶ rules/skills (durable)
        ▲                                 │
        └──────── CAPTURE ◀── CONSOLIDATE ◀── APPLY ──────────────┘
```

1. **CAPTURE** — selective, not automatic-flush. Write via `RuntimeMemoryInterface` **only** for high-value signals: `decision` (with tradeoffs), `correction`, `failure` (with root cause), `pattern` (reusable), `constraint`, `gap`. **Do NOT capture:** routine successful completions, trivial facts, things already in docs, or anything already captured in the same session. Before writing, prune stale or superseded runtime entries when the interface supports it. Agent may delete/prune stale or privacy-sensitive entries without explicit user command. `RuntimeContextInterface` and `rules/` are not written automatically; curated promotion requires human approval.
2. **CONSOLIDATE** — periodically (`learner`, weekly or `/learn --deep`, or opportunistically when agent notices bloat) distill repeated `RuntimeMemoryInterface` signals into `RuntimeContextInterface`, and repeated context signals into tracked `rules/` or workflow assets. Also prune redundant runtime memory without asking. Promotion to context or rules requires explicit human approval.
3. **APPLY** — at session start (Understand), load matching records from `RuntimeMemoryInterface` and `RuntimeContextInterface` plus relevant `rules/*.md`; adapt behavior so the same mistake never repeats.

Compound learning: each session makes the next better. "Deep" = signals are abstracted into patterns and pushed into the agent's own instructions (rules/skills), not just logged.

## Signal Taxonomy (what to capture)

| Signal | Trigger | Example (Personal) | Example (Coding) |
|--------|---------|-------------------|------------------|
| `decision` | Chose approach with tradeoffs | "User prefers concise bullets over prose — adopted for daily summaries" | "Used Events not imports for cross-module side effect — ADR-003" |
| `correction` | User pushed back / self-caught | "Don't ask before writing to memory — do it autonomously" | "Never run full suite per-edit; batch verify once" |
| `failure` | Bug/regression + root cause | "Missed weekly review because memory not loaded at Understand" | "N+1 in dashboard; forgot `->with()` — add to perf rule" |
| `pattern` | Reusable solution that worked | "Daily 08:00 triage: inbox → top3 → calendar block" | "Surgical diff + `git diff` proof before/after each edit" |
| `constraint` | Env/tooling/deploy fact (intentional) | "User timezone Asia/Jakarta, work hours 09:00-17:00 WIB" | "prettier-plugin-blade must run via Pint, not standalone" |
| `gap` | Missing doc/test/rule | "No fact file for diet preference — add via RuntimeMemoryInterface (user-health)" | "No rule for cache-key registry — add rules/cache-keys.md" |
| `preference` | Inferred user like/dislike (Personal) | "Likes Indonesian for casual, English for coding" | "Prefers pnpm, vertical slices <300 LOC" |

All signals go via `RuntimeMemoryInterface` first; only explicit or repeated signals go to `RuntimeContextInterface`.

**Unsupported stores:** The active runtime exposes no verified native interface for internal
`issues` or `refs`. Do not create local substitutes or invent names. If a workflow needs either
store, stop and report the capability gap; use a GitHub issue only for explicitly team-visible
external tracking. `rules/` remains the repository-local durable store.

## Where to Write (single source of truth)

| Destination | When | Template | Registry |
|-------------|------|----------|----------|
| **Memory storage** via RuntimeMemoryInterface | When any user/project/preference fact, correction, failure, or pattern arises | memory shape in `skills/customizing/SKILL.md` | RuntimeMemoryInterface |
| **Curated topic** via RuntimeContextInterface | Signal seen ≥2x or explicit human promotion | context shape in `skills/customizing/SKILL.md` | RuntimeContextInterface |
| **Workflow asset** `agents/{name}.md`, `commands/{name}.md`, `skills/{name}/SKILL.md` | Signal shows an agent role, command step, or skill procedure must change | triad shapes in `skills/customizing/SKILL.md` | `AGENTS.md §Available Agents` |
| **Durable constraint** `rules/[scope]-[name].md` | `context` signal repeated across projects | rule shape in `skills/customizing/SKILL.md` | `AGENTS.md §Rules Index` |
| **Reusable capability** `skills/[name]/SKILL.md` | Repeated procedural pattern | skill shape in `skills/customizing/SKILL.md` | `AGENTS.md §Available Agents` |
| **Reference** | Stable API/model/decision | No verified runtime store; stop and report | — |

Never paste a learning into chat-only summary and drop it. If it helps a future session, it goes via `RuntimeMemoryInterface` (and possibly `context/`).

## Consolidation Cadence

- **Every Summarize (selective capture)**: agent appends via `RuntimeMemoryInterface` ONLY for high-value signals (`decision`/`correction`/`failure`/`pattern`/`constraint`/`gap`). Before writing, prune stale entries (>14 days, no references, superseded); log `[gap] pruned … — reason`. One line per signal with timestamp + tag + source.
- **Every Summarize (curate)**: if signal is explicit or high-value, also update/create RuntimeContextInterface topic + RuntimeMemoryInterface storage if user fact. Also delete any memory the agent now judges unnecessary.
- **Repetition ≥ 2 (promote)**: `memory` → `context/`; `context` → `rules/`/`skills/` (requires human approval for rules). A repeated workflow correction tunes the owning `agents/`/`commands/`/`skills/` triad (surgical edit, `triad_check.py` clean). One-offs stay in `memory`/`context` to avoid rule-bloat.
- **Weekly / `/learn --deep` (hygiene — also opportunistic)**: `learner` (or any agent) queries the two verified runtime interfaces, deduplicates, redacts PII, and prunes stale or low-value entries when supported.
- **Understand phase**: before planning, query matching runtime memory/context records and cite their runtime source in the plan.

## Dedup & Hygiene

- **Runtime memory mutability**: use only operations confirmed by RuntimeMemoryInterface; prune stale or privacy-sensitive records without explicit permission when supported.
- **Context updates**: update the existing RuntimeContextInterface topic; never create duplicate topics.
- **No invented data**: fact-check commands and versions against the codebase and verified runtime documentation.
- **Self-contained**: each runtime record includes the rationale and source it needs.
- **Privacy**: redact PII/secrets before promoting RuntimeMemoryInterface data to RuntimeContextInterface or tracked rules.

## Validation (how to check the loop ran)

- [ ] RuntimeMemoryInterface updated ONLY for high-value signals this session; no routine completions or trivial facts captured.
- [ ] Stale runtime entries were pruned before writing when the interface supports pruning.
- [ ] Workflow edits surgical and `tools/triad_check.py --root . --strict` clean (if a triad was tuned).
- [ ] RuntimeContextInterface contains any approved promoted topic.
- [ ] Promoted rule is referenced from `AGENTS.md` or skill `## Skill Rules` table.
- [ ] Next session's Understand queries runtime memory/context and honors the correction (observed, not asserted).
- [ ] No duplicate runtime topics exist.
- [ ] Runtime stores remain runtime-owned; no local fallback was created.

## Anti-Patterns

- **Logging and forgetting**: writing `memory/` but never loading it at next Understand.
- **Hoarding**: keeping 500-line `notes.md` instead of promoting repeated signals to `rules/`/`skills/`.
- **Hallucinated learning**: recording a fix/path that doesn't exist or wasn't applied.
- **Over-promoting**: turning a one-off into a mandatory `rules/` — keep rare events in `memory`/`context`.
- **Asking permission for memory**: "Should I remember this?" — no; write autonomously, surface "Memory captured: …" after.
- **PII promotion**: copying raw memory with names/emails/secrets to `context/` without redaction.
