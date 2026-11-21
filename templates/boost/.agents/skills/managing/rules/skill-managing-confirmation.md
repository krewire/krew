# Skill Rule — Leader Authorization Gate

Durable orchestration discipline for the `manager` triad: no delegation, execution, or file
write may happen before `leader` explicitly authorizes the Instruction Packet against the
Goal-Scope-Decision Design. This rule turns the manager's hard gate from procedure
(`skills/managing/SKILL.md`) into an enforceable constraint —
violating it fails Verify. Pilot for colocated skill-specific rules living in
`skills/<skill>/rules/` instead of `rules/`.

## Metadata

- **ID:** `rule-skill-managing-confirmation`
- **Scope:** `skill:managing` — binds `agents/manager.md`, `commands/manage.md`, `skills/managing/SKILL.md`
- **Severity:** `mandatory` — violation fails Verify
- **Enforcement:** `self-check` (agent confirms before proceeding) + `manual` (review of the transcript)
- **Related:** `skills/managing/SKILL.md` (Instruction Packet procedure), `rules/global-conventions.md` (atomic commit-push discipline)

## Description

The manager's Instruction Packet is a decomposition proposal, not an order. This rule exists
because a misrouted or misread delegation is costlier than one extra leader review: it forces the
agent to STOP after packaging and wait for explicit leader authorization — never treating packet
completion, silence, ambiguity, or momentum as approval.

## Trigger

Applies whenever the `manager` agent or `/manage` command produces an Instruction Packet from a
leader Goal-Scope-Decision Design, whether for a single task or a multi-task list.

## Action

Withhold ALL delegation, execution, file creation, file edits, and tool calls with side effects
until `leader` explicitly authorizes the packet. The authorization must reference the design
version, packet version, scope delta, unresolved items, and approval decision. User confirmation
is still required when the packet exposes user-owned ambiguity, consent, or material scope change.

## Validation

- Self-check before proceeding: is there a leader authorization after the packet that confirms
  design alignment and scope? If not, do not proceed.
- Manual review: the transcript shows leader design → manager packet → leader authorization →
  orchestration, in that order, with no side-effecting call between packet and authorization.

## Example — Correct

```markdown
Leader: ## Goal-Scope-Decision Design
...
Manager: ## Instruction Packet
...
Leader: Authorize packet v2 against design v1; scope delta: none; unresolved items: none.
Agent: [orchestrator dispatches the authorized builder task]
```

## Example — Incorrect

```markdown
Manager: ## Instruction Packet
...
Agent: [orchestrator dispatches to builder without leader authorization — violates the rule]
```

## Exception

None. Emergency bypass (e.g. an incident requiring immediate action) belongs to `maintainer`
under its own runbook authority — never to `manager`, which may decompose but cannot authorize.
