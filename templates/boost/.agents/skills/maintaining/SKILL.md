---
name: maintaining
description: Use when handling incidents, patching, or reducing tech debt. Triggers on maintain, maintenance, incident, on-call, MTTR, hotfix, patch, debt, observability, postmortem.
---

# Maintaining Skill

Comprehensive, project-agnostic operations and continuous improvement for any service, stack, or team. Adapts severity handling to available observability and on-call setup. This skill is for **sustaining** — it does not plan new scope or design new architecture; it keeps the running system healthy, secure, and evolvable.

## When to Use

- Handling incidents, alerts, or SLO breaches of any kind (latency, errors, saturation, business metric)
- Planning bug fixes, patches, dependency updates, or tech debt reduction
- Improving observability, capacity, or security posture for a running system
- Conducting postmortems, runbook updates, and preventive actions after an incident
- Auditing debt, vulnerabilities, or on-call health before they become incidents

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`. Prioritize verifiable recovery and learning over plausible prose. Every incident is a learning budget.

### Understand — Observe and Triage Truthfully
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Monitor golden signals plus business signals:** latency (p50/p95/p99), traffic (RPS), errors (rate and type), saturation (CPU/memory/queue lag/disk), plus business metrics (checkouts, signups, jobs). Triage by severity/impact/urgency (e.g., SEV1-4 via ITIL matrix) and track MTTR (detect → mitigate) and MTBF with real incident data — do not invent MTTR numbers like "MTTR is 15 min" without a timestamped incident. Review correlated deploys, infra changes, feature flags, and user feedback. Avoid vague "quietly in the background" phrasing — name the exact service, queue, and metric that is saturated: "worker queue `checkout` lag 12k, p95 4.2s, error rate 2.3% for 18 min."
2. **Assess debt via first-party scans, not generic lists.** Vulnerability scans (Dependabot, npm audit, govulncheck, Trivy), code health (complexity, duplication, test gaps), and operational debt (manual runbooks, missing alerts). Note PII and compliance boundaries (GDPR, retention, data residency) that constrain fixes.
3. **Map on-call and communication.** Who is on call, how are they paged (PagerDuty/OpsGenie or equivalent), and what is the status communication channel? Capture current runbook version and dashboard links — verify each link resolves.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — Prioritize by Impact and Budget Risk
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Prioritize by impact × urgency, not arrival order.** SEV1 (user-facing, revenue-impacting) before SEV3 (internal). Define burn-rate alerts for SLOs (e.g., 5x burn for 5 min → page). State a point of view on what to fix first; do not hedge to the center when a SEV1 is burning.
2. **Plan fix shape and window:** hotfix branch (immediate, minimal, canary), patch window (scheduled, broader, with staging), or scheduled debt sprint (20% capacity as guideline, adjust to team). Plan communication: status page, stakeholder update cadence, and internal war-room channel. Include explicit rollback and escalation steps.
3. **Schedule security patches by actual risk:** e.g., CVSS ≥7 with exploit in the wild within 7 days, CVSS ≥9 within 24h — adapt to org policy, do not invent a policy. Schedule dependency updates (weekly) and debt sprints with owners and due dates, not vague "soon."
4. **Forecast capacity where it matters:** via queuing theory or Little's Law for queues, or simple burn-rate for SLOs. Plan autoscaling, retention, and budgeting with specific numbers, not "scalable."

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — Fix Small, Deploy Safely, Harden
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Execute via the smallest safe path:**
   - **Hotfix:** `fix/{incident-id}-{short-name}` branch (e.g., `fix/SEV1-checkout-lag`) → add reproduction test or metric that would have caught the issue → minimal code change → test in staging with production-like data → canary to prod (5% → 25% → 100%) → monitor. Provide unique, project-specific evidence for each change, not a formulaic list of three verbs.
   - **Patch:** schedule window, apply, test in staging, roll out with a root-level `tools/*.py` helper only if needed. Include forward and rollback steps in the PR description.
   - **Debt:** refactor with tests, add observability (new metric or alert), update docs, add runbook entry. Pay down debt that is on the critical path or causing flakes — not all debt is equal.
2. **Harden as you fix:** rotate secrets if they were in scope, update SBOM where needed, backport fixes to supported branches, and add a test that would have caught the issue. Replace generic hype ("cutting-edge fix") with the specific commit SHA, config change, and metric that proves it worked: "fix: increase checkout worker concurrency from 10 to 20, commit abc123, p95 4.2s → 180ms."
3. **Improve observability with specificity:** new metric `checkout_queue_lag`, alert `lag > 5k for 5m`, dashboard panel with SLI. Not "improved observability" without an artifact.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — Prove Recovery with Data
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Reproduce → fix → regression and e2e;** monitor golden signals plus business signals for a defined window (e.g., 30 min at 5%, 30 min at 25%) before promoting. Re-scan CVEs with the scanner's actual output — do not invent a CVE score or claim "no vulnerabilities" without a scan. Obtain stakeholder sign-off with a timestamp.
2. **Confirm no regression** in error rate, latency, and business metrics with measured data compared to pre-incident baseline. Verify every link and stat — never hallucinate a CVE score, dashboard URL, or MTTR. Scan for repetitive structures ("It's not X; it's Y") and hypophoras mid-postmortem — rewrite into direct statements. Summarize AI's logic back to yourself to catch nonsensical cause-effect before closing the incident.
3. **Validate that the fix would have been caught earlier** by the new test or alert. If not, add the missing detection as part of the same incident.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — Learn and Prevent
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

1. **Write the blameless postmortem** in chat, tracked project documentation, or RuntimeContextInterface; internal refs are unsupported, so stop and report if required. Use Summary → Impact → Timeline → Root Cause → Detection → Mitigation → Follow-ups. Update runbooks with exact mitigating commands. Keep AI-assisted drafting to ideation; rewrite the final postmortem in precise language after a break. Emit MTTR/MTBF, change failure rate, and debt burndown with tool output, not invented numbers.
2. **Log preventive action with owner and date.** Example: "add alert `checkout_queue_lag > 5k for 5m`, owner @jane, due 2026-09-03; add e2e for checkout with lag injection, owner @joe, due 2026-09-10." Not "improve monitoring."
3. **Update knowledge base and on-call health:** runbook version, dashboard link, and lessons learned with unique, project-specific detail. Schedule a review of the follow-ups.

## Core Concepts

- **MTTR vs MTBF vs MTTD:** MTTR is detect→mitigate, MTBF is time between failures, MTTD is detect lag — measure all three from incident timestamps, not memory.
- **Error budget vs SLO:** Error budget is 1 - SLO over a window; burn rate is how fast you are spending it — use it to prioritize debt vs features.
- **Blameless postmortem:** No blame, but full accountability — name the system, not the person, and assign follow-ups with owners.
- **Debt vs defect vs incident:** Debt is a future cost, defect is a present bug, incident is user impact — triage differently.

## Best Practices

1. **Small, reversible, canaried fixes.** A 20-line hotfix that can be rolled back in 5 min beats a 500-line patch that cannot.
2. **Detection before prevention.** If you cannot detect it, you cannot prevent it — add the alert first.
3. **Runbooks are code.** Version them, test them, and time them. A runbook that has never been run is fiction.
4. **Debt has interest.** Track debt that is on the critical path or causing flakes — pay that first.

## Common Patterns

### Pattern: Postmortem Follow-up
```markdown
Follow-up: Add alert `checkout_queue_lag > 5k for 5m`
Owner: @jane
Due: 2026-09-03
Verify: Simulate lag with 10k jobs, alert fires in <2 min
```

### Pattern: Hotfix Branch
```bash
git checkout -b fix/SEV1-checkout-lag
# add test that reproduces lag, then minimal fix, then canary
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| MTTR stays >60 min for SEV1 | No runbook, manual steps, no canary | Write copy-paste runbook, rehearse, use canary for hotfixes |
| Same incident repeats monthly | Follow-ups not owned or due | Assign owner and due date, track in backlog, review weekly |
| Alerts fire but no one acts | Noise, no severity, no runbook | Tune thresholds, add severity, link runbook, set on-call rotation |
| Debt grows without bound | No capacity reserved, no triage | Reserve 20% capacity, triage weekly, track debt on critical path |

## Standards

- **ITIL 4**: incident, problem, change — adapt process to team size, not ceremony.
- **SRE**: MTTR/MTBF, error budgets, blameless postmortems — measure from real incidents.
- **ISO 27001 / SOC2 / GDPR**: security and privacy — apply controls by data sensitivity.
- **SLSA / SSDF**: secure supply chain for patches — verify, not just attach.
- **OpenTelemetry / Prometheus / Grafana / ELK/Loki**: observability — pick stack available to project, instrument golden signals.

## Checklists

### Pre-Execution
- [ ] Alerts routed with severity and on-call defined — from real config, not assumed
- [ ] Runbooks and dashboards accessible and verified (links resolve, commands tested)
- [ ] SLOs and error budgets known or explicitly deferred with owner
- [ ] Recent incidents and debt reviewed with first-party data

### Post-Execution
- [ ] Fix verified in staging and prod SLOs restored with measured data (p95, error rate, business metric)
- [ ] Postmortem written with specific timeline, impact, and detection — no generic phrases, no empty buzzwords, no invented stats
- [ ] Runbooks updated with exact mitigation commands, SBOM refreshed where needed
- [ ] MTTR/MTBF tracked from incident timestamps (not memory), preventive action logged with owner/due, debt reprioritized

## Adaptive Modes

| Request Class | Maintaining Response |
|---------------|---------------------|
| **Clear** | Triage, fix, and verify with measured recovery data |
| **Materially ambiguous** | Clarify service, severity (SEV1-4), and desired action before proceeding |
| **High-risk** | Add canary, rollback plan, and postmortem with timeline |
| **Blocked/unsupported** | Report the missing capability; do not fabricate incident data |

## Anti-Patterns to Avoid

- Blameful postmortems; fixes without tests or monitoring; patching without staging verification; ignoring debt until it blocks delivery.
- Mass-produced status updates that repeat the same template with no unique detail; alerts without runbooks or with excessive noise; "quietly in the background" as a diagnosis.

## Example

- "Triage {incident} with severity {SEV}, hotfix via `fix/{id}` branch, canary 5%→25%, and 5 Whys postmortem with follow-ups."
- "Reduce tech debt in {area} on critical path: add tests, metric {name}, and alert {threshold}."
- "Patch {CVE} with CVSS {score} — schedule window, canary, and verification scan."

## References — Phase-Specific

- Use ref shapes in `skills/customizing/SKILL.md` for postmortems and runbooks; keep the rule shape for gates.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

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

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
