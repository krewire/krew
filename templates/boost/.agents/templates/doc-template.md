# Generic Document Template

Fallback template for documents that don't fit a specific shape. Use for standalone docs, guides, SOPs, runbooks, and any artifact without a dedicated template.

## When to Use

Use this template when:
- The document is a standalone guide, SOP, or runbook
- No specific template exists for the document type
- The document is project-specific and doesn't need the formality of a rule/ref/issue

Do NOT use this for: agents, skills, commands, rules, refs, issues, memory, context — those have dedicated templates.

## Metadata

```markdown
---
title: <Document title>
type: guide | sop | runbook | checklist | report | notes
scope: <global | skill:<name> | agent:<name>>
owner: <name or team>
created: YYYY-MM-DD
updated: YYYY-MM-DD
---
```

## Shape

```markdown
# <Title>

<Brief description of what this document covers and who should read it.>

## Overview

<What this document is about in more detail.>

## Sections

### <Section A>

<Content for section A.>

### <Section B>

<Content for section B.>

## Quick Reference

<Bulleted summary of key points or commands for fast lookup.>

## Examples

```<language>
<code or example>
```

## Related

- `context/{topic}.md`
- `rules/{name}.md`
- `refs/{name}.md`
```

## Rules

- One document = one topic; don't mix concerns
- Include a "Quick Reference" or "Summary" section so the doc is scannable
- Link to related RuntimeContextInterface, `rules/`, and `refs/` files rather than duplicating content
- Update the `updated` field when making substantive changes
