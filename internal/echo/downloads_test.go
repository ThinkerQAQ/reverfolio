package echo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDownloadClassificationConservative(t *testing.T) {
	for _, tc := range []struct{ name, kind string }{
		{"Tool-windows-amd64.zip", "package"}, {"Tool-darwin-arm64.tar.gz", "package"}, {"tool-linux-amd64", "package"}, {"Tool.dmg", "package"},
		{"SHA256SUMS", "auxiliary"}, {"Tool-windows-amd64.zip.sha256", "auxiliary"}, {"tool-source.zip.sig", "auxiliary"}, {"tool-sbom.spdx.json", "auxiliary"},
		{"ffmpeg-source.tar.xz", "source"}, {"tool-src.zip", "source"}, {"tool-source-windows.zip", "source"},
		{"website.zip", "other"}, {"project-v1.0.tar.gz", "other"}, {"tool-debug.pdb", "other"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := ClassifyAsset(tc.name); got != tc.kind {
				t.Fatalf("got %s want %s", got, tc.kind)
			}
		})
	}
}
func TestHistoricalDownloadBreakdownAndUnavailable(t *testing.T) {
	var s Snapshot
	if err := json.Unmarshal([]byte(`{"releases_available":true,"assets":[{"name":"app.msi","download_count":4},{"name":"app-source.zip","download_count":3},{"name":"SHA256SUMS","download_count":2},{"name":"ambiguous.zip","download_count":1}]}`), &s); err != nil {
		t.Fatal(err)
	}
	if d := s.Downloads; d == nil || d.Total != 10 || d.Packages != 4 || d.Source != 3 || d.Auxiliary != 2 || d.Other != 1 {
		t.Fatalf("bad totals %+v", d)
	}
	s.ReleasesOK = false
	annotateSnapshot(&s)
	if s.Downloads != nil {
		t.Fatal("failed release fetch must be unknown, not zero")
	}
}
func TestDownloadIntervalUnknownIsNotZero(t *testing.T) {
	now := time.Now().UTC()
	before := Snapshot{CapturedAt: now, ReleasesOK: true, Assets: []Asset{{ID: 1, Name: "app.exe", Downloads: 4}, {ID: 2, Name: "old.exe", Downloads: 5}}}
	after := Snapshot{CapturedAt: now.Add(time.Hour), ReleasesOK: true, Assets: []Asset{{ID: 1, Name: "app.exe", Downloads: 7}, {ID: 3, Name: "old-source.zip", Downloads: 50, CreatedAt: now.Add(-time.Hour)}}}
	interval := Compare(before, after)
	if interval.KnownDownloads == nil || interval.KnownDownloads.Packages != 3 || interval.KnownDownloads.Total != 3 || interval.UnknownAssets != 2 {
		t.Fatalf("bad interval %+v", interval)
	}
	after.ReleasesOK = false
	if d := Compare(before, after); d.KnownDownloads != nil || d.State != "unavailable" {
		t.Fatalf("bad unavailable %+v", d)
	}
}
func TestDownloadBreakdownInReportAndReplayedSnapshot(t *testing.T) {
	store := openTest(t)
	w := mustWork(t, store, "sample/downloads")
	ctx := context.Background()
	a := Snapshot{WorkID: w.ID, Key: "baseline", CapturedAt: time.Now().UTC(), ReleasesOK: true, Assets: []Asset{{ID: 1, Name: "app.exe", Downloads: 4}, {ID: 2, Name: "SHA256SUMS", Downloads: 2}}}
	if _, err := store.Save(ctx, a); err != nil {
		t.Fatal(err)
	}
	replay, err := store.SnapshotByKey(ctx, w.ID, "baseline")
	if err != nil || replay.Downloads == nil || replay.Downloads.Packages != 4 {
		t.Fatalf("replay %+v %v", replay, err)
	}
	report, err := store.GenerateReport(ctx, w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report.Markdown, "4 packages; 0 source; 2 auxiliary") {
		t.Fatal(report.Markdown)
	}
}
