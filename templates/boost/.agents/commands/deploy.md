---
name: deploy
description: Deploy and release with progressive delivery — DORA/SRE/GitOps. Use when deploy, release, CI/CD, pipeline, IaC, or infra is mentioned.
---

# Deploy Command — Artifact to Safe Release

Generic deployment command for any platform. Turns a built artifact into a progressive, observable, and reversible release. Adapts strategy to risk, compliance, and available tooling. Works with any AI agent.

Follow `skills/deploying/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Keep every promotion decision auditable and data-driven.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for artifact, environment, strategy, and rollback criteria before proceeding. Do not assume a default env.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Review scope, schedule, compliance needs, and environments. Map dev → staging → prod (or team-specific envs), regions, and dependencies. List who can approve promotion and what compliance requires. Use measured DORA data (lead time, frequency, MTTR, change failure rate) if tracked — do not invent benchmarks. Check SLOs/SLIs, error budgets, and on-call expectations. Replace vague "fast-paced world" with actual traffic ("500 RPS, 20% canary is 100 RPS") and budget.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Choose strategy fitting risk: canary (e.g., 5%→25%→100%), blue-green, rolling, or direct. For a low-risk internal tool, direct may be fine; for a payment service, canary with automated analysis is warranted — justify with blast radius and SLO impact. Define promotion gates and approval steps with specific thresholds ("promote if error rate < 0.5% and p95 < 200ms for 15 min"), not generic lists. State a point of view — do not hedge to the center. Define pipeline: build → test → lint/typecheck → security scan → push artifact → deploy to staging → smoke/e2e → deploy to prod (progressive) → smoke → monitor. Choose tools available to project: CI, IaC, containers, GitOps. Plan versioning, SBOM/signing, and audit. Define rollback criteria (error rate, latency, health checks) with quantified limits and runbooks. Avoid meaningless phrases like "unlocking creative freedom" — name exact metric and command.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Build pipeline with caching, parallelization, and OIDC for secrets where available. Keep YAML small, reviewed, and DRY — do not mass-produce repetitive jobs that copy-paste the same three steps. Provision infra via IaC with plan/apply gates. Build multi-arch images where needed, SBOM attached, signed with cosign, provenance attested. Deploy with progressive delivery (Argo Rollouts/Flagger or equivalent) and wire automated analysis to promotion. Keep root-level tools only if a helper is needed; configure observability with metrics, logs, traces, alerts, dashboards and specific SLI examples, not "cutting-edge observability."

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Validate staging: smoke + e2e + security scan; check health endpoints (`/healthz`, `/readyz`), version endpoint, and feature flag state. Verify every artifact SHA and SBOM link resolves — never hallucinate a URL, SHA, or provenance; copy the CI output verbatim. Analyze canary metrics against SLOs for the full bake time. Check provenance, SBOM, and audit logs; rehearse rollback. Scan for repetitive structures and hypophoras in runbooks — rewrite into direct commands. Summarize AI's logic back to yourself to catch nonsensical cause-effect before promoting.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Publish deployment record in chat, tracked project documentation, or RuntimeContextInterface. Do not create an internal ref artifact; stop and report if one is required. Include who/what/when, artifact SHA, SBOM link, provenance, approvals, and measured canary metrics. Update runbooks and DORA-style metrics with measured data, not invented numbers. Hand off to maintenance with explicit monitoring window and on-call handoff. If AI drafted the runbook, keep at most a few hundred words of its clearest phrasing and rewrite the rest with human judgment after a break. Schedule post-deploy review.
## Core Concepts

- **Canary vs Blue-Green vs Rolling:** Canary limits blast radius and tests in prod with real users; blue-green gives instant switchover but doubles infra; rolling is simple but not instant rollback — choose by risk and infra cost.
- **GitOps vs Push-based:** GitOps reconciles declared state (good for audit); push-based deploys on CI success (good for speed) — choose by compliance.
- **SLSA Levels:** Level 1 is provenance, Level 2 is signed provenance, Level 3 is hardened builder — pick by supply-chain risk.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{artifact}}` | string | Yes | — | Image tag, version (`v2.3.0`), or path — the artifact to release |
| `{{env}}` | string | Yes | — | Environment: `staging`, `prod`, `dev`, or team env |
| `--strategy` | enum | No | `canary` | Strategy: `canary`, `blue-green`, `rolling`, `direct` |
| `--canary` | string | No | `5%` | Canary steps: `5%` or `5,25,100%` with bake times |
| `--rollback` | string | No | `auto` | Rollback: `auto` on SLO breach or `manual` |
| `--dry-run` | flag | No | false | Render manifests without applying — show plan |

## Examples

```bash
/deploy "{artifact} to {env} with canary 5%→25%→100% and SLO-based rollback (error rate <0.5% and p95 <200ms for 15m)"
/deploy "{artifact} to {env} via GitOps with SBOM signing"
/deploy --dry-run "{artifact} to {env}"
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Canary never promotes | Analysis template too strict or metric query wrong | Check Prometheus query, lower threshold, extend bake, verify SLI |
| Rollback takes >10 min | No rehearsal, large image, no runbook | Rehearse `helm rollback`, keep image warm, write copy-paste runbook |
| SBOM missing in prod | Pipeline not attaching or registry not storing | Add syft + cosign to pipeline, verify in admission controller |

## Standards

DORA (4 keys, capabilities), SRE (SLO/SLI, error budgets), 12-Factor, GitOps (OpenGitOps), SLSA/SBOM (SPDX/CycloneDX), ISO 27001/SOC2.

## Validation

- [ ] Staging green, SBOM/signed provenance published and verified (links resolve, signatures valid), canary within SLO for full bake time, promotion with evidence and timestamp, rollback tested, no generic conclusion, no invented stats

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
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `supervisor` | Collaborates on runtime health and execution controls | Route work when the deliverable requires runtime health and execution controls. |
## Handoff

- **From previous phase:** Requires versioned input with traceability and quantified NFRs/SLOs — do not start without it.
- **To next phase:** Delivers versioned artifacts with explicit non-goals, open spikes, owners, and due dates. No TBD without an owner.

## Metrics

- Track phase-specific metrics with tool output, not invented numbers: scope stability, estimation accuracy, design review throughput, build lead time, test flake rate, deployment frequency, MTTR.


## Anti-Patterns — Phase-Specific

- Generic hedging, repetitive "It's not X; it's Y" rhetoric, and formulaic three-item lists that obscure the trade-off.
- Mass-produced boilerplate with no unique example per item; zero-example artifacts.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
