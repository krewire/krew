# Skill Rule — Laravel Pulse Conventions

Laravel Pulse (`laravel/pulse`) observability conventions for any Laravel project — dashboard,
authorization, recorders, ingest driver, and custom cards. Complements `skills/building/SKILL.md`
with monitoring/dashboard mechanics.

## Metadata

- **ID:** `rule-skill-building-pulse`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses `laravel/pulse`
- **Enforcement:** `manual` (review)
- **Related:** `rules/global-standards.md` §Maintenance, `rules/global-security.md`

---

## Description

Pulse surfaces operational health (slow requests/jobs/queries, exceptions, cache, user sessions)
on a private dashboard. Config lives in `config/pulse.php`; access is gated by an explicit gate;
ingest driver matches the deployment tier (Redis in production). No Pulse page is ever public.

## Trigger

- Setting up or extending the Pulse dashboard, recorders, or custom cards.
- Choosing the ingest strategy or authorizing access.

## Action

1. **Configure** in `config/pulse.php`: `domain` (restrict by subdomain where possible),
   `middleware` (auth group), `recorders`, `ingest` driver, retention.
2. **Authorize** via `Gate::define('viewPulse', ...)` in a service provider — only explicitly
   authorized roles/admins. Never leave the dashboard public.
3. **Choose ingest by tier**: `redis` for production (durable, low-latency), `file` for local
   development. Recorders capture: SlowRequests, SlowJobs, SlowQueries, SlowOutgoingRequests,
   Exceptions, Cache, UserSessions — enable the set that matches the project's SLOs.
4. **Custom cards** extend `Pulse\Livewire\Card`, register under `dashboard.cards`, define
   authorization via `authorize()` in the card, and surface a metric (rate, count, latency) with a
   bounded query.
5. **Set retention** so the dashboard does not grow unbounded; verify the ingest queue keeps up
   under load.

## Validation

- [ ] Dashboard accessible only via the `viewPulse` gate (authorized roles)
- [ ] Recorders match deployment needs; `redis` ingest in production, `file` in dev
- [ ] Custom cards registered with their own `authorize()`; queries bounded
- [ ] Retention configured; no unbounded data growth

## Example — Correct

```php
// AppServiceProvider
Gate::define('viewPulse', fn (User $user) => $user->hasRole(['admin', 'superadmin']));
```

```php
# config/pulse.php
'domain' => 'pulse.example.com',
'middleware' => ['web', 'auth'],
'ingest' => env('PULSE_INGEST_DRIVER', 'file'),
'recorders' => [SlowRequests::class, SlowQueries::class, Exceptions::class],
```

## Example — Incorrect

```php
// No gate — anyone with the URL sees operational data:
'domain' => null,
'middleware' => ['web'],   // auth missing
'ingest' => 'file',        // file driver in production
```

## Exception

Minimum observability may be a single `SlowRequests` recorder on low-traffic internal tools; the
gate and production ingest constraints still apply.