---
name: research
description: Conduct real-world external research via web search/fetch. Use when research, web search, market, trends, competitor, benchmark, landscape, or real-world data is mentioned. Distinct from explore (local codebase) — this is web research.
---

# Research Command — Real-World Web Research

Conducts **external, web-based research** on a topic and returns a **vision** plus **research findings**. Works with any AI agent and any domain. **Distinct from `/explore`**: explore maps a local codebase; research consults the live web (search/fetch) for the outside-the-repo landscape.

Follow `skills/researching/SKILL.md` and the mandatory workflow `Understand → Plan → Build → Verify → Summarize`.

Scope: $ARGUMENTS

If $ARGUMENTS is empty or materially ambiguous, ask what to research and — critically — **how the output should be delivered**: communicated in chat or stored through RuntimeContextInterface. Never guess the delivery format; internal refs are unsupported.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

Clarify the question, the decision it informs, and the delivery format (communicate vs file). Decide scope: markets, competitors, trends, benchmarks, tooling, time horizon.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

Use several distinct web-search queries plus targeted fetches of authoritative sources (official docs, vendor pages, reputable analyses). Note the current year and recency sensitivity.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update the linked GitHub Issue with progress and resolving commit SHA; do not create a local issue file.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

Run searches/fetches and collect **source + date** for every claim; never invent a citation, stat, or page. Produce (a) **findings** — verified, sourced evidence; (b) **vision** — a reasoned, forward-looking synthesis clearly labeled where it extrapolates.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update the GitHub Issue with gate results when present; do not create a local issue file.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

Validate key claims against more than one source where possible. Separate verified fact from assumption/projection; label conflicts and state the one you weight and why.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close/update the GitHub Issue when present; do not create a local issue file.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** GitHub Issue updated when present; tracked rules/runtime context contain evidence; next phase has explicit acceptance criteria and owners.

Either communicate the findings + vision in the response, or store them in an agreed RuntimeContextInterface topic and summarize in chat. Note sources, recency window, and open questions; stop if an internal ref is required.
## Core Concepts

- **Findings vs Vision:** findings are verified, sourced evidence; vision is the reasoned, forward-looking synthesis on top of it — labeled as interpretation.
- **Grounded citations:** every claim traces to an actual search/fetch result with a source and date — no invented data.
- **Researcher ≠ Explorer:** `/research` = live external web; `/explore` = internal local codebase. Choose by which world the question concerns.
- **Delivery by request:** confirm whether to communicate in chat or write a file before researching — then honor it.

## Examples

```bash
/research "2026 developer-experience trends: what is shifting in local-first AI tooling? Report market + key players; communicate in chat"
/research "Compare Postgres vs vector DB for RAG in 2026 — summarize findings in chat or runtime context"
/research "Competitor landscape for our dev-tool SaaS — summarize vision in chat or runtime context"
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Stuck doing local file inventory | Confused with explorer role | Refocus: this research is web-only; local code mapping belongs to `/explore` |
| A citation turns out wrong | Invented or stale source | Re-check with search/fetch; cite the live page + access date; label if unresolved |
| Vision reads like speculation | Interpretation not labeled | Mark projections clearly and separate from verified findings |
| Not sure where to deliver | Delivery not confirmed | Ask communicate-vs-file up front and honor the answer |

## Validation

- [ ] Research is web-based (search/fetch), not local codebase mapping — distinct from `/explore`
- [ ] Every claim has a source + date; no invented citations, stats, or pages
- [ ] Findings (verified) separated from vision (labeled interpretation/projection)
- [ ] Conflicts/stale data flagged, with the source you weight and why
- [ ] Output delivered in the user-requested format (communicated and/or filed to the agreed path)

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `explorer` | Collaborates on codebase research and orientation | Route work when the deliverable requires codebase research and orientation. |
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `designer` | Collaborates on architecture and API design | Route work when the deliverable requires architecture and API design. |
| `writer` | Collaborates on documentation and content quality | Route work when the deliverable requires documentation and content quality. |
| `auditor` | Collaborates on security, privacy, and quality review | Route work when the deliverable requires security, privacy, and quality review. |
