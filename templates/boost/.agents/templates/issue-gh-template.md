# GitHub Issue Template — `{title}.md`

External project issue. Tracked in GitHub. Created directly because no native internal issue store
is verified. Do not add a local `issues/` source link.

## Naming

GitHub issue IDs are assigned by GitHub (numeric, e.g. `GH-42`). The file name is the lowercase slug of the issue title.

Example: `GH-42-sql-injection-in-login-form.md`

## Shape

```markdown
---
gh-number: 42
gh-repo: OWNER/REPO
gh-state: open | closed
gh-reason: completed | not planned | duplicate  # only when closed
gh-labels: [label1, label2]
gh-assignee: username
gh-milestone: "Milestone Title"
internal-source: none (native runtime issue store unavailable)
created: YYYY-MM-DD
updated: YYYY-MM-DD
---

# {Title}

## Description

[What this issue is about.]

## Traceability

- Internal issue: unavailable; do not create a local issue file
- Parent requirement: GH-1

## Labels

`security`, `bug`, `high-priority`

## Status

| Date | State | By | Note |
|------|-------|----|------|
| YYYY-MM-DD | open | @username | Created |

## Handoff

- **From:** direct request
- **To:** GitHub issue URL
```

## Rules

- Do not invent or link an internal issue source document.
- When closing, add the `gh-state` and `gh-reason` fields
- Labels must match existing repo labels — do not create new labels ad-hoc
- Milestone must exist in the repo — create it first if needed with `gh milestone create`
