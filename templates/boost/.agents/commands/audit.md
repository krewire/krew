---
name: audit
description: Audit security, privacy, and code quality against independent benchmarks. Use when audit, security, compliance, OWASP, CWE, privacy, PII, or quality-gate review is mentioned.
---

# Audit Command — Independent Security & Quality Audit

Generic audit command for any domain. Contextual mode audits code, configuration, and repository contracts against project rules plus external benchmarks; blind mode audits only external benchmarks. It produces a severity-ranked finding list and compliance scorecard. This command **never writes code**; issue publication is opt-in.

Follow `skills/auditing/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`. Ground every finding in `file:line` evidence — no hallucinated vulnerabilities.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask for the audit scope (module, phase, or whole app) and which benchmark set to apply.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Determine which standard set and mode apply. Contextual mode reads security/PII-relevant RuntimeContextInterface topics and `rules/`; blind mode must not load project documentation before or during the audit.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Prefer existing project scanners (`scan_security.py`, `scan_violations.py`, etc.) over manual review where they cover the area (Automation-First); supplement with manual reading of high-risk seams (auth, input handling, persistence).

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Run scanners and targeted reads; gather `file:line` evidence per finding. Classify by severity (critical/high/medium/low) and category.
   - **Internal audit reports** → stop and report; no verified native runtime issue store exists.
   - **Project-level findings** (defects the team must fix) → produce issue-ready findings by default; create GitHub issues via `gh issue create` only with explicit `--publish-issues` authorization.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

No false positives that waste a fix cycle; confirm severity matches actual exploitability in this deployment.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Deliver a compliance scorecard + GitHub-issue-ready findings (severity-ranked, with location, impact, remediation, and standard trace). Never commit fixes; issue creation is opt-in.
## Anti-Patterns

- Generic boilerplate findings without code location; false positives; reporting documented intentional states as defects.

## Parameters (when invoking via tool)

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `{{scope}}` | string | No | whole-app | Module, phase, or whole application |
| `{{benchmark}}` | enum | No | `security` | `security`, `privacy`, `quality` |
| `--mode` | enum | No | `contextual` | `contextual` reads project rules; `blind` uses external benchmarks only |
| `--publish-issues` | flag | No | false | Explicitly authorize GitHub issue creation |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `builder` | Collaborates on implementation and feature delivery | Route work when the deliverable requires implementation and feature delivery. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `deployer` | Collaborates on CI/CD and release delivery | Route work when the deliverable requires CI/CD and release delivery. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `thinker` | Collaborates on structured decisions and conflict resolution | Route work when the deliverable requires structured decisions and conflict resolution. |
## Handoff

A severity-ranked finding list, each with `file:line`, impact, remediation, and trace to a standard (OWASP/CWE/ISO).

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
