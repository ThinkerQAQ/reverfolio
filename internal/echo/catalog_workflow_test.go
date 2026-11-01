package echo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPCatalogImportAndHTTPState(t *testing.T) {
	store := openTest(t)
	server := httptest.NewServer(Handler(Service{Store: store}, "catalog-fixture-token"))
	defer server.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "catalog-fixture", Version: "test"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: authTransport{"catalog-fixture-token"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	m := catalogFixture()
	preview := callMCP[CatalogImportResult](t, session, "import_catalog", CatalogImportInput{DryRun: true, Manifest: m, SelectedIDs: []string{}})
	if len(preview.Rows) != 2 {
		t.Fatal(preview)
	}
	applied := callMCP[CatalogImportResult](t, session, "import_catalog", CatalogImportInput{RequestID: "import", Manifest: m, SelectedIDs: []string{"article-one"}})
	list := callMCP[[]CatalogEntry](t, session, "list_catalog", EmptyInput{})
	if len(list) != 1 || list[0].Work.ID != applied.Rows[0].WorkID || list[0].Enabled {
		t.Fatal(list)
	}
	changed := callMCP[CatalogEntry](t, session, "set_monitoring", MonitoringInput{RequestID: "enable", WorkID: list[0].Work.ID, Enabled: true, ExpectedVersion: list[0].Version})
	if !changed.Enabled {
		t.Fatal(changed)
	}
	export := callMCP[CatalogManifest](t, session, "export_catalog", map[string]string{"source_id": "fixture-copy"})
	if len(export.Works) != 1 || !export.Works[0].Enabled || len(export.Works[0].Channels) != 2 {
		t.Fatal(export)
	}
	response, err := (&http.Client{Transport: authTransport{"catalog-fixture-token"}}).Get(server.URL + "/api/catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal("HTTP catalog unavailable")
	}
}
