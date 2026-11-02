package echo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPCommandBoundaries(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "scripts"), 0700)
	for _, n := range []string{"Setup-Reverfolio-MCP.ps1", "Manage-Reverfolio-MCP.ps1", "manage-mcp.sh"} {
		os.WriteFile(filepath.Join(root, "scripts", n), []byte(""), 0600)
	}
	for _, action := range []string{"start", "stop", "restart"} {
		cmd, err := mcpCommand("windows", root, action, "")
		if err != nil {
			t.Fatal(err)
		}
		if cmd.Args[len(cmd.Args)-1] != action {
			t.Fatal(cmd.Args)
		}
	}
	for _, action := range []string{"delete", "start; whoami", ""} {
		if _, err := mcpCommand("windows", root, action, ""); err == nil {
			t.Fatal(action)
		}
	}
	for _, id := range []string{"tunnel_x;whoami", "tunnel_", "../secret"} {
		if _, err := mcpCommand("windows", root, "setup", id); err == nil {
			t.Fatal(id)
		}
	}
	cmd, err := mcpCommand("windows", root, "setup", "tunnel_0123456789abcdef0123456789abcdef")
	if err != nil || cmd.Args[len(cmd.Args)-2] != "-TunnelId" {
		t.Fatal(cmd, err)
	}
	if tunnelReady(context.Background(), "", filepath.Join(root, "missing")) {
		t.Fatal("missing client must not be Ready")
	}
}
func TestMCPPaths(t *testing.T) {
	profile, health := mcpPaths("linux", "/home/example", "", "/tmp/config")
	if profile != filepath.Join("/tmp/config", "tunnel-client", "reverfolio.yaml") || health != filepath.Join("/tmp/config", "reverfolio", "mcp", "health.url") {
		t.Fatal(profile, health)
	}
}

func TestMCPManagementRequiresAuthorization(t *testing.T) {
	server := httptest.NewServer(Handler(Service{Store: openTest(t)}, "private-test-token"))
	defer server.Close()
	for _, method := range []string{"GET", "POST"} {
		req, _ := http.NewRequest(method, server.URL+"/api/mcp-connection", strings.NewReader("{\"action\":\"install-client\"}"))
		res, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: %d", method, res.StatusCode)
		}
	}
	req, _ := http.NewRequest("POST", server.URL+"/api/mcp-connection", strings.NewReader("{\"action\":\"setup\",\"tunnel_id\":\"tunnel_invalid\"}"))
	req.Header.Set("Authorization", "Bearer private-test-token")
	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("unconfigured install must not spawn a process: %d", res.StatusCode)
	}
}
