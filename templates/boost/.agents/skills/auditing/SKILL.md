---
name: auditing
description: Use when auditing security, privacy, or code quality against independent benchmarks. Triggers on audit, security, privacy, compliance, OWASP, CWE, PII, ISO 25010, WCAG, quality gate.
---

# Auditing Skill

Comprehensive security, quality, and repository-consistency audit. The default contextual mode reads project rules and runtime posture; explicit blind mode evaluates only external benchmarks. Works with any AI agent and any project. This skill **never writes code**. It produces findings and a compliance scorecard; external issue creation is opt-in.

## When to Use

- Security audit: OWASP Top 10, authN/authZ, injection, XSS, CSRF, rate limiting, secrets handling
- Privacy / PII audit: masking, encryption, access control, data minimization, retention
- Quality audit: cleanliness, DRY, SOLID, module boundaries, naming, static-analysis findings
- Compliance scoring against ISO 25010 / WCAG / PSR with severity ratings
- Pre-commit / pre-release quality gates
- Repository workflow and agent-contract consistency audits

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Establish the audit scope (module, phase, whole app, or repository workflow), benchmark set, and mode. Use **contextual** mode for project consistency; use **blind** mode only when independent external benchmarking is explicitly requested.
- In contextual mode, read security/PII-relevant RuntimeContextInterface topics and `rules/` to distinguish real violations from intentional states. In blind mode, do not load project rules or architecture docs before the audit.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Select the applicable standards and scanners that cover the area. Prefer existing project scanners (`scan_security.py`, `scan_violations.py`, etc.) over manual review (Automation-First). In blind mode, use only external benchmark checks and generic scanners.
- Supplement with manual reading of high-risk seams: authentication, input handling, data persistence.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Keep findings in the report by default. Create GitHub issues only when the caller explicitly enables issue publication; never create local issue/ref files.
- **Deliverable:** Evidence-backed audit findings and scorecard; no code increment is produced.
- **Measurable outcome:** Every finding has evidence, severity, impact, remediation, and benchmark trace; no secrets or debug output are introduced.

- Run scanners and targeted reads; gather concrete `file:line` evidence per finding.
- Classify each finding by severity (critical/high/medium/low) and category. Do not report documented intentional states as defects without noting them.
- **Choose the right home for the report.**
   - **Internal audit reports** (agent-owned, exploratory, working notes, compliance scorecard for personal review) → communicate in chat or use RuntimeContextInterface when appropriate. Internal issue storage is unsupported; stop and report if an internal issue artifact is required.
- **Project-level findings** (defects the team must fix, severity-ranked) → produce GitHub-issue-ready content by default. Create GitHub issues with `gh issue create` only when issue publication was explicitly enabled; each issue must include `file:line`, impact, remediation, and trace to a standard.
- When in doubt: ask the user whether to publish findings; never infer permission from audit scope.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Triple-check each finding against the real code — no hallucinated vulnerabilities or false positives.
- Confirm severity is proportionate to actual exploitability in this deployment context.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Report findings by default. If issue publication was explicitly enabled, report created GitHub issue URLs; never create local issue/ref files.
- **Deliverable:** Compliance scorecard and severity-ranked findings with explicit remediation handoff.
- **Measurable outcome:** Every finding has evidence, owner/next step, and benchmark trace; no unsupported store is created.

- Deliver a compliance scorecard + GitHub-issue-ready findings (severity-ranked, with location, impact, and concrete remediation). Never commit fixes yourself; issue creation requires explicit opt-in.

## Adaptive Modes

| Request Class | Auditing Response |
|---------------|------------------|
| **Clear** | Execute the selected contextual or blind audit and produce findings |
| **Materially ambiguous** | Clarify scope (module/phase/whole app) and benchmark set before auditing |
| **High-risk** | Add independent cross-check; never commit fixes, only report findings |
| **Blocked/unsupported** | Report the missing capability; do not fabricate findings or scores |

## Anti-Patterns

- Generic boilerplate findings without code location; false positives; equating a documented intentional design choice with a violation.

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

- **Bundle contents:** `SKILL.md`; `rules/` (4 items); `tools/` (none); `resources/` (none); `refs/` (none).
