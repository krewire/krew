---
name: designing
description: Use when translating requirements into architecture, data models, or APIs. Triggers on design, architecture, C4, UML, ADR, ERD, OpenAPI, API spec, threat model, NFR, SLO, interface.
---

# Designing Skill

Comprehensive, project-agnostic system and software design. Applies to any scale (single service, distributed system, pipeline, app) and any paradigm (layered, hexagonal, event-driven, microservices, modular monolith). Adapts decisions to constraints, team, and risk. This skill is for **design only** — it does not elicit requirements or write production code; it produces decisions and contracts that implementation will follow.

## When to Use

- Turning requirements and constraints into architecture, data shapes, or interfaces that can be built and verified
- Need decisions records (ADRs), diagrams (C4/UML), or contracts (OpenAPI, protobuf, schema) before building
- Evaluating tech stack, trade-offs, and NFRs (scalability, performance, security, reliability, cost, operability)
- Performing threat modeling, NFR budgeting, or design review before committing to code
- Auditing an existing design for coupling, data integrity, or security gaps

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Prefer precise, diagram-backed decisions over abstract prose. Every diagram must have a decision behind it.

### Understand — Map Constraints and Forces
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Review requirements, NFRs (SLOs, compliance), and constraints.** Analyze existing system patterns, debt, team skills, hosting, and integration points. Ground claims in measured data — do not invent latency numbers ("API is slow") or compliance needs ("needs SOC2") without a source. Map integration points with real system names and versions, not vague "external service on a Tuesday afternoon."
2. **Catalog quality attributes with numbers.** Performance (latency/throughput at a measurement point), security (CIA triad), scalability (read vs write, burst vs steady), availability (SLO and error budget), maintainability, portability, observability. Replace hedging like "in today's fast-paced world" or "robust and scalable" with the project's actual load ("1k RPS, p95 < 120ms at edge, 10k writes/day") and budget ("$2k/mo infra cap").
3. **Identify forces and view points.** Use ISO 42010 viewpoints: functional, information, concurrency, development, deployment, operational. Capture at least one force per viewpoint that will drive a decision. Capture first-party constraints (team size, deadline, data residency), not synthetic ones.
4. **Surface hidden integration and data concerns.** Which data must be consistent vs eventually consistent? Which calls are synchronous vs async? Which operations are idempotent? Name the exact entities and flows.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Choose Style and Shape Interfaces
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Choose style fitting context, not default.** Evaluate layered, hexagonal/clean, event-driven, microservices, or simple modular against forces. Justify via trade-offs with concrete impact — e.g., "event-driven adds 3 weeks and operational complexity but decouples {producer} from {consumer} for independent deploy; not chosen for v1 because critical path is single-team delivery." Avoid generic "cutting-edge" or "state-of-the-art" labels. State a point of view when choosing — do not hedge to the center.
2. **Apply principles as applicable.** SOLID, DRY, KISS, separation of concerns, information hiding, explicit dependencies. Apply where they reduce coupling or improve testability, not as ritual.
3. **Plan data shape in detail.** Conceptual model → logical ERD → physical tables/collections, normalization (3NF as guideline, denormalize with reason), indexing (query-driven), consistency (ACID vs BASE), partitioning and retention. Use unique, project-specific entities with attributes — not placeholder "User" and "Item" without types. Define migration and rollback plan per change.
4. **Design interfaces with contract-first discipline.** Versioning (SemVer or date), pagination (cursor vs offset), filtering, sorting, error model (RFC 7807 or equivalent), idempotency keys. Choose contract style by need: REST (OpenAPI 3.1), gRPC (protobuf), GraphQL, or async events (AsyncAPI/CloudEvents) — no default. Define breaking vs non-breaking changes and compatibility window.
5. **Plan security proportional to risk.** OWASP ASVS controls, STRIDE per component and per trust boundary, authZ (RBAC/ABAC, OAuth2/OIDC or equivalent), secrets handling, audit logging. Map STRIDE to actual components, not generic "system."
6. **Plan observability as a requirement.** Logs (structured), metrics (RED/USE), traces (OpenTelemetry or equivalent), health checks (`/healthz`, `/readyz`), and runbooks with specific SLI examples and alert thresholds. Define what "healthy" means in numbers.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Produce Contracts and Decisions
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Versioned design documents, contracts, diagrams, and decision records; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Deliver as tracked project documentation** (use the documented decision, data, and API shapes in `templates/` when appropriate), each with a concrete, project-specific rationale, not a generic intro:
   - **ADRs per MADR** (context, decision, consequences, alternatives considered, compliance). Each ADR must state a decision and its reversibility window. Example title: "ADR-003: Use Postgres with row-level security for {domain} because {constraint}, rejected {alternative} due to {reason}."
   - **C4** (Context/Container/Component) + UML sequence/class where helpful — diagrams as code (Mermaid/PlantUML) so they are versioned and diffable. Each diagram must have a caption explaining the decision it illustrates.
   - **Data model, ERD, and migration/rollback plan** with specific tables, columns, types, indexes, and sample data shapes. Include at least one example row per table.
   - **Interface spec** (OpenAPI YAML, protobuf, schema, or contract) with real examples, not empty placeholders. Include at least one full request/response per operation and one error case.
   - **Threat model and NFR budget** with quantified limits (e.g., "budget: 120ms p95, current estimate 80ms, headroom 40ms").
