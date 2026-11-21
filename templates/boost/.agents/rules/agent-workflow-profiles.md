---
id: rule-agent-workflow-profiles
scope: global
severity: mandatory
enforcement: self-check
related: [AGENTS.md, templates/agent-template.md, templates/skill-template.md]
---

# Agent Workflow Profiles — Role-Specific Phase Deliverables

## Metadata

- **ID:** `rule-agent-workflow-profiles`
- **Scope:** `global`
- **Severity:** `mandatory`
- **Enforcement:** `self-check`
- **Related:** `AGENTS.md`, `templates/agent-template.md`, `templates/skill-template.md`

## Description

Every triad follows `Understand → Plan → Build → Verify → Summarize`, but phase deliverables must
fit the triad's responsibility. A report-only, reasoning-only, coordination, tracking, planning,
or documentation agent must not be measured against implementation-only outputs such as production
code, migrations, coverage, or a deployable increment.

## Trigger

Apply when an agent, command, or skill defines phase deliverables, quality gates, or handoff content.

## Action

Select one workflow profile before execution and apply its profile-specific Build, Verify, and
Summarize outputs while preserving the five phase names.

| Profile | Applies to | Build output | Verify gate | Summarize output |
|---------|------------|--------------|-------------|------------------|
| **Delivery** | `builder`, `decomposer`, `deployer`, `maintainer`, `automator` | Code, configuration, migration, runbook, or tool increment | Tests, lint, security, rollback, and runtime evidence appropriate to the change | Changed files, test output, migration/rollback notes, and handoff |
| **Planning/design** | `planner`, `designer`, `manager`, `leader` | Requirements, design, decision, packet, or graph artifact | Traceability, measurable criteria, scope/authority, and dependency checks | Approved design or packet, open decisions, owners, and next gate |
| **Analysis/audit** | `auditor`, `explorer`, `researcher`, `thinker` | Findings, evidence map, alternatives, or recommendation | Source traceability, reasoning integrity, severity/uncertainty, and false-positive review | Finding/recommendation set, residual uncertainty, and decision input |
| **Quality** | `tester` | Test strategy, tests, reports, or quality gate result | Behavior evidence, flake diagnosis, coverage when measurable, and risk coverage | Test evidence, unresolved failures, and release recommendation |
| **Coordination** | `orchestrator`, `supervisor` | Execution graph, dispatch record, checkpoint, or runtime status | Dependency, authorization, timeout, retry, and escalation controls | Graph/status evidence, blockers, and next control decision |
| **Tracking/documentation** | `tracker`, `writer`, `housekeeper`, `customizer`, `learner` | Issue action, document, store hygiene result, asset, or learning signal | Link/inventory/registration/format and source-trace checks | URL or artifact, evidence, follow-up, and ownership |

## Validation

- The agent names its selected workflow profile before Build.
- Build output matches the profile rather than a generic implementation template.
- Verify checks the profile's evidence and does not require impossible gates.
- Handoff names the artifact, unresolved risk, owner, and next phase.

## Example — Correct

```text
Auditor profile: Analysis/audit
Build: OWASP findings with file:line evidence and CVSS severity.
Verify: Re-check exploitability and false positives; no production code is expected.
```

## Example — Incorrect

```text
Thinker profile: Delivery
Build: production code and migrations.
```

## Exception

An agent may combine profiles only when its triad explicitly documents the combination and the
handoff keeps each deliverable and gate distinct.
