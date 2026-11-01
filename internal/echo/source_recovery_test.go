package echo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCredentialWriteFailureRestoresConfiguration(t *testing.T) {
	ctx := context.Background()
	store := openTest(t)
	work, err := store.RegisterWork(ctx, WorkInput{RequestID: "site", Kind: "site", Title: "Fixture", CanonicalURL: "https://example.org/"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	s := Service{Store: store, Config: NewConfigManager(path)}
	c := DefaultConfig()
	if err = saveConfig(path, c); err != nil {
		t.Fatal(err)
	}
	source := fixtureSource("umami")
	source.WorkID = work.ID
	// A directory in place of the secret forces rename failure on every OS.
	if err = os.MkdirAll(s.credentialPath(source.ID), 0700); err != nil {
		t.Fatal(err)
	}
	_, err = s.SaveSiteSource(ctx, SaveSiteSourceInput{RequestID: "save", ExpectedVersion: c.Version, Source: source, Credential: "synthetic-only"})
	if err == nil || !strings.Contains(err.Error(), "configuration restored") {
		t.Fatalf("error %v", err)
	}
	after, err := s.Config.Read()
	if err != nil || after.Version != c.Version || len(after.Sources) != 0 {
		t.Fatalf("configuration changed: %+v %v", after, err)
	}
	leftovers, err := filepath.Glob(filepath.Join(filepath.Dir(s.credentialPath(source.ID)), ".credential-*"))
	if err != nil || len(leftovers) != 0 {
		t.Fatal("temporary credential retained", err)
	}
}

func TestScheduleChangeDuringCollectionStopsRemainingWorks(t *testing.T) {
	ctx := context.Background()
	store := openTest(t)
	mustWork(t, store, "sample/first")
	second := mustWork(t, store, "sample/second")
	path := filepath.Join(t.TempDir(), "config.toml")
	c := DefaultConfig()
	c.Collection.Interval = "1h"
	if err := saveConfig(path, c); err != nil {
		t.Fatal(err)
	}
	manager := NewConfigManager(path)
	var once sync.Once
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			_, err := manager.UpdateSettings(ctx, UpdateSettingsInput{RequestID: "disable-in-flight", Updates: []SettingChange{{Key: "collection.interval", Value: "0", ExpectedVersion: c.Version}}})
			if err != nil {
				t.Error(err)
			}
		})
		switch {
		case r.URL.Path == "/repos/sample/first":
			fmt.Fprint(w, `{"id":42,"stargazers_count":0,"forks_count":0}`)
		case strings.Contains(r.URL.Path, "/traffic/"):
			fmt.Fprint(w, `{"count":0,"uniques":0,"views":[],"clones":[]}`)
		default:
			fmt.Fprint(w, `[]`)
		}
	}))
	defer api.Close()
	s := Service{Store: store, Config: manager, GitHub: GitHub{BaseURL: api.URL}}
	s.runScheduledCollection(ctx, time.Hour, true, time.Now(), func(err error) { t.Error(err) })
	snapshots, err := store.Snapshots(ctx, second.ID)
	if err != nil || len(snapshots) != 0 {
		t.Fatal("started a subsequent work after disabling", err)
	}
}
