# Skill Rule — Spatie ActivityLog Conventions

Spatie ActivityLog (`spatie/laravel-activitylog`) audit-trail conventions for any Laravel project —
logging mutations, causer tracking, and log retrieval. Complements `skills/building/SKILL.md` and
`rules/skill-building-laravel.md` with audit mechanics.

## Metadata

- **ID:** `rule-skill-building-spatie-activitylog`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses `spatie/laravel-activitylog`
- **Enforcement:** `manual` (review)
- **Related:** `rules/global-security.md`, `rules/global-standards.md` §Maintenance

---

## Description

Audit trails record what happened, who caused it, and what state changed — for traceability and
compliance. Log at mutation points (Command Actions), capture the causer, and retrieve via the
ActivityLog API or typed helpers. Log how the user got there, not just that it happened.

## Trigger

- Logging an activity/mutation, tracking the acting user, or querying audit history.
- Configuring an audit/logging channel or a dual-channel (DB + file) strategy.

## Action

1. **Log at the mutation point** — in the Command Action after the transaction, via
   `activity()->log('...')` or `$model->log(...)` on the `LogsActivity` trait.
2. **Capture the causer** — `activity()->causedBy($user)` (defaults to the authenticated user when
   the LogsActivity/activity context is set); record the subject with `performedOn($model)`.
3. **Keep descriptions actionable** — include the change (`FR-ENR-17: finalized enrollment #42`),
   not a generic "record updated".
4. **Retrieve** via `Activity::forSubject($model)->latest()->get()`, `->forUser($user)`, or the
   model's relationship.
5. **Dual-channel strategy** (DB + file with event correlation) is a project-level choice — keep it
   behind a single logger facade; the UI consumes the DB channel, the file channel enables
   correlation/export.

## Validation

- [ ] Logs happen at the mutation point (Action), after successful persist
- [ ] Causer recorded — `causedBy($user)` or authenticated default
- [ ] Descriptions specific (what changed + which requirement/flow), not generic
- [ ] Retrieval via ActivityLog API / model relation — no raw queries against the log table
- [ ] Log channel config matches the project's audit strategy; no silently-disabled channel

## Example — Correct

```php
activity('enrollment.finalized')
    ->performedOn($enrollment)
    ->causedBy(auth()->user())
    ->withProperties(['slots' => $enrollment->slots])
    ->log('FR-ENR-17: finalized enrollment #42');
```

## Example — Incorrect

```php
// In a controller, after mutating inline, with no causer and a vague message:
activity()->log('updated');
```

## Exception

High-volume non-audit events (cron ticks, read metrics) do not need ActivityLog — keep it for
mutations with traceability value; use logs/metrics for the rest.