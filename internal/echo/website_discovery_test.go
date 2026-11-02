package echo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fixtureWebsiteTransport struct{ Responses map[string]string }

func (f fixtureWebsiteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, ok := f.Responses[req.URL.String()]
	if !ok {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: req}, nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": []string{"application/xml"}}, Request: req}, nil
}
func websiteFixture() *http.Client {
	return &http.Client{Transport: fixtureWebsiteTransport{Responses: map[string]string{
		"https://example.org/":                 `<!doctype html><html><head><link rel="alternate" type="application/rss+xml" href="/custom-feed"></head></html>`,
		"https://example.org/custom-feed":      `<?xml version="1.0"?><rss version="2.0"><channel><title>Blog</title><item><title>Go Race Conditions</title><link>https://example.org/posts/go-race?utm_source=rss</link></item><item><title>New Article</title><link>https://example.org/posts/new</link></item></channel></rss>`,
		"https://example.org/sitemap.xml":      `<?xml version="1.0"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><sitemap><loc>https://example.org/post-sitemap.xml</loc></sitemap></sitemapindex>`,
		"https://example.org/post-sitemap.xml": `<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://example.org/posts/go-race</loc></url><url><loc>https://example.org/posts/extra</loc></url><url><loc>https://example.org/about/</loc></url><url><loc>http://127.0.0.1/admin</loc></url></urlset>`,
	}}}
}

