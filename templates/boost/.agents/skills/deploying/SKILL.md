---
name: deploying
description: Use when setting up CI/CD, releasing, or provisioning infra. Triggers on deploy, deployment, CI/CD, pipeline, IaC, canary, blue-green, GitOps, SLSA, release, rollout.
---

# Deploying Skill

Comprehensive, project-agnostic release and operations for any platform, pipeline, or infra. Adapts strategy to risk, compliance, and available tooling. This skill is for **releasing** — it does not plan scope or write feature code; it takes a built artifact and makes it safely available, observable, and reversible.

## When to Use

- Creating or changing pipelines and infrastructure as code for any project (service, app, pipeline, model, or infra)
- Releasing to any environment (dev, staging, prod, regional, edge) with any progressive strategy (canary, blue-green, rolling, direct)
- Provisioning infra, managing containers/images, or configuring GitOps and progressive delivery
- Handling deployment failures, rollbacks, and compliance (SBOM, signing, audit)

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Keep every promotion decision auditable and data-driven. Prefer small, reversible releases over big-bang deploys.

### Understand — Map Environments and Risk
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Review scope, schedule, compliance needs, and environments.** Map dev → staging → prod (or team-specific envs), regions, and dependencies. List who can approve promotion and what compliance requires (change advisory, SOC2, audit log). Use measured DORA data (lead time, frequency, MTTR, change failure rate) if tracked — do not invent benchmarks like "we deploy 10x per day" without data.
2. **Check SLOs/SLIs, error budgets, and on-call expectations.** What is the SLO (e.g., 99.9% availability, p95 < 200ms)? What is the remaining error budget? Who is on call during and after the deploy? Replace vague "fast-paced world" context with the project's actual traffic ("500 RPS, 20% canary is 100 RPS") and budget ("$2k/mo infra cap").
3. **Map dependencies and blast radius.** Which services or data stores are affected? Which feature flags are involved? What is the rollback time if this fails? Name the exact services and flags.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Choose Strategy and Define Gates
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Choose strategy fitting risk, not hype:** canary (e.g., 5%→25%→100%), blue-green, rolling, or direct. For a low-risk internal tool, direct may be fine; for a payment service, canary with automated analysis is warranted — justify with blast radius and SLO impact. Define promotion gates and approval steps with specific thresholds ("promote if error rate < 0.5% and p95 < 200ms for 15 min"), not generic lists. State a point of view when choosing — do not hedge to the center.
2. **Define pipeline as stages with clear inputs/outputs:** build → test → lint/typecheck → security scan (SAST/SCA) → push artifact → deploy to staging → smoke/e2e → deploy to prod (progressive) → smoke → monitor. For each stage, define artifact, cache, and failure action. Keep stage names specific.
3. **Choose tools available to project:** CI (GitHub Actions/GitLab CI/Jenkins/CircleCI or equivalent), IaC (Terraform/OpenTofu/Pulumi), containers (Docker/Podman), orchestration (Kubernetes/Nomad/ECS or equivalent), GitOps (ArgoCD/Flux). Do not mandate a specific vendor — use what the team already runs. Document why a tool was chosen for this project.
4. **Plan versioning, supply chain, and audit:** versioning (SemVer + build metadata like `v2.3.0+sha.abc123`), SBOM (SPDX/CycloneDX via syft/trivy), signing (cosign/SLSA level 1-3), and audit requirements (who approved, when, with evidence). Define what "provenance" means for this project.
5. **Define rollback criteria and runbooks with numbers:** error rate >1% for 5 min, p95 >300ms, failed health checks (`/healthz` 5xx), or SLO burn rate >2x. Write runbooks as direct commands ("`kubectl rollout undo deploy/{service}` then `argocd app sync`"), not prose. Avoid meaningless phrases like "unlocking creative freedom" — name the exact metric and threshold and the command to run.
6. **Plan observability for the release:** which dashboards and alerts will be watched during the canary? Who will watch them and for how long? Define a explicit bake time (e.g., 15 min at 5%, 30 min at 25%).

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Implement Pipeline and Progressive Delivery
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Build pipeline with engineering quality:** caching (deps, layers, containers), parallelization (matrix per language or service), and OIDC for secrets where available (no long-lived tokens). Keep YAML small, reviewed, and DRY — do not mass-produce repetitive jobs that copy-paste the same three steps with different names. Lint pipeline YAML and test it in a branch.
2. **Provision infra via IaC with plan and apply gates.** Run `terraform plan` in CI, require approval for `apply` to prod. Manage state securely; no hardcoded secrets. Use a root-level `tools/*.py` helper only if needed; keep one-offs in `/tmp` and shared static data in `tools/tools.json`.
3. **Build artifacts with supply-chain rigor:** multi-arch images where needed, SBOM attached, signed with cosign, provenance attested. Tag with SemVer and SHA. Push to registry with retention policy.
4. **Deploy with progressive delivery:** Argo Rollouts/Flagger or equivalent (or simple canary via weighted service). Wire automated analysis (Kayenta/Flagger metrics or custom Prometheus queries) to promotion: if metrics breach SLO, auto-rollback. Keep manifests as code (Helm/Kustomize) and versioned.
5. **Configure observability for the release:** metrics (RED/USE), logs (Loki/ELK), traces (OpenTelemetry or equivalent), alerts (burn rate, error rate), and dashboards with specific SLI examples, not "cutting-edge observability." Example: "alert if `http_server_requests_seconds_count{status=~"5.."}` > 1% for 5m."

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Validate Progressively and Auditably
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Validate staging like prod:** smoke + e2e + security scan; check health endpoints (`/healthz`, `/readyz`), version endpoint, and feature flag state. Verify every artifact SHA and SBOM link resolves — never hallucinate a URL, SHA, or provenance; copy the CI output verbatim.
2. **Analyze canary metrics against SLOs** (error rate, latency, saturation) for the full bake time. If metrics are inconclusive, extend bake, do not promote on gut. Check provenance, SBOM, and audit logs; ensure each promotion has an approver and timestamp.
3. **Rehearse rollback before you need it:** `helm rollback` or `argocd app rollback` or `kubectl rollout undo` tested in staging with timing measured. Ensure runbook commands are copy-pasteable and have been run at least once. Scan for repetitive structures ("It's not X; it's Y") and hypophoras in runbooks — rewrite into direct commands.
4. **Confirm compliance and approvals:** change ticket, approver, and evidence attached. Summarize AI's logic back to yourself to catch nonsensical cause-effect before promoting — e.g., "because we use canary, latency will improve" is not causal.
5. **Take a break before the promoting decision** — fresh eyes catch slop the first pass missed.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Make Release Auditable
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

