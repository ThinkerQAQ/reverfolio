package echo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func catalogFixture() CatalogManifest {
	return CatalogManifest{SchemaVersion: 1, SourceID: "writer-one", Producer: "Example publisher", Works: []CatalogWork{
		{ExternalID: "article-one", Kind: "article", Title: "One", CanonicalURL: "https://example.org/one/", Channels: []CatalogChannel{{Platform: "blog", Language: "en", URL: "https://example.org/one/"}, {Platform: "devto", Language: "en", URL: "https://dev.to/example/one"}}},
		{ExternalID: "project-one", Kind: "project", Title: "Project", Repository: "example/project", Enabled: true},
	}}
}
func TestCatalogPreviewSelectionReplayAndRoundtrip(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	m := catalogFixture()
	preview, err := s.ImportCatalog(ctx, CatalogImportInput{DryRun: true, Manifest: m})
	if err != nil || len(preview.Rows) != 2 || preview.Rows[0].WorkID != 0 {
		t.Fatal(preview, err)
	}
	works, _ := s.Works(ctx)
	if len(works) != 0 {
		t.Fatal("preview persisted works")
	}
	in := CatalogImportInput{RequestID: "import", Manifest: m, SelectedIDs: []string{"article-one"}}
	first, err := s.ImportCatalog(ctx, in)
	if err != nil || len(first.Rows) != 1 {
		t.Fatal(first, err)
	}
	again, err := s.ImportCatalog(ctx, in)
	if err != nil || again.Rows[0].WorkID != first.Rows[0].WorkID {
		t.Fatal("replay", err)
	}
	catalog, err := s.Catalog(ctx)
	if err != nil || len(catalog) != 1 || catalog[0].Enabled || len(catalog[0].Origins) != 1 {
		t.Fatal(catalog, err)
	}
	channels, _ := s.Channels(ctx, catalog[0].Work.ID)
	if len(channels) != 2 {
		t.Fatal(channels)
	}
	exported, err := s.ExportCatalog(ctx, "portable-copy")
	if err != nil {
		t.Fatal(err)
	}
	target := openTest(t)
	_, err = target.ImportCatalog(ctx, CatalogImportInput{RequestID: "copy", Manifest: exported, SelectedIDs: []string{exported.Works[0].ExternalID}})
	if err != nil {
		t.Fatal(err)
	}
	copied, _ := target.Catalog(ctx)
	if len(copied) != 1 || copied[0].Enabled {
		t.Fatal(copied)
	}
}
func TestCatalogConflictIsAtomicAndDoesNotRetargetExternalID(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	m := catalogFixture()
	m.Works[1] = CatalogWork{ExternalID: "two", Kind: "article", Title: "Two", CanonicalURL: "https://example.org/two/", Channels: []CatalogChannel{{Platform: "devto", Language: "en", URL: "https://dev.to/example/one"}}}
	_, err := s.ImportCatalog(ctx, CatalogImportInput{RequestID: "conflict", Manifest: m, SelectedIDs: []string{"article-one", "two"}})
	if !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	works, _ := s.Works(ctx)
	if len(works) != 0 {
		t.Fatal("partially committed")
	}
	m = catalogFixture()
	_, err = s.ImportCatalog(ctx, CatalogImportInput{RequestID: "first", Manifest: m, SelectedIDs: []string{"article-one"}})
	if err != nil {
		t.Fatal(err)
	}
	m.Works[0].CanonicalURL = "https://example.org/moved/"
	_, err = s.ImportCatalog(ctx, CatalogImportInput{RequestID: "retarget", Manifest: m, SelectedIDs: []string{"article-one"}})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("source retargeted", err)
	}
}
func TestCatalogDoesNotOverwriteExistingTitleOrMonitoring(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	m := catalogFixture()
	w, err := s.RegisterWork(ctx, WorkInput{RequestID: "manual", Kind: "article", Title: "Local title", CanonicalURL: m.Works[0].CanonicalURL})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SetMonitoring(ctx, MonitoringInput{RequestID: "pause", WorkID: w.ID, Enabled: false, ExpectedVersion: 0})
	if err != nil {
		t.Fatal(err)
	}
	m.Works[0].Enabled = true
	_, err = s.ImportCatalog(ctx, CatalogImportInput{RequestID: "sync", Manifest: m, SelectedIDs: []string{"article-one"}})
	if err != nil {
		t.Fatal(err)
	}
	catalog, _ := s.Catalog(ctx)
	if catalog[0].Work.Title != "Local title" || catalog[0].Enabled || catalog[0].Version != 1 {
		t.Fatal(catalog)
	}
	_, err = s.SetMonitoring(ctx, MonitoringInput{RequestID: "stale", WorkID: w.ID, Enabled: true, ExpectedVersion: 0})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("stale update accepted", err)
	}
}
func TestCatalogSelectionAndPrivateURLValidation(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	m := catalogFixture()
	for _, selected := range [][]string{nil, {"absent"}, {"article-one", "article-one"}} {
		_, err := s.ImportCatalog(ctx, CatalogImportInput{RequestID: "invalid", Manifest: m, SelectedIDs: selected})
		if err == nil {
			t.Fatal("invalid selection accepted")
		}
	}
	m.Works[0].Channels[0].URL = "https://example.org/?api_key=private"
	_, err := s.ImportCatalog(ctx, CatalogImportInput{DryRun: true, Manifest: m})
	if err == nil {
		t.Fatal("credential URL accepted")
	}
}
func TestPausedWorkSkipsAutomaticCollection(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	w := mustWork(t, s, "example/project")
	_, err := s.SetMonitoring(ctx, MonitoringInput{RequestID: "pause", WorkID: w.ID, Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	service := Service{Store: s, GitHub: GitHub{BaseURL: "http://127.0.0.1:1"}}
	service.runScheduledCollection(ctx, time.Hour, false, time.Now(), func(err error) { t.Error(err) })
	snapshots, err := s.Snapshots(ctx, w.ID)
	if err != nil || len(snapshots) != 0 {
		t.Fatal(snapshots, err)
	}
	catalog, _ := s.Catalog(ctx)
	if !strings.Contains(catalog[0].Work.Repository, "example/") {
		t.Fatal(catalog)
	}
}
