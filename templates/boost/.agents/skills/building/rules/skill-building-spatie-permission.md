# Skill Rule — Spatie Permission RBAC Conventions

Spatie Permission (`spatie/laravel-permission`) RBAC conventions for any Laravel project — roles,
permissions, seeding, checks, and policy integration. Complements `skills/building/SKILL.md` and
`rules/skill-building-laravel.md` with authorization mechanics.

## Metadata

- **ID:** `rule-skill-building-spatie-permission`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses `spatie/laravel-permission`
- **Enforcement:** `manual` (review)
- **Related:** `rules/global-security.md`, `rules/skill-building-laravel.md`

---

## Description

RBAC is declared once (roles + permissions, seeded idempotently) and checked in two idioms:
policies/gates for object-scoped authorization, Blade directives and `User` checks for quick role
gating. Roles carry functional meaning (what a user can do), permissions stay granular.

## Trigger

- Defining or seeding roles/permissions; adding a guarded route or API.
- Checking authorization in Blade, controller, or a policy.

## Action

1. **Declare roles + permissions once**, idempotently (permission seeder). Permission names group by
   area (`module.ability`, e.g. `enrollment.create`). Never grant permissions inline in code.
2. **Assign roles** via `$user->assignRole('coordinator')`; attach permissions to roles via
   `$role->givePermissionTo([...])` — permissions attach to roles, roles attach to users.
3. **Check authorization**:
   - Object-scoped → **policies** (`$this->authorize('update', $enrollment)`) + gates.
   - Route-level → middleware alias (`->middleware('role:coordinator')` /
     `->middleware('permission:enrollment.create')`).
   - Blade → `@role('coordinator')`, `@hasrole`, `@can` (delegates to policy).
   - PHP → `$user->hasRole()`, `$user->hasPermissionTo()`.
4. Use the guard consistently (`web` by default; API guards keep their own role/permission tables
   when separate).
5. Cache flushing: `laravel-permission` busts its own permission cache on role/permission mutation;
   do not manually `forgetCachedPermissions()` unless seeding at runtime.

## Validation

- [ ] Roles/permissions seeded idempotently, grouped by area, with the correct guard
- [ ] Object checks go through policies; route checks through role/permission middleware
- [ ] Blade uses `@role`/`@hasrole`/`@can` — no ad-hoc `if (Auth::user()?->id === …)` for roles
- [ ] New permission names documented/registered; no magic strings scattered in logic

## Example — Correct

```php
// seeder
$role = Role::create(['name' => 'coordinator']);
$role->givePermissionTo(['enrollment.create', 'enrollment.read']);

// policy
public function update(User $user, Enrollment $enrollment): bool
{
    return $user->hasPermissionTo('enrollment.update')
        && ($user->hasRole('superadmin') || $user->id === $enrollment->coordinator_id);
}
```

```blade
@role('coordinator')
    <x-ts-button>Approve</x-ts-button>
@endrole
```

## Example — Incorrect

```php
// role check ad-hoc, permission bypassed, object scope ignored
if (auth()->user()->id === $enrollment->coordinator_id) {
    // approval allowed — no permission check, no policy
}
```

## Exception

Flat/functional role sets (few roles, permission-gated) are a valid simplification — the invariant
is: object-scoped checks still go through policies, never raw `if (id === …)` in controllers.