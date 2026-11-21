# Runtime Reference Template — unsupported runtime capability

This is a documentation shape only. The active runtime exposes no verified native internal ref
store. Do not create a `refs/` file or invent an interface. Stable facts may remain in tracked
`rules/` when they are rules, or be promoted through RuntimeContextInterface when they are curated
context. Stop and report any request that specifically requires an internal ref store.

## Metadata (required frontmatter)

```markdown
---
id: ref-<name>
scope: global | agent:<name> | skill:<name> | project:<name>
owner: <person or team>
related: [rules, agents, skills]
---
```

## Shape

```markdown
# <Name> — <What this ref documents>

[One paragraph: what this reference covers and why it is authoritative.]

## <Section>

<content>
```

## Ref Variants

Choose the variant that fits the content type:

### API shape

```markdown
# <Name> API — <What this API does>

## Base URL
`https://<host>/<version>/`

## Authentication
<method>: <bearer token / API key / OAuth 2.0>

## Endpoints

| Method | Path | Params | Response | Errors |
|--------|------|--------|----------|--------|
| GET | /resource | query params | 200 JSON | 400, 401, 404 |

## Request Schema

```json
{ "field": "type" }
```

## Response Schema

```json
{ "field": "type" }
```
```

### Data Model shape

```markdown
# <Name> Data Model — <What this model describes>

## Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | uuid | yes | Primary key |

## Relations

<ERD or description>

## Validation Rules

<field constraints, invariants>
```

### Decision shape

```markdown
# <Name> Decision — <What this decides>

## Context

[Why this decision is being made. Constraints, options, and who decides.]

## Options

| Option | Pros | Cons | When to use |
|--------|------|------|-------------|
| A | … | … | … |

## Decision

<Chosen option and rationale>

## Consequences

<What changes because of this decision>
```

### Pattern shape

```markdown
# <Name> Pattern — <What this pattern solves>

## When to Use

<condition>

## Structure

```
{{placeholder}}
```

## Example

```<language>
<code>
```

## Variants

<alternative forms>
```

### Configuration shape

```markdown
# <Name> Config — <What this config governs>

## File
`<path/to/config>`

## Options

| Key | Type | Default | Env | Description |
|-----|------|--------|-----|-------------|
| FOO | string | bar | FOO_BAR | Does X |

## Example

```yaml
key: value
```
```

## Rules

- Do not use local `refs/` paths or invent an unverified runtime ref interface.
- Stable content requires a verified runtime destination; otherwise stop and report.
- Link from tracked `rules/` or RuntimeContextInterface rather than duplicating facts.
