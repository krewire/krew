#!/usr/bin/env python3
"""triad_check.py — verify agent/command/skill triads.

Scans a `.agents/` (or equivalent) layout and reports:
  1. Every agent file, and whether it has a paired command + skill (the "triad").

Usage:
  python3 triad_check.py --root .
  python3 triad_check.py --root . --output /tmp/triad.json --dry-run

Exit codes: 0 = clean, 1 = findings (with --strict), 2 = usage/IO error.

Output: JSON report to stdout (or --output file). Idempotent, read-only.

In addition to pairing assets, the check verifies that every triad artifact has
an actionable Related Agents table containing registered agent names.
--dry-run skips all file writes and only reports what would happen.
"""
from __future__ import annotations

import argparse
import json
import re
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path

from tool_config import load_config
from update_related_agents import section as canonical_related_section

# Convention: agents are `<name>.md`, commands `<name>.md`, skills `<name>/SKILL.md`.
# Triad pairing is by matching stem: agent `X` -> command `X.md` + skill `X/SKILL.md`.

CONFIG = load_config()["triad"]
DIRECTORIES = CONFIG["directories"]
AGENT_DIR = DIRECTORIES["agents"]
COMMAND_DIR = DIRECTORIES["commands"]
SKILL_DIR = DIRECTORIES["skills"]
SKILL_FILE = DIRECTORIES["skill_file"]
RELATED_CONFIG = CONFIG["related_agents"]


@dataclass
class Finding:
    rule: str
    severity: str
    message: str
    file: str = ""

    def as_dict(self) -> dict:
        return {
            "id": self.rule,
            "rule": self.rule,
            "severity": self.severity,
            "category": "triad",
            "file": self.file,
            "message": self.message,
        }


@dataclass
class Report:
    scan_name: str = "triad_check"
    scan_version: str = "1.0"
    scan_type: str = "triad"
    timestamp: str = ""
    root: str = ""
    findings: list = field(default_factory=list)
    agents: dict = field(default_factory=dict)  # name -> {agent,command,skill}

    def as_dict(self) -> dict:
        missing_agents = {
            n: v for n, v in self.agents.items() if not all(v.values())
        }
        return {
            "scan_version": self.scan_version,
            "scan_name": self.scan_name,
            "scan_type": self.scan_type,
            "root": self.root,
            "timestamp": self.timestamp,
            "summary": {
                "agents": len(self.agents),
                "agents_incomplete": len(missing_agents),
                "findings": len(self.findings),
            },
            "agents": self.agents,
            "findings": [f.as_dict() for f in self.findings],
        }


def load_names(root: Path) -> dict:
    """Return {'agents': [..], 'commands': [..], 'skills': [..]} from a views root."""
    out = {"agents": [], "commands": [], "skills": []}
    for sub in ("agents", "commands"):
        d = root / sub
        if d.is_dir():
            out[sub] = sorted(p.stem for p in d.glob("*.md"))
    s = root / SKILL_DIR
    if s.is_dir():
        out["skills"] = sorted(
            p.name for p in s.iterdir() if (p / SKILL_FILE).is_file()
        )
    return out


def decompose(name: str) -> tuple[str, str]:
    """Split an asset name into (base, kind).

    Convention: command = base verb (plan), agent = base + agent morpheme
    'er'/'or' (planner, auditor), skill = base + 'ing' (planning).

    Agent/gerund endings are plain word morphemes (no hyphen):
      - skill: strip trailing 'ing'  -> building -> build, planning -> plan
      - agent: strip trailing 'er'/'or' -> builder -> build, planner -> plan,
        auditor -> audit, automator -> automate
    English spelling changes on suffixing (root doubling, silent-e) are
    normalized so the base matches the bare command verb (plan/build/explore).
    """
    stem = name
    for suffix, kind in (("ing", "skill"), ("er", "agent"), ("or", "agent"), ("", "command")):
        if stem.endswith(suffix) and len(stem) >= len(suffix):
            base = stem if suffix == "" else stem[: -len(suffix)]
            # English spelling changes on suffixing are irregular; normalize the
            # bases this repo's real triads produce so each matches its command verb.
            base = _NORM.get(base, base)
            return base, kind
    return stem, "command"


_NORM = CONFIG["normalization"]


