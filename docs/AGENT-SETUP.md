# Agent Setup

After installing Fenrir, run `fenrir setup [agent]` for your tool.

## Supported Agents

| Agent | Setup Command |
|-------|--------------|
| OpenCode | `fenrir setup opencode` |
| Claude Code | `fenrir setup claude-code` |
| Cursor | `fenrir setup cursor` |
| Windsurf | `fenrir setup windsurf` |
| Antigravity | `fenrir setup antigravity` |
| Gemini CLI | `fenrir setup gemini-cli` |
| VS Code | `fenrir setup vscode` |
| Codex | `fenrir setup codex` |
| Generic | `fenrir setup generic` |

## OpenCode

```bash
fenrir setup opencode
```

Configures: `~/.config/opencode/opencode.json`

## Claude Code

```bash
fenrir setup claude-code
```

Creates:
- `.claude/settings.json` - MCP configuration
- `FENRIR.md` - Protocol instructions

## Cursor

```bash
fenrir setup cursor
```

Creates: `.cursor/mcp.json`

## Windsurf

```bash
fenrir setup windsurf
```

Creates: `.windsurf/mcp.json`

## Antigravity

```bash
fenrir setup antigravity
```

Creates: `.antigravity/mcp.json`

## Gemini CLI

```bash
fenrir setup gemini-cli
```

Creates:
- `.gemini/settings.json` - MCP configuration
- `FENRIR.md` - Protocol instructions

## VS Code

```bash
fenrir setup vscode
```

Instructions printed for adding MCP server.

## Manual Setup

For any MCP client, add this configuration:

```json
{
  "mcpServers": {
    "fenrir": {
      "command": "fenrir",
      "args": ["mcp"]
    }
  }
}
```

## Memory Protocol

Fenrir expects you to follow this protocol:

### Start of Session
```bash
mem_session_start goal="Your goal for this session"
```

### During Session
```bash
# After any significant action
mem_save title="..." type="bugfix|decision|pattern" what="..." why="..."

# Before installing packages
pkg_check name="package-name" ecosystem="npm|pypi|cargo"

# Before architectural changes
arch_verify proposed_action="..."
```

### End of Session
```bash
mem_session_end goal="..." accomplished="..."
```

### After Context Compaction
```bash
mem_context
```

## First Use

1. Install: `brew install andragon31/tap/fenrir`
2. Setup: `fenrir setup [your-agent]`
3. Restart your AI agent
4. Start working - Fenrir will guide you
