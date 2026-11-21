---
name: researching
description: Use when conducting real-world external research via web search/fetch. Triggers on research, web search, market, trends, competitor, benchmark, landscape, real-world data, vision, findings. Distinct from exploring (local codebase).
---

# Researching Skill

Standards and conventions for conducting **real-world, external research** through web search and page fetch, producing a **vision** and **research findings** with grounded citations. Applies to any domain, stack, and intent. Works with any AI agent. **Distinct from `exploring`, which maps an internal local codebase**: this skill is exclusively about live, outside-the-repo information from the web.

## When to Use

- The question concerns the **world**, not the repo: markets, trends, competitors, benchmarks, tooling landscape, current events, best practices
- You need current/third-party information that a codebase cannot answer
- A forward-looking **vision** or research synthesis on a topic is required
- You must research and then either **communicate** the result or **file** it (per user request)

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize`; this skill adds the web-research discipline below.

### Understand — frame the research
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Restate the research question, who/what it informs, and the deliverable shape (**findings** + **vision**).
- **Confirm the delivery format up front**: communicate in chat only, or store it through RuntimeContextInterface. Internal refs are unsupported. Never guess — ask, then honor.
- Scope: markets, competitors, trends, benchmarks, tooling, time horizon for the vision.

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan — design the query strategy
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- Use multiple **distinct** web-search queries to cover the topic from different angles.
- Follow up with targeted **fetches** of authoritative pages (official docs, vendor pages, reputable third-party analyses).
- Note the current year and treat information as **recency-sensitive**: prefer up-to-date, live sources.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build — gather & synthesize
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- For every claim, record **source + date** from the actual search/fetch result. Never fabricate a citation, statistic, or page.
- Where sources conflict or look stale, flag it and state which you weight and why.
- Assemble:
  - **(a) Findings** — verified, sourced evidence.
  - **(b) Vision** — a reasoned, forward-looking synthesis, clearly marked as interpretation wherever it extends beyond the evidence.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify — cross-check
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- Validate key claims against more than one source where available.
- Separate verified fact from assumption/projection; label each. Re-run a search if a citation is uncertain.
- Confirm no citation or page is invented.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize — deliver as chosen
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- **Communicate**: present findings + vision directly in the response, with sources and recency window.
- **File**: write the research to the agreed path per the relevant template, then summarize in chat and point to the file.
- List open questions the user should verify.

## Adaptive Modes

| Request Class | Researching Response |
|---------------|---------------------|
| **Clear** | Execute the query strategy and deliver findings + vision |
| **Materially ambiguous** | Clarify the research question and delivery format before searching |
| **High-risk** | Prefer multiple independent sources; flag conflicting or stale data explicitly |
| **Blocked/unsupported** | Report the web access limitation; do not fabricate citations |

## Anti-Slop

1. Never confuse this with exploring — do not map the local codebase here (that is explorer's job).
2. Never invent a citation, statistic, or page — every claim traces to a search/fetch result with a source and date.
3. Never present a projection/vision as a verified fact — label interpretation.
4. Never rely on stale or single-source data without flagging it.
5. Never guess whether to communicate or file — ask the user and honor the answer.
6. Always check the current year and recency before trusting a source.

## Validation

- [ ] Research is web-based (search/fetch), not local codebase mapping — distinct from `/explore`
- [ ] Every claim sourced (source + date); no invented citations, stats, or pages
- [ ] Findings (verified) separated from vision (labeled interpretation/projection)
- [ ] Conflicts/stale data flagged, with the source you weight and why
- [ ] Delivered in the user-requested format (communicated and/or filed to the agreed path)

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

- **Bundle contents:** `SKILL.md`; `rules/` (none); `tools/` (none); `resources/` (none); `refs/` (none).
