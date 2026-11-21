---
name: writing
description: Use when writing, updating, syncing, or auditing any readable artifact — markdown docs, project documentation, code docblocks, and content/copy. Covers conceptual vs reference tiers, metadata and structure, link integrity, PHPDoc/docblock conventions, and doc-code-spec sync. Triggers on document, write docs, update docs, sync docs, PHPDoc, docblock, doc drift, documentation-first, SSOT, UI copy, copywriting, changelog, README.
---

# Writing Skill

Write, update, and keep every readable artifact in sync with code and specs — markdown docs, project
documentation, code docblocks, and content/copy. Enforces the **documentation-first (SSOT)**
principle: docs are the source of truth for intent and behavior, code and docs must agree, and every
doc claim must resolve to a real target. Works with any project layout (`docs/`, module docs, agent
docs & skills, devtool catalogs, README, changelog, UI copy).

## When to Use

- Writing new documentation (markdown, PHPDoc blocks, project docs).
- Updating existing docs to reflect code, spec, or convention changes.
- Adding or editing PHPDoc / docblocks on classes and methods.
- Writing or editing content/copy: UI copy, marketing, emails, changelogs, release notes, naming.
- Writing or maintaining READMEs, reference docs, conceptual docs, module docs, and devtool catalogs.
- Syncing documentation with implementation after any build, refactor, or audit.
- Resolving a doc-vs-code disagreement (never guess which is stale — check git history).
- Pre-commit review of any doc change (the doc-quality gate).

## Instructions

Follow the mandatory workflow `Understand → Plan → Build → Verify → Summarize` with anti-slop discipline.

### Understand
- **Issue tracking:** Read a linked GitHub Issue when present; if internal issue tracking is required, stop and report the unsupported runtime capability.
- **Deliverable:** Grounded understanding with file:line citations, assumption log with owners, and explicit scope boundaries.
- **Measurable outcome:** Every claim has a source; no invented stakeholders, numbers, or constraints; open questions have owners and due dates.

- Identify which artifacts are affected: reference docs (what exists and where), conceptual docs (why
  it exists and what rules govern it), content/copy (UI, marketing, email, release notes), agent
  docs & skills, or devtool catalogs.
- Locate the governing spec if one exists. A behavior change with no documented requirement stops and
  escalates — docs reflect the governing spec, never float free of it.
- When code and docs disagree, **check history first**: `git log --follow -- <file>` and `git diff`
  on both sides to see which changed last and why — both can be stale. Never silently pick one as
  "truth".

- **Anti-pattern:** Inventing stakeholders/numbers/constraints to fill gaps; carrying TBDs forward without an owner; vague hedging when quantified context exists.

### Plan
- **Issue tracking:** Use GitHub issue tools for team-visible artifacts; do not create a local issue file.
- **Deliverable:** Versioned plan/design artifact with explicit non-goals, traceability matrix, and prioritized backlog.
- **Measurable outcome:** Every requirement is SMART; every NFR has SLO + measurement point + owner; each backlog item has a unique, verifiable example.

- **Pick the tier per file.** Conceptual docs answer *why* (purpose, design principles, business
  rules, trade-offs) and must NOT contain implementation trivia (file paths, class names, schemas,
  API tables). Reference docs answer *what* (locations, contracts, schemas, catalogs) and must NOT
  contain design rationale. If a fact doesn't fit either, it is misplaced or not worth documenting —
  never bolt on a third tier.
- **Map every changed file to the doc that must change** (module reference, architecture guide,
  agent guide, devtool catalog) so no change lands undocumented.
- **Decide additions vs updates.** Never duplicate a fact — keep one authoritative location and
  cross-reference from everywhere else.
- **For content/copy:** confirm audience, tone, and channel first; match the material's existing
  voice and conventions over inventing a new style.

- **Anti-pattern:** Formulaic lists repeating the same verb; unquantified NFRs without SLO/measurement/owner; zero-example backlog; big-bang planning without time-boxed spikes.

### Build
- **Issue tracking:** Update linked tracked source documentation with progress when present; close GitHub Issue with the resolving commit SHA.
- **Deliverable:** Working, reviewed increment with tests and migration notes; stop if an internal ref is required.
- **Measurable outcome:** Lint/typecheck/tests pass from real tool output; PR <300 LOC per commit; no secrets committed; every `dd()`/debug call removed.

- Follow the document structure contract on every file:
  - `# Title — Subtitle` (exactly one H1, always first).
  - `## Description` — first H2, 1-3 sentence summary, on every doc.
  - Topical `##` content sections, `###` subsections; never skip heading levels.
  - Optional `## AI Agent Notes` — machine-readable checklists/decision tables/commands, only when
    agents will consult the doc during tasks (no prose, no rationale).
  - `## Quick References` — the standard footer, always last (not `## References`, `## See Also`,
    `## Where to Find It`).
  - `---` horizontal rules between major H2 sections.
- **No inline metadata.** History lives in git (`git log --follow -- <file>`); write descriptive
  commit messages (`type(scope): desc`). Never add `> **Last updated:**` banners or `Prior:` trails.
- **Link integrity.** All internal links relative (`[label](relative/path.md)`), anchors match exact
  headings (`#anchor` = lowercased, spaces → hyphens). Prefer cross-referencing over copying.
