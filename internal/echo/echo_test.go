package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func mustWork(t *testing.T, s *Store, repo string) Work {
	t.Helper()
	w, err := s.CreateWork(context.Background(), repo, repo)
	if err != nil {
		t.Fatal(err)
	}
	return w
}
func TestPersistenceIsolationIdempotencyAndCorrections(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "state.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a := mustWork(t, s, "sample/alpha")
	b := mustWork(t, s, "sample/beta")
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	snap := Snapshot{WorkID: a.ID, Key: "same", CapturedAt: now, ReleasesOK: true, Traffic: map[string]Traffic{"views": {Count: 4, Uniques: 2, Days: []Daily{{"2026-10-19", 4, 2}}}}}
	first, err := s.Save(ctx, snap)
	if err != nil {
		t.Fatal(err)
	}
	snap.Traffic["views"] = Traffic{Count: 99, Days: []Daily{{"2026-10-19", 99, 9}}}
	duplicate, err := s.Save(ctx, snap)
	if err != nil || duplicate.ID != first.ID {
		t.Fatalf("idempotency: %+v %v", duplicate, err)
	}
	snap.WorkID = b.ID
	if _, err = s.Save(ctx, snap); err != nil {
		t.Fatal(err)
	}
	snap.WorkID = a.ID
	snap.Key = "correction"
	snap.CapturedAt = now.Add(time.Hour)
	snap.Traffic["views"] = Traffic{Count: 3, Uniques: 2, Days: []Daily{{"2026-10-19", 3, 2}}}
	if _, err = s.Save(ctx, snap); err != nil {
		t.Fatal(err)
	}
	snap.Key = "older-completes-late"
	snap.CapturedAt = now.Add(-time.Hour)
	snap.Traffic["views"] = Traffic{Count: 100, Days: []Daily{{"2026-10-19", 100, 99}}}
	if _, err = s.Save(ctx, snap); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	days, err := s.Days(ctx, a.ID, "views")
	if err != nil || len(days) != 1 || days[0].Count != 3 {
		t.Fatalf("revision or persistence: %+v %v", days, err)
	}
	days, err = s.Days(ctx, b.ID, "views")
	if err != nil || len(days) != 1 || days[0].Count != 99 {
		t.Fatalf("repository isolation: %+v %v", days, err)
	}
	w := mustWork(t, s, "SAMPLE/ALPHA")
	if w.ID != a.ID {
		t.Fatal("repository identity should be case insensitive")
	}
}

func TestAssetLifecycle(t *testing.T) {
	before := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	previous := Snapshot{CapturedAt: before, ReleasesOK: true, Assets: []Asset{{ID: 1, Downloads: 10}, {ID: 2, Downloads: 8}, {ID: 3, Downloads: 2}}}
	current := Snapshot{CapturedAt: before.Add(time.Hour), ReleasesOK: true, Assets: []Asset{{ID: 1, Downloads: 13}, {ID: 2, Downloads: 1}, {ID: 4, Downloads: 90, CreatedAt: before.Add(-time.Hour)}, {ID: 5, Downloads: 4, CreatedAt: before.Add(time.Minute)}}}
	result := Compare(previous, current)
	states := map[int64]Change{}
	for _, a := range result.Assets {
		states[a.AssetID] = a
	}
	if *states[1].Delta != 3 || states[2].Delta != nil || states[2].State != "counter_decreased" || states[3].State != "missing_or_deleted" || states[4].Delta != nil || states[4].State != "late_baseline" || *states[5].Delta != 4 {
		t.Fatalf("bad lifecycle: %+v", result)
	}
	current.ReleasesOK = false
	if Compare(previous, current).State != "unavailable" {
		t.Fatal("failed requests cannot imply zero downloads")
	}
}

