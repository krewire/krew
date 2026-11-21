---
name: learning
description: Use when capturing, consolidating, or applying experience into agent workflow, rules, context, and memory across sessions. Triggers on learn, memory, self-improvement, capture, consolidate, apply, promote, workflow tuning, retrospective, learning loop.
---

# Learning Skill

Operationalizes the Learning Loop (`rules/self-improvement.md`) so the agent can capture, consolidate,
and apply experience across sessions across **four target stores**: workflow (`agents/`/`commands/`/`skills/`),
rules (`rules/`), RuntimeContextInterface, and RuntimeMemoryInterface. This is the agent's deep-learning mechanism:
it turns one task's outcome into durable, retrievable knowledge that changes future behavior — including
the very workflow that produces it.

## When to Use

- **End of every session** (in Summarize): capture decisions, corrections, failures, patterns.
- **On user correction**: the moment a user pushes back or redirects — record it immediately.
- **On failure/regression**: after root-causing a bug or a failed check.
- **On discovering a non-obvious constraint**: env quirk, tooling pin, deploy fact, deliberate decision.
- **On repeated signal**: a RuntimeContextInterface topic seen ≥2 times → promote to `rules/` or `skills/`.
- **On workflow drift**: an agent role, command step, or skill procedure that no longer matches how work actually happens → tune it.

## Store Routing — Which Store Gets the Signal

| Signal looks like… | Route to | Action |
|--------------------|----------|--------|
| "Agent X's Verify step omits check Y" | **Workflow** | Tune `agents/x.md` / `commands/x.md` / `skills/x/SKILL.md`; update `AGENTS.md` row |
| "Do this/never do that" behavioral rule repeated across projects | **Rules** | Write `rules/[scope]-[name].md`; reference from `AGENTS.md` |
| Project-specific fact/constraint, signal ≥2x | **Context** | Create/update a RuntimeContextInterface topic |
| Raw fact, preference, one-off, transient | **Memory** | Write/prune through RuntimeMemoryInterface |

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize` with anti-slop discipline.
This skill runs mostly inside **Summarize** and **Understand**.

### 1. CAPTURE — selective memory writes (not every session)

Decide the signal type (decision / correction / failure / pattern / constraint / gap). Agent has
autonomy through RuntimeMemoryInterface — may save/delete/prune without an explicit request when
the verified interface supports the operation.

- **Issue tracking:** Use GitHub issue tools only for team-visible findings. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Runtime memory entries written ONLY for high-value signals; stale entries pruned when supported.
- **Measurable outcome:** Memory grows only when signal is explicit or repeated; no routine successful completions logged; no trivial facts; every entry has a timestamp + tag + source.
- **Anti-pattern:** Capturing routine successful completions, trivial facts, or things already in docs. Memory must only record high-value signals: `decision` (with tradeoffs), `correction`, `failure` (with root cause), `pattern` (reusable), `constraint`, `gap`.

1. **Selective write to RuntimeMemoryInterface first (no permission needed, but WRITE SPARINGLY)**:
   - Append `YYYY-MM-DDTHH:mmZ [tag] fact — source` to a focused runtime record.
   - **Before writing, prune first when supported:** remove stale (>14 days, no references, superseded) entries.
2. **To RuntimeContextInterface when curated** (signal ≥1 if explicit/high-value, else ≥2):
   - Existing topic → update in place and record the change.
   - New topic → create a RuntimeContextInterface topic from the context shape in `skills/customizing/SKILL.md`.
3. **Write self-contained**: include paths, commands, rationale — file is read alone, no "see above".
4. **Fact-check**: verify path/command/version exists; never invent data.

```markdown
# {Topic} Context
## Description
[One paragraph: what a future agent must know — cite memory/[scope]-[desc] source.]
## AI Agent Notes
| Situation | Decision |
|-----------|----------|
| {trigger} | {action to take} |
```

### 2. CONSOLIDATE — promote, tune, prune (memory → context → durable, plus workflow)

When a RuntimeMemoryInterface signal appears a second time, a workflow asset drifts, or agent notices bloat/stale/PII — all without explicit order:

- **Issue tracking:** Use GitHub issue tools for team-visible consolidation decisions. If an internal issue record is required, stop and report the unsupported runtime capability.
- **Deliverable:** Promoted context/rule/skill, pruned memory, updated indexes.
- **Measurable outcome:** Signal seen ≥2 times before promotion; memory file count does not grow unbounded; stale entries removed, not just appended.
- **Anti-pattern:** Promoting one-offs into mandatory `rules/`; creating duplicate topics in RuntimeContextInterface; rewriting a triad instead of tuning it surgically (that is `customizer`'s job).

- **→ Workflow**: a repeated correction about how work should run → tune the owning `agents/{name}.md`, `commands/{name}.md`, or `skills/{name}/SKILL.md` minimally; verify the triad still passes `triad_check.py`; update `AGENTS.md`.
- **→ Context**: same signal ≥2x → draft/update a RuntimeContextInterface topic.
- **→ Rule**: behavioral constraint repeated across projects → write `rules/[scope}-{name}.md` (rule shape in `skills/customizing/SKILL.md`) and reference from `AGENTS.md`.
- **→ Skill step**: reusable procedure → append to `skills/[name]/SKILL.md` `## Instructions`.
- **→ Ref**: stable knowledge → stop and report; no verified native runtime ref store exists.
- **→ Prune**: stale/redundant/PII memory → delete through RuntimeMemoryInterface when supported.

