---
name: customizing
description: Use when creating or standardizing any agent asset. Triggers on customize, customize agent, scaffold, new agent, new skill, new command, triad, rename, standardization, new memory, new issue, new context, new rule, new ref, shape, convention.
---

# Customizing Skill

Single source of **agent-asset shapes and conventions** for this registry. Covers agents, commands, skills (and their variants), memory,
context, rules, refs, plus the **triad naming convention** and registration/wiring steps.

## When to Use

- Creating a new subagent and its companion command + skill
- Renaming/standardizing existing agents, commands, or skills to the triad naming convention
- Creating a new `memory/`, `issues/`, `context/`, `rules/`, or `refs/` artifact
- Registering an agent/skill in the `AGENTS.md` assets table
- Wiring a new skill path into `opencode.json` / `kilo.json` when needed

## Triad Naming Convention (Mandatory)

Every role is a triad with the **same root**:

| Type | Pattern | Example (customizing) |
|------|---------|------------------------|
| Agent | `-er` suffix | `customizer` |
| Command | bare verb | `customize` |
| Skill | active verb (`-ing`) | `customizing` |

Rules:
- **Never** reuse a name across agent/command/skill.
- Keep command and skill names matching their agent's root exactly.
- English spelling changes on suffixing are irregular; normalize bases so each matches its command verb (e.g. `planner`→`plan`, `automator`→`automate`, `customizer`→`customize`) — maintained in `_NORM` in `tools/triad_check.py`.

## Artifact Shapes

> Anti-slop discipline everywhere: specific, verifiable language; no buzzwords; no invented data; fact-check every stat and link; unique examples.

Every artifact type has a canonical template in `templates/`. Use `templates/index.md` for the full registry.

| Artifact | Template | Path |
|----------|---------|------|
| Memory note | `templates/memory-template.md` | RuntimeMemoryInterface record |
| Context topic | `templates/context-template.md` | RuntimeContextInterface topic |
| Rule | `templates/rule-template.md` | `rules/{scope}-{name}.md` |
| Ref | `templates/ref-template.md` | Unsupported until a native runtime ref store is verified |
| Agent | `templates/agent-template.md` | `agents/{name}.md` |
| Skill | `templates/skill-template.md` | `skills/{name}/SKILL.md` |
| Command | `templates/command-template.md` | `commands/{name}.md` |

### Memory — RuntimeMemoryInterface record

Full template: `templates/memory-template.md`. Rules:
- Store through RuntimeMemoryInterface; never create a local fallback.
- Every line carries an `ISO8601 [tag] fact — source`
- Do not rewrite history that hides a correction — append a new `[correction]` line
- Keep each file focused (e.g. `user-profile`, `preference-workflow`, `project-facts`, `hygiene-notes`)

### Context — RuntimeContextInterface topic

Full template: `templates/context-template.md`. Curated, must-read-before-task facts. Shape: `# <Topic> Runtime Context` + `## Description` + `## AI Agent Notes` decision table. Never duplicate a fact; update the existing runtime topic in place. Skip trivial/fluid facts.

### Rule — `rules/{scope}-{name}.md` or `skills/<skill>/rules/{name}.md`

Full template: `templates/rule-template.md`. Shared behavioral rules. Metadata: **ID** (`rule-[name]`), **Scope** (`global` | `agent:[name]` | `skill:[name]` | `phase:[Understand|Plan|Build|Verify|Summarize]`), **Severity** (`mandatory` | `advisory` | `contextual`), **Enforcement** (`automated` | `manual` | `self-check`), **Related Refs**. Each rule: Description / Trigger / Action / Validation + **Example — Correct** / **Example — Incorrect** / **Exception**.

Global rules live in `rules/`; skill-specific rules live colocated in the owning skill bundle at
`skills/<skill>/rules/`. The SKILL.md must list its colocated rules in `## Handoff`.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `automator` | Collaborates on automation and tooling | Route work when the deliverable requires automation and tooling. |
## Handoff

#### Rule Variants

| Variant | File pattern | Scope field |
|---------|-------------|------------|
| Global | `rules/global-*.md` | `global` |
| Agent | `rules/agent-*.md` | `agent:<name>` |
| Skill | skills/customizing/rules/skill-*.md | `skill:<name>` |
| Skill-local | `skills/<skill>/rules/skill-<name>-*.md` | `skill:<name>` |
| Phase | `rules/phase-*.md` | `phase:<Verify>` |
| Execution | `rules/exec-*.md` | Language-specific: Bash, Python, JS/TS, Go |

### Ref — unsupported native runtime store

Full template: `templates/ref-template.md`. No verified native runtime ref interface exists. Do not create a local ref file or invent an unverified interface; stop and report when an internal ref is required.

#### Ref Variants

