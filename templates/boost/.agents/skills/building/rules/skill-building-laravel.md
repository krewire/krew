# Skill Rule — Laravel Framework Conventions

Framework-level implementation conventions for any Laravel project. Complements the
project-agnostic `skills/building/SKILL.md` with Laravel-specific mechanics; project-specific
invariants (if any) live in a project `context/` file, not here.

## Metadata

- **ID:** `rule-skill-building-laravel`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project is Laravel (or a TALL-stack variant)
- **Enforcement:** `manual` (review) + `automated` (where a scanner exists)
- **Related:** `rules/global-conventions.md`, `rules/global-security.md`, `rules/skill-testing-pest.md`

---

## Description

Laravel conventions that keep code testable, secure, and idiomatic: strict types, constructor
injection, validated input, query-builder SQL, centralized cache keys, proper exception hierarchy,
business rules kept out of the persistence layer, side effects decoupled via events, queues,
and services, and data formatted only at the presentation boundary.

## Trigger

- Writing or reviewing Laravel controllers, Actions, Models, migrations, or Livewire components.
- Choosing error, cache, or data-access strategy in Laravel.

## Action

### Language & safety

- `declare(strict_types=1)` in every PHP file. No `dd`/`dump`/`ray`/`var_dump` in
  committed code. Use `__()` or equivalent translation helpers for every user-facing string
  (i18n-ready).
- Input must be validated before persistence: a FormRequest or validated DTO/`fromArray()` — never
  pass raw `$request->all()` straight into `create`/`update`.
- No raw SQL without bindings — use the query builder or Eloquent. Raw strings are
  an injection vector and a maintenance hazard.

### Structure

- Keep controllers/Livewire components thin: validate input, delegate the mutation to a dedicated
  Action/service class, render. Never call `Model::create()` style logic directly inside a
  component (thin-presentation only).
- Business rules belong in the domain layer (Entities where the project has them), not duplicated
  in every model/controller.
- Constructor injection only — no `app()->make()` or service-locator calls in application code.

### Side effects, queues & services

- Side effects outside the direct request path (mail, notifications, stock updates, exports) go
  through domain events — dispatch, let listeners react. Queue slow listeners so the request never
  waits on them. Do not reach into another feature's service just to trigger a side effect.
- Heavy or blocking work (PDF generation, exports, long-running queries) runs as queued jobs, not
  synchronously in the request — respond immediately with a job reference or a user-facing
  confirmation.
- Infrastructure concerns — file generation, storage writes, PDF rendering — live in a dedicated
  service class, not in model methods or controller inline code.

### Routing & middleware

- Route files scoped by feature/domain; group shared middleware (auth, role/permission, throttling)
  at the group level, not repeated per route. Livewire routes are written like any web route.
- Authorize in dedicated policies/gates, referenced from middleware (`can:`), not inlined into
  route closures.
- Read routes (`index/show`) vs mutations (`store/update/destroy`) — mutations must go through
  Actions, never controller-side `Model::...` calls.
- Bind route params to the model or DTO the route references via route-model binding
  (`/users/{user}` resolving to a typed `User $user`), never `findOrFail($id)` inside the closure;
  no `Route::` wildcards for internal operations.

### Service container & providers

- Register bindings/interfaces in a provider (`register()`), boot side effects in `boot()`. Do not
  write providers inline in routes/controllers.
- Prefer auto-discovery and constructor injection over `bind()` unless an interface needs a
  concrete swap — explicit binding only where the container cannot infer it.
- Third-party package config stays in its published config file (`config/{package}.php`) — customize
  via provider, never edit vendor code.

### Data layer

- Use the `#[Fillable]` attribute (or equivalent explicit mass-assignment control) instead of a
  `$fillable` array property.
- Every foreign key migration declares `onDelete`/`onUpdate` behavior explicitly — never rely on
  implicit default cascade semantics.
- Eager-load relations to avoid N+1; use `with()`/`load()` deliberately, and `select` only what is
  needed for wide tables.
- Validation lives in FormRequests/DTO validation rules (`rules()`/`fromArray()`), not scattered in
  controllers. Validate the shape you accept — never trust the client's declared type.
