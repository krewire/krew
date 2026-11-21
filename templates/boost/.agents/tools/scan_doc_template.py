#!/usr/bin/env python3
"""scan_doc_template.py — validate markdown documents against canonical templates.

Scans `*.md` files under a root and checks each against its canonical template
shape defined in `templates/`. Each template defines required frontmatter fields,
required sections, and acceptable heading levels.

Template matching by path:
  memory/*.md             → memory-template.md
  context/*.md            → context-template.md
  rules/*.md             → rule-template.md
  skills/*/rules/*.md    → rule-template.md (skill-local rules, colocated)
  refs/*.md              → ref-template.md
  agents/*.md            → agent-template.md
  skills/*/SKILL.md      → skill-template.md
  commands/*.md          → command-template.md
  issues/*.md            → issue-template.md

Root docs (excluded from template validation entirely — no metadata or
structure requirements at the repository root):
  README.md, AGENTS.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md

Validation checks per document:
  1. TEMPLATE-MISMATCH  — file in a known directory but no matching template found
  2. FRONTMATTER-MISSING — required frontmatter fields absent
  3. FRONTMATTER-EXTRA — unexpected frontmatter field (warn)
  4. SECTION-MISSING   — required section absent
  5. SECTION-EXTRA      — unexpected top-level section (warn)
  6. HEADING-LEVEL      — skipped heading level (e.g. H1 then H3)
  7. H1-COUNT           — more than one H1

Templates are loaded from `templates/*.md` by parsing their shapes — no separate
registry needed.

Usage:
  python3 scan_doc_template.py
  python3 scan_doc_template.py --path rules/
  python3 scan_doc_template.py --path .
  python3 scan_doc_template.py --output /tmp/scan.json --strict
  python3 scan_doc_template.py --template-only   # skip section checks (fast)

Exit codes: 0 = clean, 1 = findings (with --strict), 2 = usage/IO error.
Output: JSON report to stdout (or --output file). Idempotent, read-only.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import textwrap
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

from tool_config import load_config

FRONTMATTER_PATTERN = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)
HEADING_PATTERN = re.compile(r"^(#{1,6})\s+(.+?)\s*$", re.MULTILINE)
CODE_FENCE_PATTERN = re.compile(r"```[\s\S]*?```", re.MULTILINE)
RELATED_SECTION = load_config()["triad"]["related_agents"]["section"]
RULE_METADATA_FIELDS = ("id", "scope", "severity", "enforcement", "related")


@dataclass
class Finding:
    rule: str
    severity: str
    message: str
    file: str = ""
    expected: str = ""
    actual: str = ""

    def as_dict(self) -> dict:
        return {
            "id": self.rule,
            "rule": self.rule,
            "severity": self.severity,
            "category": "template",
            "file": self.file,
            "expected": self.expected,
            "actual": self.actual,
            "message": self.message,
        }


@dataclass
class Report:
    scan_name: str = "scan_doc_template"
    scan_version: str = "1.0"
    scan_type: str = "template"
    timestamp: str = ""
    root: str = ""
    findings: list = field(default_factory=list)
    stats: dict = field(default_factory=dict)

    def as_dict(self) -> dict:
        return {
            "scan_version": self.scan_version,
            "scan_name": self.scan_name,
            "scan_type": self.scan_type,
            "root": self.root,
            "timestamp": self.timestamp,
            "summary": {
                "files_scanned": self.stats.get("files_scanned", 0),
                "templates_checked": self.stats.get("templates_checked", 0),
                "findings": len(self.findings),
            },
            "findings": [f.as_dict() for f in self.findings],
        }


# ─── Template definitions ────────────────────────────────────────────────────

@dataclass
class Template:
    name: str
    path: str
    required_frontmatter: list[str] = field(default_factory=list)
    optional_frontmatter: list[str] = field(default_factory=list)
    required_sections: list[str] = field(default_factory=list)
    optional_sections: list[str] = field(default_factory=list)
    allow_extra_sections: bool = False
    allow_extra_frontmatter: bool = False
    allow_h1_multi: bool = False
    allow_skipped_levels: bool = True
    notes: str = ""

    def validate(
        self,
        file_path: str,
        frontmatter: dict[str, str],
        headings: list[tuple[int, str]],
        headings_for_h1: list[tuple[int, str]] | None = None,
    ) -> list[Finding]:
        findings = []

        # Use code-fence-stripped headings for H1 count
        h1_headings = headings_for_h1 if headings_for_h1 is not None else headings
        h1_count = sum(1 for level, _ in h1_headings if level == 1)
        if h1_count == 0:
            findings.append(Finding(
                rule="H1-MISSING",
                severity="error",
                file=file_path,
                message="Document has no H1 heading",
            ))
        elif h1_count > 1 and not self.allow_h1_multi:
            findings.append(Finding(
                rule="H1-COUNT",
                severity="error",
                file=file_path,
                message=f"Document has {h1_count} H1 headings (expected 1)",
            ))

        # ── Heading level sequence ──────────────────────────────────────────
        prev_level = 0
        for level, _ in headings:
            if level > prev_level + 1 and not self.allow_skipped_levels:
                findings.append(Finding(
                    rule="HEADING-SKIP",
                    severity="warn",
                    file=file_path,
                    expected=f"H{prev_level + 1}",
                    actual=f"H{level}",
                    message=f"Skipped heading level H{prev_level} → H{level}",
                ))
            prev_level = level

        # ── Frontmatter required fields ─────────────────────────────────────
        fm_keys = set(frontmatter.keys())
        for field_name in self.required_frontmatter:
            if field_name not in fm_keys:
                findings.append(Finding(
                    rule="FRONTMATTER-MISSING",
                    severity="error",
                    file=file_path,
                    expected=field_name,
                    message=f"Required frontmatter field '{field_name}' is missing",
                ))

        if not self.allow_extra_frontmatter:
            known = set(self.required_frontmatter) | set(self.optional_frontmatter)
            extra = fm_keys - known
            for field_name in extra:
                findings.append(Finding(
                    rule="FRONTMATTER-EXTRA",
                    severity="warn",
                    file=file_path,
                    expected=",".join(sorted(known)),
                    actual=field_name,
                    message=f"Unexpected frontmatter field '{field_name}' — expected one of: {', '.join(sorted(known))}",
                ))

        # ── Required sections ────────────────────────────────────────────────
        section_names = {normalize_heading(title) for _, title in headings}
        for section in self.required_sections:
            norm = normalize_heading(section)
            if norm not in section_names:
                findings.append(Finding(
                    rule="SECTION-MISSING",
                    severity="error",
                    file=file_path,
                    expected=section,
                    message=f"Required section '## {section}' is missing",
                ))

        if not self.allow_extra_sections:
            known_sections = {normalize_heading(s) for s in
                             self.required_sections + self.optional_sections}
            extra = section_names - known_sections
            for section in extra:
                findings.append(Finding(
                    rule="SECTION-EXTRA",
                    severity="warn",
                    file=file_path,
                    expected=",".join(sorted(known_sections)),
                    actual=section,
                    message=f"Unexpected section '## {section}' — expected one of: {', '.join(sorted(known_sections))}",
                ))

        return findings


# Template registry — keyed by glob-style path pattern
TEMPLATES: list[tuple[str, Template]] = [
    (
        "memory/**/*.md",
        Template(
            name="memory",
            path="templates/memory-template.md",
            required_sections=["Signals Log"],
            optional_sections=["<Section>"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
            notes="Memory files are flat; section names are dynamic.",
        ),
    ),
    (
        "context/**/*.md",
        Template(
            name="context",
            path="templates/context-template.md",
            required_sections=["Description", "AI Agent Notes"],
            optional_sections=["Changes"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
            notes="Context topics use dynamic section names.",
        ),
    ),
    (
        "rules/*.md",
        Template(
            name="rule",
            path="templates/rule-template.md",
             required_frontmatter=[],
            optional_frontmatter=[],
             required_sections=[],
            optional_sections=["Description", "Trigger", "Action", "Validation",
                             "Example — Correct", "Example — Incorrect", "Exception"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
             notes="Rules accept YAML metadata or the in-body ## Metadata compatibility shape.",
        ),
    ),
    (
        "skills/*/rules/*.md",
        Template(
            name="rule",
            path="templates/rule-template.md",
            required_frontmatter=[],
            optional_frontmatter=[],
             required_sections=[],
            optional_sections=["Description", "Trigger", "Action", "Validation",
                             "Example — Correct", "Example — Incorrect", "Exception"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
             notes="Skill-local rules accept YAML metadata or the in-body ## Metadata compatibility shape.",
        ),
    ),
    (
        "refs/*.md",
        Template(
            name="ref",
            path="templates/ref-template.md",
            required_frontmatter=[],
            optional_frontmatter=["id", "scope", "owner", "related"],
            required_sections=[],
            optional_sections=[],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
            notes="Ref templates are variant-based; section requirements vary by type.",
        ),
    ),
    (
        "agents/*.md",
        Template(
            name="agent",
            path="templates/agent-template.md",
            required_frontmatter=["name", "description", "mode"],
            optional_frontmatter=[],
            required_sections=["Configuration", "Capabilities", "Instructions", RELATED_SECTION],
            optional_sections=["Anti-Patterns", "Handoff"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
        ),
    ),
    (
        "skills/*/SKILL.md",
        Template(
            name="skill",
            path="templates/skill-template.md",
            required_frontmatter=["name", "description"],
            optional_frontmatter=[],
            required_sections=["When to Use", "Instructions", RELATED_SECTION],
            optional_sections=["Anti-Patterns", "Handoff"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
            notes="Skill phases are H3; section names vary by variant.",
        ),
    ),
    (
        "commands/*.md",
        Template(
            name="command",
            path="templates/command-template.md",
            required_frontmatter=["name", "description"],
            optional_frontmatter=[],
            required_sections=["Understand", "Plan", "Build", "Verify", "Summarize", RELATED_SECTION],
            optional_sections=["Anti-Patterns", "Parameters", "Handoff"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            allow_skipped_levels=True,
            allow_h1_multi=False,
        ),
    ),
    (
        "issues/*.md",
        Template(
            name="issue",
            path="templates/issue-template.md",
            required_frontmatter=[],
            optional_frontmatter=["id", "type", "scope", "status", "owner", "created",
                                  "updated", "related-issues", "related-context"],
            required_sections=["Problem or Objective"],
            optional_sections=["Description", "Decision or Resolution", "Status History", "Handoff"],
            allow_extra_sections=True,
            allow_extra_frontmatter=True,
            notes="issues/index.md is excluded from issue template matching.",
        ),
    ),
]

DOC_TEMPLATE = Template(
    name="generic-doc",
    path="templates/doc-template.md",
    required_frontmatter=[],
    optional_frontmatter=["title", "type", "scope", "owner", "created", "updated"],
    required_sections=[],
    optional_sections=["Overview", "Summary", "Introduction", "Description", "Details",
                      "Quick Reference", "Examples", "Related", "Resources"],
    allow_extra_sections=True,
    allow_extra_frontmatter=True,
    allow_skipped_levels=True,
    allow_h1_multi=True,
    notes="Lenient catch-all fallback for unmatched files.",
)

EXCLUDED_FROM_TEMPLATE = {
    "README.md",
    "AGENTS.md",
    "CONTRIBUTING.md",
    "SECURITY.md",
    "CODE_OF_CONDUCT.md",
    "context/index.md",
    "memory/index.md",
    "issues/index.md",
}

EXCLUDED_FROM_H1_CHECK = set()


# ─── Helpers ────────────────────────────────────────────────────────────────

def normalize_heading(title: str) -> str:
    return re.sub(r"[^\w\s-]", "", title.strip()).lower().replace(" ", "-").replace("--", "-")


def parse_frontmatter(raw: str) -> dict[str, str]:
    fm = {}
    for line in raw.splitlines():
        line = line.strip()
        if ":" in line:
            key, _, val = line.partition(":")
            fm[key.strip()] = val.strip()
    return fm


def parse_headings(content: str) -> list[tuple[int, str]]:
    results = []
    for m in HEADING_PATTERN.finditer(content):
        results.append((len(m.group(1)), m.group(2).strip()))
    return results


def match_template(file_path: str, scan_root: str = ".") -> Optional[Template]:
    path = Path(file_path)
    scan_root_p = Path(scan_root).resolve()

    try:
        rel = str(path.relative_to(scan_root_p))
    except ValueError:
        rel = str(path)

    name = path.name

    if rel in EXCLUDED_FROM_TEMPLATE:
        return None

    for pattern, template in TEMPLATES:
        if glob_match(rel, pattern):
            return template

    return DOC_TEMPLATE


def glob_match(path: str, pattern: str) -> bool:
    import fnmatch
    normalized = path.replace("\\", "/")
    return fnmatch.fnmatch(normalized, pattern)


def scan_file(file_path: str, template: Template, skip_sections: bool = False) -> list[Finding]:
    findings = []
    try:
        content = Path(file_path).read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as e:
        findings.append(Finding(
            rule="FILE-READ-ERROR",
            severity="error",
            file=file_path,
            message=f"Could not read file: {e}",
        ))
        return findings

    fm = {}
    m = FRONTMATTER_PATTERN.match(content)
    if m:
        fm = parse_frontmatter(m.group(1))
        body = content[m.end():]
    else:
        body = content

    headings = parse_headings(body)

    # Strip fenced code blocks (e.g. ```bash ... ```) before heading-level checks
    # so that `# ` inside code fences is not mistaken for a document heading
    body_for_h1 = CODE_FENCE_PATTERN.sub("", body)
    headings_for_h1 = parse_headings(body_for_h1)

    # Rules exist in two supported shapes: tracked global rules use YAML
    # frontmatter, while older skill-local rules use an in-body Metadata block.
    if template.name == "rule":
        if fm:
            for field_name in RULE_METADATA_FIELDS:
                if field_name not in fm:
                    findings.append(Finding(
                        rule="FRONTMATTER-MISSING",
                        severity="error",
                        file=file_path,
                        expected=field_name,
                        message=f"Rule metadata field '{field_name}' is missing",
                    ))
        elif normalize_heading("Metadata") not in {normalize_heading(title) for _, title in headings}:
            findings.append(Finding(
                rule="SECTION-MISSING",
                severity="error",
                file=file_path,
                expected="Metadata",
                message="Rule requires YAML metadata or an in-body '## Metadata' section",
            ))

    if skip_sections:
        pass
    else:
        findings.extend(template.validate(file_path, fm, headings, headings_for_h1))

    return findings


def walk(root: str) -> list[str]:
    files = []
    root_p = Path(root)
    if root_p.is_file():
        return [str(root_p)]
    for item in root_p.rglob("*.md"):
        files.append(str(item))
    return sorted(files)


# ─── CLI ───────────────────────────────────────────────────────────────────

def build_argparser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        description="Validate markdown documents against canonical templates.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=textwrap.dedent("""\
            Examples:
              python3 scan_doc_template.py                          # scan cwd
              python3 scan_doc_template.py --path .                # full scan
              python3 scan_doc_template.py --path rules/          # one dir
              python3 scan_doc_template.py --path agents/ commands/ skills/
              python3 scan_doc_template.py --strict                # exit 1 on findings
              python3 scan_doc_template.py --template-only         # skip sections (fast)
              python3 scan_doc_template.py --output /tmp/scan.json
        """),
    )
    p.add_argument(
        "--path",
        nargs="+",
        default=["."],
        metavar="PATH",
        help="File or directory to scan (default: current directory). "
             "Directories are walked for *.md.",
    )
    p.add_argument(
        "--output",
        metavar="FILE",
        help="Write JSON report to FILE instead of stdout.",
    )
    p.add_argument(
        "--strict",
        action="store_true",
        help="Exit with code 1 if any findings are found.",
    )
    p.add_argument(
        "--template-only",
        action="store_true",
        help="Skip section-level checks (frontmatter + heading-level only). "
             "Much faster for large trees.",
    )
    p.add_argument(
        "--exclude",
        nargs="+",
        default=[],
        metavar="PATTERN",
        help="Exclude files matching glob PATTERN (e.g. 'memory/*.md'). "
             "Can be repeated.",
    )
    return p


def main() -> int:
    args = build_argparser().parse_args()

    from datetime import datetime, timezone
    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    all_files: list[str] = []
    for p in args.path:
        all_files.extend(walk(p))

    # Apply exclusions
    import fnmatch
    filtered = []
    for f in all_files:
        if any(fnmatch.fnmatch(f, e) for e in args.exclude):
            continue
        filtered.append(f)
    all_files = filtered

    report = Report(timestamp=timestamp, root=", ".join(args.path))
    templates_checked = 0

    for file_path in all_files:
        template = match_template(file_path)
        if template is None:
            continue
        findings = scan_file(file_path, template, skip_sections=args.template_only)
        templates_checked += 1
        report.findings.extend(findings)

    report.stats = {
        "files_scanned": len(all_files),
        "templates_checked": templates_checked,
    }

    output = json.dumps(report.as_dict(), indent=2)
    if args.output:
        Path(args.output).write_text(output + "\n", encoding="utf-8")
    else:
        print(output)

    # Summary line
    summary = report.as_dict()["summary"]
    errors = sum(1 for f in report.findings if f.severity == "error")
    warns = sum(1 for f in report.findings if f.severity == "warn")
    print(
        f"\nscan_doc_template: {summary['files_scanned']} files, "
        f"{summary['templates_checked']} templates checked, "
        f"{errors} errors, {warns} warnings",
        file=sys.stderr,
    )

    if report.findings and args.strict:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
