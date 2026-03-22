# Installation

## One-liner (Recommended)

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/TU_ORG/fenrir/main/install.sh | sh
```

### Windows

```powershell
irm https://raw.githubusercontent.com/TU_ORG/fenrir/main/install.ps1 | iex
```

## Package Managers

### Homebrew (macOS / Linux)

```bash
brew install TU_ORG/tap/fenrir
```

### Scoop (Windows)

```bash
scoop bucket add fenrir https://github.com/TU_ORG/scoop-fenrir
scoop install fenrir
```

### Binary Release

Download from [GitHub Releases](https://github.com/TU_ORG/fenrir/releases):

| OS | Arch | Download |
|----|------|----------|
| macOS | Apple Silicon | `fenrir-darwin-arm64` |
| macOS | Intel | `fenrir-darwin-amd64` |
| Linux | x86_64 | `fenrir-linux-amd64` |
| Linux | ARM64 | `fenrir-linux-arm64` |
| Windows | x86_64 | `fenrir-windows-amd64.exe` |

Add to PATH:
```bash
# macOS / Linux
sudo mv fenrir /usr/local/bin/

# Windows: Add to PATH via System Properties
```

## Build from Source

```bash
git clone https://github.com/TU_ORG/fenrir
cd fenrir
go install ./cmd/fenrir
```

## Verify Installation

```bash
fenrir version
```

## Quick Setup

After installing, setup for your agent:

```bash
fenrir setup opencode    # OpenCode
fenrir setup claude-code # Claude Code
fenrir setup cursor      # Cursor
fenrir setup windsurf    # Windsurf
fenrir setup antigravity # Antigravity
fenrir setup gemini-cli  # Gemini CLI
```

See [AGENT-SETUP.md](AGENT-SETUP.md) for detailed per-agent instructions.
