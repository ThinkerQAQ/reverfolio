package echo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"strings"
	"time"
	_ "time/tzdata"
)

var ErrConflict = errors.New("conflict")
var ErrOriginalSource = errors.New("original-source evidence required")

const timestampLayout = "2006-01-02T15:04:05.000000000Z"

func stamp(t time.Time) string { return t.UTC().Format(timestampLayout) }
func randomID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// NormalizeURL preserves path identity and meaningful query parameters. It never fetches a URL.
func NormalizeURL(raw string) (string, error) {
	if len(raw) > 4096 {
		return "", ErrInvalid
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" {
		return "", ErrInvalid
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".test") || strings.HasSuffix(host, ".invalid") || !strings.Contains(host, ".") {
		return "", ErrInvalid
	}
	if ip := net.ParseIP(host); ip != nil && (!ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback()) {
		return "", ErrInvalid
	}
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	q := u.Query()
	for key := range q {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "cookie") || strings.Contains(lower, "api_key") {
			return "", ErrInvalid
		}
	}
	for k := range q {
		if strings.HasPrefix(strings.ToLower(k), "utm_") || k == "fbclid" || k == "gclid" {
			q.Del(k)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Migration 2 keeps IDs and existing foreign-key references while allowing non-repository works.
func (s *Store) migrateDomain() error {
	var version int
	if err := s.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version >= 2 {
		return nil
	}
	if _, err := s.db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
		return err
	}
	defer s.db.Exec(`PRAGMA foreign_keys=ON`)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE work_registry(id INTEGER PRIMARY KEY,title TEXT NOT NULL,repository TEXT COLLATE NOCASE UNIQUE,github_id INTEGER UNIQUE,kind TEXT NOT NULL DEFAULT 'project' CHECK(kind IN ('project','article')),canonical_url TEXT UNIQUE);
 INSERT INTO work_registry(id,title,repository,github_id) SELECT id,title,repository,github_id FROM works;
 DROP TABLE works;
 ALTER TABLE work_registry RENAME TO works;
 CREATE TABLE channel_pages(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),platform TEXT NOT NULL,language TEXT NOT NULL,url TEXT NOT NULL UNIQUE);
 CREATE TABLE commands(kind TEXT NOT NULL,request_id TEXT NOT NULL,input_hash TEXT NOT NULL,result TEXT NOT NULL,PRIMARY KEY(kind,request_id));
 CREATE TABLE mentions(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),url TEXT NOT NULL,type TEXT NOT NULL,origin TEXT NOT NULL,generation TEXT NOT NULL,actor_key TEXT NOT NULL,actor_name TEXT NOT NULL,event_key TEXT NOT NULL,state TEXT NOT NULL DEFAULT 'pending',version INTEGER NOT NULL DEFAULT 1,discovered_at TEXT NOT NULL,verified_at TEXT,UNIQUE(work_id,url));
 CREATE TABLE evidence(id INTEGER PRIMARY KEY,mention_id INTEGER NOT NULL REFERENCES mentions(id),quote TEXT NOT NULL,source_kind TEXT NOT NULL,content_hash TEXT NOT NULL,observed_at TEXT NOT NULL,captured_at TEXT NOT NULL,UNIQUE(mention_id,content_hash,source_kind));
 CREATE TABLE reviews(id INTEGER PRIMARY KEY,mention_id INTEGER NOT NULL REFERENCES mentions(id),from_state TEXT NOT NULL,to_state TEXT NOT NULL,reason TEXT NOT NULL,reviewer TEXT NOT NULL,reviewed_at TEXT NOT NULL,version INTEGER NOT NULL);
 CREATE TABLE discovery_tasks(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),query TEXT NOT NULL,state TEXT NOT NULL DEFAULT 'pending',worker TEXT NOT NULL DEFAULT '',lease_token TEXT NOT NULL DEFAULT '',lease_until TEXT NOT NULL DEFAULT '',created_at TEXT NOT NULL);
 CREATE TABLE search_runs(id INTEGER PRIMARY KEY,task_id INTEGER NOT NULL REFERENCES discovery_tasks(id),engine TEXT NOT NULL,scope TEXT NOT NULL,status TEXT NOT NULL,result_count INTEGER NOT NULL,notes TEXT NOT NULL,completed_at TEXT NOT NULL);
 CREATE TABLE historical_observations(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),external_id TEXT NOT NULL UNIQUE,metric TEXT NOT NULL,value REAL NOT NULL,unit TEXT NOT NULL,period_start TEXT NOT NULL,period_end TEXT NOT NULL,timezone TEXT NOT NULL,scope TEXT NOT NULL,source TEXT NOT NULL,observed_at TEXT NOT NULL,imported_at TEXT NOT NULL);
 PRAGMA user_version=2;`)
	if err != nil {
		return err
	}
	rows, err := tx.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	broken := rows.Next()
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return rowErr
	}
	if broken {
		return fmt.Errorf("migration would break foreign keys")
	}
	return tx.Commit()
}

// Commands are atomic and replayable; a reused request ID with changed input is rejected.
func command[T any](ctx context.Context, s *Store, kind, key string, input any, fn func(*sql.Tx) (T, error)) (T, error) {
	var zero T
	if len(key) < 1 || len(key) > 128 {
		return zero, ErrInvalid
	}
	body, err := json.Marshal(input)
	if err != nil {
		return zero, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(body))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	var oldHash, result string
	err = tx.QueryRowContext(ctx, `SELECT input_hash,result FROM commands WHERE kind=? AND request_id=?`, kind, key).Scan(&oldHash, &result)
	if err == nil {
		if hash != oldHash {
			return zero, ErrConflict
		}
		err = json.Unmarshal([]byte(result), &zero)
		return zero, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return zero, err
	}
	value, err := fn(tx)
	if err != nil {
		return zero, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return zero, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO commands VALUES(?,?,?,?)`, kind, key, hash, string(encoded)); err != nil {
		return zero, err
	}
	return value, tx.Commit()
}
func existsWork(ctx context.Context, tx *sql.Tx, id int64) error {
	var found int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM works WHERE id=?`, id).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func oneOf(v string, values ...string) bool {
	for _, x := range values {
		if v == x {
			return true
		}
	}
	return false
}

type WorkInput struct {
	RequestID    string `json:"request_id"`
	Title        string `json:"title"`
	Kind         string `json:"kind"`
	Repository   string `json:"repository,omitempty"`
	CanonicalURL string `json:"canonical_url,omitempty"`
}

func normalizeWorkInput(in WorkInput) (WorkInput, error) {
	if in.Kind == "" {
		in.Kind = "project"
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Kind == "project" {
		if in.Title == "" || len(in.Title) > 200 || !repoPattern.MatchString(in.Repository) || strings.HasSuffix(in.Repository, "/..") || strings.HasSuffix(in.Repository, "/.") {
			return WorkInput{}, ErrInvalid
		}
	} else if in.Kind == "article" || in.Kind == "site" {
		if in.Title == "" || len(in.Title) > 200 || in.Repository != "" {
			return WorkInput{}, ErrInvalid
		}
		u, err := NormalizeURL(in.CanonicalURL)
		if err != nil {
			return WorkInput{}, err
		}
		in.CanonicalURL = u
	} else {
		return WorkInput{}, ErrInvalid
	}
	return in, nil
}

func (s *Store) RegisterWork(ctx context.Context, in WorkInput) (Work, error) {
	var err error
	in, err = normalizeWorkInput(in)
	if err != nil {
		return Work{}, err
	}
	return command(ctx, s, "register_work", in.RequestID, in, func(tx *sql.Tx) (Work, error) { return registerWorkTx(ctx, tx, in) })
}
func registerWorkTx(ctx context.Context, tx *sql.Tx, in WorkInput) (Work, error) {
	var row *sql.Row
	if in.Kind == "project" {
		_, err := tx.ExecContext(ctx, `INSERT INTO works(title,repository,kind) VALUES(?,?,'project') ON CONFLICT(repository) DO NOTHING`, in.Title, in.Repository)
		if err != nil {
			return Work{}, err
		}
		row = tx.QueryRowContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works WHERE repository=?`, in.Repository)
	} else {
		_, err := tx.ExecContext(ctx, `INSERT INTO works(title,kind,canonical_url) VALUES(?,?,?) ON CONFLICT(canonical_url) DO NOTHING`, in.Title, in.Kind, in.CanonicalURL)
		if err != nil {
			return Work{}, err
		}
		row = tx.QueryRowContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works WHERE canonical_url=?`, in.CanonicalURL)
	}
	var w Work
	err := row.Scan(&w.ID, &w.Title, &w.Repository, &w.RepositoryID, &w.Kind, &w.CanonicalURL)
	if err == nil && w.Kind != in.Kind {
		return Work{}, ErrConflict
	}
	return w, err
}

type ChannelInput struct {
	RequestID string `json:"request_id"`
	WorkID    int64  `json:"work_id"`
	Platform  string `json:"platform"`
	Language  string `json:"language"`
	URL       string `json:"url"`
}
type Channel struct {
	ID       int64  `json:"id"`
	WorkID   int64  `json:"work_id"`
	Platform string `json:"platform"`
	Language string `json:"language"`
	URL      string `json:"url"`
}

func normalizeChannelInput(in ChannelInput) (ChannelInput, error) {
	u, err := NormalizeURL(in.URL)
	if err != nil || in.Platform == "" || len(in.Platform) > 80 || in.Language == "" || len(in.Language) > 32 {
		return ChannelInput{}, ErrInvalid
	}
	in.URL = u
	return in, nil
}
func (s *Store) AddChannel(ctx context.Context, in ChannelInput) (Channel, error) {
	var err error
	in, err = normalizeChannelInput(in)
	if err != nil {
		return Channel{}, err
	}
	return command(ctx, s, "add_channel", in.RequestID, in, func(tx *sql.Tx) (Channel, error) { return addChannelTx(ctx, tx, in) })
}
func addChannelTx(ctx context.Context, tx *sql.Tx, in ChannelInput) (Channel, error) {
	if err := existsWork(ctx, tx, in.WorkID); err != nil {
		return Channel{}, err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO channel_pages(work_id,platform,language,url) VALUES(?,?,?,?) ON CONFLICT(url) DO NOTHING`, in.WorkID, in.Platform, in.Language, in.URL)
	if err != nil {
		return Channel{}, err
	}
	var c Channel
	err = tx.QueryRowContext(ctx, `SELECT id,work_id,platform,language,url FROM channel_pages WHERE url=?`, in.URL).Scan(&c.ID, &c.WorkID, &c.Platform, &c.Language, &c.URL)
	if err == nil && (c.WorkID != in.WorkID || c.Platform != in.Platform || c.Language != in.Language) {
		return Channel{}, ErrConflict
	}
	return c, err
}

