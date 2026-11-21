"""Load shared static configuration for repository tools."""
from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def load_config() -> dict[str, Any]:
    path = Path(__file__).with_name("tools.json")
    return json.loads(path.read_text(encoding="utf-8"))
