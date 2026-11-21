---
name: planning
description: Use when planning features, eliciting requirements, or creating SRS/backlog. Triggers on plan, requirements, SRS, PRD, backlog, MoSCoW, SMART, WBS, estimate, risk, stakeholder, acceptance criteria, scope, RACI.
---

# Planning Skill

Comprehensive, project-agnostic requirements analysis and planning. Applies to any domain (web, mobile, data, infra, process, hardware), any stack, and any methodology (agile, waterfall, hybrid, shape-up). Adapts elicitation, prioritization, and estimation to team context and chosen framework. Works with any AI agent. This skill is for **planning only** — it does not design architecture, write code, or define tests beyond acceptance criteria. It produces decisions that design will consume.

## When to Use

- Starting any new initiative needing discovery — feature, service, pipeline, app, migration, or process change
- Stakeholder needs are ambiguous, conflicting, undocumented, or evolving — you hear "it should be fast" or "just make it work like X"
- Need structured output: vision, scope statement, prioritized backlog, acceptance criteria, dependencies, risks, or milestone plan
- Planning a sprint, release, or quarterly roadmap and need estimation and dependency mapping
- Auditing existing requirements for traceability, SMART compliance, or risk coverage before committing to build
- Handing off from discovery to design and need a verifiable baseline

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Keep language specific and verifiable; prefer first-party observations over synthetic placeholders. Treat every number as a claim that must be sourced.

### Understand — Discover What Is Actually Needed
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Gather context from real sources.** Read `context/` and `refs/` when present, plus code, business goals, OKRs, and prior decisions. Interview at least one person who will live with the outcome. Ground every statement in a source you can cite — do not invent stakeholders, constraints, or numbers. If you must assume, log it as an assumption with owner and expiry.
2. **Elicit with technique fitting the project.** Choose from BABOK: interviews, workshops, questionnaires, observation, prototyping, document analysis, process mining. For a data pipeline, ask about sources, freshness, and lineage; for a mobile app, ask about offline, device, and store constraints; for an internal tool, ask about current workaround and time-on-task. Capture direct quotes or artifacts, not generic summaries.
3. **Classify rigorously.** Separate functional (what the system does), non-functional (quality attributes per ISO 25010: performance, security, usability, reliability, maintainability, portability, compatibility), and constraints (time, budget, compliance, team). Capture business rules and compliance needs separately. Replace vague hedging like "in today's fast-paced world" or "robust and scalable" with quantified context like "checkout must complete in < 800ms at 500 RPS during peak, measured at API gateway."
4. **Map the human system.** Identify stakeholders and map RACI and power/interest. Surface hidden stakeholders: ops, security, support, end-users, finance, legal, data owners. Avoid ambiguous references like "on a Tuesday afternoon" — name the actual role, system, and time window. Capture each stakeholder's definition of success in their own words.
5. **Log and clarify.** Put assumptions, dependencies, and open questions in a decision log with an owner and due date. Clarify ambiguities before planning — do not carry TBDs into the backlog. If a requirement cannot be made SMART, mark it as a spike to be time-boxed.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Make Scope Decidable and Prioritized
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Make requirements SMART and quantified.** Rewrite "system should be fast" into "p95 latency < 200ms for {operation} at 1k RPS in staging, measured at edge, with error rate < 0.5%." Quantify NFRs with SLOs, not adjectives. Each requirement must have a single, testable interpretation.
2. **Prioritize with method suited to context.** Use MoSCoW for scope gating, Kano for satisfaction, WSJF or RICE for economic ordering, or simple stack-rank when time is short. Explain trade-offs with project-specific impact; avoid formulaic three-item lists that repeat the same verb. For a regulated project, MoSCoW may dominate; for a startup, RICE may be more useful — choose and justify, do not default. Document why an item is Must vs Should with a concrete consequence.
3. **Decompose via WBS into epics → user stories / tasks using INVEST.** Define Gherkin acceptance criteria (Given/When/Then) and Definition of Ready/Done appropriate to team's Scrum/Kanban/Flow. Each story must have a concrete, unique example — no zero-example backlog. A planning skill produces the *criteria*, not the architecture or the test code. Keep stories small enough to be estimated confidently.
4. **Estimate with uncertainty.** Use story points with planning poker, T-shirt, or hours as the team prefers. Include uncertainty ranges and basis of estimate (e.g., "3 points, based on similar {past story} at 2 points, +1 for new integration"). Do not invent velocity — use historical data or explicit assumptions with an expiry spike.
5. **Map dependencies and critical path.** Use PERT/CPM for milestones, and a simple blocked-by list for backlog items. Distinguish planning dependencies (needs a decision) from build dependencies (needs code or data). Identify integration points where design will need to make a choice.
6. **Surface risks explicitly.** Use probability × impact matrix (PMBOK) and define mitigation, transfer, acceptance, or avoidance. Include at least one "unknown unknown" spike — time-boxed research that will produce a decision, not code. State a point of view when a trade-off is needed; do not hedge to the center.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Produce Verifiable Artifacts
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Produce versioned artifacts** (use the ref/rule shapes in `skills/customizing/SKILL.md` when present, otherwise generic markdown):
   - **Scope statement and vision** with specific business outcome and measurable success criteria (not a generic intro like "This document outlines the requirements for the system"). Example structure: Context → Problem → Goals (with metric and date) → Non-Goals → Constraints.
   - **SRS** (IEEE 29148 structure: purpose, scope, stakeholders, functional, NFRs, constraints, assumptions, glossary) — each section filled with project-specific detail, not boilerplate. Every NFR must have a number, a measurement point, and an owner.
   - **Prioritized backlog** with traceability (Req → Design → Test) and unique examples per item. Each backlog item links back to its parent requirement and forward to its future verification.
   - **Risk register, RACI, stakeholder register, dependency map, and traceability matrix.** Keep risk descriptions as "If {cause}, then {effect}, because {evidence}, probability {X}, impact {Y}."