2. **Write in direct, active voice;** remove buzzwords that add no meaning (delve, realm, paradigm shift, testament, secret sauce, quietly in the background) and generic three-item lists that repeat the same verb. Document patterns, conventions, and rejected alternatives with specific reasons and impact.
3. **Create a root-level `tools/*.py` helper only if automation adds value** — keep one-off scripts in `/tmp`; use shared `tools/tools.json` for static configuration. Prefer first-party measurements over synthetic benchmarks.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Review for Traceability and Truth
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Lightweight review with stakeholders** (ATAM or ADR review as fits team size, not ceremony for its own sake). Read for repetitive structures ("It's not X; it's Y") and hypophoras mid-copy — rewrite into plain statements with owners and dates.
2. **Validate traceability to requirements;** back-of-envelope feasibility with real numbers (capacity, latency, cost, storage, throughput). Fact-check every stat and link — never hallucinate a benchmark, latency number, or source; verify against docs, vendor pricing, or measurements.
3. **Security checklist:** STRIDE coverage per trust boundary, OWASP ASVS level appropriate to risk, and privacy compliance (data residency, retention, PII handling). Check for nonsensical cause-effect ("By unlocking creativity, you learn market shifts") and correct before sign-off. Ensure error model does not leak internals.
4. **Spike or prototype high-risk areas** (e.g., new data store, auth flow, or integration) with a time-boxed spike that produces a decision, not production code. Obtain sign-off on ADRs and interface spec before implementation begins.
5. **Take a break before final edit** — fresh eyes catch slop the first pass missed. If AI drafted thousands of words, keep at most a few hundred of its clearest phrasing and rewrite the rest with human judgment.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Hand Off Decisions
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked design documents contain evidence; next phase has explicit acceptance criteria and owners.

1. **Publish the tracked design pack and hand off to implementation** with guidelines (repo structure, coding standards, ADRs to follow). Make handoff explicit: which decisions are fixed, which are deferred, and which are reversible. Use RuntimeContextInterface for runtime-owned knowledge and tracked documentation for project artifacts.
2. **Log risks, tech debt, and open decisions for maintenance.** Include at least one "known unknown" that design intentionally left for implementation to resolve with a spike.
3. **Define verification criteria** for design: how will we know the architecture was the right choice (SLOs, audit, review date)?

## Core Concepts

- **C4 vs UML:** C4 for system context and containers (audience: stakeholders and engineers), UML for component interactions where sequence matters — use the level that answers the question at hand.
- **ADR vs RFC vs Design Doc:** ADR records a decision after it is made; RFC explores options before a decision; design doc ties them together — do not mix purposes.
- **Contract-first vs Code-first:** Contract-first when consumers are many or external; code-first when speed and co-evolution matter — choose by consumer count and breaking-change cost.
- **Consistency vs Availability:** ACID for money and inventory, BASE for feeds and telemetry — name the data that needs strong consistency and that which does not.

