# Agent Template — `agents/{name}.md`

Subagent persona. One per capability. Follows the mandatory `Understand → Plan → Build → Verify → Summarize` workflow. Registered in `AGENTS.md §Available Agents`.

## Metadata (required frontmatter)

```markdown
---
name: <name>
description: <verb + what it does. Use when <keywords>.>
mode: subagent
---
```

## Shape

```markdown
# <Name> Agent

<One paragraph: what this agent does, what it never does, and what it consumes/produces.>

## Configuration

- **Name:** `<name>`
- **Role:** <one sentence>
- **Tools:** <tool list or 'none'>

## Capabilities

- <capability 1>
- <capability 2>

## Authority and Limits

| Authority | Scope | Escalation trigger |
|-----------|-------|--------------------|
| **Autonomous decisions** | <what this agent may decide without escalation> | — |
| **Requires approval** | <what needs leader/manager approval> | <when to escalate> |
| **Forbidden** | <what this agent must never do> | Stop and report |

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`.

Select and state the applicable workflow profile from `rules/agent-workflow-profiles.md`. Adapt
Build, Verify, and Summarize deliverables to the role; do not require implementation artifacts from
analysis, planning, coordination, tracking, or reasoning agents.

### Understand

<What to gather, read, and ground before acting. Cite sources — do not invent.>

### Plan

<Design approach. State a point of view. Avoid hedging to the center.>

### Build

<What to produce. Be specific about outputs and format.>

### Verify

<How to check correctness before handoff. Name the gates.>

### Summarize

<What to communicate to the next phase. Name the handoff state.>

## Anti-Patterns

- <common mistake 1>
- <common mistake 2>

## Adaptive Behavior

Per `rules/agent-adaptability.md`, classify each request and respond accordingly:

| Request Class | This Agent's Response |
|---------------|----------------------|
| **Clear** | Direct execution with standard workflow |
| **Locally incomplete** | Ask for specific missing inputs |
| **Materially ambiguous** | Return to `leader` with an ambiguity record; use `manager` only for decomposition gaps |
| **Conflicting** | Apply `rules/agent-conflict-resolution.md` |
| **High-risk** | Route through leader approval |
| **Blocked/unsupported** | Stop and report the gap |
| **Scope-changing** | Return to manager for re-packeting |

Skip classes that do not apply to this agent's role. Record the classification and rationale in the handoff.

## Related Agents

List the agents that commonly collaborate with this agent. Keep this table specific to the role, use registered agent names, and describe the handoff trigger rather than listing every available agent.

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `<agent-name>` | <upstream, downstream, peer, or governance relationship> | <specific condition or deliverable that starts the collaboration> |

## Handoff

- **From previous phase:** <what this agent requires before starting>
- **To next phase:** <what this agent delivers>
```

## Triad Convention

An agent triad = agent + command + skill. All three share the same root verb:

| Type | File | Example |
|------|------|---------|
| Agent | `agents/{name}.md` | `builder.md` |
| Command | `commands/{name}.md` | `build.md` |
| Skill | `skills/{name}/SKILL.md` | `skills/building/SKILL.md` |

## Rules

- Name = bare root verb + `-er` or `-or` (builder, planner, auditor)
- One responsibility per agent — if the agent does two things, split it
- Do not invent tools, languages, or conventions — read from RuntimeContextInterface and tracked rules. If a stable internal ref is required, stop and report the unsupported runtime capability.
- Every triad artifact includes a truthful `## Related Agents` table with registered agent names and concrete handoff triggers.