Keep RuntimeContextInterface for one-offs and rare events; promote only what repeats. Delete runtime memory freely when it no longer serves, when supported. Keep workflow edits surgical — tune what exists, don't rewrite a triad (that is `customizer`).

### 3. APPLY — load at session start

In **Understand**, before planning (no explicit order — always):

- **Issue tracking:** Read a linked GitHub Issue when present. If prior learning requires an internal issue record, stop and report the unsupported runtime capability.
- **Deliverable:** Loaded memory + context + rules; documented corrections honored.
- **Measurable outcome:** No repeated mistake from prior session; every loaded file is actually relevant to the task; stale corrections are pruned before loading.
- **Anti-pattern:** Loading every runtime memory record blindly without matching by topic; carrying stale corrections forward without pruning; skipping RuntimeContextInterface and `rules/` that are relevant to the task.

1. Query RuntimeMemoryInterface for records matching the task topic; note preferences, constraints, prior corrections.
2. Query RuntimeContextInterface for the topic matching the task.
3. Load any `rules/` entry scoped to task (global / agent / skill / phase).
4. Honor recorded corrections from `memory` + `context`: if past says "do NOT do X", skip X without re-learning.
5. If the task touches an agent's own workflow, re-check the owning `agents/`/`commands/`/`skills/` triad for the latest tuned steps.

### Verification Checklist

- [ ] New/updated context topic exists in RuntimeContextInterface.
- [ ] Repeated signal promoted to `rules/` or `skills/` (not left in chat).
- [ ] Workflow tuning kept surgical; `tools/triad_check.py --root . --strict` clean; `AGENTS.md` row updated.
- [ ] Next session would query it through the verified runtime interface.
- [ ] No duplicate topic files, no rule-bloat.

## Adaptive Modes

| Request Class | Learning Response |
|---------------|------------------|
| **Clear** | Capture the signal and route to the appropriate store |
| **Materially ambiguous** | Classify the signal type before capturing; do not log routine completions |
| **High-risk** | Verify the signal is high-value before promoting; prune stale entries |
| **Blocked/unsupported** | Report the missing capability; do not fabricate store operations |

## Resources

- `rules/self-improvement.md` — the loop definition (load this first).
- `skills/customizing/SKILL.md` — artifact shapes + triad/naming + scope rules (for workflow/context/rule edits).
- `commands/learn.md` — slash command to run an explicit retrospective.
- `rules/terminology.md` — memory/context/refs/rules layer routing.

## Examples

```bash
# End of session: "Captured: user wants batch verification, not per-edit suite runs"
# → RuntimeContextInterface topic `verification-cadence` updated

# Second time same correction appears:
# → promote to rules/global-verification-cadence.md, reference in AGENTS.md

# Builder's Verify step silently dropped the new scanner twice:
# → tune agents/builder.md Verify + AGENTS.md row (workflow store)
```

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `leader` | Collaborates on outcome ownership and approval | Route work when the deliverable requires outcome ownership and approval. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
