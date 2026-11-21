# Skill Rule — Feature Specification Structure

Self-contained feature spec document structure, project-agnostic. Complements the project-agnostic
`skills/planning/SKILL.md` (which covers elicitation, SRS, backlog, and gates) with the artifact
shape for a single-feature spec. A spec written this way IS the authoritative source for
implementation, tests, and acceptance.

## Metadata

- **ID:** `rule-skill-planning-feature-spec`
- **Scope:** `skill:planning`
- **Severity:** `contextual` — applies when producing or reviewing a feature specification document
- **Enforcement:** `manual` (review)
- **Related:** `rules/global-standards.md` §Planning/§Design, `skills/planning/SKILL.md`

---

## Description

Every feature spec follows a fixed section scaffold so requirements are traceable, verifiable, and
unambiguous — before any code exists. Sections are populated with project-specific content; no
placeholder filler. Requirements-first doctrine: no behavior ships without documented intent.

## Trigger

- Writing a new feature spec, or reviewing one before implementation starts.
- A change touches requirements/contracts and needs to be registered in the spec before code.

## Action

Use these sections (order matters; omit none except where labeled optional):

1. **Problem statement** — the concrete pain or need, with real context (who, where, measurable
   baseline), not a generic "this system is slow" claim.
2. **Goals / Non-Goals** — measurable outcome with number and date (e.g. "reduce enrollment time
   from 4 min to 90 s by Q3"); explicitly list what is out of scope. Budget the NFRs with numbers,
   measurement point, owner.
3. **User stories** — `As a <role>, I want <action>, so that <benefit>` with unique examples; a
   story without an example is not ready.
4. **Functional requirements** — each FR is one verifiable sentence with a unique ID
   (`FR-{area}{number}`), discoverable and indexable. No two FRs say the same thing.
5. **Non-functional requirements** — each NFR quantified (`NFR-{category}{number}`): performance,
   security, usability, reliability, maintainability, compatibility, portability against a standard
   (e.g. ISO 25010), each with a measurable threshold.
6. **Data / API contracts** — verifiable signatures, shapes, routes, config keys, DTO/TypeScript
   interfaces, request/response schemas. Every operation has at least one request, one response,
   and one error example.
7. **Use cases** (`UC-*`) — end-to-end flows named by the spec (happy path + named alternatives +
   rejection paths), mapping to tests later.
8. **Design decisions** (`DD-{number}`) — decisions already made, with chosen option and rationale,
   so implementation does not re-litigate them. Reference ADRs instead of duplicating.
9. **Success metrics** — how we know the feature works after ship (metric, measurement point,
   target, baseline).
10. **Roadmap** (optional) — prerequisites (with specific artifacts), build guide 1-2 sentences,
    next steps.
11. **Acceptance criteria** — checklist forms the Definition of Done and basis for tests.

Requirement ID conventions:

- `FR-{area}{number}` — functional (e.g. `FR-AUTH-03`)
- `NFR-{category}{number}` — non-functional (e.g. `NFR-PERF-01`)
- `UC-{number}` — use case
- `DD-{number}` — design decision

Cross-reference related specs/docs instead of duplicating content. Metadata block at top, IDs
matching the document. History via git, never inline `Last updated` banners.

## Validation

- [ ] All required sections present and populated — no TBD, no filler
- [ ] Every FR/NFR/UC/DD has a unique, indexable ID
- [ ] Every requirement is a single, verifiable, testable statement
- [ ] ≤1 interpretation per statement; any ambiguity split or rewritten
- [ ] Contracts have examples (request + response + error); links resolve
- [ ] Goals are quantified with number/date; NFRs budgeted
- [ ] No duplicate content across sections — cross-reference instead

## Example — Correct (requirement row)

| ID | Requirement | Source |
|----|-------------|--------|
| FR-AUTH-03 | The login endpoint must reject after 5 failed attempts within 15 min per account | ISO maximum, stakeholder interview |

## Example — Incorrect

> The system should be fast and secure. (No requirement ID, no measurable threshold, two vague
> NFRs bundled, no owner.)

## Exception

Small, well-understood changes (a field, a translation key) may follow a lightweight spec or go
straight to issue + test — but any behavior change with contracts or multiple layers gets the full
scaffold.