func TestWebsitePreviewImportMCPAndMigration(t *testing.T) {
	ctx := context.Background()
	store := openTest(t)
	service := Service{Store: store, Websites: WebsiteDiscoverer{Client: websiteFixture()}}
	server := httptest.NewServer(Handler(service, "website-fixture-token"))
	defer server.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "website-discovery-test", Version: "test"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: authTransport{"website-fixture-token"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	var version int
	if err = store.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 8 {
		t.Fatalf("schema=%d err=%v", version, err)
	}
	preview := callMCP[WebsitePreview](t, session, "preview_website", WebsitePreviewInput{WebsiteURL: "https://example.org/"})
	if len(preview.Candidates) != 3 {
		t.Fatalf("wrong candidates: %+v", preview.Candidates)
	}
	if preview.Candidates[0].URL != "https://example.org/posts/extra" || preview.Candidates[1].Title != "Go Race Conditions" || preview.Candidates[1].Format != "rss" {
		t.Fatalf("no feed preference: %+v", preview.Candidates)
	}
	works, _ := store.Works(ctx)
	if len(works) != 0 {
		t.Fatal("preview mutated works")
	}
	sources := callMCP[[]WebsiteSource](t, session, "list_website_sources", EmptyInput{})
	if len(sources) != 0 {
		t.Fatal("preview persisted source")
	}
	input := WebsiteImportInput{RequestID: "website-import-first", WebsiteURL: "https://example.org/", SelectedURLs: []string{"https://example.org/posts/go-race", "https://example.org/posts/new"}}
	result := callMCP[WebsiteImportResult](t, session, "import_website", input)
	if result.Imported != 2 || result.Existing != 0 || result.Source.ImportedWorks != 2 {
		t.Fatal(result)
	}
	result = callMCP[WebsiteImportResult](t, session, "import_website", input)
	if result.Imported != 2 || len(result.Works) != 2 {
		t.Fatal("same request id not replayed")
	}
	sources = callMCP[[]WebsiteSource](t, session, "list_website_sources", EmptyInput{})
	if len(sources) != 1 || sources[0].ImportedWorks != 2 {
		t.Fatal(sources)
	}
	preview = callMCP[WebsitePreview](t, session, "preview_website", WebsitePreviewInput{WebsiteURL: "https://example.org"})
	if preview.Candidates[1].ExistingWorkID < 1 {
		t.Fatal("existing work missing")
	}
	catalog, _ := store.Catalog(ctx)
	if len(catalog) != 2 || len(catalog[0].Origins) != 1 || catalog[0].Origins[0].Producer != "Website Discovery" {
		t.Fatalf("wrong provenance %+v", catalog)
	}
	if catalog[0].Work.Kind != "article" {
		t.Fatal(catalog[0].Work)
	}
	channels, _ := store.Channels(ctx, catalog[0].Work.ID)
	if len(channels) != 1 || channels[0].Language != "und" {
		t.Fatalf("missing generic channel %+v", channels)
	}
	// HTTP and MCP must observe the same authority.
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/api/website-sources", nil)
	request.Header.Set("Authorization", "Bearer website-fixture-token")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal(response.StatusCode)
	}
	result, err = service.ImportWebsite(ctx, WebsiteImportInput{RequestID: "website-add-third", WebsiteURL: "https://example.org/", SelectedURLs: []string{"https://example.org/posts/extra"}})
	if err != nil || result.Imported != 1 || result.Source.ImportedWorks != 3 {
		t.Fatalf("third import %+v %v", result, err)
	}
	_, err = service.ImportWebsite(ctx, WebsiteImportInput{RequestID: "website-impostor", WebsiteURL: "https://example.org/", SelectedURLs: []string{"https://example.org/not-in-feed"}})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("unobserved URL accepted %v", err)
	}
	works, _ = store.Works(ctx)
	if len(works) != 3 {
		t.Fatalf("unexpected works=%d", len(works))
	}
}
func TestWebsiteAtomAndHostGuards(t *testing.T) {
	parsed, items, _ := webDocumentCandidates([]byte(`<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Test Atom</title><link rel="alternate" href="https://example.org/atom-entry"/></entry></feed>`))
	if parsed != "atom" || len(items) != 1 || items[0].Title != "Test Atom" {
		t.Fatal(parsed, items)
	}
	for _, u := range []string{"http://127.0.0.1/", "http://localhost/", "http://169.254.169.254/latest", "http://example.org:8080/", "ftp://example.org/", "https://user:pw@example.org/", "http://internal.test/"} {
		if _, err := normalizeWebsiteURL(u); err == nil {
			t.Fatalf("unsafe website accepted %s", u)
		}
	}
	for _, v := range []string{"127.0.0.1", "10.2.3.4", "192.168.1.1", "169.254.169.254", "::1", "fc00::1", "::ffff:127.0.0.1", "::ffff:192.168.1.1", "100.64.1.1", "198.18.0.1", "2001:db8::1"} {
		if safeWebsiteIP(net.ParseIP(v)) {
			t.Fatalf("unsafe address accepted %s", v)
		}
	}
	if !safeWebsiteIP(net.ParseIP("8.8.8.8")) {
		t.Fatal("public IP rejected")
	}
	for _, s := range []string{"/tag/go/", "/about/", "/articles/", "/blog/", "/category/go/", "/assets/logo.png"} {
		if _, ok := candidateFromSitemap("https://example.org" + s); ok {
			t.Fatal("site navigation counted as article", s)
		}
	}
}
func TestWebsiteSameOriginAndResponseLimit(t *testing.T) {
	source := WebsiteDiscoverer{Client: &http.Client{Transport: fixtureWebsiteTransport{Responses: map[string]string{
		"https://example.org/": fmt.Sprintf(`<rss><channel><item><title>Entry</title><link>https://evil.example/steal</link></item><item><title>Allowed</title><link>https://example.org/good</link></item></channel></rss>`),
	}}}}
	p, err := source.Discover(context.Background(), "https://example.org/")
	if err != nil || len(p.Candidates) != 1 || p.Candidates[0].URL != "https://example.org/good" {
		t.Fatalf("cross-origin escaped %+v %v", p, err)
	}
	huge := strings.Repeat("x", maxWebBody+1)
	source = WebsiteDiscoverer{Client: &http.Client{Transport: fixtureWebsiteTransport{Responses: map[string]string{"https://example.org/": huge}}}}
	_, _, err = source.fetch(context.Background(), "https://example.org/")
	if err == nil {
		t.Fatal("oversized response was accepted")
	}
}

func TestWebsiteSchemaFiveUpgradePreservesWorks(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "old-state.sqlite")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	work, err := store.RegisterWork(context.Background(), WorkInput{RequestID: "legacy-work", Kind: "article", Title: "Preserved original", CanonicalURL: "https://example.org/existing"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.db.Exec("DROP TABLE impact_event_history; DROP TABLE impact_event_steps; DROP TABLE impact_event_mentions; DROP TABLE impact_event_works; DROP TABLE impact_events; DROP TABLE platform_points; DROP TABLE platform_runs; DROP TABLE platform_publications; DROP TABLE platform_accounts; DROP TABLE website_sources; PRAGMA user_version=5;"); err != nil {
		t.Fatal(err)
	}
	store.Close()
	upgraded, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer upgraded.Close()
	found, err := upgraded.Work(context.Background(), work.ID)
	if err != nil || found.Title != work.Title {
		t.Fatalf("migration corrupted prior work: %+v %v", found, err)
	}
	var v int
	if err = upgraded.db.QueryRow("PRAGMA user_version").Scan(&v); err != nil || v != 8 {
		t.Fatalf("schema version %v %v", v, err)
	}
	sources, err := upgraded.WebsiteSources(context.Background())
	if err != nil || len(sources) != 0 {
		t.Fatal("source migration failed", err)
	}
}
