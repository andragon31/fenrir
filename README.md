# Fenrir

**Memory & Knowledge Layer for AI Development Teams**

<p align="center">
<em>Agent-agnostic. Single binary. Zero dependencies.</em>
</p>

Fenrir gives your AI coding agent institutional memory with Progressive Disclosure for token efficiency.

```
Claude Code / OpenCode / Cursor / Windsurf / Gemini CLI / Antigravity / ...
    ↓ MCP stdio
Fenrir (single Go binary)
    ↓
SQLite + FTS5 (~/.fenrir/fenrir.db)
```

## Features

- **Progressive Disclosure** - Compact results by default, full content on demand
- **Memory Layer** - Persistent knowledge graph across sessions
- **Specs Module** - Requirements tracking in GIVEN/WHEN/THEN format
- **Auto-Inject Context** - Automatic context loading at session start
- **Incident Tracking** - Monitor and resolve project incidents

## Quick Start

### Install (One-liner)

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/andragon31/fenrir/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/andragon31/fenrir/main/install.ps1 | iex
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

## MCP Tools (24 total)

### Memory Tools (Core)
| Tool | Description |
|------|-------------|
| `mem_save` | Save structured observation to knowledge graph |
| `mem_find` | Search knowledge graph (compact by default) |
| `mem_context` | Get relevant context from previous sessions |
| `mem_timeline` | Get chronological history of a node |
| `mem_session_start` | Register session and load predictive context |
| `mem_session_end` | Close session and generate Session DNA |
| `mem_dna` | View Session DNA of past sessions |

### Memory Extensions (Layer 3)
| Tool | Description |
|------|-------------|
| `mem_get_observation` | Get full content with graph relationships |
| `mem_save_prompt` | Save original developer prompt before interpretation |
| `mem_session_checkpoint` | Create checkpoint for compaction recovery |

### Knowledge Lifecycle
| Tool | Description |
|------|-------------|
| `graph_review` | Review graph for stale/conflicting observations |
| `graph_expire` | Mark observation as expired |
| `node_authorize` | Promote observation authority level |

### Specs Module
| Tool | Description |
|------|-------------|
| `spec_save` | Save requirement in GIVEN/WHEN/THEN format |
| `spec_list` | List active specs organized by capability |
| `spec_check` | Verify if change affects active specs |
| `spec_delta` | Generate delta of specs affected by completed plan |

### Intelligence
| Tool | Description |
|------|-------------|
| `insights` | Auto-detected patterns (bugs, dependencies, patterns) |
| `trace` | Full traceability of file/decision/bug across sessions |
| `confidence_update` | Update confidence score of observation |

## CLI Reference

```bash
fenrir setup [agent]   # Setup for an AI agent
fenrir init           # Initialize in project
fenrir mcp            # Start MCP server
fenrir serve [port]   # Start HTTP API
fenrir tui            # Interactive TUI
fenrir version        # Show version
```

## Architecture

```
┌─────────────────────────────────────────────┐
│                 OpenCode                     │
│              Claude Code                     │
│                Cursor                        │
└─────────────────┬───────────────────────────┘
                  │ MCP stdio
                  ▼
┌─────────────────────────────────────────────┐
│                  Fenrir                      │
├─────────────────────────────────────────────┤
│  Memory Tools    │  Specs Module            │
│  Knowledge       │  Intelligence             │
│  Lifecycle       │  Incidents               │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│              SQLite + FTS5                    │
│           (~/.fenrir/fenrir.db)              │
└─────────────────────────────────────────────┘
```

## Documentation

- [Installation](docs/INSTALLATION.md) - All install methods
- [Agent Setup](docs/AGENT-SETUP.md) - Per-agent configuration

## License

MIT