| Variant | File pattern | Key sections |
|---------|-------------|--------------|
| API | API shape | Base URL, Authentication, Endpoints, Request/Response schemas |
| Data Model | Data model shape | Schema table, Relations, Validation Rules |
| Decision | Decision shape | Context, Options, Decision + rationale, Consequences |
| Pattern | Pattern shape | When to Use, Structure, Example, Variants |
| Configuration | Configuration shape | File, Options table, Example |

### Agent — `agents/{name}.md`

Full template: `templates/agent-template.md`. Required frontmatter: `name:`, `description:`, `mode: subagent`. Sections: `## Configuration` (Name, Role, Tools) + `## Capabilities` + `## Instructions` (Understand → Plan → Build → Verify → Summarize as H3) + `## Related Agents` + `## Anti-Patterns` + `## Handoff`. The Related Agents table must use registered agent names and concrete handoff triggers.

### Command — `commands/{name}.md`

Full template: `templates/command-template.md`. Required frontmatter: `name:`, `description:`. Sections: `Scope: $ARGUMENTS` + `If $ARGUMENTS is empty, ask …` guard + five phases as H3 + `## Related Agents` + `## Parameters` + `## Handoff`. The Related Agents table must use registered agent names and concrete invocation or handoff triggers.

### Skill — `skills/{name}/SKILL.md`

Full template: `templates/skill-template.md`. Required frontmatter: `name:`, `description:`. Sections: `## When to Use` + `## Instructions` (Understand → Plan → Build → Verify → Summarize as H3) + `## Related Agents` + `## Anti-Patterns` + `## Handoff`. The Related Agents table must use registered agent names and concrete operational handoff triggers. A skill is a colocated bundle: `SKILL.md` + `rules/` (skill-specific rules) + `tools/` + `resources/` + `refs/` in the same directory; list all bundle contents in `## Handoff`.

#### Skill Variants

| Variant | Additional sections |
|---------|-------------------|
| Process-oriented | `## Checklists` (Pre-Execution / Post-Execution) |
| Domain knowledge | `## Core Concepts`, `## Best Practices`, `## Common Patterns`, `## Troubleshooting Guide` |
| Tool integration | `## Authentication & Setup`, `## Core Operations`, `## Error Handling`, `## Rate Limiting & Performance`

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Confirm the root word; check for collisions across `agents/`, `commands/`, `skills/`.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Pick the root verb; derive agent `root-er`, command `root`, skill `root-ing`. For irregular/awkward forms choose a root that reads naturally (e.g. `learn` → `learner`/`learn`/`learning`); select the artifact shape above; reuse an existing matching asset as a richer pattern.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update a linked GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- Create the files from the shape (in this registry's directories — `agents/`, `commands/`, `skills/{name}/SKILL.md`, or the store file). In commands, include the `Scope: $ARGUMENTS` line and a 5-step workflow; in skills, `name:`, `description:`, `## When to Use`, `## Instructions`.
- In every agent, command, and skill, include a `## Related Agents` table with at least three relevant registered agents, a relationship, and a concrete handoff trigger. Keep the three tables in a triad synchronized.
- Keep triad pairs, role descriptions, related-agent mappings, and validation thresholds in `tools/tools.json`; tools must load this shared data instead of duplicating it in Python.
- Register the triad in the `AGENTS.md` assets table (Agent / Skill / Command columns). Add the skill path to `opencode.json`/`kilo.json` `skills.paths` only if it lives outside the configured `./skills` root.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Confirm no name collision (grep across all `agents/`, `commands/`, `skills/`); run `tools/triad_check.py` → zero findings.
- Confirm the `AGENTS.md` table row points to existing `SKILL.md` and command files. Check that restarting opencode picks up new agents/commands/skills.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close resolved `issues/` and GitHub Issues; update `issues/` doc with outcomes and follow-ups.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; `refs/` updated with evidence; next phase has explicit acceptance criteria and owners.

- Report the asset created, files touched, registration done, and any opencode restart required.

## Adaptive Modes

| Request Class | Customizing Response |
|---------------|---------------------|
| **Clear** | Create or rename the asset using triad conventions |
| **Materially ambiguous** | Clarify the asset type (agent/command/skill/rule/context/memory) and root name |
| **High-risk** | Verify no name collision, run triad_check, confirm AGENTS.md registration |
| **Blocked/unsupported** | Report the missing capability; do not invent unsupported artifact types |

## Anti-Patterns

- Splitting one concept across mismatched names (agent `footer`, skill `bar`); leaving an agent without its command or skill; forgetting to register in `AGENTS.md`; inventing structure instead of loading it from this skill; creating a memory/context file for a trivial/fluid fact a future agent would never need.

## Handoff

A complete, conflict-free, registered asset (triad, rule, guide, ref, context topic, or memory note) plus a short list of any restarts or follow-ups required.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (1 item).
