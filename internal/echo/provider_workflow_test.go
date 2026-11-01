package echo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPSiteProviderSharesHTTPObservations(t *testing.T) {
	store := openTest(t)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"pageviews":12,"visitors":3,"visits":7}`))
	}))
	defer provider.Close()
	service := Service{Store: store, Config: NewConfigManager(filepath.Join(t.TempDir(), "config.toml")), Providers: ProviderClient{BaseURL: provider.URL}}
	server := httptest.NewServer(Handler(service, "synthetic-mcp-token"))
	defer server.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "provider-fixture", Version: "test"}, nil)
	httpClient := &http.Client{Transport: authTransport{"synthetic-mcp-token"}}
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: httpClient}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	work := callMCP[Work](t, session, "upsert_work", WorkInput{RequestID: "site", Kind: "site", Title: "Fixture", CanonicalURL: "https://example.org/"})
	source := fixtureSource("umami")
	source.WorkID = work.ID
	status := callMCP[SiteSourcesStatus](t, session, "save_site_source", SaveSiteSourceInput{RequestID: "save", ExpectedVersion: 1, Source: source, Credential: "fixture-secret"})
	if len(status.Sources) != 1 || !status.Sources[0].CredentialLoaded {
		t.Fatal("source not saved")
	}
	run := callMCP[ProviderRun](t, session, "collect_site_source", CollectSiteSourceInput{RequestID: "collect", SourceID: source.ID})
	if run.State != "complete" || len(run.Metrics) != 3 {
		t.Fatal("provider collection failed")
	}
	response, err := httpClient.Get(server.URL + "/api/site-sources")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var shared SiteSourcesStatus
	if err = json.NewDecoder(response.Body).Decode(&shared); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || len(shared.Sources) != 1 || shared.Sources[0].LastRun.ID != run.ID {
		t.Fatal("HTTP and MCP observations differ")
	}
	report := callMCP[ReportOutput](t, session, "generate_report", WorkIDInput{WorkID: work.ID})
	if len(report.Report.ProviderRuns) != 1 {
		t.Fatal("provider report missing")
	}
}
