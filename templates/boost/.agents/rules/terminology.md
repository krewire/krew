# Terminology — Memory, Context, Refs, Rules


## Metadata

- **ID:** `rule-terminology`
- **Scope:** `global`
- **Severity:** `advisory` (misuse causes path/permission or privacy errors)
- **Enforcement:** `self-check` (agent confirms path before action)

---

## Core Vocabulary

| Term | Meaning | Canonical Store | Example |
|------|---------|---------------|---------|
| **memory** | **Autonomous, agent-owned — full autonomy** recall through RuntimeMemoryInterface. Raw, not curated; use only operations verified by the active runtime. | RuntimeMemoryInterface | A scoped runtime memory record |
| **context** | **Curated, human-approved** shared knowledge through RuntimeContextInterface. Promoted from runtime memory after ≥2 sightings. | RuntimeContextInterface | A deploy-topology runtime topic |
| **rules** | **Durable constraint** (must/should/case). Single source of truth, fails Verify if violated. Promoted from `context` when repeated. | `rules/{name}.md` | `rules/global-security.md` |
| **refs** | Unsupported internal store. No native runtime ref interface is verified; stop and report rather than creating a local fallback. | None verified | Capability-gap report |
| **agent** | Subagent persona scoped to SDLC phase (planning/design/implementation/testing/deployment/maintenance). Lives in `agents/`. | `agents/{name}.md` | `agents/builder.md` |
| **skill** | Reusable procedural capability (how-to) an agent loads on demand. | `skills/{name}/SKILL.md` | `skills/building/SKILL.md` |
| **command** | Slash entry point (`/name`) that triggers a skill/workflow. | `commands/{name}.md` | `commands/build.md` |
| **tool** | Durable helper script/binary exposed to agents. In `tools/`: root-level `*.py` plus shared configuration. One-offs → `/tmp`. | `tools/*.py` and `tools/tools.json` | `tools/triad_check.py` |

---

## Layer Distinction — The Knowledge Layers

```
RuntimeMemoryInterface ──▶ RuntimeContextInterface ──▶ rules (durable)
   raw runtime records             curated runtime topics             promotion only on repeat
   agent writes freely            human approves                   human enforces
   runtime-owned                   runtime-owned                    tracked
```

- **Never duplicate a fact — promote upward. Memory is cheap, rules are expensive.**

| Layer | Owner | Write trigger | Read trigger | Content shape | Lifecycle | Git |
|-------|-------|---------------|--------------|---------------|-----------|-----|
| **RuntimeMemoryInterface** | Agent (fully autonomous) | Any turn where agent judges a save/delete/prune adds value — **no explicit order needed** | Every `Understand` by topic | Runtime memory shape; no local path | Runtime-owned; operations depend on verified interface | Runtime-owned |
| **RuntimeContextInterface** | Agent + human | Runtime memory signal seen ≥2x or explicit user correction | `Understand` by topic match | Runtime context shape — self-contained | Update in place; one topic = one runtime record | Runtime-owned |
| **rules/** | Human | Repeated `context` signal or ADR; reviewer sign-off | Every phase (enforce) | `rules/[scope]-[name].md` from the rule shape in `skills/customizing/SKILL.md` | Versioned, referenced from `AGENTS.md` | Tracked |
| **refs** | Unsupported | No verified native interface | Stop and report | No local fallback | No artifact created | N/A |

### When to use which

- **Memory**: "User prefers `pnpm` over `npm`" — infer from correction, write through RuntimeMemoryInterface immediately, don't ask.
- **Context**: Same preference seen twice → promote to a RuntimeContextInterface topic with rationale + evidence.
- **Rules**: Preference becomes team policy after third project → `rules/global-tooling.md` (`must use pnpm`).
- **Refs**: Lockfile spec and registry URL → stop and report unless a native runtime ref store is verified.

Never duplicate a fact — promote upward. Memory is cheap, rules are expensive.

---

## Decision Table

| User says… | Agent does… |
|------------|-------------|
| "inget ini" / "remember" / inferred preference **or agent judges useful** | **Agent MAY write autonomously without explicit order** — write/delete through RuntimeMemoryInterface whenever supported; surface "Memory captured: …" or "Memory pruned: … — reason" |
| "jadikan aturan" / "promote" | Move RuntimeMemoryInterface → RuntimeContextInterface, or RuntimeContextInterface → `rules/` with approval |

## Anti-Patterns

- **Mixing layers:** writing a durable policy directly to RuntimeMemoryInterface without ever promoting → knowledge lost; or writing raw session chatter to `rules/` → rule-bloat.
- **Guessing:** `ls` the claimed path first; if it doesn't exist, report it, don't hallucinate.
- **Overloading `tools/`:** `tools/` holds durable scanners; one-off throwaway scripts → `/tmp`.
- **PII leak:** promoting raw runtime memory with real user data — redact before RuntimeContextInterface or tracked rules.

## Sync

- `AGENTS.md` and `opencode.json` live in this directory; keep them synced to `~/.config/opencode/` where the runtime expects them for loading.
- Runtime sync: memory and context remain runtime-owned. Do not create or push local fallback files; redact PII before promoting runtime memory into tracked rules.
