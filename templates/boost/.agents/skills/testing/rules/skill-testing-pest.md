# Skill Rule — Pest & Laravel Testing Conventions

Pest/Laravel-specific test-writing conventions for any Laravel project. Complements the
project-agnostic `skills/testing/SKILL.md` with framework-level mechanics; does not replace it.

## Metadata

- **ID:** `rule-skill-testing-pest`
- **Scope:** `skill:testing`
- **Severity:** `contextual` — applies when the project under test is Laravel + Pest
- **Enforcement:** `manual` (review) + `automated` (where a scanner exists)
- **Related:** `rules/global-standards.md` §Testing, `skills/testing/SKILL.md`

---

## Description

Laravel tests the real stack, not the framework. Conventions below keep suites fast, isolated, and
honest by leaning on Laravel's built-in fakes and database tools instead of hand-rolled mocks.

## Trigger

- Writing a Laravel test or leaving a `Pest.php`/`TestCase` setup convention.
- Choosing how to simulate HTTP, filesystem, queue, events, cache, or auth in a test.

## Action

### Database

- Prefer `LazilyRefreshDatabase` over `RefreshDatabase` (runs the migration only once per process;
  keeps the suite fast). Use `RefreshDatabase` only when a test relies on `seed()` between tests.
- Prefer `assertModelExists($model)` over `assertDatabaseHas('table', [...])` — asserts via the
  model, avoiding brittle column lists and honoring casts.
- Never mock Eloquent models — use factories + real database. Eloquent mocks assert implementation,
  not behavior.

### Boundary fakes (mock at the edge, never the core)

| Boundary      | Approach                         |
| ------------- | -------------------------------- |
| External HTTP | `Http::fake()`                   |
| File system   | `Storage::fake()`                |
| Queue         | `Queue::fake()`                  |
| Notifications | `Notification::fake()`           |
| Events        | `Event::fake([Specific::class])` |
| Cache         | `Cache::fake()`                  |
| Auth          | `actingAs($user)`                |

If you're reaching for `shouldReceive()`, reconsider — prefer the `fake()` methods above. Mock the
boundary where the system meets the outside world, not the business logic itself.

### Structure & naming

- One behavior per test; a test name reveals intent (`rejects expired token`, not `test1`).
- For cyclical failure paths, a test may assert an exception with `->throws(ExceptionClass::class)`.
- Group related tests with `describe()`, `group()`, or Pest's `->only()->group()` for focused runs.

### Efficiency

- Run targeted subsets during iterations: `php artisan test --compact --filter={TestName}`.
- Batch all changes, then run the module/scoped suite once. Do not run the whole suite per edit.

## Validation

- [ ] Feature tests use `LazilyRefreshDatabase` (or a justified `RefreshDatabase`)
- [ ] No Eloquent mocking anywhere; fakes used only at the boundary
- [ ] `assertModelExists()` preferred over raw `assertDatabaseHas()`
- [ ] Tests are isolated — no shared mutable state between tests
- [ ] Suites stay fast enough that targeted runs complete in seconds

## Example — Correct

```php
test('registration persists a user', function () {
    $data = RegisterUserData::from(['email' => 'a@b.co', 'name' => 'Ada']);

    $result = app(RegisterUserAction::class)->execute($data);

    expect($result->success)->toBeTrue();
    assertModelExists($result->user);
});
```

## Example — Incorrect

```php
test('creates user', function () {
    $user = Mockery::mock(User::class);           // mocking Eloquent — rejects the pattern
    // ...
    $this->assertDatabaseHas('users', [...]);     // brittle column list; no cast awareness
});
```

## Exception

A suite already established on mocks may convert incrementally rather than all at once — but new
tests follow the conventions above.