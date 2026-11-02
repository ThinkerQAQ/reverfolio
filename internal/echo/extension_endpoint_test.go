package echo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartupSettingsPersistAndSynchronizeLeastPrivilegeExtension(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "config.toml")
	cfg := DefaultConfig()
	if err := saveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	ui := filepath.Join(root, "extension", "ui")
	os.MkdirAll(ui, 0700)
	manifest := filepath.Join(root, "extension", "manifest.json")
	os.WriteFile(manifest, []byte(`{"host_permissions":["http://127.0.0.1:19452/*"],"content_security_policy":{"extension_pages":"script-src 'self'; object-src 'self'; connect-src 'self' http://127.0.0.1:19452"}}`), 0600)
	os.WriteFile(filepath.Join(ui, "index.html"), []byte(`<meta name="reverfolio-api-origin" content="http://127.0.0.1:19452">`), 0600)
	manager := NewConfigManager(cfgPath)
	_, err := manager.UpdateSettings(context.Background(), UpdateSettingsInput{RequestID: "startup", Updates: []SettingChange{{"server.listen", "127.0.0.1:24561", 1}, {"github.auth", "gh", 1}}})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := LoadConfig(cfgPath)
	if err != nil || saved.Server.Listen != "127.0.0.1:24561" || saved.GitHub.Auth != "gh" {
		t.Fatal(saved, err)
	}
	for _, path := range []string{manifest, filepath.Join(ui, "index.html")} {
		b, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(b), "127.0.0.1:24561") || strings.Contains(string(b), "19452") {
			t.Fatal(path, string(b), err)
		}
	}
	for _, listen := range []string{"0.0.0.0:24561", "example.org:24561", "127.0.0.1:0", "127.0.0.1:70000"} {
		if _, err := LocalAPIOrigin(listen); err == nil {
			t.Fatal("unsafe listener accepted", listen)
		}
	}
	_, err = manager.UpdateSettings(context.Background(), UpdateSettingsInput{RequestID: "missing-db", Updates: []SettingChange{{"storage.database", "missing.sqlite", 2}}})
	if err == nil {
		t.Fatal("missing database would silently create empty state")
	}
	saved, _ = LoadConfig(cfgPath)
	if saved.Storage.Database != cfg.Storage.Database || saved.Version != 2 {
		t.Fatal("rejected change modified config")
	}
}
