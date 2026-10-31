package echo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrInvalid = errors.New("invalid input")
var ErrNotFound = errors.New("not found")
var repoPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9_.-]+$`)

type Work struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Repository   string `json:"repository"`
	Kind         string `json:"kind"`
	CanonicalURL string `json:"canonical_url"`
	RepositoryID *int64 `json:"github_id,omitempty"`
}
type Asset struct {
	Category             string    `json:"category"`
	ClassificationReason string    `json:"classification_reason"`
	ID                   int64     `json:"id"`
	Name                 string    `json:"name"`
	Downloads            int64     `json:"download_count"`
	CreatedAt            time.Time `json:"created_at"`
}
type Daily struct {
	Day     string `json:"day"`
	Count   int64  `json:"count"`
	Uniques int64  `json:"uniques"`
}
type Traffic struct {
	Count   int64   `json:"count"`
	Uniques int64   `json:"uniques"`
	Days    []Daily `json:"days"`
}
type Snapshot struct {
	Downloads    *DownloadBreakdown         `json:"downloads"`
	Popular      map[string]json.RawMessage `json:"popular,omitempty"`
	Raw          map[string]json.RawMessage `json:"-"`
	ID           int64                      `json:"id"`
	RepositoryID int64                      `json:"github_id"`
	WorkID       int64                      `json:"work_id"`
	Key          string                     `json:"request_id"`
	CapturedAt   time.Time                  `json:"captured_at"`
	State        string                     `json:"state"`
	Stars        *int64                     `json:"stars"`
	Forks        *int64                     `json:"forks"`
	Assets       []Asset                    `json:"assets"`
	ReleasesOK   bool                       `json:"releases_available"`
	Traffic      map[string]Traffic         `json:"traffic"`
	Problems     []string                   `json:"problems"`
}
type Change struct {
	AssetID int64  `json:"asset_id"`
	Name    string `json:"name"`
	Delta   *int64 `json:"delta"`
	State   string `json:"state"`
}
type Interval struct {
	KnownDownloads *DownloadBreakdown `json:"known_downloads"`
	UnknownAssets  int                `json:"unknown_assets"`
	From           *time.Time         `json:"from"`
	To             *time.Time         `json:"to"`
	State          string             `json:"state"`
	Assets         []Change           `json:"assets"`
}

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: database path", ErrInvalid)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	// Create private runtime files before opening SQLite; do not interpolate paths into a DSN.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	f.Close()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	fail := func(err error) (*Store, error) { db.Close(); return nil, err }
	if _, err = db.Exec(`PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;`); err != nil {
		return fail(err)
	}
	var version int
	if err = db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fail(err)
	}
	if version > 3 {
		return fail(fmt.Errorf("database schema %d is newer than supported version 3", version))
	}
	if version == 0 {
		tx, err := db.Begin()
		if err != nil {
			return fail(err)
		}
		_, err = tx.Exec(`CREATE TABLE works(id INTEGER PRIMARY KEY,title TEXT NOT NULL,repository TEXT NOT NULL COLLATE NOCASE UNIQUE,github_id INTEGER UNIQUE);
   CREATE TABLE snapshots(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),request_key TEXT NOT NULL,captured_at TEXT NOT NULL,payload TEXT NOT NULL,UNIQUE(work_id,request_key));
   CREATE INDEX snapshot_work ON snapshots(work_id,id);
   CREATE TABLE traffic_days(work_id INTEGER NOT NULL REFERENCES works(id),metric TEXT NOT NULL,day TEXT NOT NULL,count INTEGER NOT NULL,uniques INTEGER NOT NULL,snapshot_id INTEGER NOT NULL REFERENCES snapshots(id),PRIMARY KEY(work_id,metric,day));
   PRAGMA user_version=1;`)
		if err != nil {
			tx.Rollback()
			return fail(err)
		}
		if err = tx.Commit(); err != nil {
			return fail(err)
		}
	}
	store := &Store{db: db}
	if err = store.migrateDomain(); err != nil {
		return fail(err)
	}
	if err = store.migrateSiteAndArchives(); err != nil {
		return fail(err)
	}
	return store, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) CreateWork(ctx context.Context, title, repository string) (Work, error) {
	if title == "" || len(title) > 200 || (!repoPattern.MatchString(repository) || strings.HasSuffix(repository, "/..") || strings.HasSuffix(repository, "/.")) {
		return Work{}, ErrInvalid
	}
	// Preserve the existing work on duplicate registration.
	_, err := s.db.ExecContext(ctx, `INSERT INTO works(title,repository) VALUES(?,?) ON CONFLICT(repository) DO NOTHING`, title, repository)
	if err != nil {
		return Work{}, err
	}
	var w Work
	err = s.db.QueryRowContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works WHERE repository=?`, repository).Scan(&w.ID, &w.Title, &w.Repository, &w.RepositoryID, &w.Kind, &w.CanonicalURL)
	return w, err
}
func (s *Store) Work(ctx context.Context, id int64) (Work, error) {
	var w Work
	err := s.db.QueryRowContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works WHERE id=?`, id).Scan(&w.ID, &w.Title, &w.Repository, &w.RepositoryID, &w.Kind, &w.CanonicalURL)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return w, err
}
func (s *Store) Works(ctx context.Context) ([]Work, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Work{}
	for rows.Next() {
		var w Work
		if err = rows.Scan(&w.ID, &w.Title, &w.Repository, &w.RepositoryID, &w.Kind, &w.CanonicalURL); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, rows.Err()
}
func (s *Store) SnapshotByKey(ctx context.Context, workID int64, key string) (Snapshot, error) {
	var id int64
	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT id,payload FROM snapshots WHERE work_id=? AND request_key=?`, workID, key).Scan(&id, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, ErrNotFound
	}
	if err != nil {
		return Snapshot{}, err
	}
	var snap Snapshot
	err = json.Unmarshal([]byte(payload), &snap)
	snap.ID = id
	return snap, err
}
func (s *Store) Save(ctx context.Context, snap Snapshot) (Snapshot, error) {
	annotateSnapshot(&snap)
	if snap.Key == "" || snap.CapturedAt.IsZero() {
		return Snapshot{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Snapshot{}, err
	}
	defer tx.Rollback()
	if snap.RepositoryID > 0 {
		var existing sql.NullInt64
		if err = tx.QueryRowContext(ctx, `SELECT github_id FROM works WHERE id=?`, snap.WorkID).Scan(&existing); err != nil {
			return Snapshot{}, err
		}
		if existing.Valid && existing.Int64 != snap.RepositoryID {
			return Snapshot{}, fmt.Errorf("%w: repository identity changed", ErrInvalid)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE works SET github_id=? WHERE id=?`, snap.RepositoryID, snap.WorkID); err != nil {
			return Snapshot{}, err
		}
	}
	payload, err := json.Marshal(snap)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO snapshots(work_id,request_key,captured_at,payload) VALUES(?,?,?,?) ON CONFLICT(work_id,request_key) DO NOTHING`, snap.WorkID, snap.Key, snap.CapturedAt.UTC().Format("2006-01-02T15:04:05.000000000Z"), string(payload))
	if err != nil {
		return Snapshot{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Snapshot{}, err
	}
	if n == 0 {
		tx.Rollback()
		return s.SnapshotByKey(ctx, snap.WorkID, snap.Key)
	}
	snap.ID, err = result.LastInsertId()
	if err != nil {
		return Snapshot{}, err
	}
	for endpoint, payload := range snap.Raw {
		if _, err = tx.ExecContext(ctx, `INSERT INTO raw_observations(snapshot_id,endpoint,payload) VALUES(?,?,?)`, snap.ID, endpoint, string(payload)); err != nil {
			return Snapshot{}, err
		}
	}
	for metric, traffic := range snap.Traffic {
		for _, day := range traffic.Days {
			_, err = tx.ExecContext(ctx, `INSERT INTO traffic_days(work_id,metric,day,count,uniques,snapshot_id) VALUES(?,?,?,?,?,?) ON CONFLICT(work_id,metric,day) DO UPDATE SET count=excluded.count,uniques=excluded.uniques,snapshot_id=excluded.snapshot_id WHERE (SELECT captured_at FROM snapshots WHERE id=excluded.snapshot_id)>=(SELECT captured_at FROM snapshots WHERE id=traffic_days.snapshot_id)`, snap.WorkID, metric, day.Day, day.Count, day.Uniques, snap.ID)
			if err != nil {
				return Snapshot{}, err
			}
		}
	}
	return snap, tx.Commit()
}
func (s *Store) Snapshots(ctx context.Context, workID int64) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,payload FROM snapshots WHERE work_id=? ORDER BY captured_at DESC,id DESC LIMIT 100`, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Snapshot{}
	for rows.Next() {
		var id int64
		var p string
		if err = rows.Scan(&id, &p); err != nil {
			return nil, err
		}
		var v Snapshot
		if err = json.Unmarshal([]byte(p), &v); err != nil {
			return nil, err
		}
		v.ID = id
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Days(ctx context.Context, workID int64, metric string) ([]Daily, error) {
	if metric != "views" && metric != "clones" {
		return nil, ErrInvalid
	}
	rows, err := s.db.QueryContext(ctx, `SELECT day,count,uniques FROM traffic_days WHERE work_id=? AND metric=? ORDER BY day`, workID, metric)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Daily{}
	for rows.Next() {
		var d Daily
		if err = rows.Scan(&d.Day, &d.Count, &d.Uniques); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Compare two complete release observations. Never infer a daily rate or a user count.
func Compare(previous, current Snapshot) Interval {
	result := Interval{From: &previous.CapturedAt, To: &current.CapturedAt, State: "ok", Assets: []Change{}}
	if !previous.ReleasesOK || !current.ReleasesOK {
		result.State = "unavailable"
		return result
	}
	old := map[int64]Asset{}
	for _, a := range previous.Assets {
		old[a.ID] = a
	}
	for _, a := range current.Assets {
		c := Change{AssetID: a.ID, Name: a.Name}
		if before, ok := old[a.ID]; ok {
			d := a.Downloads - before.Downloads
			if d < 0 {
				c.State = "counter_decreased"
			} else {
				c.State = "observed"
				c.Delta = &d
			}
			delete(old, a.ID)
		} else if a.CreatedAt.After(previous.CapturedAt) && !a.CreatedAt.After(current.CapturedAt) {
			d := a.Downloads
			c.State = "created_in_interval"
			c.Delta = &d
		} else {
			c.State = "late_baseline"
		}
		result.Assets = append(result.Assets, c)
	}
	for _, a := range old {
		result.Assets = append(result.Assets, Change{AssetID: a.ID, Name: a.Name, State: "missing_or_deleted"})
	}

	result.KnownDownloads = &DownloadBreakdown{ClassificationVersion: AssetClassificationVersion}
	for _, change := range result.Assets {
		if change.Delta == nil {
			result.UnknownAssets++
			continue
		}
		category, _ := ClassifyAsset(change.Name)
		result.KnownDownloads.add(category, *change.Delta)
	}
	return result
}
