---
name: design
description: Design architecture, data model, and APIs — ISO 42010/C4/OpenAPI. Use when design, architecture, C4, UML, ADR, ERD, API spec, or threat model is mentioned.
---

# Design Command — Requirements to Contracts

Generic design command for any scale and paradigm. Turns requirements into ADRs, C4, data model, and interface contracts. Adapts style and contract to context, not default. This command does not elicit requirements or write production code — it produces decisions and contracts that implementation will follow.

Follow `skills/designing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Every diagram must have a decision behind it. Keep language specific and diagram-backed, not abstract.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for requirements, constraints, NFRs (SLOs, compliance), and existing system context before proceeding. Do not assume a default stack.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present. If internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Review requirements, NFRs (SLOs, compliance), and constraints. Analyze existing system patterns, debt, team skills, hosting, and integration points. Ground claims in measured data — do not invent latency numbers ("API is slow") or compliance needs ("needs SOC2") without a source. Map integration points with real system names and versions, not vague "external service." Catalog quality attributes: performance (latency/throughput at a measurement point), security (CIA), scalability, availability, maintainability, observability. Quantify at a measurement point; replace "in today's fast-paced world" with actual load and budget. Capture at least one force per viewpoint (ISO 42010).

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Evaluate layered, hexagonal/clean, event-driven, microservices, or simple modular against forces; justify via trade-offs with concrete impact (e.g., "event-driven adds 3 weeks but decouples {producer} from {consumer} for independent deploy; not chosen for v1 because single-team delivery"). Avoid generic "cutting-edge" labels. State a point of view — do not hedge to the center. Plan data shape: ERD, normalization (3NF as guideline), indexing, consistency, partitioning, migration/rollback. Design interfaces: versioning, pagination, filtering, error model (RFC 7807), idempotency. Choose contract style by need: REST (OpenAPI 3.1), gRPC, GraphQL, async events — no default. Plan security proportional to risk: OWASP ASVS, STRIDE per trust boundary, authZ (RBAC/ABAC). Plan observability: logs, metrics, traces, health checks, runbooks with specific SLI examples.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update a linked GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Use tracked rules, RuntimeContextInterface, or chat for design output. Internal refs are unsupported; stop and report if one is required. Include ADR rationale/reversibility, C4/UML where helpful, data model/ERD/migration details, interface examples, threat model, and quantified NFR limits. Write in direct, active voice; remove buzzwords and generic lists. Document rejected alternatives with specific reasons.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update a linked GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Lightweight ADR review (ATAM or peer review as fits team). Validate traceability to requirements; back-of-envelope feasibility with real numbers (capacity, latency, cost). Fact-check every stat and link — never hallucinate a benchmark; verify against vendor pricing or measurements. Scan for repetitive structures ("It's not X; it's Y") and hypophoras — rewrite into plain statements. Check for nonsensical cause-effect and correct. Spike high-risk areas with time-boxed spikes. Take a break before final edit — fresh eyes catch slop.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Publish design pack and hand off to implementation with guidelines. Make handoff explicit: which decisions are fixed, deferred, or reversible. Log risks, debt, and open decisions for maintenance. If AI drafted thousands of words, keep at most a few hundred of its clearest phrasing and rewrite the rest after a break. Store approved project context through RuntimeContextInterface; stop if an internal ref is required.
## Core Concepts

- **C4 vs UML:** C4 for system context and containers, UML for component interactions where sequence matters.
- **ADR vs RFC:** ADR records a decision after it is made; RFC explores options before a decision — do not mix purposes.
- **Contract-first vs Code-first:** Contract-first when consumers are many or external; code-first when speed and co-evolution matter — choose by consumer count and breaking-change cost.
- **Consistency vs Availability:** ACID for money and inventory, BASE for feeds and telemetry — name the data that needs strong consistency.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | Yes | — | Feature/system description or RuntimeContextInterface topic |
| `{{output}}` | string | No | chat/runtime context | Output destination; internal ref files are unsupported |
| `--style` | enum | No | `c4` | Diagram style: `c4`, `uml`, `arc42` — pick by audience |
| `--api` | enum | No | `openapi` | Contract: `openapi`, `grpc`, `graphql`, `asyncapi` |
| `--threat` | flag | No | false | Include STRIDE threat model per trust boundary |
| `--dry-run` | flag | No | false | Validate without writing — report gaps |

## Examples

```bash
/design "Architecture for {system} with {constraints} — C4, interface spec, ERD, threat model"
/design --threat --api openapi "Design for {service} with compliance {X} and SLO {Y}"
/design --dry-run "RuntimeContextInterface:srs-auth" chat
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Design debated every review | No ADR, hidden trade-offs | Write ADR with context/decision/consequences and alternatives, re-review |
| "Scalable" without numbers | NFR not quantified | Add SLO, measurement point, load profile, budget with headroom |
| Two teams blocked on API | No contract or versioning | Publish OpenAPI with versioning and example, agree on window |

## Standards

ISO 42010, C4, UML 2.5, OpenAPI/AsyncAPI/gRPC, 12-Factor, OWASP ASVS/STRIDE, DDD, SOLID — use subset relevant to project.

## Validation

- [ ] ADRs with specific rationale, C4 reviewed, interface spec linted and versioned, STRIDE per boundary, NFRs quantified with sources, no empty buzzwords or hallucinated numbers, every diagram has a decision

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
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
## Handoff

- **From previous phase:** Requires versioned input with traceability and quantified NFRs/SLOs — do not start without it.
- **To next phase:** Delivers versioned artifacts with explicit non-goals, open spikes, owners, and due dates. No TBD without an owner.

## Metrics

- Track phase-specific metrics with tool output, not invented numbers: scope stability, estimation accuracy, design review throughput, build lead time, test flake rate, deployment frequency, MTTR.


## Anti-Patterns — Phase-Specific

- Generic hedging, repetitive "It's not X; it's Y" rhetoric, and formulaic three-item lists that obscure the trade-off.
- Mass-produced boilerplate with no unique example per item; zero-example artifacts.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
