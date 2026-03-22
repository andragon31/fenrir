package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup [agent]",
	Short: "Setup Fenrir for an AI agent",
	Long: `Setup Fenrir integration for various AI coding agents.

Supported agents:
  - opencode     OpenCode
  - claude-code  Claude Code
  - cursor       Cursor
  - windsurf     Windsurf
  - antigravity  Antigravity
  - gemini-cli   Gemini CLI
  - vscode       VS Code (Copilot)
  - codex        Codex
  - generic      Generic MCP client

Examples:
  fenrir setup opencode
  fenrir setup claude-code
  fenrir setup cursor`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agent := args[0]
		installer := getInstaller(agent)
		if err := installer.Install(); err != nil {
			log.Fatal("Setup failed", "agent", agent, "error", err)
		}
		log.Info("Setup complete", "agent", agent)
		fmt.Println("\nRestart your AI agent to start using Fenrir!")
	},
}

type Installer interface {
	Install() error
	Name() string
}

func getInstaller(agent string) Installer {
	switch agent {
	case "opencode":
		return &OpenCodeInstaller{}
	case "claude-code":
		return &ClaudeCodeInstaller{}
	case "cursor":
		return &CursorInstaller{}
	case "windsurf":
		return &WindsurfInstaller{}
	case "antigravity":
		return &AntigravityInstaller{}
	case "gemini-cli":
		return &GeminiCLIInstaller{}
	case "vscode":
		return &VSCodeInstaller{}
	case "codex":
		return &CodexInstaller{}
	default:
		return &GenericInstaller{}
	}
}

type OpenCodeInstaller struct{}

func (i *OpenCodeInstaller) Name() string { return "OpenCode" }

func (i *OpenCodeInstaller) Install() error {
	dir := openCodeConfigDir() + "/plugins"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create plugin dir %s: %w", dir, err)
	}

	data, err := openCodeFS.ReadFile("plugins/opencode/fenrir.ts")
	if err != nil {
		return fmt.Errorf("read embedded fenrir.ts: %w", err)
	}

	// Patch FENRIR_BIN in the installed copy
	data = patchFenrirBINLine(data, resolveBinaryPath())

	dest := filepath.Join(dir, "fenrir.ts")
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}

	// Register fenrir MCP server in opencode.json
	if err := injectOpenCodeMCP(); err != nil {
		// Non-fatal: plugin works, MCP just needs manual config
		cmd := resolveBinaryPath()
		fmt.Printf("Warning: could not auto-register MCP in opencode.json: %v\n", err)
		fmt.Printf("  Add manually to your opencode.json under \"mcp\":\n")
		fmt.Printf("  \"fenrir\": { \"type\": \"local\", \"command\": [%q, \"mcp\"], \"enabled\": true }\n", cmd)
	}

	return nil
}

type ClaudeCodeInstaller struct{}

func (i *ClaudeCodeInstaller) Name() string { return "Claude Code" }

func (i *ClaudeCodeInstaller) Install() error {
	exe := resolveBinaryPath()
	
	dir := filepath.Join(os.Getenv("USERPROFILE"), ".claude", "mcp")
	if home, err := os.UserHomeDir(); err == nil {
		dir = filepath.Join(home, ".claude", "mcp")
	}
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create mcp dir %s: %w", dir, err)
	}

	entry := map[string]interface{}{
		"command": exe,
		"args":    []string{"mcp"},
	}
	data, _ := json.MarshalIndent(entry, "", "  ")
	
	dest := filepath.Join(dir, "fenrir.json")
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}

	// Add allowlist to ~/.claude/settings.json
	settingsPath := filepath.Join(filepath.Dir(filepath.Dir(dir)), "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var config map[string]json.RawMessage
		if json.Unmarshal(data, &config) == nil {
			var perms map[string]interface{}
			if raw, exists := config["permissions"]; exists {
				json.Unmarshal(raw, &perms)
			} else {
				perms = make(map[string]interface{})
			}
			
			allow, _ := perms["allow"].([]interface{})
			tools := []string{
				"mem_save", "mem_find", "mem_context", "mem_timeline",
				"mem_session_start", "mem_session_end", "pkg_check", "pkg_audit",
			}
			
			for _, t := range tools {
				found := false
				for _, existing := range allow {
					if existing == t { found = true; break }
				}
				if !found { allow = append(allow, t) }
			}
			perms["allow"] = allow
			permsJSON, _ := json.Marshal(perms)
			config["permissions"] = json.RawMessage(permsJSON)
			finalJSON, _ := json.MarshalIndent(config, "", "  ")
			os.WriteFile(settingsPath, finalJSON, 0644)
		}
	}

	fmt.Println("Claude Code: MCP registered & tools allowlisted")
	return nil
}