1. **Publish deployment record** in chat, tracked project documentation, or RuntimeContextInterface. Internal refs are unsupported; stop and report if required. Include who/what/when, artifact SHA, SBOM link, provenance, approvals, and measured canary metrics. Update runbooks and DORA-style metrics with measured data, not invented numbers. Hand off to maintenance with explicit monitoring window and on-call handoff. If AI drafted the runbook, keep at most a few hundred words of its clearest phrasing and rewrite the rest with human judgment.
2. **Schedule post-deploy review and error-budget check.** Log at least one improvement: faster rollback, better canary analysis, or reduced pipeline time.
3. **Define next release criteria** and what is explicitly out of scope for this deployment.

## Core Concepts

- **Canary vs Blue-Green vs Rolling:** Canary limits blast radius and tests in prod with real users; blue-green gives instant switchover but doubles infra; rolling is simple but not instant rollback — choose by risk and infra cost.
- **GitOps vs Push-based:** GitOps reconciles declared state (good for audit and drift); push-based deploys on CI success (good for speed) — choose by compliance and team maturity.
- **SLSA Levels:** Level 1 is provenance, Level 2 is signed provenance, Level 3 is hardened builder — pick by supply-chain risk and compliance.
- **Progressive delivery:** Promotion is not time-based, it is metric-based — promote when SLOs are met, not when the clock says so.

## Best Practices

