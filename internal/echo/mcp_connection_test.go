package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("unconfigured install must not spawn a process: %d", res.StatusCode)
	}
}