func TestGitHubPartialFailuresAndPagination(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer fixture-secret" {
			t.Error("missing authorization")
		}
		switch {
		case r.URL.Path == "/repos/sample/alpha":
			fmt.Fprint(w, `{"id":42,"stargazers_count":2,"forks_count":1}`)
		case strings.HasSuffix(r.URL.Path, "/releases"):
			if r.URL.Query().Get("page") == "1" {
				releases := make([]any, 100)
				for i := range releases {
					releases[i] = map[string]any{"draft": true}
				}
				json.NewEncoder(w).Encode(releases)
			} else {
				fmt.Fprint(w, `[{"id":7}]`)
			}
		case strings.HasSuffix(r.URL.Path, "/releases/7/assets"):
			fmt.Fprint(w, `[{"id":1,"name":"app.zip","download_count":7,"created_at":"2026-10-01T00:00:00Z"}]`)
		case strings.HasSuffix(r.URL.Path, "/views"):
			http.Error(w, "fixture-secret should never be persisted", 403)
		case strings.HasSuffix(r.URL.Path, "/clones"):
			fmt.Fprint(w, `{"count":3,"uniques":2,"clones":[{"timestamp":"2026-10-19T00:00:00Z","count":3,"uniques":2}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	store := openTest(t)
	work := mustWork(t, store, "sample/alpha")
	service := Service{Store: store, GitHub: GitHub{BaseURL: server.URL, Client: server.Client(), Token: "fixture-secret"}}
	snap, err := service.Collect(context.Background(), work.ID, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if snap.State != "partial" || !snap.ReleasesOK || len(snap.Assets) != 1 || snap.Stars == nil || *snap.Stars != 2 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
	if _, ok := snap.Traffic["views"]; ok {
		t.Fatal("403 must not turn into zero views")
	}
	n := calls
	again, err := service.Collect(context.Background(), work.ID, "request-1")
	if err != nil || again.ID != snap.ID || calls != n {
		t.Fatal("retry should return persisted result without more requests")
	}
	encoded, _ := json.Marshal(snap)
	if strings.Contains(string(encoded), "fixture-secret") {
		t.Fatal("secret leaked into snapshot")
	}
}

func TestHTTPAuthorizationHostAndValidation(t *testing.T) {
	store := openTest(t)
	handler := Handler(Service{Store: store}, "fixture-api-token")
	cases := []struct {
		method, path, host, auth, body string
		want                           int
	}{
		{"GET", "/api/works", "127.0.0.1:18440", "", "", 401},
		{"GET", "/api/works", "evil.example", "Bearer fixture-api-token", "", 403},
		{"POST", "/api/works", "127.0.0.1:18440", "Bearer fixture-api-token", `{"title":"Demo","repository":"../bad"}`, 400},
		{"POST", "/api/works", "127.0.0.1:18440", "Bearer fixture-api-token", `{"title":"Demo","repository":"sample/alpha"} {}`, 400},
		{"POST", "/api/works", "127.0.0.1:18440", "Bearer fixture-api-token", `{"title":"Demo","repository":"sample/alpha"}`, 200},
		{"GET", "/api/works/999/snapshots", "127.0.0.1:18440", "Bearer fixture-api-token", "", 404},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		req.Host = c.host
		req.Header.Set("Authorization", c.auth)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Fatalf("%s %s: got %d want %d: %s", c.method, c.path, rec.Code, c.want, rec.Body)
		}
	}
	req := httptest.NewRequest("GET", "/api/works", nil)
	req.Host = "localhost:18440"
	req.Header.Set("Authorization", "Bearer fixture-api-token")
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatal("cross-origin request accepted")
	}
}

func TestRepositoryIdentityCannotBeReassigned(t *testing.T) {
	s := openTest(t)
	w := mustWork(t, s, "sample/alpha")
	ctx := context.Background()
	snap := Snapshot{WorkID: w.ID, RepositoryID: 12, Key: "first", CapturedAt: time.Now().UTC()}
	if _, err := s.Save(ctx, snap); err != nil {
		t.Fatal(err)
	}
	snap.RepositoryID = 99
	snap.Key = "second"
	if _, err := s.Save(ctx, snap); err == nil {
		t.Fatal("repository replacement must not silently reuse history")
	}
	stored, err := s.Work(ctx, w.ID)
	if err != nil || stored.RepositoryID == nil || *stored.RepositoryID != 12 {
		t.Fatal("stable identity was overwritten")
	}
}

func TestSubsecondTrafficRevisionOrder(t *testing.T) {
	s := openTest(t)
	w := mustWork(t, s, "sample/nanosecond")
	ctx := context.Background()
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	for i, at := range []time.Time{now, now.Add(time.Nanosecond)} {
		_, err := s.Save(ctx, Snapshot{WorkID: w.ID, Key: fmt.Sprint(i), CapturedAt: at, Traffic: map[string]Traffic{"views": {Days: []Daily{{"2026-10-19", int64(5 - i), 1}}}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	days, err := s.Days(ctx, w.ID, "views")
	if err != nil || days[0].Count != 4 {
		t.Fatalf("subsecond correction lost: %+v %v", days, err)
	}
	snaps, err := s.Snapshots(ctx, w.ID)
	if err != nil || snaps[0].Key != "1" {
		t.Fatal("snapshot ordering must include fractions")
	}
}
