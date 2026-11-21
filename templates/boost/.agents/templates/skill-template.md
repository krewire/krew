# Skill Template — `skills/{name}/SKILL.md`

Reusable procedural capability. Loaded on demand by agents. Follows the mandatory `Understand → Plan → Build → Verify → Summarize` workflow. One skill per directory.

## Metadata (required frontmatter)

```markdown
---
name: <name>
description: <Use when … Triggers on <keywords>.>
---
```

## Shape

```markdown
# <Name> Skill

<One paragraph: what this skill does and when to load it.>

## When to Use

- <condition 1>
- <condition 2>

## Adaptive Modes

Per `rules/agent-adaptability.md`, adapt the workflow depth to the request class:

| Mode | Trigger | Workflow |
|------|---------|----------|
| **Minimal** | Clear class, single specialist, low risk | Full `Understand → Plan → Build → Verify → Summarize` without orchestrator routing |
| **Standard** | Clear class with cross-phase dependencies | Full workflow with appropriate delegation |
| **Escalate** | Materially ambiguous, high-risk, or conflicting | Return to `leader`; route decomposition gaps to `manager` only after leader design |
| **Stop** | Blocked/unsupported capability or scope-changing request | Report gap; do not invent fallbacks |

When ambiguity is detected during Understand, record it and route per the materiality test in `rules/agent-adaptability.md`.

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`.

Select and state the applicable workflow profile from `rules/agent-workflow-profiles.md`. Adapt
Build, Verify, and Summarize deliverables to the skill's role; do not inherit implementation-only
gates when the skill produces analysis, planning, coordination, tracking, or documentation.

### Understand

<What to gather before proceeding. Sources, context, constraints.>

### Plan

<Design approach, choices, trade-offs. Justify the approach.>

### Build

<The actual output — code, doc, decision, plan, audit result, etc. Be specific.>

### Verify

<Checklist or gates before handing off.>

### Summarize

<Handoff to next phase. Name the deliverable and its location.>

## Anti-Patterns

- <mistake 1>
- <mistake 2>

## Related Agents

List the agents that load, feed, govern, or consume this skill. Keep the list role-specific, use registered agent names, and explain the operational handoff trigger.

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `<agent-name>` | <upstream, downstream, peer, or governance relationship> | <specific condition or deliverable that starts the collaboration> |

## Handoff

- **From previous phase:** <required input>
- **To next phase:** <deliverable>
```

## Skill Variants

Add shape-specific sections after `## When to Use`:

| Variant | Additional sections |
|---------|-------------------|
| **Process-oriented** | `## Checklists` (Pre-Execution / Post-Execution) |
| **Domain knowledge** | `## Core Concepts`, `## Best Practices`, `## Common Patterns`, `## Troubleshooting Guide` |
| **Tool integration** | `## Authentication & Setup`, `## Core Operations`, `## Error Handling`, `## Rate Limiting & Performance` |

## Skill Bundle Layout

A skill is a colocated bundle — everything the agent needs lives in `skills/<name>/`:

```
skills/<name>/
├── SKILL.md      # procedure (this file)
├── rules/        # skill-specific durable rules (scope `skill:<name>`)
├── tools/        # toolbox helpers owned by this skill
├── resources/    # checklists, templates, examples owned by this skill
└── refs/         # mini-SSOT references owned by this skill
```

Global rules stay in `rules/`; only skill-specific rules colocate. The SKILL.md must list
its bundle contents (rules + tools + resources + refs) in `## Handoff` so the loader sees the
whole skillset without searching.

## Rules

- Skill name = bare root verb + `-ing` (building, planning, auditing)
- Trigger on the root verb: "plan", "audit", "design"
- Load on demand — agents load the skill when the trigger phrase appears
- Do not hard-code project names, secrets, or environment-specific values
