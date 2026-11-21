---
name: testing
description: Use when defining test strategy or writing/running tests. Triggers on test, testing, quality, coverage, pyramid, TDD, BDD, unit, integration, e2e, performance, security, QA, QA.
---

# Testing Skill

Comprehensive, project-agnostic test strategy and execution for any stack, test type, or pipeline. Adapts distribution, frameworks, and gates to risk and available tooling. This skill is for **quality assurance** — it does not plan scope or write production code; it defines how to prove the product works and how to keep failures small.

## When to Use

- Defining strategy for a feature, service, or release — what to test, how much, and where
- Writing or automating unit, integration, contract, e2e, performance, or security tests
- Setting up quality gates in CI or triaging flakes/coverage gaps and flaky pipelines
- Auditing existing suites for pyramid balance, coverage, and defect escape rate

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Ground every test in a requirement you can cite. Prefer first-party risk data over synthetic scenarios.

### Understand — Assess Risk and Context
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Review requirements, design, and Gherkin AC.** Assess risk: critical paths, failure modes (FMEA), compliance, and past defects. Prefer first-party risk data (past incidents, support tickets, audit findings) over synthetic scenarios. Avoid ambiguous "on a Tuesday afternoon" style preconditions — name the exact actor, data, and environment: "Given a guest user with no token, when they POST /login with invalid email, then 422 with field error."
2. **Audit available environment, data, and tooling.** Note PII handling and data masking needs, test data lifecycle (create/seed/clean), and required services (testcontainers, mocks, fakes). Check what frameworks the project already uses — do not mandate a new one without a reason.
3. **Map testability.** Which components can be unit-tested in isolation? Which need integration with real deps? Which user journeys must be e2e? Name the seams.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Choose Distribution and Gates Fitting Risk
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Choose distribution fitting risk:** Test Pyramid/Trophy as guideline (e.g., many unit, fewer integration, fewest e2e; quadrants as helpful), not a rigid rule. For a data pipeline, integration may dominate; for a UI-heavy app, e2e may be higher — justify by risk. Include at least one concrete, unique example per type — no zero-example test plan.
2. **Select types needed:**
   - **Unit:** isolated, fast (<100ms), table-driven where relevant — tests a single behavior, not a class.
   - **Integration:** with real deps via testcontainers or fakes — tests interaction, not just logic.
   - **Contract:** Pact or equivalent where teams integrate — tests consumer/provider compatibility.
   - **E2e:** user journeys via Page Object/screen object, stable selectors, deterministic data — tests the deployed system.
   - **Performance:** load/stress/soak/soak for SLOs (p95, error rate <1%) with specific load profiles, not "cutting-edge performance."
   - **Security:** SAST/DAST per OWASP Testing Guide, dependency scanning.
   - **Accessibility:** WCAG 2.2 where user-facing.
   Include only those relevant to the project's risk. For each, define a concrete, unique example and the data it needs.
3. **Define gates that are actually enforced:** coverage targets (e.g., line ≥80%, branch ≥70% as team standard), flake budget (<1% or 1/100 runs), accessibility thresholds (WCAG AA). Quantify, do not hedge with "in today's fast-paced world." Define what happens when the gate fails (block merge vs warn).
4. **Choose frameworks available to project** (Vitest/Jest, pytest, go test, JUnit, Playwright/Cypress, k6/JMeter, ZAP or equivalents — pick by ecosystem and performance, not preference). Choose one per language and stick to it unless there is a specific gap.
5. **Plan CI for speed and stability:** parallelize by file or shard, cache deps and containers, retry with quarantine for flakes (quarantine → issue link, not silent retry), and affected-test detection (only run changed + dependent tests). Avoid repetitive generic justifications — name the specific job and time saved: "shard e2e into 4 jobs, save 6 min per run."

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Implement Suites That Are Fast, Isolated, and Deterministic
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Implement suites in project's test directories** (use a root-level `tools/*.py` helper only when helper automation is needed — otherwise use `tests/`, `__tests__/`, or `src/**/*.test.*`):
   - **Unit:** table-driven, isolated, <100ms, no network or file I/O. One behavior per test, name reveals intent: `rejects expired token`, not `test1`.
   - **Integration:** with testcontainers or fakes; contract tests where teams integrate; clean data per test (transaction rollback or truncate).
   - **E2e:** Page Object or screen object, stable selectors (`data-testid`), deterministic data (seeded, not random), isolated per test (new user per e2e).
   - **Performance:** scenarios for SLOs (p95, error rate) with specific load profiles and assertions, not "cutting-edge performance." Ramp: 10 → 100 → 500 RPS.
   - **Security:** OWASP checklist and scans (ZAP) with reproducible steps.
2. **Seed fixtures, factories, and deterministic data.** Use factories (e.g., `makeUser({overrides})`) and fixed seeds, not random data that hides failures. Mask PII. Document procedures in tracked documentation or RuntimeContextInterface; internal refs are unsupported.
3. **Replace generic hype** ("state-of-the-art test suite") with measurable scope: "42 unit, 12 integration (testcontainers), 5 e2e (Playwright), k6 p95 < 200ms at 100 RPS, ZAP no HIGH."

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Run, Measure, and Quarantine Honestly
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Run the full gate from real tool output:** lint + typecheck + test + coverage + mutation if available. Report the tool's actual numbers — never invent coverage (e.g., "87.3% line, 72% branch from c8") or latency stats; verify every link resolves. Fail the build when the gate is red — do not adjust numbers to pass.
2. **Validate performance SLOs and security scans;** check accessibility where applicable (axe, Lighthouse). For perf, assert on p95 and error rate, not just "no errors."
3. **Regression and flake handling:** run affected tests; quarantine flakes with issue links and a due date — do not silently retry and hide them. Mark quarantined tests with `// quarantine: #123, flake rate 2/50, due 2026-09-10`.
4. **Scan for slop in test code itself:** repetitive structures ("It's not X; it's Y") and hypophoras in test names ("And honestly?") — rewrite to plain intent (`should reject expired token`). Summarize AI's logic back to yourself to catch nonsensical cause-effect before merging.
5. **Report coverage, flaky rate, and defect log with first-party data.** Keep AI-assisted drafting to ideation — rewrite the final report in your own precise language after a break.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Make Quality Visible
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

