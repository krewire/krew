# Multi-Runtime Agent Binding
<!-- Class-level reference for agent-agnostic registry optimization -->
<!-- Use when adding/updating runtime context files (.hermes.md, CLAUDE.md, .openai.md, .github/copilot-instructions.md, .openclaw.md) -->

procedure:
  1. Confirm agent-agnostic: never break opencode.json / kilo.json / AGENTS.md
  2. Add per-runtime file at standard path (cwd-only per each runtime spec)
  3. Update master .hermes.md binding list (line 21-style) — preserve all previous bindings
  4. Verify co-existence: grep all runtime files for same root; no duplicate keys
  5. For OpenClaw: config is global (~/.openclaw/openclaw.json), never repo-local; workspace reads AGENTS.md + bootstrap from <workspace>; skills auto-discovered from <workspace>/skills (no skills.paths needed); optional extras via agents.defaults.contextFiles
  6. Check repo privacy (README.md / LICENSE) before exposing context

bindings_table:
  hermes: .hermes.md + ~/.hermes/config.yaml (profile-safe, HERMES_HOME)
  claude: CLAUDE.md (cwd-only, priority over AGENTS.md per hermes skill)
  openai_gpt: .openai.md (marker, no proprietary path needed)
  opencode: opencode.json (preserved)
  kilo: kilo.json (instructions + skills.paths)
  copilot: .github/copilot-instructions.md (standard Copilot path)
  openclaw: .openclaw.md + AGENTS.md (workspace-root; global config at ~/.openclaw/openclaw.json)
  mcp: settings.json (github wrapper)

pitfalls:
  - Do NOT replace runtime configs; co-exist is the rule.
  - Do NOT hardcode ~/.hermes; always profile-safe.
  - Before writing, check README.md license line (private vs public); never expose PII.
  - Verifying means grep, not assumption.
