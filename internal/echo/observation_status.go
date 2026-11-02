package echo

import "time"

// MetricObservation retains each metric's own source time; absent is different from zero.
type MetricObservation struct {
	Value      int64     `json:"value"`
	ObservedAt time.Time `json:"observed_at"`
}

func summarizeObservations(work Work, snapshots []Snapshot, changes Interval) SourceSummary {
	out := SourceSummary{WorkID: work.ID, Title: work.Title, Repository: work.Repository, Changes: changes, ValidMetrics: map[string]MetricObservation{}}
	if len(snapshots) > 0 {
		latest := snapshots[0]
		out.Latest = &latest
	}
	put := func(key string, value int64, at time.Time) {
		if _, ok := out.ValidMetrics[key]; !ok {
			out.ValidMetrics[key] = MetricObservation{value, at}
		}
	}
	releases := []Snapshot{}
	for _, snap := range snapshots {
		if snap.ReleasesOK {
			releases = append(releases, snap)
		}
		if snap.Stars != nil {
			put("stars", *snap.Stars, snap.CapturedAt)
		}
		if snap.Forks != nil {
			put("forks", *snap.Forks, snap.CapturedAt)
		}
		if snap.ReleasesOK && out.LastRelease == nil {
			release := snap
			annotateSnapshot(&release)
			out.LastRelease = &release
			put("downloads", release.Downloads.Total, snap.CapturedAt)
			put("packages", release.Downloads.Packages, snap.CapturedAt)
		}
		for _, key := range []string{"views", "clones"} {
			if t, ok := snap.Traffic[key]; ok {
				put(key, t.Count, snap.CapturedAt)
			}
		}
	}
	if len(releases) > 1 {
		out.Changes = Compare(releases[1], releases[0])
	}
	return out
}