1. **Publish report and traceability (Req → Test)** in chat, tracked project documentation, or RuntimeContextInterface. Internal refs are unsupported; stop and report if required. Communicate release readiness with specific measured evidence. Do not write a generic conclusion.
2. **Propose improvements with owners:** gaps (e.g., "missing contract test for {consumer} → {provider}, owner @jane, due 2026-09-10"), tooling (shard e2e), and data (seed performance dataset).
3. **Update knowledge base with unique, project-specific examples** — not generic lessons. Example: "flaky e2e for checkout was isolated fixture shared between tests; fixed by per-test user factory."

## Core Concepts

- **Pyramid vs Trophy vs Diamond:** Pyramid optimizes for speed and isolation; Trophy adds integration and static; Diamond emphasizes contract — choose by where defects actually escape.
- **Testcontainers vs mocks vs fakes:** Testcontainers for real integration, mocks for interaction verification, fakes for speed — use the lightest that still catches the risk.
- **Flake budget:** A flake rate >1% means the suite is unreliable — quarantine, fix, or delete, do not ignore.
- **Deterministic data:** Seeded factories beat random data; random hides failures and makes reproduction impossible.

## Best Practices

1. **One behavior, one test.** A test that asserts three things will be skipped when one fails.
2. **Fast and isolated.** A unit test that hits the network is an integration test — rename and move it.
3. **Stable selectors.** Use `data-testid`, not CSS or text that changes with copy.
4. **Quarantine, don't ignore.** A flaky test that is retried silently is a lying test — mark it and file an issue.

## Common Patterns

### Pattern: Table-Driven Unit (Go)
```go
tests := []struct{name string; token string; want error}{
  {"expired", expiredToken, ErrExpired},
  {"invalid", "bad", ErrInvalid},
}
for _, tc := range tests {
  t.Run(tc.name, func(t *testing.T) {
    err := Validate(tc.token)
    if !errors.Is(err, tc.want) { t.Fatalf("got %v, want %v", err, tc.want) }
  })
}
```

### Pattern: Factory with Overrides (TS)
```ts
const makeUser = (overrides = {}) => ({
  id: "u-" + nanoid(),
  email: "test+" + nanoid() + "@example.com",
  role: "user",
  ...overrides,
})
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Suite takes >10 min | No sharding, no cache, too many e2e | Shard by file, cache containers, move some e2e to integration |
| Flakes 2-3 per run | Shared mutable fixture or random data | Isolate per test, seed deterministically, reset state via transaction |
| Coverage drops but no failures | New code without tests or uncovered branches | Add unit for new branch, gate in CI, not locally only |
| E2e fails only in CI | Timing or selector fragility | Use `data-testid`, wait for network idle, not fixed timeout |

## Standards

- **ISTQB / ISO 29119**: test process — use as checklist, adapt to team.
- **Test Pyramid / Trophy / Diamond**: distribution — adapt to risk and defect history.
- **TDD / BDD (Gherkin)**: development flow — use where it helps design and communication.
- **OWASP Testing Guide / ASVS**: security — apply level by risk.
- **WCAG 2.2**: accessibility — include when user-facing, test with axe.
- **SLO/SLI**: performance gates — define per NFR, measure at the same point as prod.

## Checklists

### Pre-Execution
- [ ] Acceptance criteria and risk areas identified from real sources, not synthetic
- [ ] Environment and fixtures ready, PII masked, deterministic seeding chosen
- [ ] Frameworks chosen by project ecosystem — not assumed, not duplicated

### Post-Execution
- [ ] Distribution fits risk, coverage gates met with tool output — not invented numbers
- [ ] CI green with parallelization, sharding, retries, and flake quarantine (with issue links)
- [ ] Performance and security scans passed where needed — reports verified, links resolve
- [ ] Traceability (Req → Test) updated, no generic intro/conclusion, no empty buzzwords, no hallucinated stats

## Adaptive Modes

| Request Class | Testing Response |
|---------------|-----------------|
| **Clear** | Execute the test strategy with the chosen distribution |
| **Materially ambiguous** | Clarify scope, risk areas, and critical paths before testing |
| **High-risk** | Increase coverage depth, add security and performance scans |
| **Blocked/unsupported** | Report the missing capability; do not fabricate test results |

## Anti-Patterns to Avoid

- Ice-cream cone (many e2e, few unit) or flaky, slow suites that teach the team to ignore CI.
- Tests without assertions or with mocked system under test (tests that cannot fail).
- Mass-produced test boilerplate that repeats the same three verbs and no unique data; tests that share mutable state.

## Example

- "Create strategy for {service} with coverage gate 80% and contract + e2e + perf validation at 100 RPS, p95 < 200ms, ZAP 0 HIGH."
- "Add unit and integration tests for {feature} with factories, testcontainers, and CI sharding."
- "Audit {existing suite} for pyramid balance and flake rate; produce findings and quarantine plan."

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |

- **Bundle contents:** `SKILL.md`; `rules/` (2 items); `tools/` (none); `resources/` (none); `refs/` (none).
