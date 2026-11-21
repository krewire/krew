#!/usr/bin/env python3
"""Populate synchronized Related Agents tables across every triad."""
from __future__ import annotations

import argparse
from pathlib import Path

from tool_config import load_config


CONFIG = load_config()["triad"]
RELATED: dict[str, list[str]] = CONFIG["related"]
ROLE: dict[str, str] = CONFIG["roles"]
PAIR: dict[str, dict[str, str]] = CONFIG["pairs"]
RELATED_CONFIG = CONFIG["related_agents"]
SECTION = f"## {RELATED_CONFIG['section']}"
TABLE_HEADER = (
    "| Agent | Relationship | Use / Handoff Trigger |\n"
    "|-------|--------------|----------------------|"
)
BUNDLE_MARKER = "- **Bundle contents:**"


def section(root: str) -> str:
    relationship = RELATED_CONFIG["relationship"]
    trigger = RELATED_CONFIG["trigger_prefix"]
    rows = [
        f"| `{name}` | {relationship} {ROLE[name]} | {trigger} {ROLE[name]}. |"
        for name in RELATED[root]
    ]
    return "\n".join((SECTION, "", "This triad collaborates with the following registered agents:", "", TABLE_HEADER, *rows, ""))


def replace_section(content: str, replacement: str) -> str:
    marker = SECTION + "\n"
    start = content.find(marker)
    if start < 0:
        handoff = content.find("## Handoff")
        if handoff < 0:
            return content.rstrip() + "\n\n" + replacement.rstrip() + "\n"
        return content[:handoff].rstrip() + "\n\n" + replacement.rstrip() + "\n" + content[handoff:]
    end = content.find("\n## ", start + len(marker))
    if end < 0:
        end = len(content)
    old_section = content[start:] if end >= len(content) else content[start:end]
    preserved = [
        line for line in old_section.splitlines() if line.startswith(BUNDLE_MARKER)
    ]
    base = replacement.rstrip()
    if preserved:
        # Keep skill bundle manifest outside the canonical table so both tools stay idempotent.
        base = base + "\n\n" + "\n".join(preserved)
    return content[:start].rstrip() + "\n\n" + base + "\n" + content[end + 1:]


def triad_files(root: Path, name: str) -> tuple[Path, Path, Path]:
    pair = PAIR[name]
    dirs = CONFIG["directories"]
    return (
        root / dirs["agents"] / f"{name}.md",
        root / dirs["commands"] / f"{pair['command']}.md",
        root / dirs["skills"] / pair["skill"] / dirs["skill_file"],
    )


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", default=".")
    parser.add_argument("--check", action="store_true", help="Report drift without writing files")
    args = parser.parse_args()
    root = Path(args.root).resolve()
    changed = 0
    missing = 0
    for name in sorted(RELATED):
        for path in triad_files(root, name):
            if not path.is_file():
                print(f"missing: {path}")
                missing += 1
                continue
            old = path.read_text(encoding="utf-8")
            new = replace_section(old, section(name))
            if old != new:
                changed += 1
                if not args.check:
                    path.write_text(new, encoding="utf-8")
    print(f"related_agents: {'would change' if args.check else 'changed'} {changed} files; {missing} missing")
    return 1 if missing or (args.check and changed > 0) else 0


if __name__ == "__main__":
    raise SystemExit(main())
