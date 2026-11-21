---
name: track
description: Open, close, list, or view GitHub issues. Use when track, issue, github issue, open issue, close issue, list issues, issue status.
---

# Track Command — GitHub issue management

Manages GitHub issues: open, close, list, and view. Follow `skills/tracking/SKILL.md` and the
mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: `$ARGUMENTS` — one of: `open [--title TITLE] [--body BODY] [--label LABEL] [--assignee USER] [--milestone TITLE] [--repo OWNER/REPO]`, `close ISSUE_NUMBER [--reason completed|not planned|duplicate] [--repo OWNER/REPO]`, `list [--state open|closed|all] [--label LABEL] [--assignee USER] [--milestone TITLE] [--repo OWNER/REPO]`, or `view ISSUE_NUMBER [--repo OWNER/REPO]`. Local issue-file linking is unsupported.

If `$ARGUMENTS` is empty or materially ambiguous, ask for the action and target issue.

### Understand
- **Issue tracking:** Read a linked GitHub Issue for context; do not create or read a local issue document.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Identify the action and target. Parse `$ARGUMENTS` into subcommand + flags. For `open`, extract title, body, labels, and assignee from direct input; local internal issue files are unsupported. Verify `gh` authentication with `gh auth status`. Determine the target repo: use `--repo` flag, current directory git remote, or this repository's remote (`reasvyn/ai-agent-kit`).

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; internal issue storage is unsupported.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Select the `gh issue` subcommand that matches the action:

| Action | Subcommand |
|--------|-----------|
| `open` | `gh issue create --title ... --body ... --label ... --assignee ... --milestone ...` |
| `close` | `gh issue close N --reason ...` |
| `list` | `gh issue list --state ... --label ... --assignee ... --milestone ...` |
| `view` | `gh issue view N` |

For `open`, compose the body with a traceability section based on direct input, a tracked document, or an audit report; no local issue store is used.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Execute the `gh issue` command. Capture the full output.

- `open`: append `--json number,url` to get structured output; extract issue URL and number.
- `close`: confirm the returned output shows `closed`.
- `list`: use `--json number,title,state,labels,updatedAt,url` for structured data; format as a markdown table.
- `view`: use `--json` with all fields for structured output; render as a readable summary.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Confirm `gh auth status` passes with no error.
- `open`: fetch the issue URL and verify title + labels match the input.
- `close`: run `gh issue view N --json state` and confirm `"state": "CLOSED"`.
- `list`: confirm row count matches expected filters.
- `view`: confirm the returned issue number matches the requested one.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue closed or updated; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

- `open`: report `GH-{number}` and the URL; note the linked internal doc.
- `close`: report the closed issue number and reason.
- `list`: present the table with count summary.
- `view`: present the full issue with metadata.

## Anti-Patterns

- Never require a local issue source document; link only the GitHub issue and its direct request context.
- Never close without `--reason`.
- Never assume current directory is the target repo — verify with `gh repo view`.

## Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--title` | string | For `open` | — | Issue title |
| `--body` | string | No | — | Issue body (use `\n` for newlines) |
| `--label` | string | No | — | Labels (comma-separated or repeat flag) |
| `--assignee` | string | No | — | GitHub username |
| `--milestone` | string | No | — | Milestone title |
| `--repo` | string | No | current repo | `OWNER/REPO` |
| `--link` | string | No | — | Unsupported legacy option; do not use a local issue path |
| `--reason` | string | For `close` | — | `completed`, `not planned`, or `duplicate` |
| `--state` | string | For `list` | `open` | `open`, `closed`, or `all` |

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `manager` | Collaborates on intent clarification and task-graph packaging | Route work when the deliverable requires intent clarification and task-graph packaging. |
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `maintainer` | Collaborates on operations, incidents, and technical debt | Route work when the deliverable requires operations, incidents, and technical debt. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
## Handoff

Delivers the issue URL, number, and state. Do not link to a local internal issue document.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.
