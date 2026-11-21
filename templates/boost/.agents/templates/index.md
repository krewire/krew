# Templates Index — Document Shape Registry

This directory holds canonical templates for tracked artifacts in this registry. Runtime-owned
memory and context are not repository files. Internal issues and refs have no verified native
runtime store and must not be generated until that capability exists.

Generate from these templates via `/customize`.

## Registry

| Template | Path | Use for |
|----------|------|---------|
| Memory | `templates/memory-template.md` | RuntimeMemoryInterface shape; no repository file |
| Issue (internal) | `templates/issue-template.md` | Unsupported until a native runtime issue store is verified |
| Issue (GitHub) | `templates/issue-gh-template.md` | External GitHub issue — promoted from internal issue via `tracker` |
| Context | `templates/context-template.md` | RuntimeContextInterface shape; no repository file |
| Rule | `templates/rule-template.md` | Durable behavioral constraints — fails Verify if violated |
| Rule (skill-local) | `templates/rule-template.md` | Skill-specific constraints, colocated in `skills/<skill>/rules/` |
| Ref | `templates/ref-template.md` | Unsupported until a native runtime ref store is verified |
| Agent | `templates/agent-template.md` | Subagent persona — role, capabilities, workflow |
| Skill | `templates/skill-template.md` | Reusable procedural capability — loaded on demand |
| Command | `templates/command-template.md` | Slash entry point — triggers a skill or workflow |
| Generic doc | `templates/doc-template.md` | Standalone guide, SOP, runbook, checklist, report, notes |

## Layer Routing

| Document | Goes to | Git |
|----------|---------|-----|
| Memory | RuntimeMemoryInterface | Runtime-owned |
| Issue (internal) | Unsupported runtime capability | Not created |
| Issue (GitHub) | GitHub (external) | Tracked |
| Context | RuntimeContextInterface | Runtime-owned |
| Rule | `rules/` | Tracked |
| Rule (skill-local) | `skills/<skill>/rules/` | Tracked |
| Ref | Unsupported runtime capability | Not created |
| Agent | `agents/` | Tracked |
| Skill | `skills/{name}/SKILL.md` | Tracked |
| Command | `commands/` | Tracked |
| Generic doc | Project-specific | Tracked |

## Quick Reference

```
RuntimeMemoryInterface:user-profile → memory-template.md
Internal issue request              → stop and report capability gap
RuntimeContextInterface:tool-preferences → context-template.md
rules/global-security.md           → rule-template.md
Internal ref request               → stop and report capability gap
agents/builder.md                  → agent-template.md
skills/building/SKILL.md           → skill-template.md
commands/build.md                  → command-template.md
```

## Naming

- Memory: RuntimeMemoryInterface record with the memory shape (no path)
- Issue: no local naming convention; stop until a native runtime issue store is verified
- Context: RuntimeContextInterface topic with the context shape (no path)
- Rule: `rules/{scope}-{name}.md` (scope: global, agent:, skill:, phase:)
- Rule (skill-local): `skills/<skill>/rules/skill-<name>-*.md` (scope: `skill:<name>`)
- Ref: no local naming convention; stop until a native runtime ref store is verified
- Agent/Skill/Command: shared root verb (builder/building/build, planner/planning/plan)
