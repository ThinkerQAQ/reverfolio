package echo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMCPRequiresSavedConfigBeforeWritingCredentials(t *testing.T) {
	t.Setenv("REVERFOLIO_CONFIG", filepath.Join(t.TempDir(), "missing.toml"))
	if err := validateMCPConfig(); err == nil {
		t.Fatal("missing config accepted")
	}
}
func TestMCPDataDirectoryFollowsConfigWithoutOpeningSQLite(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config.toml")
	if err := os.WriteFile(config, []byte("version = 1\n[storage]\ndatabase = \"private/store.sqlite\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REVERFOLIO_CONFIG", config)
	t.Setenv("REVERFOLIO_DATA_DIR", "")
	actual, err := ConfigDir()
	if err != nil || actual != filepath.Join(root, "private") {
		t.Fatal(actual, err)
	}
	if _, err := os.Stat(filepath.Join(actual, "store.sqlite")); !os.IsNotExist(err) {
		t.Fatal("management opened SQLite")
	}
}