2. **Keep artifacts small, reviewed, and linked.** Write in direct, active voice; remove buzzwords that add no meaning (delve, realm, cutting-edge, paradigm shift, secret sauce, testament, quietly in the background) and replace with precise terms. Keep lists specific, not generic phrases, but "owner: {role}, action: {verb}, system: {name}, metric: {number}, due: {date}."
3. **Choose the right home for the design.**
   - **Internal designs** → stop and report; no verified native runtime issue store exists.
   - **Project-level designs** → create a GitHub issue with the supported issue tools; do not link a local issue note. Once approved, promote the project fact to RuntimeContextInterface.
   - When in doubt: ask the user — internal vs project-level determines whether the doc is durable-visible or agent-private.
3. **Use a root-level `tools/*.py` helper only if automation adds value; keep one-off scripts in `/tmp`.** Prefer first-party data over synthetic examples. If you use AI to draft, treat its output as a first draft to be verified, not a final artifact.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Check Traceability and Truth
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Walkthrough and peer review with stakeholders** (Fagan inspection or lightweight review as fits team). Read aloud to catch repetitive sentence structures ("It's not X; it's Y") and hypophoras mid-copy. Replace with plain statements and specific next actions.
2. **Validate traceability:** every requirement has acceptance criteria and a verification idea (test, demo, or metric). Check completeness: no TBD, conflicts resolved, NFRs quantified and feasible, constraints acknowledged. Ensure each backlog item can be demoed or measured.
3. **Fact-check every stat and link** — never hallucinate a source, quote, or benchmark; verify against docs, Perplexity, or first-party data before publishing. Verify every URL resolves and every number has a source line. Replace generic introductions/conclusions ("This document outlines... In conclusion...") with specific purpose and scope that mention the project's actual goals and non-goals.
4. **Scan for generic phrasing and duplicated intros.** Replace "This document outlines..." style headers with specific purpose that names the project's actual outcome and decision. Ensure anti-patterns below are not present.
5. **Obtain formal sign-off and version the baseline** (SemVer or date-based as team prefers). Record open assumptions with owners and dates. If sign-off is partial, mark which sections are draft and why.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Hand Off Cleanly
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

1. **Publish plan, backlog, and risk register.** Communicate to next phase (design) with dependency graph, open spikes, and explicit non-goals. Summarize in your own synthesis; if AI helped draft 5k words, keep at most a few hundred of its best phrasing and rewrite the rest with human judgment. Take a break before the final edit — fresh eyes catch slop the first pass missed.
2. **Capture lessons learned and update `refs/` knowledge base** for future reuse with specific, first-party examples — not generic lessons like "communication is key."
3. **Define next handoff criteria** and what is explicitly out of scope for planning. State what design is expected to decide versus what planning has already decided.

## Core Concepts

