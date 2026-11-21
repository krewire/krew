# Runtime Context Template — `RuntimeContextInterface`

Curated, must-read-before-task facts. Promoted from RuntimeMemoryInterface after ≥2 sightings or
explicit user correction. Runtime-owned. One topic = one record. Update in place; never duplicate.

## Shape

```markdown
# <Topic> Runtime Context

## Description

[One paragraph: what a future agent must know. Cite the originating RuntimeMemoryInterface record
and any human-reviewed rationale.]

## AI Agent Notes

| Situation | Decision |
|-----------|----------|
| {trigger} | {action to take} |

## Changes

- YYYY-MM-DD: <who> — <what changed and why>
```

## Rules

- File = one topic only; update in place, bump `## Changes`
- Never duplicate a fact — if it belongs elsewhere, link to it
- Store through RuntimeContextInterface; do not create a local fallback or index file.
- Skip trivial/fluid facts — record only where a future agent would make a costly wrong assumption without this

## When to Create

- Same signal appears in RuntimeMemoryInterface ≥2×
- User explicitly corrects or confirms a decision: "always do X"
- A project-specific constraint discovered during work that the team must respect

## When NOT to Create

- One-off observations unlikely to recur
- Trivial preferences ("user prefers dark mode") — keep in RuntimeMemoryInterface
- Facts that belong in `rules/` (durable, team-enforced) or an internal ref store (unsupported until verified)
