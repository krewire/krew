# Skill Rule — TallStack UI Frontend Conventions

TALL-stack (Tailwind, Alpine, Livewire, Laravel) frontend conventions for any Laravel project using
Livewire + TallStackUI + Tailwind CSS v4 + Vite. Complements the project-agnostic
`skills/building/SKILL.md` with the UI/component layer mechanics.

## Metadata

- **ID:** `rule-skill-building-tallstack`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses Livewire + TallStackUI + Tailwind v4
- **Enforcement:** `manual` (review) + `automated` (where a scanner/prettier exists)
- **Related:** `rules/global-standards.md` §Design (WCAG), `rules/skill-building-laravel.md`,
  `rules/skill-building-medialibrary.md`, `rules/skill-testing-pest.md`

---

## Description

A TALL-stack frontend has three responsibility layers that must not bleed into each other:
**Livewire** owns component state/actions (thin, delegating to Actions), **Blade/TallStackUI**
owns presentation (no business logic), and **Tailwind v4 + Vite** owns styling/build (utilities +
CSS-first theme, no custom CSS by default). Layout, responsiveness, dark mode, tables, and file
uploads each have a mandated mechanism; replacing them with hand-rolled approximations is where
frontend debt accumulates. Conventions below keep each layer idiomatic.

## Trigger

- Building or modifying a Livewire component, Blade view, TallStackUI usage, or Tailwind theming.
- Adding Vite entry points/plugins or debugging build/HMR.
- Implementing list/CRUD views or file upload flows.

## Action

### Livewire components (mechanics)

- **Thin component** — validate input, call a Command Action (via method/constructor injection),
  read-only entities for render; never `Model::create()` or business rules inside a component.
- Component properties are simple scalars/arrays or Form Objects; complex forms use a Form Object
  (`#[Validate]` / `form()`), not raw component props.
- Derived values for Blade come from Alpine/computed expressions or small read-only methods — no
  `@if` with inline business expressions and no model mutation in the template.
- Livewire owns derived metrics and visibility flags (`public int $completionRate`, `public bool $isEditable`),
  computed in `mount()` or `#[Computed]` getters, so templates only render precomputed values.
- Alpine.js owns lightweight client-side UI state (`x-data="{ open: false }"` for toggles, dropdowns,
  local filtering) and never re-implements server business rules.
- File uploads: `WithFileUploads` in the component, then delegate to the Action for persistence
  (see `rules/skill-building-medialibrary.md` where MediaLibrary is used).

### Blade & TallStackUI (presentation)

- Prefer TallStackUI components (`x-ts-*`: table, card, modal, input, select, badge, toast) over
  hand-rolled HTML; the component library is the primary UI vocabulary.
- Customize components via `TallStackUi::customize()` (v3+) — **never** the legacy
  `TallStackUi::personalize()` which no longer exists.
- Blade is a trivial binding layer: `{{ $value }}`, `@foreach`, and `@if` gates. **No `@php` blocks**
  with calculations, array assembly, or business branching — they are untyped, untestable, and
  invisible to automated scans. Every calculation happens in Livewire before rendering.
- `@if` is only a trivial gate on an already-computed flag (`@if ($isVisible)`), never an expression
  with business logic (`@if ($user->role === 'admin' && $stats['x'] > 0)`). Prefer Livewire boolean
  properties (server-driven) and Alpine `x-show`/`x-if` (client-driven) for visibility toggles.
- Role-based visibility uses Spatie RBAC Blade directives (`@role`, `@hasrole`, `@can`) instead of
  `hasRole()` calls inline in the template.
- Blade holds presentation only: layout, responsive structure, dark mode toggle, translations
  (`__()`), and accessibility (WCAG 2.1/2.2 AA: contrast, focus, keyboard, reflow, icon labels).
- No inline `style=""`; no custom CSS when a utility or `x-ts-*` component suffices. If a design
  genuinely cannot be expressed with utilities/components, document why in the PR.

### Layout, Responsive & Dark Mode

- Use the project's shared layout shell (drawer sidebar + navbar header) for navigation chrome —
  never hand-rolled per-page chrome.
- Content width uses a container pattern (`max-w-7xl mx-auto`).
- **Mobile-first** responsive design: base classes for mobile, then `sm:`/`md:`/`lg:` for larger
  viewports. Test at mobile (320px), tablet, and desktop before shipping; no horizontal scrolling or
  overlap at 320px. Avoid desktop-only layouts (`hidden md:block` for everything meaningful).
- Dark mode uses the dual signal `data-theme` attribute + `.dark` class, kept in sync by a small
  theme helper (reading/writing the persisted preference). Prefer TallStackUI's `<x-theme-switch>`
  for the toggle — do not hand-roll a new toggle. Persistence (e.g. localStorage) is part of the
  feature; a toggle that flips the class but not persist the preference is half a feature.
