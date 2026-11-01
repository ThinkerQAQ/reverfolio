package echo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Backup creates a consistent standalone database, including when writers are active.
// Paths are chosen by the service, never supplied by an HTTP/MCP client.
func (s *Store) Backup(ctx context.Context) (string, error) {
	f, err := os.CreateTemp("", "reverfolio-backup-*.sqlite")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if err = f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	if _, err = s.db.ExecContext(ctx, `VACUUM INTO ?`, path); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

type Review struct {
	ID         int64  `json:"id"`
	MentionID  int64  `json:"mention_id"`
	From       string `json:"from"`
	To         string `json:"to"`
	Reason     string `json:"reason"`
	Reviewer   string `json:"reviewer"`
	ReviewedAt string `json:"reviewed_at"`
	Version    int64  `json:"version"`
}
type SearchRun struct {
	ID          int64  `json:"id"`
	TaskID      int64  `json:"task_id"`
	Engine      string `json:"engine"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	ResultCount int    `json:"result_count"`
	Notes       string `json:"notes"`
	CompletedAt string `json:"completed_at"`
}
type Audit struct {
	ActorAliases []ActorAlias            `json:"actor_aliases"`
	ActorHistory []ActorLinkHistory      `json:"actor_history"`
	Reviews      []Review                `json:"reviews"`
	SearchRuns   []SearchRun             `json:"search_runs"`
	History      []HistoricalObservation `json:"history"`
}

func (s *Store) Audit(ctx context.Context) (Audit, error) {
	out := Audit{Reviews: []Review{}, SearchRuns: []SearchRun{}, History: []HistoricalObservation{}}
	rows, err := s.db.QueryContext(ctx, `SELECT id,mention_id,from_state,to_state,reason,reviewer,reviewed_at,version FROM reviews ORDER BY id`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var r Review
		if err = rows.Scan(&r.ID, &r.MentionID, &r.From, &r.To, &r.Reason, &r.Reviewer, &r.ReviewedAt, &r.Version); err != nil {
			rows.Close()
			return out, err
		}
		out.Reviews = append(out.Reviews, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT id,task_id,engine,scope,status,result_count,notes,completed_at FROM search_runs ORDER BY id`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var r SearchRun
		if err = rows.Scan(&r.ID, &r.TaskID, &r.Engine, &r.Scope, &r.Status, &r.ResultCount, &r.Notes, &r.CompletedAt); err != nil {
			rows.Close()
			return out, err
		}
		out.SearchRuns = append(out.SearchRuns, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT work_id,external_id,metric,value,unit,period_start,period_end,timezone,scope,source,observed_at FROM historical_observations ORDER BY id`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var h HistoricalObservation
		var start, end, observed string
		if err = rows.Scan(&h.WorkID, &h.ExternalID, &h.Metric, &h.Value, &h.Unit, &start, &end, &h.Timezone, &h.Scope, &h.Source, &observed); err != nil {
			return out, err
		}
		h.PeriodStart, _ = time.Parse(timestampLayout, start)
		h.PeriodEnd, _ = time.Parse(timestampLayout, end)
		h.ObservedAt, _ = time.Parse(timestampLayout, observed)
		out.History = append(out.History, h)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	out.ActorAliases, out.ActorHistory, err = s.IdentityAudit(ctx)
	return out, err
}
func (s Service) registerOperations(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/audit", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.Audit(r.Context())
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("POST /api/backup", func(w http.ResponseWriter, r *http.Request) {
		path, err := s.Store.Backup(r.Context())
		if err != nil {
			http.Error(w, "backup failed", 500)
			return
		}
		defer os.Remove(path)
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "backup unavailable", 500)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/vnd.sqlite3")
		w.Header().Set("Content-Disposition", `attachment; filename="reverfolio-backup.sqlite"`)
		io.Copy(w, f)
	})
}

// Scheduler uses a stable UTC interval key, so restarting the process does not
// duplicate an already persisted observation in the same interval.
func (s Service) runScheduledCollection(ctx context.Context, fallback time.Duration, followConfig bool, at time.Time, onError func(error)) {
	interval := fallback
	var scheduledVersion int64
	if s.Config != nil && s.Config.Path != "" {
		config, err := s.Config.Read()
		if err != nil {
			onError(fmt.Errorf("could not read collection schedule"))
			return
		}
		scheduledVersion = config.Version
	}
	// A configuration change invalidates the remaining batch. An already running
	// request may finish; the next scheduler tick uses the new configuration.
	current := func() bool {
		if ctx.Err() != nil {
			return false
		}
		if scheduledVersion == 0 {
			return true
		}
		config, err := s.Config.Read()
		if err != nil {
			onError(fmt.Errorf("could not read collection schedule"))
			return false
		}
		return config.Version == scheduledVersion
	}
	if followConfig && s.Config != nil && s.Config.Path != "" {
		config, err := s.Config.Read()
		if err != nil {
			onError(fmt.Errorf("could not read collection schedule"))
			return
		}
		interval, err = time.ParseDuration(config.Collection.Interval)
		if err != nil {
			onError(fmt.Errorf("invalid collection interval"))
			return
		}
	}
	if interval == 0 {
		return
	}
	if interval < time.Minute {
		onError(fmt.Errorf("invalid collection interval"))
		return
	}
	works, err := s.Store.Works(ctx)
	if err != nil {
		onError(err)
		return
	}
	bucket := at.UTC().Unix() / int64(interval/time.Second)
	for _, work := range works {
		if !current() {
			return
		}
		enabled, checkErr := s.Store.MonitoringEnabled(ctx, work.ID)
		if checkErr != nil {
			onError(checkErr)
			return
		}
		if !enabled || work.Kind != "project" {
			continue
		}
		child, cancel := context.WithTimeout(ctx, 90*time.Second)
		_, err = s.Collect(child, work.ID, fmt.Sprintf("scheduled:%d:%d", int64(interval/time.Second), bucket))
		cancel()
		if err != nil && ctx.Err() == nil {
			onError(fmt.Errorf("scheduled collection failed for work %d", work.ID))
		}
	}
	if s.Config != nil {
		c, err := s.Config.Read()
		if err != nil {
			onError(err)
			return
		}
		for _, source := range c.Sources {
			if !current() {
				return
			}
			enabled, checkErr := s.Store.MonitoringEnabled(ctx, source.WorkID)
			if checkErr != nil {
				onError(checkErr)
				return
			}
			if !enabled || !source.Enabled {
				continue
			}
			raw, _ := json.Marshal(source)
			credentialState := "missing"
			if s.credentialPresent(source) {
				credentialState = "present"
			}
			key := fmt.Sprintf("scheduled:%d:%d:%x:%s", int64(interval/time.Second), bucket, sha256.Sum256(raw), credentialState)
			child, cancel := context.WithTimeout(ctx, 90*time.Second)
			_, err := s.CollectSiteSource(child, CollectSiteSourceInput{SourceID: source.ID, RequestID: key})
			cancel()
			if err != nil && ctx.Err() == nil {
				onError(fmt.Errorf("scheduled site collection failed for source %s", source.ID))
			}
		}
	}
}

// Re-read the nonsecret schedule every minute, including while disabled.
// Explicit CLI interval overrides remain fixed; stdio MCP never starts this loop.
func (s Service) RunScheduler(ctx context.Context, interval time.Duration, followConfig bool, onError func(error)) {
	tick := func() { s.runScheduledCollection(ctx, interval, followConfig, time.Now(), onError) }
	tick()
	timer := time.NewTicker(time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			tick()
		}
	}
}

// RestoreDatabase is an offline operation. Validate and migrate a temporary copy
// before placing it at a new destination; never overwrite an existing database.
func RestoreDatabase(ctx context.Context, source, destination string) error {
	if _, err := os.Stat(destination); err == nil {
		return ErrConflict
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(destination), ".reverfolio-restore-*.sqlite")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = io.Copy(tmp, io.LimitReader(in, 1024*1024*1024+1)); err != nil {
		tmp.Close()
		return err
	}
	size, err := tmp.Seek(0, io.SeekEnd)
	tmp.Close()
	if err != nil || size > 1024*1024*1024 {
		return ErrInvalid
	}
	probe, err := sql.Open("sqlite", name)
	if err != nil {
		return err
	}
	var integrity string
	err = probe.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity)
	var version int
	if err == nil {
		err = probe.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version)
	}
	probe.Close()
	if err != nil || integrity != "ok" || version < 1 || version > 6 {
		return ErrInvalid
	}
	store, err := Open(name)
	if err != nil {
		return err
	}
	defer store.Close()
	var n int
	if err = store.db.QueryRowContext(ctx, `SELECT count(*) FROM works`).Scan(&n); err != nil {
		return err
	}
	rows, err := store.db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	bad := rows.Next()
	rowErr := rows.Err()
	rows.Close()
	if bad || rowErr != nil {
		return ErrInvalid
	}
	store.Close()
	// Hard-link placement is atomic and fails if another writer created destination.
	return os.Link(name, destination)
}
