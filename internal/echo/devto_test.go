package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func devtoFixture(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected mutation: %s", r.Method)
			w.WriteHeader(405)
			return
		}
		if r.Header.Get("Accept") != foremAccept || r.Header.Get("api-key") != "test-private-forem-key" {
			w.WriteHeader(401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/users/me":
			fmt.Fprint(w, `{"id":7,"username":"sampledevto"}`)
		case "/api/articles/me/published":
			if r.URL.Query().Get("page") != "1" {
				fmt.Fprint(w, "[]")
				return
			}
			fmt.Fprint(w, `[
    {"id":101,"title":"Original Go article","url":"https://dev.to/sampledevto/go-post","canonical_url":"https://example.org/original","page_views_count":12,"public_reactions_count":3,"comments_count":1},
    {"id":102,"title":"Standalone DEV.to post","url":"https://dev.to/sampledevto/standalone","canonical_url":"https://dev.to/sampledevto/standalone","positive_reactions_count":0,"comments_count":0},
    {"id":103,"title":"Unselected article","url":"https://dev.to/sampledevto/unselected","page_views_count":50}
   ]`)
		case "/api/followers/users":
			if r.URL.Query().Get("page") != "1" {
				fmt.Fprint(w, "[]")
				return
			}
			fmt.Fprint(w, `[{"id":31,"user_id":81},{"id":32,"user_id":82}]`)
		case "/api/analytics/totals":
			if r.URL.Query().Get("article_id") != "102" {
				t.Error("unselected article analytics")
				w.WriteHeader(400)
				return
			}
			fmt.Fprint(w, `{"page_views":{"total":40},"reactions":{"total":0},"comments":{"total":0}}`)
		default:
			t.Errorf("unexpected API endpoint %s", r.URL.String())
			w.WriteHeader(404)
		}
	}))
}
func TestDEVToOfficialAPIE2EMCPAndHTTP(t *testing.T) {
	ctx := context.Background()
	upstream := devtoFixture(t)
	defer upstream.Close()
	store := openTest(t)
	configPath := filepath.Join(t.TempDir(), "config.toml")
	service := Service{Store: store, Config: &ConfigManager{Path: configPath}, DEVTo: DEVToClient{Client: upstream.Client(), BaseURL: upstream.URL}}
	original, err := store.RegisterWork(ctx, WorkInput{RequestID: "original", Kind: "article", Title: "Keep original title", CanonicalURL: "https://example.org/original"})
	if err != nil {
		t.Fatal(err)
	}
	h := httptest.NewServer(Handler(service, "mcp-fixture-secret"))
	defer h.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "DEV.to integration", Version: "test"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: h.URL + "/mcp", HTTPClient: &http.Client{Transport: authTransport{"mcp-fixture-secret"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	var v int
	store.db.QueryRow("PRAGMA user_version").Scan(&v)
	if v != 7 {
		t.Fatal(v)
	}
	account, err := service.DEVToConnect(ctx, ConnectDEVToInput{RequestID: "connect-1", APIKey: "test-private-forem-key"})
	if err != nil || account.ID != "devto_7" || !account.CredentialPresent {
		t.Fatalf("connect %+v %v", account, err)
	}
	accounts := callMCP[[]PlatformAccount](t, session, "list_platform_accounts", EmptyInput{})
	if len(accounts) != 1 || accounts[0].Username != "sampledevto" {
		t.Fatalf("accounts %+v", accounts)
	}
	raw, _ := json.Marshal(accounts)
	if strings.Contains(string(raw), "test-private-forem-key") {
		t.Fatal("key leaked through MCP")
	}
	secret := service.platformCredentialPath(account.ID)
	info, err := os.Stat(secret)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("credential must be a regular private runtime file: %+v %v", info, err)
	}
	// Windows os.FileMode reports the readonly attribute, not POSIX ACL
	// permissions. POSIX systems must enforce owner-only 0600.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("credential requires 0600 on POSIX: %v", info.Mode().Perm())
	}
	preview := callMCP[DEVToPreview](t, session, "preview_devto", DEVToPreviewInput{AccountID: account.ID})
	if len(preview.Articles) != 3 || preview.Articles[0].ExistingWorkID != original.ID {
		t.Fatalf("canonical association wrong %+v", preview)
	}
	works, _ := store.Works(ctx)
	if len(works) != 1 {
		t.Fatal("preview wrote works")
	}
	result := callMCP[ImportDEVToResult](t, session, "import_devto", ImportDEVToInput{RequestID: "import-1", AccountID: account.ID, SelectedArticleIDs: []int64{101, 102}})
	if result.Created != 1 || result.Linked != 1 || len(result.Publications) != 2 {
		t.Fatalf("import %+v", result)
	}
	result = callMCP[ImportDEVToResult](t, session, "import_devto", ImportDEVToInput{RequestID: "import-1", AccountID: account.ID, SelectedArticleIDs: []int64{101, 102}})
	if result.Created != 1 || len(result.Publications) != 2 {
		t.Fatal("idempotency result", result)
	}
	works, _ = store.Works(ctx)
	if len(works) != 2 {
		t.Fatal("unselected article got imported", len(works))
	}
	channels, _ := store.Channels(ctx, original.ID)
	if len(channels) != 1 || channels[0].URL != "https://dev.to/sampledevto/go-post" {
		t.Fatalf("cross-published URL missing %+v", channels)
	}
	collected := callMCP[PlatformRun](t, session, "collect_devto", CollectDEVToInput{RequestID: "collect-1", AccountID: account.ID})
	if collected.State != "complete" || collected.SavedPoints != 7 {
		t.Fatalf("collection %+v", collected)
	}
	dashboard := callMCP[PlatformDashboard](t, session, "get_platform_dashboard", EmptyInput{})
	if len(dashboard.Accounts) != 1 || len(dashboard.Publications) != 2 {
		t.Fatalf("wrong dashboard %+v", dashboard)
	}
	if dashboard.Accounts[0].Latest[0].Metric != "followers" || dashboard.Accounts[0].Latest[0].Value != 2 {
		t.Fatal("follower count wrong", dashboard.Accounts[0])
	}
	stats := map[int64]map[string]int64{}
	for _, pub := range dashboard.Publications {
		v := map[string]int64{}
		for _, p := range pub.Latest {
			v[p.Metric] = p.Value
		}
		stats[pub.Publication.RemoteArticleID] = v
	}
	if stats[101]["views"] != 12 || stats[102]["views"] != 40 || stats[102]["reactions"] != 0 {
		t.Fatalf("article metrics incorrectly mapped %+v", stats)
	}
	collected = callMCP[PlatformRun](t, session, "collect_devto", CollectDEVToInput{RequestID: "collect-1", AccountID: account.ID})
	if collected.SavedPoints != 7 {
		t.Fatal("idempotent run should return same result", collected)
	}
	var points int
	store.db.QueryRow("SELECT COUNT(*) FROM platform_points").Scan(&points)
	if points != 7 {
		t.Fatalf("duplicate snapshots on replay %d", points)
	}
	req, _ := http.NewRequest("GET", h.URL+"/api/platform-dashboard", nil)
	req.Header.Set("Authorization", "Bearer mcp-fixture-secret")
	response, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal("HTTP platform dashboard", response.StatusCode)
	}
	response, e = http.Get(h.URL + "/api/platform-dashboard")
	if e != nil {
		t.Fatal(e)
	}
	response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatal("platform dashboard lacks authentication", response.StatusCode)
	}
	bad := callMCP[[]PlatformAccount](t, session, "list_platform_accounts", EmptyInput{})
	if len(bad) != 1 {
		t.Fatal("account count changed")
	}
	_, err = service.DEVToImport(ctx, ImportDEVToInput{RequestID: "unselected-import", AccountID: account.ID, SelectedArticleIDs: []int64{999}})
	if err != ErrInvalid {
		t.Fatalf("foreign article import accepted %v", err)
	}
}

