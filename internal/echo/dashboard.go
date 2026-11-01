package echo

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type DashboardBar struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}
type DashboardObservation struct {
	At        time.Time `json:"at"`
	Stars     *int64    `json:"stars"`
	Forks     *int64    `json:"forks"`
	Downloads *int64    `json:"downloads"`
}
type DashboardProject struct {
	Work         Work                   `json:"work"`
	AttemptAt    *time.Time             `json:"attempt_at"`
	AttemptState string                 `json:"attempt_state"`
	Problems     []string               `json:"problems"`
	TrafficAt    *time.Time             `json:"traffic_at"`
	Traffic      map[string]Traffic     `json:"traffic"`
	Days         map[string][]Daily     `json:"days"`
	Popular      map[string]any         `json:"popular"`
	Downloads    *DownloadBreakdown     `json:"downloads"`
	DownloadAt   *time.Time             `json:"download_at"`
	Changes      Interval               `json:"changes"`
	Releases     []DashboardBar         `json:"releases"`
	Platforms    []DashboardBar         `json:"platforms"`
	Observations []DashboardObservation `json:"observations"`
}
type Dashboard struct {
	GeneratedAt time.Time          `json:"generated_at"`
	Projects    []DashboardProject `json:"projects"`
	Providers   []ProviderRun      `json:"providers"`
}

func sortedBars(m map[string]int64) []DashboardBar {
	out := []DashboardBar{}
	for k, v := range m {
		out = append(out, DashboardBar{Label: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Value == out[j].Value {
			return out[i].Label < out[j].Label
		}
		return out[i].Value > out[j].Value
	})
	return out
}
func platformForAsset(a Asset) string {
	if a.Category != "package" {
		return a.Category
	}
	name := strings.ToLower(a.Name)
	for _, entry := range []struct{ key, label string }{{"windows", "Windows"}, {"darwin", "macOS"}, {"macos", "macOS"}, {"linux", "Linux"}, {"extension", "Extension"}} {
		if strings.Contains(name, entry.key) {
			return entry.label
		}
	}
	if strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi") {
		return "Windows"
	}
	return "Unknown platform"
}
func (s *Store) Dashboard(ctx context.Context) (Dashboard, error) {
	out := Dashboard{GeneratedAt: time.Now().UTC(), Projects: []DashboardProject{}, Providers: []ProviderRun{}}
	works, err := s.Works(ctx)
	if err != nil {
		return out, err
	}
	for _, w := range works {
		if w.Kind != "project" {
			continue
		}
		snapshots, err := s.dashboardSnapshots(ctx, w.ID)
		if err != nil {
			return out, err
		}
		p := DashboardProject{Work: w, Problems: []string{}, Traffic: map[string]Traffic{}, Days: map[string][]Daily{}, Popular: map[string]any{}, Releases: []DashboardBar{}, Platforms: []DashboardBar{}, Observations: []DashboardObservation{}}
		if len(snapshots) > 0 {
			at := snapshots[0].CapturedAt
			p.AttemptAt = &at
			p.AttemptState = snapshots[0].State
			p.Problems = snapshots[0].Problems
		}
		releaseSnapshots := []Snapshot{}
		for index, snap := range snapshots {
			point := DashboardObservation{At: snap.CapturedAt, Stars: snap.Stars, Forks: snap.Forks}
			if snap.Downloads != nil {
				total := snap.Downloads.Total
				point.Downloads = &total
			}
			if index < 100 {
				p.Observations = append(p.Observations, point)
			}
			for _, key := range []string{"views", "clones"} {
				if _, ok := p.Traffic[key]; !ok {
					if v, found := snap.Traffic[key]; found {
						p.Traffic[key] = v
						at := snap.CapturedAt
						if p.TrafficAt == nil || at.After(*p.TrafficAt) {
							p.TrafficAt = &at
						}
					}
				}
			}
			for _, key := range []string{"referrers", "paths"} {
				if _, ok := p.Popular[key]; !ok && snap.Popular[key] != nil {
					var v any
					if json.Unmarshal(snap.Popular[key], &v) == nil {
						p.Popular[key] = v
					}
				}
			}
			if snap.ReleasesOK {
				releaseSnapshots = append(releaseSnapshots, snap)
			}
		}
		if len(releaseSnapshots) > 0 {
			latest := releaseSnapshots[0]
			p.Downloads = latest.Downloads
			at := latest.CapturedAt
			p.DownloadAt = &at
			releases, platforms := map[string]int64{}, map[string]int64{}
			for _, a := range latest.Assets {
				tag := a.ReleaseTag
				if tag == "" {
					tag = "Unknown release (legacy snapshot)"
				}
				releases[tag] += a.Downloads
				platforms[platformForAsset(a)] += a.Downloads
			}
			p.Releases = sortedBars(releases)
			p.Platforms = sortedBars(platforms)
		}
		if len(releaseSnapshots) > 1 {
			p.Changes = Compare(releaseSnapshots[1], releaseSnapshots[0])
		}
		for _, key := range []string{"views", "clones"} {
			p.Days[key], err = s.Days(ctx, w.ID, key)
			if err != nil {
				return out, err
			}
		}
		out.Projects = append(out.Projects, p)
	}
	runs, err := s.providerRunsQuery(ctx, `SELECT id,payload FROM provider_runs WHERE id IN (SELECT MAX(id) FROM provider_runs GROUP BY source_id UNION SELECT MAX(id) FROM provider_runs WHERE state IN ('complete','empty') GROUP BY source_id) ORDER BY id DESC`)
	if err != nil {
		return out, err
	}
	seen := map[string]bool{}
	successful := map[string]bool{}
	for _, run := range runs {
		if !seen[run.SourceID] {
			out.Providers = append(out.Providers, run)
			seen[run.SourceID] = true
			if run.State == "complete" || run.State == "empty" {
				successful[run.SourceID] = true
			}
			continue
		}
		if !successful[run.SourceID] && (run.State == "complete" || run.State == "empty") {
			out.Providers = append(out.Providers, run)
			successful[run.SourceID] = true
		}
	}
	return out, nil
}

func (s *Store) dashboardSnapshots(ctx context.Context, workID int64) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,payload FROM snapshots WHERE work_id=? AND id IN (
 SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? ORDER BY captured_at DESC,id DESC LIMIT 100)
 UNION SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? AND json_extract(payload,'$.releases_available')=1 ORDER BY captured_at DESC,id DESC LIMIT 2)
 UNION SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? AND json_extract(payload,'$.traffic.views') IS NOT NULL ORDER BY captured_at DESC,id DESC LIMIT 1)
 UNION SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? AND json_extract(payload,'$.traffic.clones') IS NOT NULL ORDER BY captured_at DESC,id DESC LIMIT 1)
 UNION SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? AND json_extract(payload,'$.popular.paths') IS NOT NULL ORDER BY captured_at DESC,id DESC LIMIT 1)
 UNION SELECT id FROM (SELECT id FROM snapshots WHERE work_id=? AND json_extract(payload,'$.popular.referrers') IS NOT NULL ORDER BY captured_at DESC,id DESC LIMIT 1)
 ) ORDER BY captured_at DESC,id DESC`, workID, workID, workID, workID, workID, workID, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Snapshot{}
	for rows.Next() {
		var snap Snapshot
		var raw string
		if err = rows.Scan(&snap.ID, &raw); err != nil {
			return nil, err
		}
		id := snap.ID
		if err = json.Unmarshal([]byte(raw), &snap); err != nil {
			return nil, err
		}
		snap.ID = id
		out = append(out, snap)
	}
	return out, rows.Err()
}
