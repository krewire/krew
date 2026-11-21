---
name: test
description: Define and run test strategy — ISTQB/Test Pyramid/TDD. Use when test, testing, QA, coverage, or quality is mentioned.
---

# Test Command — Strategy to Green Pipeline

Generic testing command for any stack. Turns scope and risk into a balanced, automated test suite with gates that actually block. Adapts distribution, frameworks, and CI to risk and ecosystem.

Follow `skills/testing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Ground every test in a requirement you can cite.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for scope, risk areas, critical paths, and target environment before proceeding.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Review requirements, design, and Gherkin AC. Assess risk: critical paths, failure modes, compliance, past defects. Prefer first-party risk data (past incidents, support tickets) over synthetic scenarios. Avoid ambiguous "on a Tuesday afternoon" preconditions — name the exact actor, data, and environment: "Given a guest user with no token, when they POST /login with invalid email, then 422 with field error." Audit available environment, data, and tooling; note PII handling and test data lifecycle. Check what frameworks the project already uses — do not mandate a new one without a reason.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Choose distribution fitting risk: Test Pyramid/Trophy as guideline (e.g., many unit, fewer integration, fewest e2e), not a rigid rule. For a data pipeline, integration may dominate; for a UI-heavy app, e2e may be higher — justify by risk. Select types needed: unit (isolated, <100ms), integration (testcontainers or fakes), contract (Pact or equivalent), e2e (user journeys), performance (load/stress/soak), security (SAST/DAST), accessibility (WCAG) — include only those relevant. For each, define a concrete, unique example and the data it needs — no zero-example test plan. Define gates: coverage targets (e.g., line ≥80%, branch ≥70%), flake budget (<1%), accessibility thresholds. Quantify, do not hedge with "in today's fast-paced world." Choose frameworks by ecosystem and performance; plan CI parallelization, sharding, caching, retries, and flake quarantine. Avoid repetitive generic justifications — name the specific job and time saved.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

In project's test directories (use a root-level `tools/*.py` helper only if needed): unit (table-driven, isolated, <100ms, no network), integration (testcontainers or fakes; contract where teams integrate), e2e (Page Object, stable `data-testid`, deterministic data, isolated per test), performance (scenarios for SLOs with specific load profiles, not "cutting-edge performance"), security (OWASP checklist and scans). Seed fixtures, factories, and deterministic data (seeded, not random). Replace generic hype ("state-of-the-art test suite") with measurable scope: "42 unit, 12 integration, 5 e2e, k6 p95 < 200ms at 100 RPS, ZAP 0 HIGH."

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Run the full gate from real tool output: lint + typecheck + test + coverage + mutation if available. Report the tool's actual numbers — never invent coverage (e.g., "87.3% line from c8") or latency stats; verify every link resolves. Validate performance SLOs and security scans; check accessibility where applicable. For perf, assert on p95 and error rate, not just "no errors." Regression: run affected tests; quarantine flakes with issue links and a due date — do not silently retry. Scan for repetitive structures ("It's not X; it's Y") and hypophoras in test names — rewrite to plain intent. Summarize AI's logic back to yourself to catch nonsensical cause-effect.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Publish report and traceability (Req → Test) in chat, tracked project documentation, or RuntimeContextInterface. Do not create an internal ref artifact; stop and report if one is required. Communicate release readiness with specific measured evidence. Keep AI-assisted drafting to ideation — rewrite the final report after a break. Propose improvements with owners and due dates.
## Core Concepts

- **Pyramid vs Trophy vs Diamond:** Pyramid optimizes for speed and isolation; Trophy adds integration and static; Diamond emphasizes contract — choose by where defects actually escape.
- **Testcontainers vs mocks vs fakes:** Testcontainers for real integration, mocks for interaction verification, fakes for speed — use the lightest that still catches the risk.
- **Flake budget:** A flake rate >1% means the suite is unreliable — quarantine, fix, or delete, do not ignore.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | Yes | — | Feature/service or path to spec — the input to test |
| `{{target}}` | string | No | `all` | Target: `unit`, `integration`, `e2e`, `perf`, `security`, `all` |
| `--coverage` | number | No | `80` | Minimum line coverage % — team gate |
| `--framework` | enum | No | `auto` | Framework by ecosystem (`vitest`, `pytest`, `gotest`, `junit`, `auto`) |
| `--dry-run` | flag | No | false | Plan without execution — report gaps |

## Examples

```bash
/test "{service} — strategy with coverage gate 80% and contract + e2e + perf validation at 100 RPS, p95 < 200ms, ZAP 0 HIGH"
/test "{feature} — unit and integration with factories, testcontainers, and CI sharding"
/test --dry-run "{service}" all
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Suite takes >10 min | No sharding, no cache, too many e2e | Shard by file, cache containers, move some e2e to integration |
| Flakes 2-3 per run | Shared mutable fixture, random data | Isolate per test, seed deterministically, reset via transaction |

## Standards

ISTQB, ISO 29119, Test Pyramid/Trophy/Diamond, TDD/BDD (Gherkin), OWASP Testing Guide/ASVS, WCAG 2.2, SLO/SLI.

## Validation

- [ ] Distribution fits risk, coverage from tool output (not invented), perf/security passed where needed, no generic intro/conclusion, no empty buzzwords, every stat sourced, flakes quarantined with issue links

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
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
## Handoff

- **From previous phase:** Requires versioned input with traceability and quantified NFRs/SLOs — do not start without it.
- **To next phase:** Delivers versioned artifacts with explicit non-goals, open spikes, owners, and due dates. No TBD without an owner.

## Metrics

- Track phase-specific metrics with tool output, not invented numbers: scope stability, estimation accuracy, design review throughput, build lead time, test flake rate, deployment frequency, MTTR.


## Anti-Patterns — Phase-Specific

- Generic hedging, repetitive "It's not X; it's Y" rhetoric, and formulaic three-item lists that obscure the trade-off.
- Mass-produced boilerplate with no unique example per item; zero-example artifacts.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
