package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGSCDetailDimensionsAndExactURLAttribution(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Dimensions  []string `json:"dimensions"`
			Aggregation string   `json:"aggregationType"`
			Limit       int      `json:"rowLimit"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.Aggregation != "auto" || body.Limit != 25000 {
			t.Error("query contract")
		}
		if len(body.Dimensions) == 1 {
			fmt.Fprint(w, `{"rows":[{"keys":["2026-10-25"],"clicks":2,"impressions":8}]}`)
		} else {
			fmt.Fprint(w, `{"rows":[{"keys":["2026-10-25","https://example.org/article/"],"clicks":1,"impressions":3},{"keys":["2026-10-25","https://example.org/other/"],"clicks":1,"impressions":5}]}`)
		}
	}))
	defer server.Close()
	c := fixtureSource("gsc")
	rows, problem := (ProviderClient{BaseURL: server.URL}).collectDetails(context.Background(), c, "fixture", now)
	if problem != "" || len(rows) != 6 {
		t.Fatalf("details: %s %+v", problem, rows)
	}
	s := openTest(t)
	ctx := context.Background()
	work, err := s.RegisterWork(ctx, WorkInput{RequestID: "a", Kind: "article", Title: "Same title", CanonicalURL: "https://example.org/article/"})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.attributeDetails(ctx, rows, work.ID); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.PageURL == "https://example.org/article/" && row.WorkID != work.ID {
			t.Fatal("exact URL not attributed")
		}
		if row.PageURL == "https://example.org/other/" && row.WorkID != 0 {
			t.Fatal("inferred ownership")
		}
	}
}
func TestUmamiSeriesAndPageWindowStaySeparate(t *testing.T) {
	c := fixtureSource("umami")
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/websites/website-id/pageviews" {
			if r.URL.Query().Get("unit") != "day" {
				t.Error("missing daily unit")
			}
			fmt.Fprint(w, `{"pageviews":[{"x":"2026-10-25 00:00:00","y":9}],"sessions":[{"x":"2026-10-25 00:00:00","y":4}]}`)
			return
		}
		if r.URL.Query().Get("type") != "path" {
			t.Error("page ranking dimension")
		}
		fmt.Fprint(w, `[{"x":"/article/","y":7}]`)
	}))
	defer server.Close()
	rows, problem := (ProviderClient{BaseURL: server.URL}).collectDetails(context.Background(), c, "fixture", now)
	if problem != "" || len(rows) != 3 {
		t.Fatalf("series: %s %+v", problem, rows)
	}
	if rows[0].PeriodEnd.Sub(rows[0].PeriodStart) != 24*time.Hour || rows[2].PeriodEnd.Sub(rows[2].PeriodStart) != 7*24*time.Hour {
		t.Fatal("window totals became daily")
	}
	s := openTest(t)
	ctx := context.Background()
	site, err := s.RegisterWork(ctx, WorkInput{RequestID: "site", Kind: "site", Title: "Site", CanonicalURL: "https://example.org/"})
	if err != nil {
		t.Fatal(err)
	}
	article, err := s.RegisterWork(ctx, WorkInput{RequestID: "article", Kind: "article", Title: "Article", CanonicalURL: "https://example.org/article/"})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.attributeDetails(ctx, rows, site.ID); err != nil {
		t.Fatal(err)
	}
	if rows[2].WorkID != article.ID || rows[2].PageURL != "https://example.org/article/" {
		t.Fatal("explicit site URL resolution failed")
	}
}
func TestOptionalDetailFailurePreservesCollectedRows(t *testing.T) {
	c := fixtureSource("gsc")
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			fmt.Fprint(w, `{"rows":[{"keys":["2026-10-25"],"clicks":0,"impressions":0}]}`)
		} else {
			w.WriteHeader(503)
		}
	}))
	defer server.Close()
	rows, problem := (ProviderClient{BaseURL: server.URL}).collectDetails(context.Background(), c, "fixture", time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC))
	if problem != "provider_http_503" || len(rows) != 2 || rows[0].Value != 0 {
		t.Fatal("partial rows or true zeros lost", problem, rows)
	}
}
