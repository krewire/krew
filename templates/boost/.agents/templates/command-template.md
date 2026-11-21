# Command Template — `commands/{name}.md`

Slash entry point. Triggers a skill or a multi-step workflow. One per agent/skill. Follows the mandatory `Understand → Plan → Build → Verify → Summarize` workflow.

## Metadata (required frontmatter)

```markdown
---
name: <name>
description: <verb + what it does. Use when <keywords>.>
---
```

## Shape

```markdown
# <Name> Command — <purpose>

<One paragraph: what this command does. Follow `skills/<name>/SKILL.md` and the
mandatory workflow `Understand → Plan → Build → Verify → Summarize`.>

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask <clarifying question>. Do not invent scope.

### Understand

<What to discover before acting.>

### Plan

<Design approach or approach selection.>

### Build

<The output — list, plan, artifact, report, etc.>

### Verify

<How to check the output is complete and correct.>

### Summarize

<What to deliver and how to communicate it.>

## Anti-Patterns

- <common mistake 1>
- <common mistake 2>

## Related Agents

Document the agents that this command normally routes to or receives work from. Use registered agent names and concrete invocation or handoff triggers.

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `<agent-name>` | <upstream, downstream, peer, or governance relationship> | <specific condition or deliverable that starts the collaboration> |

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{param}}` | string | No | default | <description> |
| `--flag` | flag | No | false | <description> |

## Handoff

<What the next phase or user receives.>
```

## Rules

- Command name = bare root verb (plan, build, audit, explore)
- Include `Scope: $ARGUMENTS` line
- Include `If $ARGUMENTS is empty, ask …` guard
- Five phases as H3 — `### Understand`, `### Plan`, `### Build`, `### Verify`, `### Summarize`
- Reference the matching skill with `Follow skills/<name>/SKILL.md`
