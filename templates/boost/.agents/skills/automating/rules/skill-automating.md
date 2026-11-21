# Skill Rule — Scripting & Automation Discipline

Project-agnostic standards for devtool scripts and automation: one canonical CLI, predictable
output contracts, tested and performance-bounded. Complements `skills/automating/SKILL.md`; the
scripts themselves are code and follow `rules/global-standards.md` §Code Quality Metrics.

## Metadata

- **ID:** `rule-skill-automating`
- **Scope:** `skill:automating` — binds when creating or modifying automation scripts/devtools
- **Severity:** `contextual` — applies while authoring or extending scripts
- **Enforcement:** `manual` (review) + `automated` (syntax checks, contract tests)
- **Related:** `rules/global-standards.md` §Code Quality Metrics, §Testing; `rules/skill-testing.md` §Quality & Diagnosis; `skills/automating/SKILL.md`

---

## Script Structure

Scripts follow one canonical shape so helpers, report wiring, and CLI behavior are shared instead
of reinvented per script. Deviation means duplicated logic that drifts.

| Rule | Description | Validation |
|------|-------------|------------|
| **SC-01** Canonical template | One template (argument parsing, shared helpers, report wiring); scripts differ only in the scan logic and constants. | Diff against template |
| **SC-02** Parser-first CLI | Real argument parsing, typed signatures, `--help`; never hand-rolled `sys.argv` slicing; no silently ignored flags. | `--help` smoke test |
| **SC-03** Clean exit codes | 0 on success/no findings; 1 when `--strict` finds issues; nonzero on error — the exit code is the batch/CI signal. | Exit-code contract test |
| **SC-04** Idempotent re-runs | Same input → same findings; never mutates project files and never destroys prior artifacts — a repeat run is safe. | Run twice, diff output |
| **SC-05** Pure scanner functions | Scanner takes the file list + optional scope and returns findings; no stdout writes, no re-discovery, no side effects inside scanners. | Review |
| **SC-06** Compiled patterns | Pre-compiled module-level `RE_*` regex constants, explicit about what they target. | Review; perf smoke |
| **SC-07** Standard findings | Finding fields: deterministic id (`{rule}-{seq}`), project-relative path, real line number, concrete suggestion, authoritative reference. | Contract test |

Template skeleton:

```python
#!/usr/bin/env python3
"""{Name} — one-line description."""

from __future__ import annotations
import argparse, json, sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent   # project root
SRC_DIR = ROOT / "src"                           # adjust to the project's source root
OUTPUT_DIR = Path(__file__).parent / "outputs"
SCAN_NAME = "{name}"

@dataclass
class Finding: ...    # standardized finding fields (SC-07)

def read_file(path: Path) -> str: ...            # one read path, errors tolerated
def relative_path(path: Path) -> str: ...        # project-relative, portable across clones
def scan_something(files: list[Path], scope: str | None = None) -> list[Finding]: ...
def build_report(findings, scope, start_time): ...
def parse_args() -> argparse.Namespace: ...      # standard flags (IN-01)
def print_summary(result) -> None: ...
def main() -> None: ...

if __name__ == "__main__":
    main()
```

## Script Interface

Every script exposes the same CLI so agents and CI can drive any script without learning per-script
quirks. A script that invents its own flags or writes reports somewhere unexpected breaks the
automation discipline.

| Rule | Description | Validation |
|------|-------------|------------|
| **IN-01** Standard flags | Support `--scope/-m` (target a sub-area), `--output/-o`, `--format/-f`, `--verbose/-v`, `--quiet/-q`, `--strict/-s`, `--json`. | `--help` parse |
| **IN-02** stdout/stderr contract | Data on stdout only via the requested format; `--quiet` suppresses findings noise; diagnostics/errors go to stderr; the exit code is the batch signal. | Pipe tests |
| **IN-03** Valid JSON out | `--json` emits schema-stable JSON on stdout that round-trips through standard parsers. | Pipe to a JSON parser |
| **IN-04** Strict gate | `--strict` flips the exit code to 1 when findings exist — the pipeline gate. | Known-finding fixture |
| **IN-05** Output naming | Default output `outputs/{timestamp}-{name}.json` — unique per run, history preserved, never silently overwritten. | Inspect outputs/ |
| **IN-06** Output location | Reports land in a dedicated gitignored `outputs/` directory (committed `.gitkeep`); never the repo root, never a fixed name like `report.json`. | Review |
| **IN-07** Discoverable | Scripts documented in a README/table so the automation-first workflow finds them. | Doc-link check |

Why the flags exist: `--scope` keeps a run cost-bounded (full-corpus is the exception); `--output`
lets CI park reports where needed; `--format`/`--json` keep the human and the pipe happy; `--quiet`
is the batch mode where the exit code is the signal; `--strict` is the pipeline gate.

