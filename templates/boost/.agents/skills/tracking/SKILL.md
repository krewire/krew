---
name: tracking
description: Use when opening, closing, listing, or viewing GitHub issues. Triggers on track, issue, github issue, open issue, close issue, list issues, issue status.
---

# Tracking Skill

Opens, closes, lists, and views GitHub issues via `gh`. Uses direct requests, tracked project documentation, and audit reports as sources; internal `issues/` storage is unsupported. Loads on demand when the user mentions tracking, issues, or GitHub issue management.

## When to Use

- User asks to open, close, list, or view a GitHub issue
- A tracked design document or audit report is ready to be promoted to a GitHub issue
- The user wants to check the status of open issues
- A blocker has been resolved and the issue needs to be closed

## Instructions

Follow the mandatory workflow: `Understand → Plan → Build → Verify → Summarize`.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

1. **Parse the action** from `$ARGUMENTS`:
   - `open` — create a new issue
   - `close` — close an existing issue
   - `list` — list issues with filters
   - `view` — show a specific issue
2. **Identify the repo.** Priority: `--repo OWNER/REPO` flag > current directory git remote > `reasvyn/ai-agent-kit`.
3. **Auth check.** Run `gh auth status` — if it fails, abort with a clear message.
4. **For `open`:** If a tracked document or audit report is provided, extract `title`, `body`, `labels`, `assignee`, and `milestone` from it. Otherwise parse `--title`, `--body`, `--label`, `--assignee`, and `--milestone` from flags.
5. **For `close`:** Extract the issue number and `--reason`. Default reason is `completed`.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; internal issue storage is unsupported.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

1. **Choose the subcommand:**

```
gh issue create   --title TEXT --body TEXT [--label LABEL] [--assignee USER]
                  [--milestone TITLE] [--repo OWNER/REPO]
gh issue close    N --reason completed|not planned|duplicate [--repo OWNER/REPO]
gh issue list     [--state open|closed|all] [--label LABEL] [--assignee USER]
                  [--milestone TITLE] [--author USER] [--repo OWNER/REPO]
                  [--json ...]
gh issue view     N [--repo OWNER/REPO] [--json ...]
```

2. **For `open`:** Build the body by combining `--body` with a traceability section:
   ```
   ## Traceability

   Source: tracked design document or audit report reference
   ```
3. **For `list`:** Use `--json number,title,state,labels,updatedAt,url` for structured output; fall back to default table format if `--json` is not supported.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked tracked source document with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

1. **Execute the command** via `bash`.
2. **Capture output.** For `gh issue create`, use `--jq '.number'` or parse the returned URL for the issue number.
3. **Format results:**
   - `open` → `GH-{number}: {url}`
   - `close` → `GH-{number} closed ({reason})`
   - `list` → markdown table with columns: `#`, `Title`, `State`, `Labels`, `Updated`
   - `view` → structured summary with all metadata

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the tracked source document with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

1. **Auth:** `gh auth status` must pass.
2. **For `open`:** `gh issue view {number} --json number,title,labels` — confirm the created issue matches input.
3. **For `close`:** `gh issue view {number} --json state` — confirm `"state": "CLOSED"`.
4. **For `list`:** confirm row count and filter correctness.
5. **For `view`:** confirm returned number matches requested number.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close resolved GitHub Issues; update the tracked source document with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; source traceability is recorded; next phase has explicit acceptance criteria and owners.

Report the outcome clearly:

- `open`: `Opened GH-{number}: {title}\n{url}` + traceability note.
- `close`: `Closed GH-{number}: {title} ({reason})`.
- `list`: table + `N issues (X open, Y closed)`.
- `view`: full issue summary with all metadata.

## Adaptive Modes

| Request Class | Tracking Response |
|---------------|------------------|
| **Clear** | Execute the issue action with gh CLI |
| **Materially ambiguous** | Clarify the action (open/close/list/view) and target issue |
| **High-risk** | Verify repo and auth before executing; confirm destructive actions |
| **Blocked/unsupported** | Report the missing capability; do not fabricate issue data |

## Anti-Patterns

- **Never** require a local issue source document; link only the direct request context.
- **Never** close without `--reason` — missing reason makes it ambiguous.
- **Never** use `--body-file` with a path outside the repo — validate the path first.
- **Never** assume the current directory is the right repo — always run `gh repo view` to confirm.

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

- **From:** tracked design document, audit report, or user request.
- **To:** issue URL and number for documentation; update the source document when it is tracked.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