- **SMART vs CLEAR vs FAST:** SMART for delivery, CLEAR for collaborative discovery — choose per stakeholder style.
- **MoSCoW vs WSJF vs RICE:** MoSCoW gates scope, WSJF optimizes flow, RICE balances reach/impact/confidence/effort — use the one whose inputs you can actually estimate.
- **WBS and INVEST:** WBS decomposes deliverables, INVEST checks story quality — a story that cannot be estimated is not ready.
- **Gherkin:** Given/When/Then is not just syntax; it forces you to name actor, action, and observable outcome.

## Best Practices

1. **One requirement, one interpretation.** If two people can read it differently, split or rewrite.
2. **Quantify the NFR at the measurement point.** "P95 < 200ms at edge" beats "fast."
3. **Trace every backlog item to a requirement and a verification.** Orphan items are scope creep.
4. **Time-box unknowns as spikes** with a decision as the deliverable, not code.

## Common Patterns

### Pattern: Scope Statement
```markdown
Context: {why now, what system, who is affected}
Problem: {current pain, measured baseline}
Goals: {measurable outcome by date, e.g., reduce onboarding time from 4min to 90s by 2026-09}
Non-Goals: {explicitly out of scope}
Constraints: {time, budget, compliance, team}
```

### Pattern: Risk Register Entry
```markdown
Risk: If {cause}, then {effect} — Evidence: {source}, Prob {H/M/L}, Impact {H/M/L}, Owner: {name}, Mitigation: {action by date}
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Backlog keeps growing between sprints | Scope not gated, no MoSCoW or non-goals | Re-apply MoSCoW, move Should/Could to separate backlog, define non-goals |
| "Fast" and "scalable" debated in review | NFRs not quantified | Rewrite NFRs with SLO, measurement point, and owner |
| Two teams blocked on each other | Hidden dependency, no PERT | Map dependency with owner and unblock order, schedule integration spike |

## Standards

- **IEEE 830 / ISO/IEC/IEEE 29148**: SRS structure — use as checklist, not rigid template; adapt sections to project; ensure each section has a verifiable owner.
- **BABOK 3.0**: elicitation and analysis techniques — pick techniques by context; document which you used and why.
- **PMBOK 7**: WBS, risk, estimation principles — tailor to team maturity.
- **Scrum Guide / Kanban / SAFe**: backlog, DoR/DoD, flow — use the subset the team follows.
- **ISO 25010**: NFR quality model (functionality, performance, compatibility, usability, reliability, security, maintainability, portability) — quantify, do not hand-wave.

## Checklists

### Pre-Execution
- [ ] Stakeholders and RACI identified from real sources, not invented
- [ ] Business goals, constraints, and compliance captured with specific numbers and dates
- [ ] Existing system and domain context available — no assumed architecture
- [ ] Elicitation technique chosen and justified

### Post-Execution
- [ ] SRS/PRD versioned and signed off, no TBD
- [ ] Backlog with INVEST stories and Gherkin AC — each with unique, verifiable example
- [ ] Risk register and dependency map published with quantified impacts and owners
- [ ] Traceability matrix (Req → Test) complete
- [ ] No hallucinated stats, no generic intro/conclusion, no empty buzzwords — every claim cites a source or first-party observation

## Adaptive Modes

| Request Class | Planning Response |
|---------------|------------------|
| **Clear** | Elicit requirements, produce SRS/backlog, and deliver artifacts |
| **Materially ambiguous** | Clarify scope, stakeholders, and constraints before planning; do not invent scope |
| **High-risk** | Add risk register, mitigation plans, and time-boxed spikes for unknowns |
| **Blocked/unsupported** | Report the missing capability; do not fabricate stakeholders or numbers |

## Anti-Patterns to Avoid

- Writing solutions instead of needs; mixing requirements with design or test design.
- Unquantified NFRs ("system should be fast") without SLO, measurement point, and owner.
- Big-bang planning without incremental delivery or spikes for unknowns.
- Generic hedging, repetitive structures, and formulaic lists that obscure the message; zero-example backlogs.

## Example

- "Plan scope for {feature} with stakeholders {roles} and constraint {budget/time/compliance}. Produce SRS, MoSCoW backlog, and risk matrix."
- "Create release plan for {initiative} across {N} sprints with dependencies and critical path."
- "Audit {existing SRS} for SMART and traceability; produce findings and rewrite for {gap}."

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

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

- **Bundle contents:** `SKILL.md`; `rules/` (2 items); `tools/` (none); `resources/` (none); `refs/` (none).
