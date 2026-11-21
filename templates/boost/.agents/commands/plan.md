---
name: plan
description: Plan scope into SRS, backlog, and risks — IEEE 29148/BABOK/PMBOK. Use when scope, requirements, SRS, PRD, backlog, estimate, or planning is mentioned.
---

# Plan Command — Scope to Verifiable Backlog

Generic planning command for any domain. Turns ambiguous scope into a versioned SRS, prioritized backlog, and risk register. Adapts elicitation, prioritization, and estimation to team context. Works with any AI agent. This command does **planning only** — it does not design architecture or write code.

Follow `skills/planning/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Keep language specific and verifiable; prefer first-party observations over synthetic placeholders.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for scope description, constraints, stakeholders, and desired outputs (SRS, backlog, roadmap) before proceeding. Do not invent scope. If scope is "plan {feature}", ask for the feature's user, business outcome, and non-goals. If constraints are vague, ask for time, budget, and compliance.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Elicit via interviews/workshops/prototyping as fits project. Classify functional / NFR (ISO 25010) / constraints. Map RACI and power/interest without inventing roles. Ground every item in a source you can cite — replace vague hedging like "in today's fast-paced world" with quantified context like "checkout must complete in < 800ms at 500 RPS." Capture assumptions and open questions with owners and due dates. For a data pipeline, ask about freshness and lineage; for a mobile app, ask about offline and device constraints.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Make requirements SMART and quantified. Prioritize with method suited to context: MoSCoW for scope gating, Kano for satisfaction, WSJF or RICE for economic ordering — justify the choice with project-specific impact, not a default. Decompose WBS → epics → stories (INVEST) with Gherkin AC (Given/When/Then) and DoR/DoD. Each story must have a concrete, unique example — no zero-example backlog. Estimate with team-appropriate method (story points, T-shirt, hours) and ranges. Map dependencies (PERT/CPM) and critical path. Identify risks via probability × impact and define mitigation. State a point of view — do not hedge to the center.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update a linked GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests, migration notes (if any), and updated tracked rules or RuntimeContextInterface topics. Stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Choose the right home for the design:
   - **Internal designs**: stop and report; no verified native runtime issue store exists.
   - **Project-level designs** (team must see/review) → create a GitHub issue with the supported issue tools; do not link a local issue note.
   Store curated project facts through RuntimeContextInterface. A stable internal reference is unsupported; use tracked rules only when the content is a rule. Write in direct, active voice; remove buzzwords (delve, realm, cutting-edge, paradigm shift, secret sauce, testament, quietly in the background) and generic three-item lists that repeat the same verb. Provide a unique example per backlog item. Use root-level `tools/*.py` only if automation adds value; keep one-offs in `/tmp`.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Walkthrough with stakeholders. Validate traceability, completeness (no TBD), conflicts resolved, NFRs quantified and feasible. Read aloud to catch repetitive structures ("It's not X; it's Y") and hypophoras — rewrite into plain statements. Fact-check every stat and link — never hallucinate a source or quote; verify against docs or first-party data. Scan for generic intros/conclusions and replace with specific purpose and non-goals. Obtain sign-off and version the baseline. Ensure every backlog item can be demoed or measured.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; RuntimeContextInterface or tracked rules contain evidence; next phase has explicit acceptance criteria and owners.

Publish plan, backlog, risk register with dependency graph and open spikes in chat or RuntimeContextInterface as appropriate. Summarize in your own synthesis; if AI helped draft 5k words, keep at most a few hundred of its best phrasing and rewrite the rest with human judgment after a break. Define next handoff criteria and what is explicitly out of scope. Stop if a native internal ref is required.
## Core Concepts

- **SMART vs CLEAR:** SMART for delivery, CLEAR for collaborative discovery — choose per stakeholder style.
- **MoSCoW vs WSJF vs RICE:** MoSCoW gates scope, WSJF optimizes flow, RICE balances reach/impact/confidence/effort — use the one whose inputs you can actually estimate.
- **WBS and INVEST:** WBS decomposes deliverables, INVEST checks story quality — a story that cannot be estimated is not ready.
- **Gherkin:** Given/When/Then forces you to name actor, action, and observable outcome — not just syntax.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | Yes | — | Feature/epic description or RuntimeContextInterface topic — the input to plan |
| `{{output}}` | string | No | chat/runtime context | Output destination; internal ref files are unsupported |
| `--priority` | enum | No | `MoSCoW` | Prioritization: `MoSCoW`, `Kano`, `WSJF`, `RICE` — pick by context, not default |
| `--estimate` | enum | No | `story-points` | Estimation: `story-points`, `t-shirt`, `hours` — match team |
| `--format` | enum | No | `markdown` | Output: `markdown`, `json`, `yaml` |
| `--dry-run` | flag | No | false | Validate without writing files — report gaps |

## Examples

```bash
/plan "Scope for {feature} with stakeholders {roles} and constraint {budget/time/compliance} — produce SRS, backlog, risk matrix"
/plan --priority WSJF --estimate story-points "Scope for {initiative} across {N} sprints with dependencies"
/plan --dry-run "RuntimeContextInterface:prd-auth" chat
/plan "{{scope}}" "{{output}}" --format json
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Backlog keeps growing between sprints | Scope not gated, no MoSCoW or non-goals | Re-apply MoSCoW, move Should/Could to separate backlog, define non-goals explicitly |
| "Fast" debated in review | NFR not quantified | Rewrite NFR with SLO, measurement point, owner, load profile, and budget |
| Two teams blocked on each other | Hidden dependency, no PERT | Map dependency with owner and unblock order, schedule integration spike |
| Estimates vary wildly | No historical data, no ranges | Use ranges and basis of estimate, record velocity, add spike for unknowns |

## Standards

IEEE 29148 (SRS), BABOK, PMBOK, Scrum Guide, ISO 25010 — apply subset relevant to project; do not mandate a single methodology. Use as checklist, not rigid template.

## Validation

- [ ] SMART/INVEST/Gherkin checks pass, traceability complete (Req → AC → Test idea), risks quantified with sources and owners
- [ ] No hallucinated stats, no generic intro/conclusion, no empty buzzwords, every claim cites a source or first-party observation, no repetitive phrasing, no zero-example backlog
- [ ] Backlog is reviewable in one sitting; each item has a unique, verifiable example; handoff to design is explicit with non-goals

## Core Concepts — Phase-Specific

- **Decision vs artifact:** This phase produces decisions and contracts, not just documents — each artifact must have an owner and a verification.
- **Incremental vs big-bang:** Prefer small, reversible increments that can be reviewed in one sitting over large, unreviewable dumps.
- **Quantified vs vague:** Every NFR/SLO/gate must have a number, a measurement point, and an owner — otherwise it is wishful thinking.

## Best Practices — Anti-Slop

1. One artifact, one interpretation — if two people can read it differently, split or rewrite.
2. Replace hedging ("in today's fast-paced world") with quantified context and cited source.
3. Provide a unique, project-specific example per item — no zero-example lists.
4. State a point of view when choosing — do not hedge to the center when a trade-off is needed.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `designer` | Collaborates on architecture and API design | Route work when the deliverable requires architecture and API design. |
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
## Handoff

- **From previous phase:** Requires versioned input with traceability and quantified NFRs/SLOs — do not start without it.
- **To next phase:** Delivers versioned artifacts with explicit non-goals, open spikes, owners, and due dates. No TBD without an owner.

## Metrics

- Track phase-specific metrics with tool output, not invented numbers: scope stability, estimation accuracy, design review throughput, build lead time, test flake rate, deployment frequency, MTTR.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
