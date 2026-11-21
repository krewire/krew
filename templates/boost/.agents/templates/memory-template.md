# Runtime Memory Template — `RuntimeMemoryInterface`

Runtime-owned and agent-owned. One focused runtime record per topic. No PII, no secrets. This is a
shape, not a repository file path.

## Shape

```markdown
# <Scope> — <Short description>

> **Source:** inferred from chat + explicit statements
> **Privacy:** no secrets, no raw PII

## <Section>

- Key: value (provenance: date + source)

## Signals Log

Append — one line per signal, with provenance.

- YYYY-MM-DDTHH:mmZ [decision|correction|failure|pattern|constraint|gap|preference|observation] fact — source
```

## Rules

- Store through `RuntimeMemoryInterface`; do not create a local fallback.
- Every line carries an `ISO8601 [tag] fact — source`
- Do not rewrite history that hides a correction — append a new `[correction]` line
- Keep each runtime record focused (for example: user profile, workflow preference, project fact).

## Scopes

| Scope | Use for |
|-------|---------|
| `user` | User profile, preferences, language |
| `preference` | Workflow/tool preferences |
| `project` | Project-specific facts, decisions, context |
| `tooling` | Environment, tooling, pin, quirk |
| `work` | Current session notes, active tasks |
| `hygiene` | Registry maintenance, stale entries |

## Variants

- **Runtime signal log** — append to the existing RuntimeMemoryInterface record or create a new scoped record.
- **Tombstone** — when supported by the runtime, record `[gap] pruned <reason> — YYYY-MM-DD`.
