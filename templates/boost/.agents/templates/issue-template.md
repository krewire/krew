# Internal Issue Template — unsupported runtime capability

This is a documentation shape only. The active runtime exposes no verified native internal issue
store. Do not create an `issues/` file or invent an interface. Stop and report the capability gap.
For team-visible work, use the explicitly supported external GitHub issue tools.

## Naming

```
issues/{type}-{seq}-{scope}-{short-description}.md
```

| Segment | Format | Example |
|---------|--------|---------|
| `type` | `plan` \| `audit` \| `task` \| `fix` \| `docs` \| `refactor` | `plan` |
| `seq` | Zero-padded 3-digit integer | `001` |
| `scope` | `design` \| `security` \| `quality` \| `workflow` \| `config` | `design` |
| `short-description` | kebab-case, max 5 words | `auth-redesign` |

Example: `issues/plan-001-design-auth-redesign.md`

## Metadata (required frontmatter)

```markdown
---
id: I-<type>-<seq>         # e.g., I-plan-001
type: plan | audit | task | fix | docs | refactor
scope: design | security | quality | workflow | config
status: draft | in_review | approved | superseded | archived
owner: <name or 'agent'>
created: YYYY-MM-DD
updated: YYYY-MM-DD
related-issues: [I-plan-002, I-audit-003]
related-context: [context/auth.md]
---
```

## Shape

```markdown
# Issue {id} — {Title}

> **Type:** {type} | **Scope:** {scope} | **Status:** {status}
> **Owner:** {owner} | **Created:** {YYYY-MM-DD}

## Problem or Objective

[What is the issue? Why does it exist? What triggers it?]

## Description

[Detailed background, context, and intent. What the issue is trying to solve or decide.]

## Decision or Resolution

[For plans: the chosen approach and rationale. For audits: findings summary and recommendation.]

## Status History

| Date | Status | By | Note |
|------|--------|----|------|
| YYYY-MM-DD | draft | owner | Opened |

## Handoff

- **From:** <who/when triggered this>
- **To:** <next owner or next phase>
```

## Layer Routing

| Scope | Internal (gitignored) | External (tracked) |
|-------|----------------------|-------------------|
| Personal work, drafts, exploratory | `issues/` | — |
| Team-visible decisions | — | GitHub issue (`gh issue create`) |
| Approved cross-team patterns | RuntimeContextInterface | — |

## Routing Rule (blocked)

- **Internal** → stop and report: no verified runtime issue store exists.
- **Project-level** → use GitHub issue tools (team-visible, external only).
- **Approved cross-team** → promote through RuntimeContextInterface after team acknowledgement.
- **Audits that matter to the team** → use GitHub issue tools; do not link to a local issue file.

## ID Allocation

No local sequence is allocated. Do not create an index or local issue artifact.

## Rules

- Do not create `issues/` or an internal issue index.
- Never invent an unverified runtime issue interface.
- If a workflow requires an internal issue, stop and report the missing runtime capability.