func (s *Store) Channels(ctx context.Context, workID int64) ([]Channel, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,work_id,platform,language,url FROM channel_pages WHERE work_id=? ORDER BY id`, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Channel{}
	for rows.Next() {
		var c Channel
		if err = rows.Scan(&c.ID, &c.WorkID, &c.Platform, &c.Language, &c.URL); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type MentionInput struct {
	RequestID  string    `json:"request_id"`
	WorkID     int64     `json:"work_id"`
	URL        string    `json:"url"`
	Type       string    `json:"type"`
	Origin     string    `json:"origin"`
	Generation string    `json:"generation"`
	ActorKey   string    `json:"actor_key,omitempty"`
	ActorName  string    `json:"actor_name,omitempty"`
	EventKey   string    `json:"event_key,omitempty"`
	Quote      string    `json:"quote"`
	SourceKind string    `json:"source_kind"`
	ObservedAt time.Time `json:"observed_at"`
}
type Evidence struct {
	ID         int64  `json:"id"`
	Quote      string `json:"quote"`
	SourceKind string `json:"source_kind"`
	Hash       string `json:"content_hash"`
	ObservedAt string `json:"observed_at"`
	CapturedAt string `json:"captured_at"`
}
type Mention struct {
	ID           int64      `json:"id"`
	WorkID       int64      `json:"work_id"`
	URL          string     `json:"url"`
	Type         string     `json:"type"`
	Origin       string     `json:"origin"`
	Generation   string     `json:"generation"`
	ActorKey     string     `json:"actor_key"`
	ActorName    string     `json:"actor_name"`
	EventKey     string     `json:"event_key"`
	State        string     `json:"state"`
	Version      int64      `json:"version"`
	DiscoveredAt string     `json:"discovered_at"`
	VerifiedAt   *string    `json:"verified_at"`
	Evidence     []Evidence `json:"evidence"`
}

const mentionColumns = `id,work_id,url,type,origin,generation,actor_key,actor_name,event_key,state,version,discovered_at,verified_at`

func scanMention(row interface{ Scan(...any) error }) (Mention, error) {
	var m Mention
	err := row.Scan(&m.ID, &m.WorkID, &m.URL, &m.Type, &m.Origin, &m.Generation, &m.ActorKey, &m.ActorName, &m.EventKey, &m.State, &m.Version, &m.DiscoveredAt, &m.VerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	m.Evidence = []Evidence{}
	return m, err
}
func (s *Store) SubmitMention(ctx context.Context, in MentionInput) (Mention, error) {
	u, err := NormalizeURL(in.URL)
	if err != nil {
		return Mention{}, err
	}
	in.URL = u
	if !oneOf(in.Type, "mention", "citation", "repost", "rewrite", "comment", "adoption", "aggregation", "publication") || !oneOf(in.Origin, "self", "third_party", "automated", "unknown") || !oneOf(in.Generation, "human", "ai", "unknown") || !oneOf(in.SourceKind, "web", "api", "secondary_report") || strings.TrimSpace(in.Quote) == "" || len(in.Quote) > 1200 || len(in.ActorName) > 200 || len(in.EventKey) > 200 || in.ObservedAt.IsZero() || in.ObservedAt.After(time.Now().Add(5*time.Minute)) {
		return Mention{}, ErrInvalid
	}
	if in.ActorKey != "" {
		in.ActorKey, err = NormalizeURL(in.ActorKey)
		if err != nil {
			return Mention{}, ErrInvalid
		}
	}
	return command(ctx, s, "submit_mention", in.RequestID, in, func(tx *sql.Tx) (Mention, error) {
		if err := existsWork(ctx, tx, in.WorkID); err != nil {
			return Mention{}, err
		}
		now := stamp(time.Now())
		_, err := tx.ExecContext(ctx, `INSERT INTO mentions(work_id,url,type,origin,generation,actor_key,actor_name,event_key,discovered_at) VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(work_id,url) DO NOTHING`, in.WorkID, in.URL, in.Type, in.Origin, in.Generation, in.ActorKey, in.ActorName, in.EventKey, now)
		if err != nil {
			return Mention{}, err
		}
		m, err := scanMention(tx.QueryRowContext(ctx, `SELECT `+mentionColumns+` FROM mentions WHERE work_id=? AND url=?`, in.WorkID, in.URL))
		if err != nil {
			return Mention{}, err
		}
		if m.Type != in.Type || m.Origin != in.Origin || m.Generation != in.Generation || m.ActorKey != in.ActorKey || m.EventKey != in.EventKey {
			return Mention{}, ErrConflict
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(in.Quote)))
		_, err = tx.ExecContext(ctx, `INSERT INTO evidence(mention_id,quote,source_kind,content_hash,observed_at,captured_at) VALUES(?,?,?,?,?,?) ON CONFLICT(mention_id,content_hash,source_kind) DO NOTHING`, m.ID, in.Quote, in.SourceKind, hash, stamp(in.ObservedAt), now)
		if err != nil {
			return Mention{}, err
		}
		m.Evidence, err = readEvidence(ctx, tx, m.ID)
		return m, err
	})
}
func readEvidence(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, id int64) ([]Evidence, error) {
	rows, err := q.QueryContext(ctx, `SELECT id,quote,source_kind,content_hash,observed_at,captured_at FROM evidence WHERE mention_id=? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Evidence{}
	for rows.Next() {
		var e Evidence
		if err = rows.Scan(&e.ID, &e.Quote, &e.SourceKind, &e.Hash, &e.ObservedAt, &e.CapturedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (s *Store) Mentions(ctx context.Context, workID int64) ([]Mention, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+mentionColumns+` FROM mentions WHERE (?=0 OR work_id=?) ORDER BY id`, workID, workID)
	if err != nil {
		return nil, err
	}
	out := []Mention{}
	for rows.Next() {
		m, e := scanMention(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Evidence, err = readEvidence(ctx, s.db, out[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

type ReviewInput struct {
	RequestID       string `json:"request_id"`
	MentionID       int64  `json:"mention_id"`
	ExpectedVersion int64  `json:"expected_version"`
	State           string `json:"state"`
	Reason          string `json:"reason"`
	Reviewer        string `json:"reviewer"`
}

func (s *Store) ReviewMention(ctx context.Context, in ReviewInput) (Mention, error) {
	if !oneOf(in.State, "verified", "rejected", "pending", "gone") || strings.TrimSpace(in.Reason) == "" || len(in.Reason) > 2000 || in.Reviewer == "" || len(in.Reviewer) > 200 {
		return Mention{}, ErrInvalid
	}
	return command(ctx, s, "review_mention", in.RequestID, in, func(tx *sql.Tx) (Mention, error) {
		m, err := scanMention(tx.QueryRowContext(ctx, `SELECT `+mentionColumns+` FROM mentions WHERE id=?`, in.MentionID))
		if err != nil {
			return Mention{}, err
		}
		if m.Version != in.ExpectedVersion {
			return Mention{}, ErrConflict
		}
		if in.State == "verified" {
			var n int
			if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM evidence WHERE mention_id=? AND source_kind IN ('web','api')`, m.ID).Scan(&n); err != nil {
				return Mention{}, err
			}
			if n == 0 {
				return Mention{}, fmt.Errorf("%w: %w", ErrInvalid, ErrOriginalSource)
			}
		}
		now := stamp(time.Now())
		var verified any
		if in.State == "verified" {
			verified = now
		}
		if _, err = tx.ExecContext(ctx, `UPDATE mentions SET state=?,version=version+1,verified_at=? WHERE id=?`, in.State, verified, m.ID); err != nil {
			return Mention{}, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO reviews(mention_id,from_state,to_state,reason,reviewer,reviewed_at,version) VALUES(?,?,?,?,?,?,?)`, m.ID, m.State, in.State, in.Reason, in.Reviewer, now, m.Version+1); err != nil {
			return Mention{}, err
		}
		m.State = in.State
		m.Version++
		m.VerifiedAt = nil
		if verified != nil {
			m.VerifiedAt = &now
		}
		m.Evidence, err = readEvidence(ctx, tx, m.ID)
		return m, err
	})
}

type TaskInput struct {
	RequestID string `json:"request_id"`
	WorkID    int64  `json:"work_id"`
	Query     string `json:"query"`
}
type Task struct {
	ID         int64  `json:"id"`
	WorkID     int64  `json:"work_id"`
	Query      string `json:"query"`
	State      string `json:"state"`
	Worker     string `json:"worker"`
	LeaseUntil string `json:"lease_until"`
	LeaseToken string `json:"lease_token,omitempty"`
}

func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.WorkID, &t.Query, &t.State, &t.Worker, &t.LeaseUntil)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return t, err
}

const taskColumns = `id,work_id,query,state,worker,lease_until`

func (s *Store) CreateTask(ctx context.Context, in TaskInput) (Task, error) {
	if strings.TrimSpace(in.Query) == "" || len(in.Query) > 2000 {
		return Task{}, ErrInvalid
	}
	return command(ctx, s, "create_task", in.RequestID, in, func(tx *sql.Tx) (Task, error) {
		if err := existsWork(ctx, tx, in.WorkID); err != nil {
			return Task{}, err
		}
		r, err := tx.ExecContext(ctx, `INSERT INTO discovery_tasks(work_id,query,created_at) VALUES(?,?,?)`, in.WorkID, in.Query, stamp(time.Now()))
		if err != nil {
			return Task{}, err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return Task{}, err
		}
		return scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM discovery_tasks WHERE id=?`, id))
	})
}
func (s *Store) Tasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns+` FROM discovery_tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Task{}
	for rows.Next() {
		t, e := scanTask(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type ClaimInput struct {
	RequestID string `json:"request_id"`
	TaskID    int64  `json:"task_id"`
	Worker    string `json:"worker"`
}

func (s *Store) ClaimTask(ctx context.Context, in ClaimInput) (Task, error) {
	if in.Worker == "" || len(in.Worker) > 200 {
		return Task{}, ErrInvalid
	}
	return command(ctx, s, "claim_task", in.RequestID, in, func(tx *sql.Tx) (Task, error) {
		task, err := scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM discovery_tasks WHERE id=?`, in.TaskID))
		if err != nil {
			return Task{}, err
		}
		now := stamp(time.Now())
		if task.State != "pending" && task.State != "partial" && task.State != "failed" && (task.State != "running" || task.LeaseUntil > now) {
			return Task{}, ErrConflict
		}
		task.State = "running"
		task.Worker = in.Worker
		task.LeaseToken = randomID()
		task.LeaseUntil = stamp(time.Now().Add(10 * time.Minute))
		_, err = tx.ExecContext(ctx, `UPDATE discovery_tasks SET state='running',worker=?,lease_token=?,lease_until=? WHERE id=?`, task.Worker, task.LeaseToken, task.LeaseUntil, task.ID)
		return task, err
	})
}

type CompleteInput struct {
	RequestID   string `json:"request_id"`
	TaskID      int64  `json:"task_id"`
	LeaseToken  string `json:"lease_token"`
	Engine      string `json:"engine"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	ResultCount int    `json:"result_count"`
	Notes       string `json:"notes"`
}

func (s *Store) CompleteTask(ctx context.Context, in CompleteInput) (Task, error) {
	if !oneOf(in.Status, "complete", "partial", "failed") || in.Engine == "" || in.Scope == "" || len(in.Scope) > 4000 || len(in.Notes) > 4000 || in.ResultCount < 0 {
		return Task{}, ErrInvalid
	}
	return command(ctx, s, "complete_task", in.RequestID, in, func(tx *sql.Tx) (Task, error) {
		var token, state, until string
		err := tx.QueryRowContext(ctx, `SELECT lease_token,state,lease_until FROM discovery_tasks WHERE id=?`, in.TaskID).Scan(&token, &state, &until)
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		if err != nil {
			return Task{}, err
		}
		if token == "" || token != in.LeaseToken || state != "running" || until < stamp(time.Now()) {
			return Task{}, ErrConflict
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO search_runs(task_id,engine,scope,status,result_count,notes,completed_at) VALUES(?,?,?,?,?,?,?)`, in.TaskID, in.Engine, in.Scope, in.Status, in.ResultCount, in.Notes, stamp(time.Now())); err != nil {
			return Task{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE discovery_tasks SET state=?,lease_token='',lease_until='' WHERE id=?`, in.Status, in.TaskID); err != nil {
			return Task{}, err
		}
		return scanTask(tx.QueryRowContext(ctx, `SELECT `+taskColumns+` FROM discovery_tasks WHERE id=?`, in.TaskID))
	})
}

type HistoricalObservation struct {
	WorkID      int64     `json:"work_id"`
	ExternalID  string    `json:"external_id"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Timezone    string    `json:"timezone"`
	Scope       string    `json:"scope"`
	Source      string    `json:"source"`
	ObservedAt  time.Time `json:"observed_at"`
}
type ImportInput struct {
	RequestID     string                  `json:"request_id"`
	SchemaVersion int                     `json:"schema_version"`
	DryRun        bool                    `json:"dry_run"`
	Rows          []HistoricalObservation `json:"rows"`
}
type ImportResult struct {
	Valid    bool              `json:"valid"`
	DryRun   bool              `json:"dry_run"`
	Inserted int               `json:"inserted"`
	Existing int               `json:"existing"`
	Rows     []ImportRowResult `json:"rows"`
}
type ImportRowResult struct {
	ExternalID string `json:"external_id"`
	State      string `json:"state"`
}

func (s *Store) ImportHistory(ctx context.Context, in ImportInput) (ImportResult, error) {
	for i := range in.Rows {
		in.Rows[i].PeriodStart = in.Rows[i].PeriodStart.UTC()
		in.Rows[i].PeriodEnd = in.Rows[i].PeriodEnd.UTC()
		in.Rows[i].ObservedAt = in.Rows[i].ObservedAt.UTC()
	}
	if in.SchemaVersion != 1 || len(in.Rows) == 0 || len(in.Rows) > 500 {
		return ImportResult{}, ErrInvalid
	}
	validate := func(tx *sql.Tx) (ImportResult, error) {
		result := ImportResult{Valid: true, DryRun: in.DryRun, Rows: []ImportRowResult{}}
		seen := map[string]bool{}
		for _, row := range in.Rows {
			_, zoneErr := time.LoadLocation(row.Timezone)
			if row.ExternalID == "" || len(row.ExternalID) > 200 || seen[row.ExternalID] || row.Metric == "" || len(row.Metric) > 100 || row.Unit == "" || row.Scope == "" || row.Source == "" || row.ObservedAt.IsZero() || row.PeriodStart.IsZero() || row.PeriodEnd.Before(row.PeriodStart) || zoneErr != nil || row.Value < 0 || math.IsNaN(row.Value) || math.IsInf(row.Value, 0) {
				return ImportResult{}, ErrInvalid
			}
			seen[row.ExternalID] = true
			if err := existsWork(ctx, tx, row.WorkID); err != nil {
				return ImportResult{}, err
			}
			var previous HistoricalObservation
			var start, end, observed string
			err := tx.QueryRowContext(ctx, `SELECT work_id,external_id,metric,value,unit,period_start,period_end,timezone,scope,source,observed_at FROM historical_observations WHERE external_id=?`, row.ExternalID).Scan(&previous.WorkID, &previous.ExternalID, &previous.Metric, &previous.Value, &previous.Unit, &start, &end, &previous.Timezone, &previous.Scope, &previous.Source, &observed)
			state := "new"
			if err == nil {
				previous.PeriodStart, _ = time.Parse(timestampLayout, start)
				previous.PeriodEnd, _ = time.Parse(timestampLayout, end)
				previous.ObservedAt, _ = time.Parse(timestampLayout, observed)
				a, _ := json.Marshal(previous)
				b, _ := json.Marshal(row)
				if string(a) != string(b) {
					return ImportResult{}, ErrConflict
				}
				state = "existing"
				result.Existing++
			} else if !errors.Is(err, sql.ErrNoRows) {
				return ImportResult{}, err
			} else {
				result.Inserted++
			}
			result.Rows = append(result.Rows, ImportRowResult{row.ExternalID, state})
		}
		if in.DryRun {
			return result, nil
		}
		for _, row := range in.Rows {
			_, err := tx.ExecContext(ctx, `INSERT INTO historical_observations(work_id,external_id,metric,value,unit,period_start,period_end,timezone,scope,source,observed_at,imported_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(external_id) DO NOTHING`, row.WorkID, row.ExternalID, row.Metric, row.Value, row.Unit, stamp(row.PeriodStart), stamp(row.PeriodEnd), row.Timezone, row.Scope, row.Source, stamp(row.ObservedAt), stamp(time.Now()))
			if err != nil {
				return ImportResult{}, err
			}
		}
		return result, nil
	}
	if in.DryRun {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return ImportResult{}, err
		}
		defer tx.Rollback()
		return validate(tx)
	}
	return command(ctx, s, "import_history", in.RequestID, in, validate)
}

type SourceSummary struct {
	ValidMetrics map[string]MetricObservation `json:"valid_metrics"`
	LastRelease  *Snapshot                    `json:"last_release"`
	WorkID       int64                        `json:"work_id"`
	Title        string                       `json:"title"`
	Repository   string                       `json:"repository"`
	Latest       *Snapshot                    `json:"latest"`
	Changes      Interval                     `json:"last_comparable_changes"`
}
type Report struct {
	ProviderRuns []ProviderRun           `json:"provider_runs"`
	Sources      []SourceSummary         `json:"sources"`
	History      []HistoricalObservation `json:"history"`

	GeneratedAt            string    `json:"generated_at"`
	WorkID                 int64     `json:"work_id"`
	Works                  []Work    `json:"works"`
	Mentions               []Mention `json:"mentions"`
	VerifiedPages          int       `json:"verified_pages"`
	ThirdPartyEvents       int       `json:"third_party_events"`
	VerifiedActorURLs      int       `json:"verified_actor_urls"`
	VerifiedActors         int       `json:"verified_actors"`
	Pending                int       `json:"pending"`
	HistoricalObservations int       `json:"historical_observations"`
	Notes                  []string  `json:"notes"`
}

func (s *Store) Report(ctx context.Context, workID int64) (Report, error) {
	result := Report{GeneratedAt: stamp(time.Now()), WorkID: workID, Notes: []string{"Downloads, clones, clicks and visitor identifiers are not verified users.", "Actor counts use explicit audited URL groups; display names alone are not merged.", "Searches cover recorded queries and sources, not the entire web."}}
	var err error
	result.Works, err = s.Works(ctx)
	if err != nil {
		return result, err
	}
	if workID != 0 {
		w, err := s.Work(ctx, workID)
		if err != nil {
			return result, err
		}
		result.Works = []Work{w}
	}
	result.Sources = []SourceSummary{}
	for _, work := range result.Works {
		if work.Kind != "project" {
			continue
		}
		overview, err := s.Overview(ctx, work.ID)
		if err != nil {
			return result, err
		}
		snapshots, err := s.dashboardSnapshots(ctx, work.ID)
		if err != nil {
			return result, err
		}
		summary := summarizeObservations(work, snapshots, overview.Changes)
		result.Sources = append(result.Sources, summary)
	}
	audit, err := s.Audit(ctx)
	if err != nil {
		return result, err
	}
	runs, err := s.ProviderRuns(ctx)
	if err != nil {
		return result, err
	}
	result.ProviderRuns = []ProviderRun{}
	seenRuns := map[string]bool{}
	for _, run := range runs {
		if !seenRuns[run.SourceID] && (workID == 0 || run.WorkID == workID) {
			result.ProviderRuns = append(result.ProviderRuns, run)
			seenRuns[run.SourceID] = true
		}
	}
	result.History = []HistoricalObservation{}
	for _, row := range audit.History {
		if workID == 0 || row.WorkID == workID {
			result.History = append(result.History, row)
		}
	}
	result.Mentions, err = s.Mentions(ctx, workID)
	if err != nil {
		return result, err
	}
	identities := map[string]string{}
	for _, a := range audit.ActorAliases {
		identities[a.Alias] = a.IdentityURL
	}
	actorURLs := map[string]bool{}
	actors := map[string]bool{}
	events := map[string]bool{}
	for _, m := range result.Mentions {
		if m.State == "pending" {
			result.Pending++
		}
		if m.State != "verified" {
			continue
		}
		result.VerifiedPages++
		if m.Origin != "third_party" {
			continue
		}
		if m.ActorKey != "" {
			actorURLs[m.ActorKey] = true
			key := m.ActorKey
			if identity := identities[key]; identity != "" {
				key = identity
			}
			actors[key] = true
		}
		key := m.EventKey
		if key == "" {
			key = fmt.Sprint("mention:", m.ID)
		}
		events[fmt.Sprint(m.WorkID, ":", key)] = true
	}
	result.VerifiedActorURLs = len(actorURLs)
	result.VerifiedActors = len(actors)
	result.ThirdPartyEvents = len(events)
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM historical_observations WHERE (?=0 OR work_id=?)`, workID, workID).Scan(&result.HistoricalObservations)
	return result, err
}
