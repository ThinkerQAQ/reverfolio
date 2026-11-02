//go:build windows

package echo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMCPSetupPromptsAreVisibleAndLifecycleActionsStayHidden(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Setup-Reverfolio-MCP.ps1", "Manage-Reverfolio-MCP.ps1"} {
		if err := os.WriteFile(filepath.Join(root, "scripts", name), []byte("# fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, action := range []string{"setup", "start", "stop", "restart"} {
		cmd, err := mcpCommand("windows", root, action, "tunnel_00000000000000000000000000000000")
		if err != nil {
			t.Fatal(err)
		}
		attr := cmd.SysProcAttr
		if attr == nil {
			t.Fatalf("%s has no console policy", action)
		}
		if action == "setup" {
			if attr.CreationFlags&0x10 == 0 || attr.HideWindow {
				t.Fatal("setup must show its credential prompt")
			}
		} else if !attr.HideWindow || attr.CreationFlags&0x08000000 == 0 {
			t.Fatalf("%s must not flash a console", action)
		}
	}
}
