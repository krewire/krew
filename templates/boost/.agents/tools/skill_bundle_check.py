#!/usr/bin/env python3
"""Validate and optionally add skill bundle manifests to SKILL.md files."""
from __future__ import annotations

import argparse
from pathlib import Path


MARKER = "- **Bundle contents:**"


def bundle_manifest(skill_dir: Path) -> str:
    parts = ["`SKILL.md`"]
    for name in ("rules", "tools", "resources", "refs"):
        directory = skill_dir / name
        if directory.is_dir():
            count = len(list(directory.iterdir()))
            parts.append(f"`{name}/` ({count} item{'s' if count != 1 else ''})")
        else:
            parts.append(f"`{name}/` (none)")
    return f"{MARKER} " + "; ".join(parts) + "."


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", default=".")
    parser.add_argument("--fix", action="store_true", help="Insert or replace the manifest")
    args = parser.parse_args()
    skills = Path(args.root).resolve() / "skills"
    drift = 0
    missing = 0
    for skill_dir in sorted(p for p in skills.iterdir() if p.is_dir()):
        path = skill_dir / "SKILL.md"
        if not path.is_file():
            continue
        content = path.read_text(encoding="utf-8")
        manifest = bundle_manifest(skill_dir)
        lines = content.splitlines()
        matches = [i for i, line in enumerate(lines) if line.startswith(MARKER)]
        expected = matches != [len(lines) - 1] or lines[-1] != manifest
        if expected:
            drift += 1
            if args.fix:
                lines = [line for line in lines if not line.startswith(MARKER)]
                path.write_text("\n".join(lines).rstrip() + "\n\n" + manifest + "\n", encoding="utf-8")
        if not matches:
            missing += 1
    print(f"skill_bundle: {'would change' if not args.fix else 'changed'} {drift} files; {missing} missing manifests")
    return 1 if drift and not args.fix else 0


if __name__ == "__main__":
    raise SystemExit(main())
