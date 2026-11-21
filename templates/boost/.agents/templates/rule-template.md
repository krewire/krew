# Rule Template — `rules/{scope}-{name}.md` or `skills/<skill>/rules/{name}.md`

Durable behavioral constraint. Global rules live in `rules/`; skill-specific rules live colocated
in the owning skill bundle at `skills/<skill>/rules/`. Fails Verify if violated. Promoted from
RuntimeContextInterface after repeated signal or explicit human decision.

## Metadata (required)

```markdown
---
id: rule-<name>
scope: global | agent:<name> | skill:<name> | phase:<Understand|Plan|Build|Verify|Summarize>
severity: mandatory | advisory | contextual
enforcement: automated | manual | self-check
related: [rules, agents, skills, refs]
---
```

Global rules use the YAML frontmatter shape above. Skill-local rules may use the compatible
in-body `## Metadata` block when they predate frontmatter adoption; the scanner accepts either
shape but requires all five fields: ID, Scope, Severity, Enforcement, and Related.

## Shape

```markdown
# <Name> — <What this rule governs>

## Description

[One paragraph: what this rule is, why it exists, and what behavior it enforces. Cite the context/{topic}.md source that promoted this.]

## Trigger

[When does this rule apply? Name the condition, action, or phase.]

## Action

[What must be done — or must NOT be done. Be prescriptive. One action per rule.]

## Validation

[How to verify compliance. Automated (tool/command), manual (checklist), or self-check (what the agent confirms before proceeding).]

## Example — Correct

```<language or format>
<code or command that satisfies the rule>
```

## Example — Incorrect

```<language or format>
<code or command that violates the rule — <brief explanation of why>
```

## Exception

[Any documented, intentional state that differs from the rule. Note the source and owner. Example: "Pre-existing intentional skip of X — tracked in context/{topic}.md"]
```

## Rule Variants

| Variant | File pattern | Scope field |
|---------|-------------|------------|
| Global | `rules/global-<name>.md` | `global` |
| Agent | `rules/agent-<name>.md` | `agent:<name>` |
| Skill | `rules/skill-<name>.md` | `skill:<name>` |
| Skill-local | `skills/<skill>/rules/skill-<name>-*.md` | `skill:<name>` |
| Phase | `rules/phase-<name>.md` | `phase:<Verify>` |
| Execution | `rules/exec-<name>.md` | Language-specific: Bash, Python, JS/TS, Go |

## Rules

- One rule per file; one action per rule
- Severity `mandatory` fails Verify; `advisory` warns; `contextual` only applies in specific contexts
- Register the rule in `AGENTS.md §Rules Index`
- Global rules go in `rules/`; skill-specific rules go colocated in `skills/<skill>/rules/`
- Never write rules for one-off facts — those belong in RuntimeMemoryInterface or RuntimeContextInterface