type GeminiCLIInstaller struct{}
func (i *GeminiCLIInstaller) Name() string { return "Gemini CLI" }
func (i *GeminiCLIInstaller) Install() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gemini", "settings.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	var config map[string]json.RawMessage
	data, err := os.ReadFile(configPath)
	if err == nil { json.Unmarshal(data, &config) } else { config = make(map[string]json.RawMessage) }

	var mcpServers map[string]interface{}
	if raw, exists := config["mcpServers"]; exists { json.Unmarshal(raw, &mcpServers) } else { mcpServers = make(map[string]interface{}) }

	mcpServers["fenrir"] = map[string]interface{}{ "command": resolveBinaryPath(), "args": []string{"mcp"} }
	mcpJSON, _ := json.Marshal(mcpServers)
	config["mcpServers"] = json.RawMessage(mcpJSON)
	finalJSON, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(configPath, finalJSON, 0644); err != nil {
		return err
	}

	// Write system.md
	systemPath := filepath.Join(home, ".gemini", "system.md")
	os.WriteFile(systemPath, []byte(fenrirProtocolMarkdown), 0644)
	fmt.Println("Gemini CLI: system.md updated with Fenrir protocol")

	return nil
}

type CursorInstaller struct{}
func (i *CursorInstaller) Name() string { return "Cursor" }
func (i *CursorInstaller) Install() error {
	appData := os.Getenv("APPDATA")
	if appData == "" { home, _ := os.UserHomeDir(); appData = filepath.Join(home, "Library", "Application Support") }
	dir := filepath.Join(appData, "Cursor", "User", "globalStorage", "cursor-retrieval")
	os.MkdirAll(dir, 0755)
	configPath := filepath.Join(dir, "mcpServers.json")
	var config map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil { json.Unmarshal(data, &config) } else { config = map[string]interface{}{"mcpServers": make(map[string]interface{})} }
	mcpServers, _ := config["mcpServers"].(map[string]interface{})
	if mcpServers == nil { mcpServers = make(map[string]interface{}); config["mcpServers"] = mcpServers }
	mcpServers["fenrir"] = map[string]interface{}{ "type": "command", "command": resolveBinaryPath(), "args": []string{"mcp"}, "env": make(map[string]string) }
	finalJSON, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(configPath, finalJSON, 0644)
}

type WindsurfInstaller struct{}
func (i *WindsurfInstaller) Name() string { return "Windsurf" }
func (i *WindsurfInstaller) Install() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	var config map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil { json.Unmarshal(data, &config) } else { config = map[string]interface{}{"mcpServers": make(map[string]interface{})} }
	mcpServers, _ := config["mcpServers"].(map[string]interface{})
	if mcpServers == nil { mcpServers = make(map[string]interface{}); config["mcpServers"] = mcpServers }
	mcpServers["fenrir"] = map[string]interface{}{ "command": resolveBinaryPath(), "args": []string{"mcp"} }
	finalJSON, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(configPath, finalJSON, 0644)
}

type AntigravityInstaller struct{}
func (i *AntigravityInstaller) Name() string { return "Antigravity" }
func (i *AntigravityInstaller) Install() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gemini", "antigravity", "mcp_servers.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	var mcpServers map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil { json.Unmarshal(data, &mcpServers) } else { mcpServers = make(map[string]interface{}) }
	mcpServers["fenrir"] = map[string]interface{}{ "command": resolveBinaryPath(), "args": []string{"mcp"} }
	finalJSON, _ := json.MarshalIndent(mcpServers, "", "  ")
	return os.WriteFile(configPath, finalJSON, 0644)
}

type VSCodeInstaller struct{}
func (i *VSCodeInstaller) Name() string { return "VS Code" }
func (i *VSCodeInstaller) Install() error {
	fmt.Println("VS Code: Use the 'Generic' instructions or install an MCP bridge extension.")
	return (&GenericInstaller{}).Install()
}

type CodexInstaller struct{}
func (i *CodexInstaller) Name() string { return "Codex" }
func (i *CodexInstaller) Install() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codex", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Inyectar bloque MCP
	data, _ := os.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "[mcp_servers.fenrir]") {
		block := fmt.Sprintf("\n[mcp_servers.fenrir]\ncommand = %q\nargs = [\"mcp\"]\n", resolveBinaryPath())
		content += block
	}

	// Inyectar instrucciones de memoria
	instrPath := filepath.Join(home, ".codex", "fenrir_instructions.md")
	os.WriteFile(instrPath, []byte(fenrirProtocolMarkdown), 0644)
	
	if !strings.Contains(content, "model_instructions_file") {
		content = "model_instructions_file = " + fmt.Sprintf("%q", instrPath) + "\n" + content
	}

	os.WriteFile(configPath, []byte(content), 0644)
	fmt.Println("Codex: Config updated and instructions file created")
	return nil
}

type GenericInstaller struct{}
func (i *GenericInstaller) Name() string { return "Generic MCP" }
func (i *GenericInstaller) Install() error {
	fmt.Printf("Generic MCP setup:\nCommand: %s\nArgs: mcp\n", resolveBinaryPath())
	return nil
}
