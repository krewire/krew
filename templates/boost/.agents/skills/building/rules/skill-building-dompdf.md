# Skill Rule — DomPDF (Laravel) Conventions

DOMPDF via `barryvdh/laravel-dompdf` conventions for any Laravel project generating PDFs — Blade
rendering, paper/options, asset embedding, and unicode. Complements `skills/building/SKILL.md` with
PDF generation mechanics.

## Metadata

- **ID:** `rule-skill-building-dompdf`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses `barryvdh/laravel-dompdf`
- **Enforcement:** `manual` (review)
- **Related:** `rules/skill-building-blade` (see `rules/skill-building-tallstack.md`), `rules/global-security.md`

---

## Description

Generate PDFs from Blade views with explicit paper size, orientation, and options; embed assets
(images, CSS, unicode fonts) so the PDF renders consistently offline. Output streams or downloads —
never write user-generated PDF content unescaped.

## Trigger

- Generating a report/invoice/certificate PDF from data or a Blade view.
- Embedding images/fonts or handling unicode (non-Latin) in PDF output.

## Action

1. **Render from Blade** — `Pdf::loadView('pages.report', $data)`; keep the view presentation-only
   (reuse `rules/skill-building-tallstack.md` §Blade: no business logic in the template).
2. **Set paper & orientation** explicitly — `->setPaper('a4', 'portrait'|'landscape')` and options
   (`isHtml5ParserEnabled`, `isRemoteEnabled`, margin) via config or per-call.
3. **Embed assets** — reference CSS/images with absolute/public paths that DOMPDF can resolve;
   enable remote access only where needed (weigh the SSRF/leak risk).
4. **Handle unicode** — declare a unicode TTF font in `config/dompdf.php` font-family for non-Latin
   scripts; CP1252/DejaVu defaults mangle most non-ASCII.
5. **Stream/download** — `->stream()` for preview, `->download()` for file delivery; never debug-dump
   (`dd($pdf)`) the blob in committed code.

## Validation

- [ ] Paper size/orientation set; HTML5 parser enabled where needed
- [ ] CSS/images resolve from absolute paths; unicode font declared for non-Latin content
- [ ] Blade view presentation-only; escaped dynamic content
- [ ] `stream()`/`download()` chosen; no debug output of the blob
- [ ] Remote access enabled only when required, never by default

## Example — Correct

```php
$pdf = Pdf::loadView('reports.grade-card', ['report' => $report])
    ->setPaper('a4', 'landscape')
    ->setOption('isHtml5ParserEnabled', true);

return $pdf->download("grade-card-{$report->id}.pdf");
```

## Example — Incorrect

```php
// Raw HTML string with hardcoded CSS, no options, no escaping:
$pdf = Pdf::loadHTML("<h1>{$report->title}</h1><style>h1{color:red}</style>");
return $pdf->stream();
```

## Exception

For high-fidelity/very large documents, consider a headless-browser renderer; DOMPDF is the
lightweight default for structured reports.