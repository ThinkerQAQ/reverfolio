package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocalUIAutomaticSessionAndAPIProtection(t *testing.T) {
	service := Service{Store: openTest(t)}
	server := httptest.NewServer(Handler(service, "secret-api-token-for-tests"))
	defer server.Close()
	client := server.Client()
	get := func(path string) (*http.Response, []byte) {
		t.Helper()
		res, err := client.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		return res, body
	}
	home, markup := get("/")
	if home.StatusCode != 200 {
		t.Fatalf("home: %d", home.StatusCode)
	}
	if strings.Contains(string(markup), "Local API token") || strings.Contains(string(markup), "secret-api-token-for-tests") {
		t.Fatal("UI leaked token field or credential")
	}
	if !strings.Contains(string(markup), "app.js") {
		t.Fatal("standalone UI assets missing")
	}
	if len(home.Cookies()) != 1 || !home.Cookies()[0].HttpOnly || home.Cookies()[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("local HttpOnly session not issued")
	}
	if css, _ := get("/assets/app.css"); css.StatusCode != 200 {
		t.Fatalf("CSS %d", css.StatusCode)
	}
	if js, _ := get("/assets/app.js"); js.StatusCode != 200 {
		t.Fatalf("JS %d", js.StatusCode)
	}
	if icon, data := get("/assets/favicon.svg"); icon.StatusCode != 200 || !strings.Contains(string(data), "<svg") {
		t.Fatalf("Brand favicon missing: HTTP %d", icon.StatusCode)
	}
	anon, _ := get("/api/works")
	if anon.StatusCode != 401 {
		t.Fatalf("unauthenticated API returned %d", anon.StatusCode)
	}
	cookie := home.Cookies()[0]
	request := func(method, path, body string, headers map[string]string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(cookie)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res
	}
	if r := request("GET", "/api/works", "", nil); r.StatusCode != 200 {
		t.Fatalf("cookie GET %d", r.StatusCode)
	}
	body := `{"request_id":"cookie-works-create","title":"Demo","kind":"project","repository":"local/demo"}`
	if r := request("POST", "/api/actions/upsert_work", body, map[string]string{"Content-Type": "application/json"}); r.StatusCode != 401 {
		t.Fatalf("cookie POST without header %d", r.StatusCode)
	}
	if r := request("POST", "/api/actions/upsert_work", body, map[string]string{"Content-Type": "application/json", "X-Reverfolio-UI": "1", "Sec-Fetch-Site": "cross-site"}); r.StatusCode != 401 {
		t.Fatalf("cross-site cookie POST %d", r.StatusCode)
	}
	if r := request("POST", "/api/actions/upsert_work", body, map[string]string{"Content-Type": "application/json", "X-Reverfolio-UI": "1", "Origin": "https://hostile.example"}); r.StatusCode != 403 {
		t.Fatalf("cross-origin cookie POST %d", r.StatusCode)
	}
	if r := request("POST", "/api/actions/upsert_work", body, map[string]string{"Content-Type": "application/json", "X-Reverfolio-UI": "1", "Origin": server.URL, "Sec-Fetch-Site": "same-origin"}); r.StatusCode != 200 {
		t.Fatalf("same-origin UI mutation %d", r.StatusCode)
	}
	if r := request("POST", "/mcp", "{}", map[string]string{"X-Reverfolio-UI": "1"}); r.StatusCode != 401 {
		t.Fatalf("MCP must not accept UI cookie: %d", r.StatusCode)
	}
	requestWithBearer, err := http.NewRequest("GET", server.URL+"/api/works", nil)
	if err != nil {
		t.Fatal(err)
	}
	requestWithBearer.Header.Set("Authorization", "Bearer secret-api-token-for-tests")
	result, err := client.Do(requestWithBearer)
	if err != nil {
		t.Fatal(err)
	}
	result.Body.Close()
	if result.StatusCode != 200 {
		t.Fatalf("bearer authorization broken: %d", result.StatusCode)
	}
}
