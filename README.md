# Fenrir

**AI Governance & Memory Layer for Development Teams**

<p align="center">
<em>Agent-agnostic. Single binary. Zero dependencies.</em>
</p>

Fenrir gives your AI coding agent institutional memory. No more repeating the same mistakes.

```
Claude Code / OpenCode / Cursor / Windsurf / Gemini CLI / Antigravity / ...
    ↓ MCP stdio
Fenrir (single Go binary)
    ↓
SQLite + FTS5 (~/.fenrir/fenrir.db)
```

## Features

- **Memory Layer** - Persistent knowledge graph across sessions
- **Package Validation** - CVE checks, license analysis, typosquatting detection
- **Drift Detection** - Monitor architectural violations
- **Security Shield** - Audit logs, prompt injection protection
- **Git Sync** - Share memories with your team via git

## Quick Start

### Install (One-liner)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/TU_ORG/fenrir/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/TU_ORG/fenrir/main/install.ps1 | iex
```

### Setup Your Agent

| Agent | Command |
|-------|---------|
| OpenCode | `fenrir setup opencode` |
| Claude Code | `fenrir setup claude-code` |
| Cursor | `fenrir setup cursor` |
| Windsurf | `fenrir setup windsurf` |
| Antigravity | `fenrir setup antigravity` |
| Gemini CLI | `fenrir setup gemini-cli` |

That's it. No Node.js, no Python, no Docker.

## MCP Tools (22 total)

| Module | Tools |
|--------|-------|
| **Memory** | mem_save, mem_find, mem_context, mem_timeline, mem_session_start, mem_session_end, mem_dna |
| **Validator** | pkg_check, pkg_license, pkg_audit |
| **Enforcer** | arch_save, arch_verify, arch_drift, policy_check, predict |
| **Shield** | audit_log, session_audit, inject_guard, fenrir_stats |
| **Intelligence** | insights, trace, confidence_update |

## CLI Reference

```bash
fenrir setup [agent]   # Setup for an AI agent
fenrir init           # Initialize in project
fenrir mcp            # Start MCP server
fenrir serve [port]    # Start HTTP API
fenrir tui            # Interactive TUI

fenrir search <query> # Search memories
fenrir stats          # Show statistics
fenrir drift          # Show drift scores
fenrir insights       # Auto-detected patterns

fenrir session list   # List sessions
fenrir sync           # Git sync
fenrir version        # Show version
```

## Documentation

- [Installation](docs/INSTALLATION.md) - All install methods
- [Agent Setup](docs/AGENT-SETUP.md) - Per-agent configuration

## License

MIT