- Colors come from `@theme` CSS variables — never hardcoded hex or default Tailwind palette values.
  A hardcoded `bg-white` must always have a `dark:` counterpart.

### Tables & File Uploads

- **Tables:** list views use TallStackUI's `x-ts-table` with sorting and pagination — do not
  hand-roll an HTML `<table>` because "the data is simple". The component ships accessible header
  semantics (`scope`, `aria-sort`); bulk-selection header checkboxes carry an explicit
  `aria-label`. Wire columns to a paginated, eager-loaded query — an un-paginated listing grows
  unbounded with the dataset. Reuse the project's shared CRUD record-manager base component where
  one exists instead of copying table markup from another feature.
- **File uploads:** upload-capable components use Livewire's `WithFileUploads` trait and pass the
  file to the Action; the Action performs the media-library call (collection, conversions,
  retrieval) inside its transaction — never a component side effect. Validate mime/size in the
  component for UX and re-validate in the Action for enforcement. Raw `store()`/local-storage calls
  bypass collections and conversions entirely. Media specifics come from
  `rules/skill-building-medialibrary.md`.

### Tailwind CSS v4 (styling)

- **CSS-first config** — no `tailwind.config.js`; theme tokens live in `resources/css/app.css` via
  `@theme` (`--color-*`), with `@import` for layers (Tailwind v4 syntax — do not mix in v3
  `@layer`/`tailwind.config.js` style).
- Check `resources/css/app.css` for existing tokens before adding a new color/spacing value.
- Use the **semantic palette** (e.g. `--color-primary-*`, `--color-base-*`) — never arbitrary
  `bg-blue-500`, `text-red-500`, or `bg-[#123456]`, which bypass theming/dark mode. Use
  `bg-[var(--color-primary)]` only when a semantic utility does not exist; otherwise use `bg-primary`.
- Palette tokens can be overridden at runtime via CSS variables without recompilation — hardcoded
  colors break brand theming and dark mode, and custom CSS duplicates framework behavior.
- Dark mode via `.dark` class variant (`dark:bg-base-200`) with a `dark:` counterpart for any
  `bg-white`/`bg-black`; the dual signal is `data-theme` + `.dark`.
- Prefer utilities (`mt-4`, `flex`, `gap-2`, `rounded-lg`) over custom CSS.

```css
/* resources/css/app.css */
@import "tailwindcss";
@theme {
    --color-primary: var(--brand-primary, #4f46e5);
    --color-base-100: var(--brand-base-100, #ffffff);
}
```

### Vite (build)

- Entries declared in `vite.config.js` with `laravel-vite-plugin` and `@tailwindcss/vite`; CSS entry
  pointed at the Tailwind layer. No entry without a matching input file.
- `@vite` directive renders manifest assets; use `npm run build` before Verify for any Blade/CSS change,
  and `npx prettier --check` for non-PHP files.

## Validation

- [ ] Livewire thin — mutations delegated to Actions; no `Model::*` in components
- [ ] Blade presentation-only; `x-ts-*` primacy; `__()` for user strings
- [ ] No `@php` business logic; `@if` gates only precomputed flags (no inline expressions)
- [ ] Derived metrics/visibility computed in Livewire or Alpine, not in the template
- [ ] Layout uses the shared shell; responsive tested at 320px/tablet/desktop; no horizontal scroll
- [ ] Dark mode: `data-theme` + `.dark` synced, toggle persists the preference, `dark:` counterpart
  present wherever a light background is hardcoded
- [ ] Tables via `x-ts-table` with sorting/pagination/accessible headers; queries paginated + eager-loaded
- [ ] Uploads via `WithFileUploads` + Action-side media persistence
- [ ] No arbitrary hex/default Tailwind colors; no inline `style=`; no custom CSS when utilities/components suffice
- [ ] Tailwind v4 CSS-first (`@theme` in `app.css`, `@import` layers), not `tailwind.config.js`
- [ ] Vite entries/plugins correct; `npm run build` clean; `npx prettier --check` clean

## Example — Correct

```blade
<x-ts-card title="{{ __('Enrollment') }}">
    <x-ts-input wire:model="form.name" label="Name" />
    <x-ts-button onclick="Livewire.dispatch('open')">{{ __('Save') }}</x-ts-button>
</x-ts-card>
```

Tints use the semantic palette only, e.g. `bg-primary text-base-content dark:bg-base-200`.

## Example — Incorrect

```blade
{{-- business logic inline, custom HTML, arbitrary color --}}
@if ($program->slots()->count() > $program->capacity)
    <div style="background: #123456">{{ $program->name }}</div>
@endif
```

## Exception

A legacy component may migrate incrementally from hand-rolled HTML to `x-ts-*`; transitional shim
classes in `@layer components` are tolerated while migrating, never extended for new designs.