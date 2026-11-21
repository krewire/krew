#!/usr/bin/env python3
"""scan_broken_links.py — validate markdown internal + external links.

Scans every `*.md` file under a root and reports broken links:
  1. Internal links `[label](path)` — `path` must resolve to an existing file or
     directory. Relative paths are resolved against the file's directory.
  2. Anchor links `[label](path#anchor)` — `#anchor` must match a heading in the
     target file (case-insensitive, normalized like GitHub's slugger).
  3. External links `[label](http(s)://...)` — HTTP HEAD/GET; non-2xx is broken.
     Skipped with --skip-external.

Notes:
  - `~/...` and `../...` paths are reported as LINK-CROSS-REPO, not broken, when
    they point outside the scan root.
  - Code-fence links (```...```) are skipped.
  - HTML `<a href="...">` and bare URLs are out of scope.
  - Caches external results at `.cache/scan_broken_links.json` to keep reruns
    fast and polite (rate-limited).

Usage:
  python3 scan_broken_links.py
  python3 scan_broken_links.py --skip-external
  python3 scan_broken_links.py --path README.md
  python3 scan_broken_links.py --path docs/ agents/ rules/
  python3 scan_broken_links.py --path .
  python3 scan_broken_links.py --check-external --timeout 5
  python3 scan_broken_links.py --output /tmp/links.json --strict
  python3 scan_broken_links.py --dry-run

With no --path, the script scans every *.md under the current working directory.
--path takes one or more files or directories; each directory is walked for
*.md. Use --path to scope a quick check without walking the whole tree. Links
are resolved relative to each source file (not the path you pass).

Exit codes: 0 = clean, 1 = findings (with --strict), 2 = usage/IO error.
Output: JSON report to stdout (or --output file). Idempotent, read-only.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable

# Markdown link with optional anchor:  [label](path)
LINK_PATTERN = re.compile(r"(?<!!)\[([^\]]+)\]\(([^)\s]+)(?:\s+\"[^\"]*\")?\)")
# Markdown image link: ![label](path) — also validated.
IMAGE_PATTERN = re.compile(r"!\[([^\]]*)\]\(([^)\s]+)(?:\s+\"[^\"]*\")?\)")
HEADING_PATTERN = re.compile(r"^(#{1,6})\s+(.+?)\s*$", re.MULTILINE)
CODE_FENCE_PATTERN = re.compile(r"```.*?```", re.DOTALL)
# All inline code spans (captures the full span for overlap checking).
INLINE_CODE_PATTERN = re.compile(r"`[^`\n]+`")
EXTERNAL_SCHEMES = ("http://", "https://", "mailto:")


@dataclass
class Finding:
    rule: str
    severity: str
    message: str
    file: str = ""
    link: str = ""

    def as_dict(self) -> dict:
        return {
            "id": self.rule,
            "rule": self.rule,
            "severity": self.severity,
            "category": "link",
            "file": self.file,
            "link": self.link,
            "message": self.message,
        }


@dataclass
class Report:
    scan_name: str = "scan_broken_links"
    scan_version: str = "1.0"
    scan_type: str = "link"
    timestamp: str = ""
    root: str = ""
    external_checked: bool = False
    findings: list = field(default_factory=list)
    stats: dict = field(default_factory=dict)

    def as_dict(self) -> dict:
        return {
            "scan_version": self.scan_version,
            "scan_name": self.scan_name,
            "scan_type": self.scan_type,
            "root": self.root,
            "timestamp": self.timestamp,
            "external_checked": self.external_checked,
            "summary": {
                "files_scanned": self.stats.get("files_scanned", 0),
                "links_total": self.stats.get("total", 0),
                "internal": self.stats.get("internal", 0),
                "external": self.stats.get("external", 0),
                "anchors": self.stats.get("anchors", 0),
                "findings": len(self.findings),
            },
            "findings": [f.as_dict() for f in self.findings],
        }


def find_code_spans(text: str) -> list[tuple[int, int]]:
    """Return (start, end) of every fenced and inline code span in text."""
    spans: list[tuple[int, int]] = []
    for m in CODE_FENCE_PATTERN.finditer(text):
        spans.append((m.start(), m.end()))
    for m in INLINE_CODE_PATTERN.finditer(text):
        spans.append((m.start(), m.end()))
    return spans


def in_code(pos: int, spans: list[tuple[int, int]]) -> bool:
    return any(start <= pos < end for start, end in spans)


def extract_links(text: str) -> list[tuple[str, str]]:
    """Return (label, target) for real markdown links/images — skips code spans.

    Detects links on the original text so we can use span positions from
    `find_code_spans` directly. This correctly skips:
      - Links inside fenced code blocks (```...```)
      - Links whose label is an inline code span (`` [`label`](url) ``)
    """
    spans = find_code_spans(text)
    out: list[tuple[str, str]] = []

    for m in LINK_PATTERN.finditer(text):
        if in_code(m.start(), spans):
            continue
        # group(1) = label, group(2) = target — regex already captured these.
        out.append((m.group(1), m.group(2)))

    for m in IMAGE_PATTERN.finditer(text):
        if in_code(m.start(), spans):
            continue
        out.append((m.group(1), m.group(2)))

    return out


def slugify_anchor(text: str) -> str:
    """GitHub-style anchor slug: lowercase, replace whitespace with -, strip punctuation."""
    text = text.lower()
    text = re.sub(r"[^\w\s\-]", "", text, flags=re.UNICODE)
    text = re.sub(r"\s+", "-", text.strip())
    return text


def extract_headings(text: str) -> set[str]:
    """Return the set of all heading anchors in a markdown file."""
    headings: set[str] = set()
    for _, title in HEADING_PATTERN.findall(text):
        headings.add(slugify_anchor(title))
    return headings


def is_external(target: str) -> bool:
    return target.startswith(EXTERNAL_SCHEMES)


def resolve_internal(source: Path, target: str, root: Path) -> tuple[Path | None, str | None]:
    """Resolve an internal link target. Returns (resolved_path, anchor_or_None).

    - If target is a relative path, resolve against source.parent.
    - If target starts with `/`, resolve against root.
    - Strips a `#anchor` before resolving the path.
    """
    anchor: str | None = None
    if "#" in target:
        target, anchor = target.split("#", 1)
    if not target:
        return None, anchor
    if target.startswith("/"):
        return (root / target.lstrip("/")).resolve(), anchor
    return (source.parent / target).resolve(), anchor


def is_outside_root(path: Path, root: Path) -> bool:
    try:
        path.relative_to(root)
        return False
    except ValueError:
        return True


def check_external(url: str, timeout: int, cache: dict) -> tuple[int, str]:
    """Return (status_code, message). Uses cache for repeatability across runs."""
    if url in cache:
        return cache[url]["status"], cache[url]["message"]
    try:
        req = urllib.request.Request(url, method="HEAD")
        req.add_header("User-Agent", "scan_broken_links/1.0 (+local dev tool)")
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            status = resp.getcode()
            message = "ok" if 200 <= status < 400 else f"HTTP {status}"
    except urllib.error.HTTPError as e:
        status = e.code
        # Some hosts reject HEAD; fall back to GET with Range header.
        if status in (403, 405):
            try:
                req = urllib.request.Request(url, method="GET")
                req.add_header("Range", "bytes=0-0")
                with urllib.request.urlopen(req, timeout=timeout) as resp:
                    status = resp.getcode()
                    message = "ok" if 200 <= status < 400 else f"HTTP {status}"
            except Exception as e2:
                message = f"HTTP {status} (GET fallback: {e2})"
        else:
            message = f"HTTP {status}"
    except urllib.error.URLError as e:
        status = 0
        message = f"URL error: {e.reason}"
    except TimeoutError:
        status = 0
        message = "timeout"
    except Exception as e:
        status = 0
        message = f"error: {e}"
    cache[url] = {"status": status, "message": message}
    return status, message


def load_cache(root: Path) -> dict:
    cache_path = root / ".cache" / "scan_broken_links.json"
    if cache_path.is_file():
        try:
            return json.loads(cache_path.read_text())
        except (OSError, json.JSONDecodeError):
            return {}
    return {}


def save_cache(root: Path, cache: dict) -> None:
    if not cache:
        return
    cache_path = root / ".cache" / "scan_broken_links.json"
    try:
        cache_path.parent.mkdir(parents=True, exist_ok=True)
        cache_path.write_text(json.dumps(cache, indent=2))
    except OSError as e:
        print(f"[warn] could not write cache: {e}", file=sys.stderr)


def scan_file(
    path: Path,
    root: Path,
    headings_cache: dict,
    check_ext: bool,
    timeout: int,
    ext_cache: dict,
) -> tuple[list[Finding], dict]:
    """Scan a single markdown file. Returns (findings, stats)."""
    findings: list[Finding] = []
    stats = {"total": 0, "internal": 0, "external": 0, "anchors": 0}
    try:
        text = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as e:
        findings.append(Finding(
            rule="READ-ERROR",
            severity="error",
            message=f"could not read file: {e}",
            file=str(path.relative_to(root)),
        ))
        return findings, stats

    for _label, target in extract_links(text):
        stats["total"] += 1
        # Cross-repo: path starting with `~` or `~/`
        if target.startswith("~") or target.startswith("../") and ".." in Path(target).parts[:2]:
            # Note: only flag truly cross-repo style paths. A single `..` that
            # still resolves inside root is fine; resolve and verify.
            if target.startswith("~"):
                findings.append(Finding(
                    rule="LINK-CROSS-REPO",
                    severity="info",
                    message=(
                        f"cross-repo reference '{target}' points outside this repo — "
                        f"references must stay relative within the scan root"
                    ),
                    file=str(path.relative_to(root)),
                    link=target,
                ))
                continue
        if is_external(target):
            stats["external"] += 1
            if check_ext:
                status, message = check_external(target, timeout, ext_cache)
                if status < 200 or status >= 400:
                    findings.append(Finding(
                        rule="LINK-EXTERNAL-BROKEN",
                        severity="error",
                        message=f"external link broken: {message}",
                        file=str(path.relative_to(root)),
                        link=target,
                    ))
        else:
            stats["internal"] += 1
            resolved, anchor = resolve_internal(path, target, root)
            if resolved is None:
                continue  # bare `#anchor` to current file
            # External-repo style that escapes root via `../` chains → mark info.
            try:
                resolved.relative_to(root)
                in_root = True
            except ValueError:
                in_root = False
            if not in_root:
                findings.append(Finding(
                    rule="LINK-CROSS-REPO",
                    severity="info",
                    message=(
                        f"path '{target}' resolves outside scan root — "
                        f"cross-repo path references are discouraged"
                    ),
                    file=str(path.relative_to(root)),
                    link=target,
                ))
                continue
            if not resolved.exists():
                findings.append(Finding(
                    rule="LINK-INTERNAL-BROKEN",
                    severity="error",
                    message=f"path '{target}' does not resolve to an existing file or directory",
                    file=str(path.relative_to(root)),
                    link=target,
                ))
                continue
            if anchor:
                stats["anchors"] += 1
                if resolved.is_file() and resolved.suffix.lower() in {".md", ".markdown"}:
                    cache_key = str(resolved)
                    if cache_key not in headings_cache:
                        try:
                            headings_cache[cache_key] = extract_headings(resolved.read_text(encoding="utf-8"))
                        except OSError:
                            headings_cache[cache_key] = set()
                    if slugify_anchor(anchor) not in headings_cache[cache_key]:
                        findings.append(Finding(
                            rule="LINK-ANCHOR-BROKEN",
                            severity="error",
                            message=f"anchor '#{anchor}' not found in {resolved.name}",
                            file=str(path.relative_to(root)),
                            link=f"{target}#{anchor}",
                        ))
    return findings, stats


def build_report(
    root: Path,
    paths: list[Path],
    check_ext: bool,
    timeout: int,
) -> Report:
    rep = Report()
    rep.timestamp = time.strftime("%Y%m%d%H%M%S")
    rep.root = str(root)
    rep.external_checked = check_ext
    headings_cache: dict = {}
    ext_cache = load_cache(root)
    files = 0
    totals = {"total": 0, "internal": 0, "external": 0, "anchors": 0}
    for p in paths:
        if not p.is_file() or p.suffix.lower() not in {".md", ".markdown"}:
            continue
        files += 1
        findings, stats = scan_file(p, root, headings_cache, check_ext, timeout, ext_cache)
        rep.findings.extend(findings)
        for k, v in stats.items():
            totals[k] = totals.get(k, 0) + v
    rep.stats = {"files_scanned": files, **totals}
    save_cache(root, ext_cache)
    return rep


def iter_markdown(root: Path, paths: list[Path] | None = None) -> Iterable[Path]:
    """Yield markdown files under root (or under given paths).

    Skips hidden directories BELOW root (e.g. `.git`, `.cache`, `node_modules`)
    but allows the root itself to be a hidden directory (e.g. `.agents`).
    Each yielded path is confirmed to live inside root.
    """
    if paths:
        bases: list[Path] = paths
    else:
        bases = [root]
    seen: set[Path] = set()
    for base in bases:
        if not base.exists():
            continue
        if base.is_file():
            candidates: list[Path] = [base]
        else:
            candidates = sorted(base.rglob("*.md"))
        # Strip root prefix from each path so we only consider sub-directories
        # below root when checking for hidden dirs.
        try:
            base_rel = base.relative_to(root)
        except ValueError:
            base_rel = Path(".")
        for p in candidates:
            if p.suffix.lower() not in {".md", ".markdown"}:
                continue
            try:
                rel = p.relative_to(root)
            except ValueError:
                continue
            # Sub-directories below root: skip if any starts with '.'.
            if any(part.startswith(".") for part in rel.parts[:-1]):
                continue
            # Also skip if the user passed a base that itself is hidden and
            # is not the root — only the root is allowed to be a dotdir.
            if base_rel != Path(".") and any(part.startswith(".") for part in base_rel.parts):
                continue
            if p in seen:
                continue
            seen.add(p)
            yield p


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(
        description="Validate markdown internal and external links."
    )
    p.add_argument("--output", help="Write JSON report to this file instead of stdout")
    p.add_argument("--strict", action="store_true", help="Exit 1 if any error/warning finding")
    p.add_argument("--dry-run", action="store_true", help="Read-only scan; do not write cache")
    p.add_argument(
        "--path",
        nargs="*",
        dest="path",
        help="Files or directories to scan (relative to cwd, or absolute); default: all *.md under cwd",
    )
    p.add_argument(
        "--check-external",
        action="store_true",
        default=True,
        help="Validate external http(s) links (default: true; toggle off via --skip-external)",
    )
    p.add_argument(
        "--skip-external",
        action="store_true",
        help="Skip external link validation (overrides --check-external)",
    )
    p.add_argument("--timeout", type=int, default=10, help="HTTP timeout in seconds (default 10)")
    args = p.parse_args(argv)

    cwd = Path(os.getcwd()).resolve()
    check_ext = args.check_external and not args.skip_external
    if args.path:
        user_paths = [Path(p).expanduser().resolve() for p in args.path]
        targets = list(iter_markdown(cwd, user_paths))
    else:
        targets = list(iter_markdown(cwd))

    # The "root" used for cross-repo detection is the smallest common
    # ancestor of the targets — usually cwd, but if the user passes a single
    # subtree (e.g. `--path docs/`), we use that subtree so internal links
    # outside it are still flagged as cross-repo.
    if len(targets) == 1 and targets[0].is_dir():
        root = targets[0]
    else:
        root = cwd

    rep = build_report(root, targets, check_ext, args.timeout)
    if args.dry_run:
        # Inspect-before-apply guard: do not write cache when dry-run.
        if args.output:
            print(f"[dry-run] would write report (findings={len(rep.findings)}) to {args.output}")
        else:
            print(f"[dry-run] read-only scan; findings={len(rep.findings)}; nothing written")
        return 0 if not args.strict or not rep.findings else 1

    payload = rep.as_dict()
    text = json.dumps(payload, indent=2, ensure_ascii=False)
    if args.output:
        out = Path(args.output).expanduser()
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(text + "\n")
    else:
        print(text)

    if args.strict:
        bad = [f for f in rep.findings if f.severity in ("error", "warning")]
        return 1 if bad else 0
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
