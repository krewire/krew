# Skill Rule — Spatie MediaLibrary Conventions

Spatie MediaLibrary implementation conventions for any Laravel project accepting file uploads or
managing media. Complements `skills/building/SKILL.md` and `rules/skill-building-laravel.md` with
media persistence mechanics.

## Metadata

- **ID:** `rule-skill-building-medialibrary`
- **Scope:** `skill:building`
- **Severity:** `contextual` — applies when the project uses `spatie/laravel-medialibrary`
- **Enforcement:** `manual` (review)
- **Related:** `rules/skill-building-laravel.md`, `rules/skill-testing-pest.md`,
  `rules/global-security.md`

---

## Description

All file storage goes through MediaLibrary collections — never raw `Storage::put()` for
user/media files. Each collection declares its own validation (max size, MIME), conversions live on
the collection, and uploads are handled in the domain layer (Action), not the presentation layer.

## Trigger

- Implementing uploads, image handling, document storage, or any media feature.
- Retrieving or deleting stored media; adding image conversions.

## Action

1. **Define collections on the model** via `registerMediaCollections()` — name, `singleFile()`
   where appropriate, per-collection accept rules, and conversions in `registerMediaConversions()`.
2. **Validate server-side** — MIME types (not just file extension) and max size on the collection;
   a client-side check is never trusted.
3. **Sanitize filenames** with `Str::slug()` before storing.
4. **Upload from the domain layer** — the Command Action receives the validated upload and calls
   `$model->addMedia($file)->toMediaCollection('name')`; Livewire only holds the pending upload
   via `WithFileUploads`.
5. **Retrieve** via `getFirstMediaUrl('collection')` / `getFirstMediaUrl('collection', 'thumb')`,
   or `getMedia('documents')` for multiple. Match the method to the collection's cardinality —
   `getFirstMediaUrl()` is for single-file collections; on a multi-file collection it silently
   ignores every item but the first. MediaLibrary owns URL generation — never build media URLs with
   `Storage::url()` or `asset('storage/...')`, and precompute `->getUrl()` results in the query
   layer instead of letting the view call it on raw `Media` models.
6. **Delete** via `clearMediaCollection('collection')` (all media) or `deleteMedia($mediaId)` (one
   item). Never remove `media` rows with raw DB or disk deletes — MediaLibrary tracks those records
   and the conversions and generated variants they leave behind.

### Key configuration

| Setting             | Location                              | Default       |
| ------------------- | ------------------------------------- | ------------- |
| Max file size       | Per collection                        | 10 MB         |
| Accepted MIME types | Per collection                        | model-specific |
| Image conversions   | `registerMediaConversions()`          | thumb 150×150 |
| Storage disk        | `config/media-library.php`            | `public`      |
| Queue               | `config/media-library.php`            | `default`     |

## Validation

- [ ] Upload via MediaLibrary — no `Storage::put()` for user/media files
- [ ] MIME validated server-side (not just extension)
- [ ] Filename sanitized with `Str::slug()`
- [ ] Collection registered on the model; max size + accepted types defined on it
- [ ] Conversions defined where image processing is needed
- [ ] Upload handled in Command Action, not Livewire
- [ ] `addMedia(` appears only in the domain layer — no media calls from the presentation layer
- [ ] Deletion via `clearMediaCollection()` / `deleteMedia()` — no raw DB or disk deletes
- [ ] Media URLs come from MediaLibrary accessors — no `Storage::url()` / `asset('storage/...')`,
      no `Storage` facade in views
- [ ] `getFirstMediaUrl()` used only on single-file collections

## Example — Correct

```php
public function registerMediaCollections(): void
{
    $this->addMediaCollection('avatar')
        ->singleFile()
        ->acceptsFile(fn (File $file) => in_array($file->mimeType, ['image/jpeg', 'image/png']))
        ->registerMediaConversions(function (Media $media): void {
            $this->addMediaConversion('thumb')->width(150)->height(150)->sharpen(10);
        });
}

// In the Action:
$user->addMedia($uploadedFile)->toMediaCollection('avatar');
// Render:
$thumb = $user->getFirstMediaUrl('avatar', 'thumb');
// Delete:
$user->clearMediaCollection('avatar');
```

## Example — Incorrect

```php
Storage::put('avatars/' . $request->file('avatar')->getClientOriginalName(), $file->get());
// ↑ raw storage, unsanitized name, no MIME check, no conversions

DB::table('media')->where('model_id', $user->id)->delete();
// ↑ raw DB delete — leaves orphaned rows and stale conversions

Storage::url('avatars/' . $media->file_name);
// ↑ hand-built URL — breaks when the disk config changes; MediaLibrary owns URLs
```

## Exception

Internal/system-generated files (exports, logs destined for download) may use `Storage::` where the
content is not user media — but any file a user uploads goes through MediaLibrary.