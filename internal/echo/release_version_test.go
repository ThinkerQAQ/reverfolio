package echo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseVersionMatchesExtension(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "extension", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Version != CurrentVersion {
		t.Fatalf("extension version %q, app and Bridge %q", manifest.Version, CurrentVersion)
	}
}
