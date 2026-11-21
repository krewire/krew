# Global Language — English-only for Artifacts, User-Language for Communication


Applies to **every** artifact in this registry. Private repo, but English is the single source of truth for collaboration, search, and cross-ecosystem reuse. User-facing chat is bilingual and follows the user's last message.

## Metadata

- **ID:** `rule-global-language`
- **Scope:** `global`
- **Severity:** `mandatory`
- **Enforcement:** `manual` (review) + `automated` (lint/scanner for non-ASCII docs if needed)
- **Related:** `rules/global-conventions.md`, `AGENTS.md §Conventions`, and the verified agent-agnostic-memory runtime skill

---

## Rule

### 1. English-only for all stored artifacts (mandatory)

Every file committed to git **MUST be in English**:

- `*.md` (AGENTS.md, README.md, rules/, refs/, context/, memory/ when tracked, skills/, commands/, agents/)
- Code comments, docblocks (PHPDoc/TSDoc), commit messages, branch names, PR titles/bodies
- `tools/` scripts, `refs/` specs (OpenAPI, ADR)
- `context/` topic files and `rules/` constraints
- `memory/` files (except files that record user verbatim, which may keep the original language but English is preferred)

**Why:** searchability (`grep`/`rg` in one language), cross-repo/ecosystem reuse, and to keep `memory/` → `context/` → `rules/` promotion language-uniform.

### 2. Exception: direct communication with the user

When the **agent talks directly to the user** (CLI response, chat, inline comment in a PR that addresses the user), it **SHOULD follow the user's last message language**:

- User writes Indonesian → answer in Indonesian (concise, CLI-native).
- User writes English → answer in English.
- Mixed/bilingual user → match the dominant language of the last turn, keep artifacts in English, but wrap user-facing summary in user's language.

Examples:

| Stored artifact (EN) | User-facing chat (ID if user used ID) |
|----------------------|----------------------------------------|
| Commit: `feat(memory): add flat-memory pruning` | Agent to user: "Berhasil — `memory/` sudah di-prune 3 file lama." |
| Doc: `context/tool-preferences.md` in EN | Chat: "Ditambahkan ke `context/tool-preferences.md` (dalam bahasa Inggris agar konsisten)." |
| Code comment: `// retry with jitter for idempotency, see ADR-003` | Chat: "Saya menambahkan retry dengan jitter — alasan dicatat di ADR-003." |

### 3. Non-English detection

- In Build/Verify, scan `*.md` and `*.md` headers for non-ASCII sentences outside code blocks; flag for translation. Runtime memory may quote user verbatim, but distilled runtime memory/context records must be English.
- A `memory/private-*.md` file (local-only, gitignored) is the only place allowed to be fully in user language (contains PII) — but English is still preferred if user is comfortable.

### 4. How to comply

- Write docs/code in English first, then add a user-facing Indonesian summary in chat if the user's last message was Indonesian.
- Do not translate `rules/`, `AGENTS.md`, or `refs/` to Indonesian — keep them English, even when explaining to the user (explain in Indonesian, point to English file).
- When quoting user verbatim in a `memory/` file, keep original language + add `[EN: translation]` if needed for search.

---

## Validation

- [ ] `grep -P "[^\x00-\x7F]" --include="*.md"` outside code blocks shows no Indonesian sentences in `rules/`, `refs/`, `context/`, `AGENTS.md`, `README.md` (except this rule's examples table).
- [ ] Commit messages and branch names are English (`feat(scope): desc`).
- [ ] Agent responses in last 10 turns follow user's language, while committed files remain English.

## Anti-Patterns

- **Bilingual docs**: half Indonesian, half English in the same `*.md` → split: doc stays EN, chat is ID.
- **Indonesian commit messages**: `docs: sinkronisasi memori` → `docs(memory): sync autonomous memory layer`.
- **Translating rules to user language and committing**: keep rule file EN, explain rule in chat in user's language.

## References

- `rules/global-conventions.md` §Documentation
- `AGENTS.md §Conventions`
