package echo

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestDashboardRetainsFailureAndLatestAvailableMetrics(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/project")
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	stars := int64(3)
	_, err := s.Save(ctx, Snapshot{WorkID: w.ID, Key: "good", CapturedAt: at, State: "complete", Stars: &stars, ReleasesOK: true, Downloads: &DownloadBreakdown{Total: 8}, Assets: []Asset{{Name: "app-windows.zip", ReleaseTag: "v1", Category: "package", Downloads: 8}}, Traffic: map[string]Traffic{"views": {Count: 10, Uniques: 4, Days: []Daily{{Day: "2026-09-30T00:00:00Z", Count: 10, Uniques: 4}}}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Save(ctx, Snapshot{WorkID: w.ID, Key: "failed", CapturedAt: at.Add(time.Hour), State: "failed", Problems: []string{"traffic_http_403"}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	p := d.Projects[0]
	if p.AttemptState != "failed" || p.Traffic["views"].Count != 10 || p.Downloads.Total != 8 || !p.DownloadAt.Equal(at) || len(p.Days["views"]) != 1 || p.Releases[0].Label != "v1" || p.Platforms[0].Label != "Windows" {
		t.Fatalf("lost scope or fallback: %+v", p)
	}
	for _, run := range []ProviderRun{{SourceID: "fixture", State: "complete", CapturedAt: at}, {SourceID: "fixture", State: "failed", CapturedAt: at.Add(time.Hour)}} {
		raw, _ := json.Marshal(run)
		_, err = s.db.ExecContext(ctx, "INSERT INTO provider_runs(source_id,request_id,config_hash,captured_at,state,payload) VALUES(?,?,?,?,?,?)", run.SourceID, run.State, "fixture", stamp(run.CapturedAt), run.State, string(raw))
		if err != nil {
			t.Fatal(err)
		}
	}
	d, err = s.Dashboard(ctx)
	if err != nil || len(d.Providers) != 2 || d.Providers[0].State != "failed" || d.Providers[1].State != "complete" {
		t.Fatalf("provider fallback: %+v %v", d.Providers, err)
	}
}
