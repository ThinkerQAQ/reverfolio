package echo

import (
	"context"
	"testing"
	"time"
)

func TestReportPreservesEachValidMetricBeyondFailureHistory(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/observations")
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	zero := int64(0)
	_, err := s.Save(ctx, Snapshot{WorkID: w.ID, Key: "valid", CapturedAt: at, State: "complete", Stars: &zero, ReleasesOK: true, Assets: []Asset{{ID: 1, Name: "app.zip", Downloads: 7}}, Traffic: map[string]Traffic{"clones": {Count: 4}}})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 105; i++ {
		_, err = s.Save(ctx, Snapshot{WorkID: w.ID, Key: stamp(at.Add(time.Duration(i) * time.Hour)), CapturedAt: at.Add(time.Duration(i) * time.Hour), State: "failed"})
		if err != nil {
			t.Fatal(err)
		}
	}
	report, err := s.Report(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}
	x := report.Sources[0]
	if x.Latest.State != "failed" || x.ValidMetrics["downloads"].Value != 7 || x.ValidMetrics["clones"].Value != 4 || x.LastRelease == nil {
		t.Fatalf("%+v", x)
	}
	star, found := x.ValidMetrics["stars"]
	if !found || star.Value != 0 || !star.ObservedAt.Equal(at) {
		t.Fatal("zero lost or source time replaced", x)
	}
	if _, found := x.ValidMetrics["forks"]; found {
		t.Fatal("missing field invented")
	}
}
