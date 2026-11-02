package echo

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func (s *Store) migrateDiscovery() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= 9 {
		return nil
	}
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS discovery_run_details(search_run_id INTEGER PRIMARY KEY REFERENCES search_runs(id),checked_result TEXT NOT NULL,queries TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS discovery_run_mentions(search_run_id INTEGER NOT NULL REFERENCES search_runs(id),mention_id INTEGER NOT NULL REFERENCES mentions(id),PRIMARY KEY(search_run_id,mention_id));
 CREATE TABLE IF NOT EXISTS work_search_identifiers(work_id INTEGER NOT NULL REFERENCES works(id),kind TEXT NOT NULL,value TEXT NOT NULL,PRIMARY KEY(work_id,kind,value));
 PRAGMA user_version=9;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

type RenewDiscoveryInput struct {
	RequestID  string `json:"request_id"`
	TaskID     int64  `json:"task_id"`
	LeaseToken string `json:"lease_token"`
}

func (s *Store) RenewDiscovery(ctx context.Context, in RenewDiscoveryInput) (Task, error) {
	if in.LeaseToken == "" {
		return Task{}, ErrInvalid
	}
	return command(ctx, s, "renew_discovery", in.RequestID, in, func(tx *sql.Tx) (Task, error) {
		now := time.Now()
		r, err := tx.ExecContext(ctx, "UPDATE discovery_tasks SET lease_until=? WHERE id=? AND state='running' AND lease_token=? AND lease_until>?", stamp(now.Add(10*time.Minute)), in.TaskID, in.LeaseToken, stamp(now))
		if err != nil {
			return Task{}, err
		}
		n, err := r.RowsAffected()
		if err != nil {
			return Task{}, err
		}
		if n != 1 {
			return Task{}, ErrConflict
		}
		task, err := scanTask(tx.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM discovery_tasks WHERE id=?", in.TaskID))
		return task, err
	})
}
func recordDiscoveryDetails(ctx context.Context, tx *sql.Tx, in CompleteInput, runID int64) error {
	var workID int64
	var taskQuery string
	if err := tx.QueryRowContext(ctx, "SELECT work_id,query FROM discovery_tasks WHERE id=?", in.TaskID).Scan(&workID, &taskQuery); err != nil {
		return err
	}
	checked := in.CheckedResult
	if checked == "" {
		checked = "incomplete"
		if in.Status == "complete" && in.ResultCount == 0 {
			checked = "not_found_in_checked_scope"
		}
		if in.ResultCount > 0 {
			if len(in.MentionIDs) > 0 {
				checked = "found"
			} else {
				checked = "legacy_unlinked"
			}
		}
	} else if !oneOf(checked, "found", "not_found_in_checked_scope", "incomplete") {
		return ErrInvalid
	}
	if checked == "not_found_in_checked_scope" && (in.Status != "complete" || in.ResultCount != 0 || len(in.MentionIDs) > 0) {
		return ErrInvalid
	}
	if checked == "found" && (in.ResultCount == 0 || len(in.MentionIDs) == 0) {
		return ErrInvalid
	}
	if len(in.MentionIDs) > 100 || len(in.MentionIDs) > in.ResultCount || len(in.Queries) > 50 {
		return ErrInvalid
	}
	queries := in.Queries
	if len(queries) == 0 {
		queries = []string{taskQuery}
	}
	for _, q := range queries {
		if strings.TrimSpace(q) == "" || len(q) > 2000 {
			return ErrInvalid
		}
	}
	raw, err := json.Marshal(queries)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO discovery_run_details VALUES(?,?,?)", runID, checked, string(raw)); err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, id := range in.MentionIDs {
		if seen[id] {
			return ErrInvalid
		}
		seen[id] = true
		var actual int64
		if err = tx.QueryRowContext(ctx, "SELECT work_id FROM mentions WHERE id=?", id).Scan(&actual); err != nil {
			return ErrInvalid
		}
		if actual != workID {
			return ErrInvalid
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO discovery_run_mentions VALUES(?,?)", runID, id); err != nil {
			return err
		}
	}
	return nil
}

type WorkSearchIdentifier struct {
	WorkID int64  `json:"work_id"`
	Kind   string `json:"kind"`
	Value  string `json:"value"`
}
type SaveSearchIdentifierInput struct {
	RequestID  string               `json:"request_id"`
	Identifier WorkSearchIdentifier `json:"identifier"`
}

func (s *Store) SaveSearchIdentifier(ctx context.Context, in SaveSearchIdentifierInput) (WorkSearchIdentifier, error) {
	id := in.Identifier
	if !oneOf(id.Kind, "historical_name", "historical_url", "translated_title", "phrase") || strings.TrimSpace(id.Value) == "" || len(id.Value) > 2000 {
		return id, ErrInvalid
	}
	if id.Kind == "historical_url" {
		if _, err := NormalizeURL(id.Value); err != nil {
			return id, err
		}
	}
	return command(ctx, s, "save_search_identifier", in.RequestID, in, func(tx *sql.Tx) (WorkSearchIdentifier, error) {
		if err := existsWork(ctx, tx, id.WorkID); err != nil {
			return id, err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO work_search_identifiers VALUES(?,?,?) ON CONFLICT DO NOTHING", id.WorkID, id.Kind, id.Value)
		return id, err
	})
}

type DiscoveryRunDetail struct {
	RunID         int64    `json:"run_id"`
	TaskID        int64    `json:"task_id"`
	WorkID        int64    `json:"work_id"`
	Engine        string   `json:"engine"`
	Scope         string   `json:"scope"`
	Status        string   `json:"status"`
	CheckedResult string   `json:"checked_result"`
	Queries       []string `json:"queries"`
	MentionIDs    []int64  `json:"mention_ids"`
	CompletedAt   string   `json:"completed_at"`
}
type DiscoveryCoverage struct {
	Runs             []DiscoveryRunDetail   `json:"runs"`
	Identifiers      []WorkSearchIdentifier `json:"identifiers"`
	SuggestedQueries map[int64][]string     `json:"suggested_queries"`
}

func (s *Store) DiscoveryCoverage(ctx context.Context) (DiscoveryCoverage, error) {
	out := DiscoveryCoverage{Runs: []DiscoveryRunDetail{}, Identifiers: []WorkSearchIdentifier{}, SuggestedQueries: map[int64][]string{}}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id,r.task_id,t.work_id,r.engine,r.scope,r.status,COALESCE(d.checked_result,'legacy_unknown'),COALESCE(d.queries,'[]'),r.completed_at FROM search_runs r JOIN discovery_tasks t ON t.id=r.task_id LEFT JOIN discovery_run_details d ON d.search_run_id=r.id ORDER BY r.id DESC`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var r DiscoveryRunDetail
		var raw string
		if err = rows.Scan(&r.RunID, &r.TaskID, &r.WorkID, &r.Engine, &r.Scope, &r.Status, &r.CheckedResult, &raw, &r.CompletedAt); err != nil {
			rows.Close()
			return out, err
		}
		if err = json.Unmarshal([]byte(raw), &r.Queries); err != nil {
			rows.Close()
			return out, err
		}
		r.MentionIDs = []int64{}
		out.Runs = append(out.Runs, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return out, err
	}
	for i := range out.Runs {
		r := &out.Runs[i]
		rows, err = s.db.QueryContext(ctx, "SELECT mention_id FROM discovery_run_mentions WHERE search_run_id=? ORDER BY mention_id", r.RunID)
		if err != nil {
			return out, err
		}
		for rows.Next() {
			var id int64
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return out, err
			}
			r.MentionIDs = append(r.MentionIDs, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return out, err
		}
	}
	works, err := s.Works(ctx)
	if err != nil {
		return out, err
	}
	quote := func(v string) string { return "\"" + strings.ReplaceAll(v, "\"", "\\\"") + "\"" }
	for _, w := range works {
		out.SuggestedQueries[w.ID] = []string{quote(w.Title)}
		if w.CanonicalURL != "" {
			out.SuggestedQueries[w.ID] = append(out.SuggestedQueries[w.ID], quote(w.CanonicalURL))
		}
		if w.Repository != "" {
			out.SuggestedQueries[w.ID] = append(out.SuggestedQueries[w.ID], quote("github.com/"+w.Repository))
		}
	}
	rows, err = s.db.QueryContext(ctx, "SELECT work_id,kind,value FROM work_search_identifiers ORDER BY work_id,kind,value")
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id WorkSearchIdentifier
		if err = rows.Scan(&id.WorkID, &id.Kind, &id.Value); err != nil {
			return out, err
		}
		out.Identifiers = append(out.Identifiers, id)
		out.SuggestedQueries[id.WorkID] = append(out.SuggestedQueries[id.WorkID], quote(id.Value))
	}
	return out, rows.Err()
}
func (s Service) registerDiscoveryCoverage(mux *http.ServeMux) {
	actionRoute(mux, "renew_discovery_task", func(ctx context.Context, in RenewDiscoveryInput) (any, error) { return s.Store.RenewDiscovery(ctx, in) })
	actionRoute(mux, "save_work_search_identifier", func(ctx context.Context, in SaveSearchIdentifierInput) (any, error) {
		return s.Store.SaveSearchIdentifier(ctx, in)
	})
	mux.HandleFunc("GET /api/discovery-coverage", func(w http.ResponseWriter, r *http.Request) {
		data, err := s.Store.DiscoveryCoverage(r.Context())
		if err != nil {
			http.Error(w, "coverage unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})
}