1. **Small, frequent, reversible deploys.** A 5% canary that fails is a win; a 100% deploy that fails is an incident.
2. **Health checks are not optional.** `/healthz` for process, `/readyz` for dependencies — both must be distinct.
3. **SBOM and signing are not theater.** Attach them, verify them in admission, and alert when they are missing.
4. **Runbooks are code.** Test them, version them, and time them.

## Common Patterns

### Pattern: Canary Steps (Argo Rollouts)
```yaml
strategy:
  canary:
    steps:
    - setWeight: 5
    - pause: {duration: 15m}
    - analysis: {templates: [{templateName: error-rate}]}
    - setWeight: 25
    - pause: {duration: 30m}
    - analysis: {templates: [{templateName: latency}]}
```

### Pattern: Health Endpoints
```yaml
livenessProbe: {httpGet: {path: /healthz, port: 8080}, periodSeconds: 10}
readinessProbe: {httpGet: {path: /readyz, port: 8080}, periodSeconds: 5}
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Canary never promotes | Analysis template too strict or metric query wrong | Check Prometheus query, lower threshold, extend bake, verify SLI |
| Rollback takes >10 min | No rehearsal, large image, no runbook | Rehearse `helm rollback`, keep image warm, write copy-paste runbook |
| SBOM missing in prod | Pipeline not attaching or registry not storing | Add syft + cosign to pipeline, verify in admission controller |
| Pipeline slow (30m+) | No cache, sequential jobs | Cache deps/layers, parallelize per service, use matrix |

## Standards

- **DORA** (4 keys, capabilities) — measure and improve, not just report.
- **SRE** (SLO/SLI, error budgets, toil) — adapt to service criticality; burn rate alerts, not static thresholds.
- **12-Factor / GitOps (OpenGitOps)** — build/release/run, declarative reconciliation — adapt to hosting model.
- **SLSA / SBOM (SPDX/CycloneDX)** — supply chain — apply by compliance need; verify, don't just attach.
- **ISO 27001 / SOC2** — security and compliance — include where required, with evidence.

## Checklists

### Pre-Execution
- [ ] SLOs and rollback criteria defined with specific numbers, owners, and runbook commands
- [ ] IaC reviewed, secrets via vault/OIDC, no hardcoded credentials, state secured
- [ ] Pipeline with caching and parallelization, YAML linted, no mass-produced duplication
- [ ] Canary steps and analysis templates defined and tested in staging

### Post-Execution
- [ ] Staging green, SBOM/signed provenance published and verified (links resolve, signatures valid)
- [ ] Canary within SLO for full bake time, promotion automated or approved with evidence and timestamp
- [ ] Rollback tested and timed, runbooks updated, metrics emitted — no generic conclusion, no invented stats
- [ ] Deployment record published with SHA, SBOM, approver, and metrics

## Adaptive Modes

| Request Class | Deploying Response |
|---------------|-------------------|
| **Clear** | Execute the deployment strategy with progressive delivery |
| **Materially ambiguous** | Clarify artifact, environment, strategy, and rollback criteria before deploying |
| **High-risk** | Reduce blast radius, add canary, require rollback plan and approval |
| **Blocked/unsupported** | Report the missing capability; do not deploy without rollback plan |

## Anti-Patterns to Avoid

- Long-lived branches without integration; manual, unrepeatable deploys; no health checks or observability; deploys without rollback plan.
- Hardcoded secrets, unaudited artifacts, or "it worked on my machine" as verification.
- Hype-driven "state-of-the-art" claims without measurement; big-bang deploys that maximize blast radius.

## Example

- "Release {artifact} to {env} with canary 5%→25%→100% and SLO-based rollback (error rate <0.5% and p95 <200ms for 15m)."
- "Set up pipeline for {project} with test, scan, SBOM signing, and GitOps deploy via ArgoCD."
- "Audit {pipeline} for DORA and SLSA; produce findings and improvement plan."

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `supervisor` | Collaborates on runtime health and execution controls | Route work when the deliverable requires runtime health and execution controls. |

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
