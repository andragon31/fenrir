package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andragon31/fenrir/internal/graph"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

func getDefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.fenrir"
}

func getGraph() (*graph.Graph, func()) {
	dataDir := viper.GetString("data_dir")
	if dataDir == "" {
		dataDir = getDefaultDataDir()
	}

	g, err := graph.New(dataDir)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}

	return g, func() { g.Close() }
}

func resolveBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "fenrir"
	}
	if res, err := filepath.EvalSymlinks(exe); err == nil {
		return res
	}
	return exe
}

func patchFenrirBINLine(src []byte, absBin string) []byte {
	const marker = `const FENRIR_BIN = process.env.FENRIR_BIN ?? "fenrir"`
	replacement := fmt.Sprintf(`const FENRIR_BIN = process.env.FENRIR_BIN ?? Bun.which("fenrir") ?? %q`, absBin)
	return []byte(strings.Replace(string(src), marker, replacement, 1))
}

func openCodeConfigDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, "Library", "Application Support")
	}
	return filepath.Join(appData, "OpenCode")
}

func injectOpenCodeMCP() error {
	dir := openCodeConfigDir()
	configPath := filepath.Join(dir, "opencode.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Remove comments if any (manual simple strip)
	cleanData := stripJSONC(data)

	var config map[string]interface{}
	if err := json.Unmarshal(cleanData, &config); err != nil {
		return err
	}

	mcpBlock, _ := config["mcp"].(map[string]interface{})
	if mcpBlock == nil {
		mcpBlock = make(map[string]interface{})
	}

	mcpBlock["fenrir"] = map[string]interface{}{
		"type":    "local",
		"command": []string{resolveBinaryPath(), "mcp"},
		"enabled": true,
	}

	config["mcp"] = mcpBlock
	output, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(configPath, output, 0644)
}

func stripJSONC(data []byte) []byte {
	var out []byte
	i := 0
	for i < len(data) {
		if data[i] == '"' {
			out = append(out, data[i])
			i++
			for i < len(data) && data[i] != '"' {
				if data[i] == '\\' && i+1 < len(data) {
					out = append(out, data[i], data[i+1])
					i += 2
					continue
				}
				out = append(out, data[i])
				i++
			}
			if i < len(data) {
				out = append(out, data[i])
				i++
			}
			continue
		}
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '/' {
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(data) && data[i] == '/' && data[i+1] == '*' {
			i += 2
			for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			if i+1 < len(data) {
				i += 2
			} else {
				i = len(data)
			}
			continue
		}
		out = append(out, data[i])
		i++
	}
	return out
}

func driftBar(score float64) string {
	width := 20
	filled := int(score * float64(width))
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func writeProjectConfig(projectName string) {
	config := map[string]interface{}{
		"project": projectName,
		"modules": []string{"."},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(".fenrir/config.json", data, 0644)
}