Example: `python3 scan_things.py --scope payments --strict` scans one area and fails the pipeline
on any finding; `python3 scan_things.py --json | jq '.summary'` feeds machines.

## Testing & Performance

A script that silently scans the wrong files is worse than no script — it certifies false
confidence. Scripts must be tested, and must stay fast enough to be run often; slow or wrong
scripts die from disuse.

| Rule | Description | Validation |
|------|-------------|------------|
| **TP-01** Scripts are tested | At minimum a manual smoke: scoped run, JSON schema, quiet mode, strict exit codes. An automated suite is preferred. | Smoke checklist |
| **TP-02** Test the contract, not "no crash" | Assert schema fields, severity enum, id uniqueness, and scoping — not merely a zero exit code. | Suite assertions |
| **TP-03** Test scoping explicitly | Run against a case with a known finding (exit 1) and a clean case (exit 0); testing only the full corpus hides scoping bugs. | Known-fixture runs |
| **TP-04** Size/perf bounded | Full-scan target < 30s; pre-compiled regexes, `pathlib`, stream large files, cache repeated reads, no duplicate full-corpus scans. | `time` smoke test |
| **TP-05** Regression-checked | The automated suite passes after every script edit and before skill integration. | Post-edit run |

Performance guidance:

| Guideline | Why |
|-----------|-----|
| Pre-compile regexes (module-level `RE_*`) | Recompiling per file turns seconds into minutes on large trees |
| `pathlib` over `os.path` | Consistent join/glob, no hidden path chains |
| Stream large files | Process line-by-line instead of loading whole dumps into memory |
| Cache file reads / share discovery | One discovery pass; pass content between scanners instead of re-reading |
| < 30s full scan | A scanner that takes minutes is only run on request — the discipline dies from disuse |

## Reuse-Before-Create & Skill Integration

Scripts exist to serve the agent-skill layer. Before writing a new script, check what already exists;
a duplicate scanner drifts from the canonical rule the moment the rule text changes.

| Rule | Description |
|------|-------------|
| **RB-01** Reuse before create | Check `tools/README.md` and skill `## Automation Scripts` tables first. If the pattern is covered, use the existing tool — never write a second script for the same rule family. |
| **RB-02** No one-offs in `tools/` | A script used a handful of times (one-time conversion, temporary data fix) goes to `/tmp` (e.g. `/tmp/migrate_x.py`), runs, then is discarded. `tools/` is only for durable, reusable devtools. |
| **RB-03** Justified first | Only add a scanner/script when a governing spec (or a documented automation need) justifies it. |
| **RB-04** Interpret in skills, not scripts | Scripts produce raw findings; skills do the interpretation (severity judgment, issue bodies, prioritization). The flow is one-directional and repeatable: detect → measure → read → recommend → act. A script that recommends beyond its finding duplicates skill logic; a skill that re-scans bypasses the script and they drift. |
| **RB-05** Discovery layer | New scripts are documented in `## Automation Scripts` tables of the relevant skills and in `tools/README.md` — a script that appears in no table is invisible to the automation-first discipline. |

### Skill Handoffs

| Condition | Action |
|-----------|--------|
| Creating/modifying a script in `tools/` | Follow this rule's structure/interface/testing sections and the canonical template |
| A skill needs a new automation | Check `tools/README.md` + skill tables first (reuse-before-create) |
| Script is a quality gate | Integrate it into the quality-gate commands it serves |
| Multi-scanner initiative | Size-triage it (M/L); stage per script, inform the user if it crosses into L |

### How to Add a New Script

1. Create `tools/scan_{name}.py` following the canonical template (SC-01).
2. Add an `## Automation Scripts` entry to the relevant skills.
3. Test per §Testing & Performance (scoped run, JSON schema, strict exit codes).
4. Verify the output schema matches the standard before any consumer depends on it.
5. Add to the `tools/README.md` table.
6. Commit with a descriptive message per the project's commit conventions.

The script is the artifact, but the *integration* is what makes it useful — table entries and README
before merge, schema verified before integration. Verification includes `git status` + `git diff`
showing only intended files (script + integration-table edits), and no two scripts targeting the same
rule family.

## Validation

- [ ] `--help` parses all standard flags; nothing silently ignored
- [ ] Exit codes: 0 clean, 1 strict-with-findings, nonzero on error
- [ ] `--json` output round-trips through a parser; schema stable
- [ ] Idempotent: two runs produce comparable results and preserve prior artifacts
- [ ] Smoke-tested for scope, strict, quiet, and a known finding; automated suite passes after edits
- [ ] Full-scan wall time within budget (< 30s or trending there)
- [ ] Reuse checked first; script exists in skill tables + README; findings stay raw; no duplicate rule-family scanner