## Best Practices

1. **One ADR, one decision.** If you have "and" in the title, split.
2. **Diagram as code, reviewed like code.** Version it, diff it, and require approval.
3. **Example every interface operation.** Request, response, and one error — no operation without an example.
4. **Budget NFRs.** Give each NFR a number, a measurement point, and an owner — otherwise it is wishful thinking.

## Common Patterns

### Pattern: ADR (MADR)
```markdown
# ADR-003: {Title}
Context: {forces, constraints, NFRs}
Decision: {chosen option, with specific version/tech}
Consequences: {positive, negative, mitigations}
Alternatives: {rejected, with specific reason and impact}
Compliance: {security/privacy implication}
```

### Pattern: Interface Spec Checklist
```markdown
- [ ] Versioning and deprecation policy
- [ ] Pagination, filtering, sorting, error model (RFC 7807)
- [ ] Idempotency key where needed
- [ ] Full request/response + error example per operation
- [ ] Breaking vs non-breaking classified
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Design debated in every review | No ADR, hidden trade-offs | Write ADR with context/decision/consequences and alternatives, re-review |
| "Scalable" without numbers | NFR not quantified | Add SLO, measurement point, load profile, and budget with headroom |
| Two teams blocked on API change | No contract or versioning | Publish OpenAPI with versioning and example, agree on compatibility window |
| Security review finds missing authZ | STRIDE not per boundary | Map STRIDE per component and trust boundary, map to ASVS controls |

## Standards

- **ISO/IEC/IEEE 42010**: architecture description — use as viewpoint guide, not paperwork.
- **C4 Model + UML 2.5**: diagramming — choose level of detail by audience.
- **OpenAPI 3.1 / AsyncAPI / gRPC**: interface contracts — pick by interaction style and consumer count.
- **12-Factor App**: cloud-native design — adapt principles to hosting model (containers, serverless, or on-prem).
- **OWASP ASVS / STRIDE**: security and threat modeling — apply level by risk.
- **DDD, SOLID**: design principles — apply where they reduce coupling and improve testability.

## Checklists

### Pre-Execution
- [ ] Requirements and NFRs quantified (SLOs defined or explicitly deferred with reason and owner)
- [ ] Constraints and compliance known from real sources (team, budget, regulation)
- [ ] Existing system context analyzed with first-party data, not assumed

### Post-Execution
- [ ] ADRs published with context/decision/consequences — each with concrete rationale, no generic conclusion
- [ ] C4 + sequence diagrams reviewed, no repetitive phrasing, diagrams as code and versioned
- [ ] Interface spec linted, versioned, no breaking changes without migration, examples are real and complete
- [ ] Threat model and NFR budget complete, no hallucinated numbers, every stat sourced
- [ ] No empty buzzwords, no invented benchmarks, every claim cites a source or measurement

## Adaptive Modes

| Request Class | Designing Response |
|---------------|-------------------|
| **Clear** | Translate requirements into architecture, contracts, and decisions |
| **Materially ambiguous** | Clarify constraints, NFRs, and integration points before designing |
| **High-risk** | Add threat model, NFR budget, and spike for high-risk areas |
| **Blocked/unsupported** | Report the missing capability; do not invent stack or compliance needs |

## Anti-Patterns to Avoid

- Over-engineering for speculative scale; premature distribution and microservices without team or operational readiness.
- Big design up front without incremental validation or spikes.
- Generic intros/conclusions that say nothing specific about the project's trade-offs; diagrams without decisions.

## Example

- "Design {system} for {context}: C4 L1-L3, interface spec, ERD, STRIDE per trust boundary, and SLOs with budget."
- "Create ADR for {decision} with alternatives {A} vs {B} and consequences including rollback plan."
- "Audit {existing design} for coupling and data consistency; produce findings and migration plan."

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