func TestDEVToRejectsUnownedAndUnreachableAndAbsentMetrics(t *testing.T) {
	for _, item := range []devtoArticle{{ID: 10, Title: "A", URL: "http://localhost/private"}, {ID: 1, Title: "Bad", URL: "https://attacker.invalid/path"}, {ID: 0, Title: "A", URL: "https://dev.to/author/ok"}} {
		if _, err := normalizeDevtoArticle(item); err == nil {
			t.Fatal("untrusted URL accepted", item)
		}
	}
	t.Run("bad token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) }))
		defer srv.Close()
		store := openTest(t)
		svc := Service{Store: store, Config: &ConfigManager{Path: filepath.Join(t.TempDir(), "config.toml")}, DEVTo: DEVToClient{Client: srv.Client(), BaseURL: srv.URL}}
		_, err := svc.DEVToConnect(context.Background(), ConnectDEVToInput{RequestID: "bad", APIKey: "wrong-private-key"})
		if err == nil {
			t.Fatal("invalid API key accepted")
		}
		rows, _ := svc.PlatformAccounts(context.Background())
		if len(rows) != 0 {
			t.Fatal("invalid account saved")
		}
	})
	t.Run("missing api fields are not zero", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/users/me":
				fmt.Fprint(w, `{"id":2,"username":"sampleuser"}`)
			case "/api/articles/me/published":
				fmt.Fprint(w, `[{"id":8,"title":"Missing","url":"https://dev.to/sampleuser/missing"}]`)
			case "/api/followers/users":
				fmt.Fprint(w, `[]`)
			case "/api/analytics/totals":
				fmt.Fprint(w, `{}`)
			default:
				w.WriteHeader(404)
			}
		}))
		defer srv.Close()
		store := openTest(t)
		svc := Service{Store: store, Config: &ConfigManager{Path: filepath.Join(t.TempDir(), "config.toml")}, DEVTo: DEVToClient{Client: srv.Client(), BaseURL: srv.URL}}
		a, err := svc.DEVToConnect(context.Background(), ConnectDEVToInput{RequestID: "connect-test", APIKey: "valid-api-key"})
		if err != nil {
			t.Fatal(err)
		}
		_, err = svc.DEVToImport(context.Background(), ImportDEVToInput{RequestID: "import-test", AccountID: a.ID, SelectedArticleIDs: []int64{8}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = svc.DEVToCollect(context.Background(), CollectDEVToInput{RequestID: "collect-test", AccountID: a.ID})
		if err != nil {
			t.Fatal(err)
		}
		d, err := svc.PlatformDashboard(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(d.Accounts) != 1 || len(d.Accounts[0].Latest) != 1 || d.Accounts[0].Latest[0].Value != 0 {
			t.Fatal("empty followers must be exact zero", d)
		}
		if len(d.Publications) != 1 || len(d.Publications[0].Latest) != 0 {
			t.Fatalf("missing article metrics became zero: %+v", d.Publications)
		}
	})
}
