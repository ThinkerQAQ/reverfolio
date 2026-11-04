package echo

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var nativeMCPConfigPath string

// SetMCPConfigPath is called once at startup, before serving requests.
func SetMCPConfigPath(path string) error {
	if path == "" || strings.ContainsAny(path, "\r\n\x00") {
		return errors.New("MCP requires a saved config.toml")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	nativeMCPConfigPath = absolute
	return nil
}
func mcpConfigPath() string {
	if nativeMCPConfigPath != "" {
		return nativeMCPConfigPath
	}
	if path := os.Getenv("REVERFOLIO_CONFIG"); path != "" {
		absolute, _ := filepath.Abs(path)
		return absolute
	}
	exe, _ := os.Executable()
	return filepath.Clean(filepath.Join(filepath.Dir(exe), "..", "config.toml"))
}
func mcpConfigArguments() []string { return []string{"--config", mcpConfigPath()} }
func ConfigDir() (string, error) {
	if override := os.Getenv("REVERFOLIO_DATA_DIR"); override != "" {
		return filepath.Abs(override)
	}
	path := mcpConfigPath()
	config, err := NewConfigManager(path).Read()
	if err != nil {
		return "", err
	}
	database := config.Storage.Database
	if !filepath.IsAbs(database) {
		database = filepath.Join(filepath.Dir(path), database)
	}
	return filepath.Dir(database), nil
}
func mcpStdioCommand(executable, config string) string {
	quote := func(value string) string { return strconv.Quote(strings.ReplaceAll(value, "\\", "/")) }
	return quote(executable) + " --mcp-stdio --config " + quote(config)
}

func validateMCPConfig() error {
	path := mcpConfigPath()
	if strings.ContainsAny(path, "\r\n\x00") {
		return errors.New("invalid MCP configuration path")
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return errors.New("MCP requires an existing config.toml")
	}
	_, err = LoadConfig(path)
	return err
}
