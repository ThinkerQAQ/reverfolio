package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrustedExtensionLocalUIAccess(t *testing.T) {
	server := httptest.NewServer(Handler(Service{Store: openTest(t), ExtensionToken: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}, "long-test-secret-no-browser-exposure"))
	defer server.Close()
	client := server.Client()
	call := func(method, path, origin, mode, header string) (int, http.Header) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		if origin == TrustedReverfolioExtensionOrigin {
			req.Header.Set("X-Reverfolio-Extension-Access", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
		}
		if mode != "" {
			req.Header.Set("Sec-Fetch-Mode", mode)
		}
		if header != "" {
			req.Header.Set("X-Reverfolio-UI", header)
		}
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		return res.StatusCode, res.Header
	}
	if code, h := call("OPTIONS", "/api/settings", TrustedReverfolioExtensionOrigin, "cors", ""); code != 204 || h.Get("Access-Control-Allow-Origin") != TrustedReverfolioExtensionOrigin {
		t.Fatalf("extension preflight %d, origin=%q", code, h.Get("Access-Control-Allow-Origin"))
	}
	if code, h := call("GET", "/api/works", TrustedReverfolioExtensionOrigin, "cors", ""); code != 200 || h.Get("Access-Control-Allow-Origin") != TrustedReverfolioExtensionOrigin {
		t.Fatalf("extension dashboard API %d", code)
	}
	for _, method := range []string{"GET", "OPTIONS", "POST"} {
		code, h := call(method, "/api/works", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "cors", "1")
		if code == 200 || code == 204 || h.Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("unknown extension %s authorized: %d", method, code)
		}
	}
	if code, _ := call("GET", "/api/works", TrustedReverfolioExtensionOrigin, "no-cors", ""); code != 401 {
		t.Fatalf("invalid fetch mode authorized %d", code)
	}
	// Chromium host-permission fetches can omit Origin; require the separate
	// local capability AND site=none to authenticate that browser transport.
	r, err := http.NewRequest("GET", server.URL+"/api/works", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Sec-Fetch-Mode", "cors")
	r.Header.Set("Sec-Fetch-Site", "none")
	r.Header.Set("X-Reverfolio-Extension-Access", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	res, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("Origin-less trusted extension GET %d", res.StatusCode)
	}
	r2, err := http.NewRequest("GET", server.URL+"/api/works", nil)
	if err != nil {
		t.Fatal(err)
	}
	r2.Header.Set("Sec-Fetch-Mode", "cors")
	r2.Header.Set("Sec-Fetch-Site", "none")
	res2, err := client.Do(r2)
	if err != nil {
		t.Fatal(err)
	}
	res2.Body.Close()
	if res2.StatusCode != 401 {
		t.Fatalf("Origin-less request without capability authorized %d", res2.StatusCode)
	}
	if code, _ := call("GET", "/api/works", TrustedReverfolioExtensionOrigin, "", ""); code != 401 {
		t.Fatalf("nonbrowser spoof authorized %d", code)
	}
	if code, _ := call("POST", "/api/actions/upsert_work", TrustedReverfolioExtensionOrigin, "cors", ""); code != 401 {
		t.Fatalf("missing UI mutation header authorized %d", code)
	}
	if code, _ := call("POST", "/mcp", TrustedReverfolioExtensionOrigin, "cors", "1"); code != 401 {
		t.Fatalf("MCP must require bearer %d", code)
	}
	if code, _ := call("GET", "/api/works", "https://malicious.example", "cors", ""); code != 401 {
		t.Fatalf("foreign website authorized %d", code)
	}
}