- **PHPDoc/docblock conventions.** Native types are the primary documentation; docblocks supplement
  only what the signature cannot express (business meaning, side effects, exceptions). Never
  duplicate `@param`/`@return` next to typed signatures; never add
  `@author`/`@version`/`@created`/`@package`; `@throws` lists specific exception types, not generic
  ones; one-line for simple methods, multi-line for complex. During implementation, `builder` writes
  docblocks inline following these conventions; `writer` owns the conventions and audits passes.
- **Content/copy quality.** Prefer plain, specific language; one idea per sentence; active voice for
  UI/marketing; never invent numbers, quotes, or testimonials; adapt tone to the channel (UI terse,
  changelog factual, marketing persuasive).
- **Edit, don't rewrite.** Target only the changed section; preserve unrelated content, nuance, and
  cross-references. Use targeted edits over full rewrites; when a rewrite seems necessary, read the
  whole file first.
- Mirror spec amendments in any guide, skill, or devtool that documents or automates the change.

- **Anti-pattern:** Vague comments like "delve into the realm"; mixing refactor and feature in one commit; hardcoded secrets or debug output committed.

### Verify
- **Issue tracking:** Update tracked source documentation with gate results when present; close GitHub Issue only after all acceptance criteria pass.
- **Deliverable:** Verified artifact with real gate outputs (lint/typecheck/test/SAST/SCA), not invented numbers.
- **Measurable outcome:** All local gates pass; coverage meets project floor; no hallucinated stats or links; every URL resolves.

- **Markdown-only changes:** run `tools/scan_broken_links.py` (no test suite needed for doc changes).
- Run the **doc-quality gate** on every doc change before commit — all four, together:
  1. **Structural:** H1 format, `## Description` first, `## Quick References` footer, no skipped
     heading levels.
  2. **Tier separation:** no implementation details in conceptual docs, no rationale in reference docs.
  3. **Integrity:** every `[text](path)` resolves, every anchor matches a heading, no duplicated content.
  4. **Metadata/PHPDoc:** history in git (no inline dates), docblocks not duplicating native types.
- Run `pint` (or the project formatter) only if PHP/PHPDoc files were touched.
- For content/copy: proofread for tone accuracy, factual correctness, and consistency with existing copy.
- `git diff` on each touched doc — prove only the intended sections changed, nothing dropped.

- **Anti-pattern:** Invented coverage numbers or stats; `// TODO` without an issue ID; hallucinated links or sources; claiming gates passed without running them.

### Summarize
- **Issue tracking:** Close resolved GitHub Issues; update tracked source documentation with outcomes and follow-ups when present.
- **Deliverable:** Closed issues, updated changelog/docs, explicit handoff to next phase with demo/rollback instructions.
- **Measurable outcome:** Linked GitHub Issues closed or updated; tracked documentation or RuntimeContextInterface contains evidence; next phase has explicit acceptance criteria and owners.

- Report: files written/updated, tier decisions, drift findings and how they were resolved (which side
  was stale and why), and any doc gaps raised as follow-ups.

## Adaptive Modes

| Request Class | Writing Response |
|---------------|-----------------|
| **Clear** | Write or update the artifact with SSOT discipline |
| **Materially ambiguous** | Clarify the artifact tier (conceptual/reference/copy) and governing spec |
| **High-risk** | Check git history before editing; verify doc-code-spec sync |
| **Blocked/unsupported** | Report the missing capability; do not fabricate content |

## Anti-Patterns

- **Full-file rewrites** for small edits — silently drops edge-case notes and cross-references.
- **Code-as-truth** when a doc is stale — deciding "fix the doc to match code" without checking git
  history; both sides can be stale.
- **Kitchen-sink docs** — one file trying to be both manual and index, mixing rationale with schemas.
- **Duplicated facts** — copy-pasting a fact into a second doc instead of cross-referencing the
  authoritative one.
- **Inline "Last updated" banners** — history belongs in git, not the file.
- **Docs deferred to "later"** — a change without its doc update is an incomplete change.
- **Docblocks duplicating type hints** — `@param string $x` next to `string $x` is dead weight that
  drifts.
- **Fabricated copy** — invented stats, quotes, or testimonials in marketing content; a fabricated
  voice where an established one exists.

## Related Agents

This triad collaborates with the following registered agents:

| Agent | Relationship | Use / Handoff Trigger |
|-------|--------------|----------------------|
| `planner` | Collaborates on requirements and project planning | Route work when the deliverable requires requirements and project planning. |
| `researcher` | Collaborates on external research and evidence gathering | Route work when the deliverable requires external research and evidence gathering. |
| `tester` | Collaborates on test strategy and quality assurance | Route work when the deliverable requires test strategy and quality assurance. |
| `housekeeper` | Collaborates on memory, context, and rules hygiene | Route work when the deliverable requires memory, context, and rules hygiene. |
| `customizer` | Collaborates on agent-asset creation and triad wiring | Route work when the deliverable requires agent-asset creation and triad wiring. |
| `learner` | Collaborates on workflow learning and signal promotion | Route work when the deliverable requires workflow learning and signal promotion. |
## Handoff

Docs and copy in sync with code and specs; links verified; quality gate passed; drift findings
resolved or reported as follow-ups.

- **Anti-pattern:** Unreviewed AI-generated bulk left in docs/changelog; empty buzzwords in handoff; closing issues without evidence; missing demo/rollback instructions.

- **Bundle contents:** `SKILL.md`; `rules/` (1 item); `tools/` (none); `resources/` (none); `refs/` (none).
