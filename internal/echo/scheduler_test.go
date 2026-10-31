package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduleHotReloadDisableAndRestartIdempotency(t *testing.T) {
	var requests atomic.Int64
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch {
		case r.URL.Path == "/repos/sample/scheduled":
			fmt.Fprint(w, `{"id":42,"stargazers_count":0,"forks_count":0}`)
		case strings.HasSuffix(r.URL.Path, "/traffic/views"):
			fmt.Fprint(w, `{"count":0,"uniques":0,"views":[]}`)
		case strings.HasSuffix(r.URL.Path, "/traffic/clones"):
			fmt.Fprint(w, `{"count":0,"uniques":0,"clones":[]}`)
		default:
			fmt.Fprint(w, `[]`)
		}
	}))
	defer api.Close()
	store := openTest(t)
	work := mustWork(t, store, "sample/scheduled")
	cfg := DefaultConfig()
	cfg.Collection.Interval = "0"
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := saveConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	manager := NewConfigManager(path)
	service := Service{Store: store, GitHub: GitHub{BaseURL: api.URL}, Config: manager}
	ctx := context.Background()
	now := time.Now().UTC()
	onError := func(err error) { t.Fatal(err) }
	service.runScheduledCollection(ctx, time.Hour, true, now, onError)
	if requests.Load() != 0 {
		t.Fatal("disabled schedule made requests")
	}
	update := func(value string) {
		t.Helper()
		c, err := manager.Read()
		if err != nil {
			t.Fatal(err)
		}
		_, err = manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: randomID(), Updates: []SettingChange{{Key: "collection.interval", Value: value, ExpectedVersion: c.Version}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	update("1h")
	service.runScheduledCollection(ctx, time.Hour, true, now, onError)
	count := requests.Load()
	if count == 0 {
		t.Fatal("enabled schedule did not collect")
	}
	snapshots, err := store.Snapshots(ctx, work.ID)
	if err != nil || len(snapshots) != 1 || snapshots[0].State != "complete" {
		t.Fatalf("snapshots %+v error %v", snapshots, err)
	}
	// A fresh service (process restart) uses the same persisted interval key.
	restarted := Service{Store: store, GitHub: GitHub{BaseURL: api.URL}, Config: NewConfigManager(path)}
	restarted.runScheduledCollection(ctx, time.Hour, true, now, onError)
	if requests.Load() != count {
		t.Fatal("restart fetched the same interval again")
	}
	update("0")
	service.runScheduledCollection(ctx, time.Hour, true, now.Add(time.Hour), onError)
	if requests.Load() != count {
		t.Fatal("disabled schedule did not stop")
	}
	// An explicit CLI interval is independent of the saved schedule.
	service.runScheduledCollection(ctx, time.Hour, false, now.Add(time.Hour), onError)
	if requests.Load() <= count {
		t.Fatal("CLI interval override ignored")
	}
}
func TestScheduleValidationAndSafeRuntimeStatus(t *testing.T) {
	for _, v := range []string{"-1h", "1s", "1000h", "not-a-duration"} {
		if validateSetting("collection.interval", v) == nil {
			t.Fatalf("accepted %s", v)
		}
	}
	for _, v := range []string{"0", "0s", "1m", "24h", "168h"} {
		if err := validateSetting("collection.interval", v); err != nil {
			t.Fatalf("rejected %s: %v", v, err)
		}
	}
	s := Service{GitHub: GitHub{Token: "private-credential"}}
	status, err := s.RuntimeStatus(context.Background())
	if err != nil || !status.GitHubCredentialLoaded {
		t.Fatalf("status %+v %v", status, err)
	}
	raw, _ := json.Marshal(status)
	if strings.Contains(string(raw), "private-credential") {
		t.Fatal("runtime leaked credential")
	}
}