def build_report(root: Path) -> Report:
    rep = Report()
    rep.timestamp = time.strftime("%Y%m%d%H%M%S")
    rep.root = str(root)
    main = load_names(root)
    reps = {"agent": set(main["agents"]), "command": set(main["commands"]), "skill": set(main["skills"])}

    # Build base -> names index using the naming convention so an agent pairs
    # with the matching command/skill by base stem (planner -> plan -> planning).
    def index(base: str, kind: str, names: set) -> bool:
        wanted = {n for n in names if decompose(n)[0] == base}
        return any(decompose(n)[1] == kind for n in wanted)

    for a in sorted(reps["agent"]):
        base = decompose(a)[0]
        has_cmd = index(base, "command", reps["command"])
        has_skill = index(base, "skill", reps["skill"])
        rep.agents[a] = {
            "agent": True,
            "command": has_cmd,
            "skill": has_skill,
        }
        if not has_cmd:
            rep.findings.append(Finding(
                rule="TRIAD-COMMAND",
                severity="warning",
                message=f"Agent '{a}' (base '{base}') has no paired command",
                file=f"{COMMAND_DIR}/{a}.md",
            ))
        if not has_skill:
            rep.findings.append(Finding(
                rule="TRIAD-SKILL",
                severity="warning",
                message=f"Agent '{a}' (base '{base}') has no paired skill",
                file=f"{SKILL_DIR}/{a}/SKILL.md",
            ))

    # Commands / skills whose base matches no agent are orphans.
    agent_bases = {decompose(a)[0] for a in reps["agent"]}
    for c in sorted(reps["command"]):
        if decompose(c)[0] not in agent_bases:
            rep.findings.append(Finding(
                rule="ORPHAN-COMMAND",
                severity="info",
                message=f"Command '{c}' has base with no matching agent",
                file=f"{COMMAND_DIR}/{c}.md",
            ))
    for s in sorted(reps["skill"]):
        if decompose(s)[0] not in agent_bases:
            rep.findings.append(Finding(
                rule="ORPHAN-SKILL",
                severity="info",
                message=f"Skill '{s}' has base with no matching agent",
                file=f"{SKILL_DIR}/{s}/SKILL.md",
            ))

    # Related Agents is a synchronized triad contract, not just a prose section.
    # Validate each paired artifact independently so drift in one surface is visible.
    for agent in sorted(reps["agent"]):
        base = decompose(agent)[0]
        related = {n for n in reps["agent"] if n != agent}
        artifact_paths = (
            root / AGENT_DIR / f"{agent}.md",
            root / COMMAND_DIR / f"{base}.md",
            root / SKILL_DIR / next((n for n in reps["skill"] if decompose(n)[0] == base), base) / SKILL_FILE,
        )
        for path in artifact_paths:
            check_related_agents(path, agent, related, rep)
    return rep


def check_related_agents(path: Path, owner: str, registered: set[str], report: Report) -> None:
    """Validate the required Related Agents markdown table in one artifact."""
    if not path.is_file():
        return
    content = path.read_text(encoding="utf-8")
    marker = f"## {RELATED_CONFIG['section']}"
    start = content.find(marker)
    if start < 0:
        report.findings.append(Finding(
            rule="RELATED-AGENTS-SECTION",
            severity="error",
            message="Missing required Related Agents section",
            file=path.as_posix(),
        ))
        return
    end = content.find("\n## ", start + len(marker))
    section = content[start:] if end < 0 else content[start:end]
    # Strip skill bundle manifest lines before drift comparison; they are owned by skill_bundle_check.
    section_for_drift = "\n".join(
        line for line in section.splitlines() if not line.startswith("- **Bundle contents:**")
    )
    if not section_for_drift.strip().startswith(canonical_related_section(owner).strip()):
        report.findings.append(Finding(
            rule="RELATED-AGENT-DRIFT",
            severity="error",
            message="Related Agents section differs from tools/tools.json canonical mapping",
            file=path.as_posix(),
        ))
    rows = re.findall(r"^\|\s*`([^`]+)`\s*\|([^|]+)\|([^|]+)\|\s*$", section, re.MULTILINE)
    if len(rows) < RELATED_CONFIG["minimum_rows"]:
        report.findings.append(Finding(
            rule="RELATED-AGENTS-TABLE",
            severity="error",
            message=(
                f"Related Agents table must contain at least "
                f"{RELATED_CONFIG['minimum_rows']} data rows"
            ),
            file=path.as_posix(),
        ))
    for name, relationship, trigger in rows:
        if name not in registered:
            report.findings.append(Finding(
                rule="RELATED-AGENT-UNKNOWN",
                severity="error",
                message=f"Related agent '{name}' is not a registered agent",
                file=path.as_posix(),
            ))
        if name == owner:
            report.findings.append(Finding(
                rule="RELATED-AGENT-SELF",
                severity="error",
                message=f"Related Agents table must not self-reference '{owner}'",
                file=path.as_posix(),
            ))
        if not relationship.strip() or not trigger.strip():
            report.findings.append(Finding(
                rule="RELATED-AGENT-CONTENT",
                severity="error",
                message="Each Related Agents row needs a relationship and handoff trigger",
                file=path.as_posix(),
            ))


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description="Verify agent/command/skill triads.")
    p.add_argument("--root", required=True, help="Path to .agents (or equivalent) to scan")
    p.add_argument("--output", help="Write JSON report to this file instead of stdout")
    p.add_argument("--strict", action="store_true", help="Exit 1 if any finding of severity>=warning")
    p.add_argument("--dry-run", action="store_true",
                   help="Do not write any files; print what would happen (inspect-before-apply guard)")
    args = p.parse_args(argv)

    root = Path(args.root).expanduser()
    if not root.is_dir():
        print(f"error: --root is not a directory: {root}", file=sys.stderr)
        return 2

    rep = build_report(root)
    payload = rep.as_dict()
    text = json.dumps(payload, indent=2, ensure_ascii=False)
    if args.dry_run:
        # Inspect-before-apply guard: never write when unexpected behavior is
        # a risk. Report the outcome and what --output would have done.
        if args.output:
            print(f"[dry-run] would write {len(text) + 1} bytes to {args.output}")
        else:
            print("[dry-run] read-only scan (no --output given); nothing to write")
    elif args.output:
        out = Path(args.output).expanduser()
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(text + "\n")
    else:
        print(text)

    if args.strict:
        bad = [f for f in rep.findings if f.severity in ("warning", "error")]
        return 1 if bad else 0
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
