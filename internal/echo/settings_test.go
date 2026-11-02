package echo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestTOMLSettingsTreePersistsAndVersions(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := DefaultConfig()
	if err := saveConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	manager := NewConfigManager(path)
	settings, err := manager.Settings(ctx)
	if err != nil || len(settings) != len(preferenceKeys) {
		t.Fatalf("settings %+v %v", settings, err)
	}
	changes := []SettingChange{{Key: "appearance.compact", Value: "true", ExpectedVersion: 1}, {Key: "overview.evidence_limit", Value: "8", ExpectedVersion: 1}}
	got, err := manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: "batch1", Updates: changes})
	if err != nil || len(got) != 2 || got[0].Version != 2 {
		t.Fatalf("updated=%+v error=%v", got, err)
	}
	onDisk, err := LoadConfig(path)
	if err != nil || onDisk.Version != 2 || !onDisk.Interface.Compact || onDisk.Overview.EvidenceLimit != "8" {
		t.Fatalf("saved %+v %v", onDisk, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "[interface]") || !strings.Contains(string(raw), "[overview]") {
		t.Fatal("configuration not TOML")
	}
	if _, err = manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: "stale", Updates: changes}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale write: %v", err)
	}
	if _, err = manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: "secret", Updates: []SettingChange{{Key: "github.token", Value: "shhh", ExpectedVersion: 2}}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("secret key was accepted: %v", err)
	}
	if _, err = manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: "partial", Updates: []SettingChange{{Key: "overview.show_github", Value: "false", ExpectedVersion: 2}, {Key: "appearance.compact", Value: "NO", ExpectedVersion: 2}}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("partially invalid batch accepted: %v", err)
	}
	after, _ := LoadConfig(path)
	if !after.Overview.ShowGitHub || after.Version != 2 {
		t.Fatal("validation partially wrote a batch")
	}
}

func TestConfigExampleParseAndRejectUnknownKeys(t *testing.T) {
	_, err := LoadConfig("../../../config.example.toml")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[server]\nlisten = \"127.0.0.1:1\"\n[missing]\na = \"secret\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = LoadConfig(path); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown TOML keys not rejected: %v", err)
	}
	var cfg ConfigFile
	if _, err = toml.Decode("version = 1\n[overview]\nshow_github = true\n", &cfg); err != nil {
		t.Fatal(err)
	}
}
