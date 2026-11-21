---
name: maintain
description: Maintain, triage incidents, and reduce debt — ITIL/SRE. Use when maintain, incident, MTTR, hotfix, patch, debt, or observability is mentioned.
---

# Maintain Command — Sustain and Learn

Generic maintenance command for any service. Turns incidents, alerts, and debt into small, safe fixes and learning. Adapts severity handling to available observability and on-call setup. Works with any AI agent.

Follow `skills/maintaining/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Prioritize verifiable recovery and learning over plausible prose.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for service, severity (SEV1-4), and desired action (triage/hotfix/patch/debt/observe) before proceeding.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Monitor golden signals: latency (p50/p95/p99), traffic (RPS), errors (rate and type), saturation (CPU/memory/queue lag/disk), plus business metrics. Triage by severity/impact/urgency (e.g., SEV1-4 via ITIL matrix) and track MTTR (detect → mitigate) and MTBF with real incident data — do not invent MTTR numbers like "MTTR is 15 min" without a timestamped incident. Review correlated deploys, infra changes, feature flags, and user feedback. Avoid vague "quietly in the background" phrasing — name the exact service, queue, and metric that is saturated: "worker queue `checkout` lag 12k, p95 4.2s, error rate 2.3% for 18 min." Assess debt via vulnerability scans and code health checks. Map on-call and communication: who is on call, how are they paged, and what is the status channel?

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Prioritize by impact × urgency, not arrival order. SEV1 (user-facing, revenue-impacting) before SEV3 (internal). Define burn-rate alerts for SLOs (e.g., 5x burn for 5 min → page). State a point of view on what to fix first; do not hedge to the center when a SEV1 is burning. Plan fix shape and window: hotfix branch (immediate, minimal, canary), patch window (scheduled, broader, with staging), or scheduled debt sprint (20% capacity as guideline). Plan communication: status page, stakeholder update cadence, and war-room channel. Include explicit rollback and escalation steps. Schedule security patches by actual risk: e.g., CVSS ≥7 with exploit in the wild within 7 days, CVSS ≥9 within 24h — adapt to org policy.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Execute via the smallest safe path: `fix/{incident-id}-{short-name}` branch (e.g., `fix/SEV1-checkout-lag`) → add reproduction test or metric that would have caught the issue → minimal code change → test in staging with production-like data → canary to prod (5% → 25% → 100%) → monitor. Provide unique, project-specific evidence for each change, not a formulaic list of three verbs. Apply patches, update dependencies, refactor with tests, add observability (new metric or alert), rotate secrets, refresh SBOM where needed. Replace generic hype ("cutting-edge fix") with the specific commit SHA, config change, and metric that proves it worked: "fix: increase checkout worker concurrency from 10 to 20, commit abc123, p95 4.2s → 180ms." Use a root-level `tools/*.py` helper only if needed; keep one-offs in `/tmp`.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Reproduce → fix → regression and e2e; monitor golden signals plus business signals for a defined window (e.g., 30 min at 5%, 30 min at 25%) before promoting. Re-scan CVEs with the scanner's actual output — do not invent a CVE score or claim "no vulnerabilities" without a scan. Obtain stakeholder sign-off with a timestamp. Confirm no regression in error rate, latency, and business metrics with measured data compared to pre-incident baseline. Verify every link and stat — never hallucinate a CVE score, dashboard URL, or MTTR. Scan for repetitive structures ("It's not X; it's Y") and hypophoras mid-postmortem — rewrite into direct statements. Summarize AI's logic back to yourself to catch nonsensical cause-effect before closing the incident.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Write the blameless postmortem in chat, tracked project documentation, or RuntimeContextInterface; internal refs are unsupported, so stop and report if one is required. Use first-party timeline, Summary → Impact → Timeline → Root Cause → Detection → Mitigation → Follow-ups. Update runbooks with exact mitigating commands. Keep AI-assisted drafting to ideation; rewrite the final postmortem in precise language after a break. Emit MTTR/MTBF, change failure rate, and debt burndown with tool output, not invented numbers. Log preventive action with owner and date.
## Core Concepts

- **MTTR vs MTBF vs MTTD:** MTTR is detect→mitigate, MTBF is time between failures, MTTD is detect lag — measure all three from incident timestamps, not memory.
- **Error budget vs SLO:** Error budget is 1 - SLO over a window; burn rate is how fast you are spending it — use it to prioritize debt vs features.
- **Blameless postmortem:** No blame, but full accountability — name the system, not the person, and assign follow-ups with owners.
- **Debt vs defect vs incident:** Debt is a future cost, defect is a present bug, incident is user impact — triage differently.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | Yes | — | Service, alert, or SEV id — the incident or area to maintain |
| `{{action}}` | string | No | `triage` | Action: `triage`, `hotfix`, `patch`, `debt`, `observe` |
| `--severity` | enum | No | `SEV3` | Severity: `SEV1`, `SEV2`, `SEV3`, `SEV4` — per org matrix |
| `--window` | string | No | `now` | Maintenance window or `now` for hotfix |
| `--dry-run` | flag | No | false | Plan without applying — show triage and follow-ups |

## Examples

```bash
/maintain "{incident} with severity {SEV} — triage, hotfix via fix/{id} branch, canary, and 5 Whys postmortem"
/maintain "{service} debt scan and patch — schedule window, canary, verification scan"
/maintain "{service} observability improvement — add metric {name} and alert {threshold}"
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| MTTR stays >60 min for SEV1 | No runbook, manual steps, no canary | Write copy-paste runbook, rehearse, use canary for hotfixes |
| Same incident repeats monthly | Follow-ups not owned or due | Assign owner and due date, track in backlog, review weekly |
| Alerts fire but no one acts | Noise, no severity, no runbook | Tune thresholds, add severity, link runbook, set on-call rotation |

## Standards

ITIL 4, SRE (MTTR/MTBF, error budgets, blameless postmortems), ISO 27001/SOC2/GDPR, SLSA/SSDF, OpenTelemetry/Prometheus/Grafana/ELK.

## Validation

- [ ] MTTR tracked from incident timestamps (not memory), postmortem with specific timeline and impact, fix verified in staging and prod SLOs restored with measured data, runbooks updated, no generic phrases, no empty buzzwords, no hallucinated stats

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
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `automator` | Collaborates on automation and tooling | Route work when the deliverable requires automation and tooling. |
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
