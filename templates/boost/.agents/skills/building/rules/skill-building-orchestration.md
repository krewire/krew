# Skill Rule — Orchestrator Skill Delegation & Coordination

An orchestrator skill is a conductor, not a do-everything skill. It owns the build order, the spec
anchors, and the quality gates; specialized work is delegated to sub-skills that each own one
concern. This rule defines when to delegate, what to load before touching each concern, and how to
hand off upstream and downstream without duplicating work. Complements the project-agnostic
`skills/building/SKILL.md`.

## Metadata

- **ID:** `rule-skill-building-orchestration`
- **Scope:** `global` — governs any orchestrator agent/skill coordinating specialized sub-skills
- **Severity:** `contextual` — binds while orchestrating a multi-concern implementation
- **Enforcement:** `manual` (review)
- **Related:** `rules/skill-building-refactoring.md`, `rules/skill-testing.md`, `rules/skill-writing.md` (doc-sync downstream), `rules/global-standards.md` §Planning

---

## Delegate Each Concern to Its Specialized Sub-Skill

**What it enforces:** before writing code for a specialized concern, load the matching sub-skill
from the skill map — never implement that concern from general knowledge. The rule fires per concern
group, not once per feature.

**Why it matters:** each sub-skill encodes project-specific contracts the orchestrator cannot hold in
full (component table wiring, media-library collections, recorder setup). Writing a concern without
its skill produces code that looks right but violates conventions, which scanners and code review
then reject — a full rework cycle that delegation avoids.

**How to apply:** load the sub-skill before the first file of that concern is written — e.g. livewire
components → the livewire/component skill; file uploads/media → the media-library skill; UI/styling/
layout → the tailwind/UI skill; observability dashboards → the observability skill; refactoring
verification → the refactoring skill.

**Pitfalls to avoid:**

- Writing the presentation of a feature without its UI skill because "it's just markup" — component/
  palette conventions are project-specific.
- Handling a specialized integration inline in the main action without its skill — registration and
  config will be wrong.
- Loading every sub-skill unconditionally "for safety" — load only the skills the feature actually
  touches (unneeded skill loads bloat context).

**Verification:** for each specialized concern in the feature, the matching skill was loaded before
that code was written; no concern was implemented from generic knowledge.

---

## Upstream and Downstream Handoffs

**What it enforces:** the orchestrator knows where work comes from and where it goes. Upstream:
spec/planning output and already-refactored code. Downstream: the test suite and the doc sync.

**Why it matters:** the implementation pipeline is only as strong as its boundaries. If the
orchestrator silently starts from stale specs or finishes without feeding tests and doc sync, the
feature lands with untested requirements and drifting docs — the two defects the whole pipeline
exists to prevent.

**How to apply:** on entry, confirm the upstream contract is satisfied (spec exists with requirement
IDs; refactored code is already in its extracted shape). On completion, hand off to testing (write/
verify the requirement-traced tests) and doc sync (align docs with the shipped code). During the
build, hand off per concern to the sub-skill table above.

**Pitfalls to avoid:**

- Treating the feature as done after the code works in the browser — the pipeline's downstream
  (tests + docs) is part of done.
- Re-implementing upstream work (e.g., rewriting an action that a refactor already extracted).

**Verification:** the final report names the downstream handoffs that were completed (test-suite
result, doc sync); no upstream contract was re-derived in the orchestrator.

---

## M/L Tasks — Stage by Layer and Verify Each Stage

**What it enforces:** for M/L features, the build is staged by layer/concern and each stage is
verified before the next begins.

**Why it matters:** an orchestrator that throws all sub-skills at a large feature in parallel
produces an unreviewable diff and compounds cross-layer errors. Staging lets each layer settle and be
verified independently — exactly what the orchestrator is positioned to enforce, since the sub-skills
each see only their slice.

**How to apply:** classify the feature size. For M, stage internally and batch verification with one
checkpoint before commit. For L, split into sessions, inform the user with a session plan, and end
each session with `git status`/`git diff` review, targeted verification, and a short report.

**Pitfalls to avoid:**

- Starting the next layer before the current layer's targeted check passes.
- Letting an L feature expand into one giant uncommitted change set.

**Verification:** the transcript shows per-stage verification before advancing; L features have a
user-approved session plan.

---

## Version Bump / Release Orchestration

**What it enforces:** a release is a coordinated change across many version-bearing artifacts —
never a single-edit "bump the version number".

**Why it matters:** every deployment pipeline keys off a tag; a version that is updated in some
places but not others produces stale docs, mismatched app-reporting, and in the worst case a build
pulled from the wrong ref because the tag is missing or stale.

**How to apply (pre-flight checklist):**

1. `grep` for the version string across `composer.json`, `package.json` (+ lock), app config/version
   service, docs (spec index, refs, README), and CI workflows — never rely on a hard-coded list.
2. `composer.json:version` + `package.json:version` (+ lock) are primary; keep app config/version
   fallback in sync.
3. Docs with explicit version bands (spec build order, module refs, README badges) — no stragglers.
4. `context/`/`rules/`/`skills/` must never hard-code the version number — always re-grep to find
   stragglers.
5. After bump: create the matching `git tag vX.Y.Z` **and push it** — pipelines (release workflows,
   build-from-`GIT_URL#tag`) depend on the tag existing; verify with `git describe --tags` and on
   the target host.

**Pitfalls to avoid:**

- Bumping `version` without creating/pushing the matching tag (deploy builds from the wrong ref).
- Hard-coding "current version" in agent docs so the checklist itself goes stale.
- Assuming every QA stage runs locally — tag-driven pipelines derive stage from the tag, so
  pre-release tags (`-dev`, `-beta`, `-rc`) gate gradually and only the final production tag deploys.

**Verification:** the tag exists and is pushed; `git describe --tags` matches the bump intent;
grep finds no stragglers.