- Reach for a FormRequest when rules are complex, need custom messages, or carry authorization
  (`authorize()`); inline `$request->validate()` is fine for one or two simple rules with no auth
  logic.
- Query building uses the Eloquent query builder — parameterized, chainable. Favor Eloquent
  aggregates and scopes (`withCount()`, model scopes like `scopeActive`) over `DB::raw()` joins and
  duplicated ad-hoc `where` clauses.

### Data formatting — storage vs presentation

Data is stored in canonical form and formatted only at the presentation boundary. No presentation
formatting leaks into storage, Entities, or DTOs.

- **Store canonical, format at boundary:**
  - Dates: UTC in DB (`datetime`/`timestamp`); format for display via locale-aware helpers
    (`Carbon::locale(...)->isoFormat()`), never `date('d-m-Y', strtotime($model->date))`.
  - Numbers/currency: raw `decimal`/`int` in DB; format via locale-aware number helpers or
    translation keys with placeholders (`__('...', [':amount' => $value])`), never `number_format`
    scattered through views.
  - JSON: typed structs/DTOs in code, `json` column at storage; never hand-rolled
    `json_encode` inline without a defined shape.
- **Precision:** money as `decimal(15,2)` (or typed struct/integer cents when the platform has a
  money type) — never `float`; percentages as typed integers when they drive business rules.
- **i18n:** all display strings via the translation helper; dates/numbers via locale-aware helpers.
- **No formatting in persistence shapes:** persistence arrays/values return canonical values;
  display formatting lives in the presentation layer's getters/computed properties. Keep
  `toPersistable…()`-style canonical output separate from display-formatted output.
- PDF/CSV exports reuse the same presentation helpers — no duplicated formatting in the exporter.

Pitfalls: storing a formatted string ("Rp 1.000.000") instead of the number; `float` money in
migrations/DTOs; formatting a persistence array that is also used for storage.

### Cache & errors

- Cache keys live in a central registry — a dedicated config file or key map — never inline
  `'cache_key'` strings scattered through code. Cache reads/writes belong in the service/action
  layer (`Cache::remember`), not in controllers or views.
- Throw the project's domain exception type for business-rule rejections, not a generic
  `RuntimeException` — so the UI layer can catch and translate it predictably.

### API responses & views

- JSON responses use API Resources (`new UserResource($user)`, `UserResource::collection($users)`)
  so one shape definition is reused everywhere; do not hand-build response arrays.
- No Eloquent queries in Blade — views render data prepared by the controller/component.
- Prefer collection chains (`filter()`, `map()`, `values()`) over imperative loops for
  transform-and-filter work.

## Validation

- [ ] `declare(strict_types=1)` present in every PHP file
- [ ] No debug calls in committed code
- [ ] Translation helper used for user-facing strings
- [ ] Create/update receive only validated/whitelisted input (FormRequest or DTO), never `$request->all()`
- [ ] No raw SQL without bindings
- [ ] Controllers/Livewire thin; mutations delegated to Actions
- [ ] Constructor injection only
- [ ] `#[Fillable]` used; FK migrations declare `onDelete`/`onUpdate`
- [ ] N+1 avoided via eager loading
- [ ] Dates/numbers/currency stored canonical (UTC, decimal not float), formatted only at the presentation boundary with locale-aware helpers
- [ ] Cache keys centralized; domain rejections use the domain exception
- [ ] Side effects dispatched via events; heavy work queued; infrastructure in services
- [ ] API Resources for JSON responses; no Eloquent queries in Blade

## Example — Correct

```php
class CreateCourseRequest extends FormRequest
{
    public function rules(): array
    {
        return [
            'name' => ['required', 'string', 'max:255'],
            'study_program_id' => ['required', 'exists:study_programs,id'],
        ];
    }
}

class CreateCourseAction
{
    public function __construct(private CourseRepository $courses) {}

    public function execute(array $validated): Course
    {
        return $this->courses->create($validated);
    }
}
```

## Example — Incorrect

```php
// Livewire or controller doing persistence directly with raw input:
public function save(Request $request)
{
    Course::create($request->all());          // raw input + persistence in presentation layer
    Cache::put('course_list', $rows);         // inline cache key
}
```

## Exception

Workspace overlays may add stricter project rules; they never relax a mandatory global rule without
an ADR and reviewer sign-off.