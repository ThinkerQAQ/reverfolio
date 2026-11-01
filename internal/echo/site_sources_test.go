package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixtureSource(provider string) SiteSource {
	c := SiteSource{ID: provider + "-fixture", Provider: provider, WorkID: 1, Timezone: "UTC", LookbackDays: 7, LagDays: 1, AuthMode: "token"}
	switch provider {
	case "umami":
		c.Endpoint = "https://api.umami.is/v1"
		c.Property = "website-id"
	case "gsc":
		c.Endpoint = "https://www.googleapis.com/webmasters/v3"
		c.Property = "sc-domain:example.org"
		c.Timezone = "America/Los_Angeles"
		c.LagDays = 3
	case "bing":
		c.Endpoint = "https://ssl.bing.com/webmaster/api.svc/json"
		c.Property = "https://example.org/"
	}
	return c
}
func TestOfficialProviderContractsAndDateScopes(t *testing.T) {
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	for _, provider := range []string{"umami", "gsc", "bing"} {
		t.Run(provider, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch provider {
				case "umami":
					if r.Header.Get("Authorization") != "Bearer private-credential" || r.URL.Path != "/websites/website-id/stats" || r.URL.Query().Get("timezone") != "UTC" {
						t.Error("Umami request contract", r.URL.Path)
					}
					fmt.Fprint(w, `{"pageviews":{"value":12},"visitors":{"value":3},"visits":{"value":7}}`)
				case "gsc":
					var b map[string]any
					json.NewDecoder(r.Body).Decode(&b)
					if r.Method != "POST" || b["startDate"] != "2026-10-22" || b["endDate"] != "2026-10-28" || b["aggregationType"] != "byProperty" || b["dataState"] != "final" || r.Header.Get("Authorization") != "Bearer private-credential" {
						t.Errorf("GSC contract %+v", b)
					}
					fmt.Fprint(w, `{"rows":[{"clicks":2,"impressions":100,"ctr":0.02,"position":8}]}`)
				case "bing":
					if r.URL.Query().Get("apikey") != "private-credential" || r.URL.Query().Get("siteUrl") != "https://example.org/" {
						t.Error("Bing request contract")
					}
					ms := time.Date(2026, 10, 28, 7, 0, 0, 0, time.UTC).UnixMilli()
					fmt.Fprintf(w, `{"d":[{"Clicks":0,"Impressions":4,"Date":"/Date(%d-0700)/"}]}`, ms)
				}
			}))
			defer server.Close()
			metrics, problem := (ProviderClient{BaseURL: server.URL}).collect(context.Background(), fixtureSource(provider), "private-credential", now)
			if problem != "" || len(metrics) == 0 {
				t.Fatalf("metrics %+v %s", metrics, problem)
			}
			for _, m := range metrics {
				if strings.Contains(m.Scope, "private-credential") || m.PeriodEnd.Before(m.PeriodStart) || m.Timezone == "" {
					t.Fatalf("bad scoped metric %+v", m)
				}
			}
			if provider == "bing" && (metrics[0].Timezone != "-07:00" || metrics[0].SourceDate == "") {
				t.Fatal("Bing source offset lost")
			}
		})
	}
}
func TestProviderMissingZeroEmptyErrorsAndNoCredentialLeaks(t *testing.T) {
	for _, tc := range []struct {
		provider, body, problem string
		count                   int
	}{{"umami", `{"pageviews":0,"visitors":0,"visits":0}`, "", 3}, {"umami", `{"pageviews":0,"visitors":0}`, "required_metric_missing", 0}, {"gsc", `{}`, "", 0}, {"gsc", `{"error":{"message":"private-credential"}}`, "provider_error_response", 0}, {"bing", `{"d":[]}`, "", 0}, {"bing", `{}`, "response_invalid", 0}} {
		t.Run(tc.provider+tc.body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, tc.body) }))
			defer server.Close()
			metrics, problem := (ProviderClient{BaseURL: server.URL}).collect(context.Background(), fixtureSource(tc.provider), "private-credential", time.Now())
			if problem != tc.problem || len(metrics) != tc.count {
				t.Fatalf("%s %+v", problem, metrics)
			}
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "private-credential", 401) }))
	defer server.Close()
	_, problem := (ProviderClient{BaseURL: server.URL}).collect(context.Background(), fixtureSource("bing"), "private-credential", time.Now())
	if problem != "provider_http_401" {
		t.Fatal("unsafe provider error", problem)
	}
}
func TestSourcePersistenceReplayAndWriteOnlyCredentials(t *testing.T) {
	ctx := context.Background()
	store := openTest(t)
	work, err := store.RegisterWork(ctx, WorkInput{RequestID: "site", Kind: "site", Title: "Fixture", CanonicalURL: "https://example.org/"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	s := Service{Store: store, Config: NewConfigManager(path)}
	source := fixtureSource("umami")
	source.WorkID = work.ID
	status, err := s.SaveSiteSource(ctx, SaveSiteSourceInput{RequestID: "save", ExpectedVersion: 1, Source: source})
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := s.CollectSiteSource(ctx, CollectSiteSourceInput{SourceID: source.ID, RequestID: "blocked"})
	if err != nil || blocked.State != "blocked" || blocked.Problem != "credential_missing" || len(blocked.Metrics) != 0 {
		t.Fatalf("blocked %+v %v", blocked, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"pageviews":12,"visitors":3,"visits":7}`)
	}))
	defer server.Close()
	s.Providers.BaseURL = server.URL
	status, err = s.SaveSiteSource(ctx, SaveSiteSourceInput{RequestID: "key", ExpectedVersion: status.Version, Source: source, Credential: "private-credential"})
	if err != nil {
		t.Fatal(err)
	}
	config, _ := os.ReadFile(path)
	raw, _ := json.Marshal(status)
	if strings.Contains(string(config), "private-credential") || strings.Contains(string(raw), "private-credential") {
		t.Fatal("credential exposed")
	}
	a, err := s.CollectSiteSource(ctx, CollectSiteSourceInput{SourceID: source.ID, RequestID: "complete"})
	if err != nil || a.State != "complete" {
		t.Fatalf("collect %+v %v", a, err)
	}
	b, err := s.CollectSiteSource(ctx, CollectSiteSourceInput{SourceID: source.ID, RequestID: "complete"})
	if err != nil || a.ID == 0 || a.ID != b.ID {
		t.Fatalf("idempotent IDs %d %d %v", a.ID, b.ID, err)
	}
	report, err := store.GenerateReport(ctx, work.ID)
	if err != nil || len(report.Report.ProviderRuns) != 1 || !strings.Contains(report.Markdown, "Site analytics observations") {
		t.Fatal("provider missing from report", err)
	}
	source.Endpoint = "https://another.example.org/api"
	if _, err = s.SaveSiteSource(ctx, SaveSiteSourceInput{RequestID: "retarget", ExpectedVersion: status.Version, Source: source}); err == nil {
		t.Fatal("retargeted an existing secret without replacement")
	}
}
func TestSourceValidationAndBingDate(t *testing.T) {
	c := fixtureSource("umami")
	c.ID = "../escape"
	if c.validate() == nil {
		t.Fatal("path traversal accepted")
	}
	c = fixtureSource("gsc")
	c.Timezone = "UTC"
	if c.validate() == nil {
		t.Fatal("GSC must preserve PT dates")
	}
	if _, ok := parseBingDate("/Date(1-9999)/"); ok {
		t.Fatal("bad date offset accepted")
	}
	for _, raw := range []string{"null", `{"value":null}`, `{"value":-1}`} {
		if _, ok := jsonCount(json.RawMessage(raw)); ok {
			t.Fatal("invalid count accepted")
		}
	}
}
