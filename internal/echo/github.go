package echo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type GitHub struct {
	Client  *http.Client
	BaseURL string
	Token   string
}

func (g GitHub) getOnce(ctx context.Context, path string, out any) error {
	base := g.BaseURL
	if base == "" {
		base = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("User-Agent", "OpenEcho")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub request failed")
	}
	defer res.Body.Close()
	// Never persist remote error bodies, request headers, or tokens.
	if res.StatusCode != http.StatusOK {
		retry, _ := strconv.Atoi(res.Header.Get("Retry-After"))
		return apiStatusError{res.StatusCode, time.Duration(retry) * time.Second}
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 8*1024*1024+1))
	if err != nil || len(data) > 8*1024*1024 {
		return fmt.Errorf("GitHub response exceeds limit or could not be read")
	}
	if err = json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("invalid GitHub response")
	}
	captureRaw(ctx, path, data)
	return nil
}
func (g GitHub) Collect(ctx context.Context, work Work, key string) Snapshot {
	snap := Snapshot{WorkID: work.ID, Key: key, CapturedAt: time.Now().UTC(), State: "complete", Traffic: map[string]Traffic{}, Assets: []Asset{}, Problems: []string{}}
	snap.Raw = map[string]json.RawMessage{}
	ctx = context.WithValue(ctx, rawCaptureKey{}, snap.Raw)
	fail := func(source string, err error) {
		snap.Problems = append(snap.Problems, source+": "+err.Error())
		snap.State = "partial"
	}
	prefix := "/repos/" + work.Repository
	var repo struct {
		ID    int64  `json:"id"`
		Stars *int64 `json:"stargazers_count"`
		Forks *int64 `json:"forks_count"`
	}
	if err := g.get(ctx, prefix, &repo); err != nil {
		fail("repository", err)
	} else if repo.ID == 0 || repo.Stars == nil || repo.Forks == nil {
		fail("repository", fmt.Errorf("missing required fields"))
	} else {
		snap.Stars = repo.Stars
		snap.Forks = repo.Forks
		snap.RepositoryID = repo.ID
	}
	snap.ReleasesOK = true
	for page := 1; page <= 100; page++ {
		var releases []struct {
			ID     int64   `json:"id"`
			Draft  bool    `json:"draft"`
			Assets []Asset `json:"assets"`
		}
		if err := g.get(ctx, prefix+"/releases?per_page=100&page="+strconv.Itoa(page), &releases); err != nil {
			fail("releases", err)
			snap.ReleasesOK = false
			break
		}
		for _, release := range releases {
			if release.Draft {
				continue
			}
			if release.ID <= 0 {
				snap.ReleasesOK = false
				fail("releases", fmt.Errorf("missing release identity"))
				continue
			}
			for assetPage := 1; assetPage <= 100; assetPage++ {
				var assets []Asset
				if err := g.get(ctx, prefix+"/releases/"+strconv.FormatInt(release.ID, 10)+"/assets?per_page=100&page="+strconv.Itoa(assetPage), &assets); err != nil {
					snap.ReleasesOK = false
					fail("releases", err)
					break
				}
				for _, a := range assets {
					if a.ID <= 0 || a.Downloads < 0 || a.CreatedAt.IsZero() {
						snap.ReleasesOK = false
						fail("releases", fmt.Errorf("invalid asset"))
						break
					}
					snap.Assets = append(snap.Assets, a)
				}
				if len(assets) < 100 {
					break
				}
				if assetPage == 100 {
					snap.ReleasesOK = false
					fail("releases", fmt.Errorf("asset pagination limit reached"))
				}
			}
		}
		if len(releases) < 100 {
			break
		}
		if page == 100 {
			snap.ReleasesOK = false
			fail("releases", fmt.Errorf("pagination limit reached"))
		}
	}
	seen := map[int64]bool{}
	for _, asset := range snap.Assets {
		if seen[asset.ID] {
			snap.ReleasesOK = false
			fail("releases", fmt.Errorf("duplicate asset in paginated response"))
			break
		}
		seen[asset.ID] = true
	}
	if !snap.ReleasesOK {
		snap.Assets = []Asset{}
	}
	for _, metric := range []string{"views", "clones"} {
		var data struct {
			Count   *int64 `json:"count"`
			Uniques *int64 `json:"uniques"`
			Views   []struct {
				Timestamp time.Time `json:"timestamp"`
				Count     int64     `json:"count"`
				Uniques   int64     `json:"uniques"`
			} `json:"views"`
			Clones []struct {
				Timestamp time.Time `json:"timestamp"`
				Count     int64     `json:"count"`
				Uniques   int64     `json:"uniques"`
			} `json:"clones"`
		}
		if err := g.get(ctx, prefix+"/traffic/"+metric+"?per=day", &data); err != nil {
			fail(metric, err)
			continue
		}
		if data.Count == nil || data.Uniques == nil {
			fail(metric, fmt.Errorf("missing totals"))
			continue
		}
		t := Traffic{Count: *data.Count, Uniques: *data.Uniques, Days: []Daily{}}
		rows := data.Views
		if metric == "clones" {
			rows = data.Clones
		}
		valid := rows != nil && *data.Count >= 0 && *data.Uniques >= 0
		for _, row := range rows {
			if row.Timestamp.IsZero() || row.Count < 0 || row.Uniques < 0 {
				valid = false
				break
			}
			t.Days = append(t.Days, Daily{row.Timestamp.UTC().Format(time.DateOnly), row.Count, row.Uniques})
		}
		if valid {
			snap.Traffic[metric] = t
		} else {
			fail(metric, fmt.Errorf("invalid daily data"))
		}
	}
	snap.Popular = map[string]json.RawMessage{}
	for _, kind := range []string{"referrers", "paths"} {
		var items []json.RawMessage
		if err := g.get(ctx, prefix+"/traffic/popular/"+kind, &items); err != nil {
			fail(kind, err)
			continue
		}
		if items == nil {
			fail(kind, fmt.Errorf("invalid popular-content response"))
			continue
		}
		body, err := json.Marshal(items)
		if err != nil {
			fail(kind, err)
			continue
		}
		snap.Popular[kind] = body
	}
	if snap.Stars == nil && !snap.ReleasesOK && len(snap.Traffic) == 0 {
		snap.State = "failed"
	}
	return snap
}

type Service struct {
	Store  *Store
	GitHub GitHub
	Config *ConfigManager
}

func (s Service) Collect(ctx context.Context, id int64, key string) (Snapshot, error) {
	if len(key) < 1 || len(key) > 128 {
		return Snapshot{}, ErrInvalid
	}
	work, err := s.Store.Work(ctx, id)
	if err != nil {
		return Snapshot{}, err
	}
	if work.Kind != "project" {
		return Snapshot{}, ErrInvalid
	}
	previous, err := s.Store.SnapshotByKey(ctx, id, key)
	if err == nil {
		return previous, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Snapshot{}, err
	}
	return s.Store.Save(ctx, s.GitHub.Collect(ctx, work, key))
}
