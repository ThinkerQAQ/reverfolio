package echo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ImpactEvidence struct {
	MentionID int64  `json:"mention_id"`
	Role      string `json:"role"`
}
type ImpactStep struct {
	Kind      string `json:"kind"`
	Label     string `json:"label"`
	At        string `json:"at"`
	MentionID int64  `json:"mention_id"`
}
type ImpactReview struct {
	Version  int64  `json:"version"`
	Reason   string `json:"reason"`
	Reviewer string `json:"reviewer"`
	At       string `json:"at"`
}
type ImpactEvent struct {
	History        []ImpactReview   `json:"history"`
	ID             int64            `json:"id"`
	Key            string           `json:"event_key"`
	Title          string           `json:"title"`
	Kind           string           `json:"kind"`
	Direction      string           `json:"direction"`
	TargetURL      string           `json:"target_url"`
	State          string           `json:"state"`
	EffectiveState string           `json:"effective_state"`
	Outcome        string           `json:"outcome"`
	Summary        string           `json:"summary"`
	Version        int64            `json:"version"`
	WorkIDs        []int64          `json:"work_ids"`
	Evidence       []ImpactEvidence `json:"evidence"`
	Steps          []ImpactStep     `json:"steps"`
}
type SaveImpactInput struct {
	RequestID       string      `json:"request_id"`
	Event           ImpactEvent `json:"event"`
	ExpectedVersion int64       `json:"expected_version"`
	Reason          string      `json:"reason"`
	Reviewer        string      `json:"reviewer"`
}

func (s *Store) migrateImpact() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= 8 {
		return nil
	}
	_, err = tx.Exec(`CREATE TABLE impact_events(id INTEGER PRIMARY KEY,event_key TEXT NOT NULL UNIQUE,version INTEGER NOT NULL,payload TEXT NOT NULL);
 CREATE TABLE impact_event_works(event_id INTEGER NOT NULL REFERENCES impact_events(id),work_id INTEGER NOT NULL REFERENCES works(id),PRIMARY KEY(event_id,work_id));
 CREATE TABLE impact_event_mentions(event_id INTEGER NOT NULL REFERENCES impact_events(id),mention_id INTEGER NOT NULL REFERENCES mentions(id),role TEXT NOT NULL,PRIMARY KEY(event_id,mention_id));
 CREATE TABLE impact_event_steps(event_id INTEGER NOT NULL REFERENCES impact_events(id),position INTEGER NOT NULL,mention_id INTEGER NOT NULL REFERENCES mentions(id),payload TEXT NOT NULL,PRIMARY KEY(event_id,position));
 CREATE TABLE impact_event_history(event_id INTEGER NOT NULL REFERENCES impact_events(id),version INTEGER NOT NULL,payload TEXT NOT NULL,reason TEXT NOT NULL,reviewer TEXT NOT NULL,changed_at TEXT NOT NULL,PRIMARY KEY(event_id,version));
 INSERT INTO impact_events(event_key,version,payload) SELECT DISTINCT event_key,1,json_object('event_key',event_key,'title',event_key,'kind','unclassified','direction','unknown','state','pending','outcome','not_confirmed','work_ids',json('[]'),'evidence',json('[]'),'steps',json('[]')) FROM mentions WHERE event_key!='';
 INSERT INTO impact_event_works SELECT DISTINCT e.id,m.work_id FROM impact_events e JOIN mentions m ON m.event_key=e.event_key;
 INSERT INTO impact_event_mentions SELECT e.id,m.id,'context' FROM impact_events e JOIN mentions m ON m.event_key=e.event_key;
 PRAGMA user_version=8;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func validateImpact(in SaveImpactInput) error {
	e := in.Event
	if e.ID < 0 || in.ExpectedVersion < 0 || len(strings.TrimSpace(e.Title)) == 0 || len(e.Title) > 240 || len(strings.TrimSpace(e.Key)) == 0 || len(e.Key) > 200 || len(e.Summary) > 4000 || len(in.Reason) == 0 || len(in.Reason) > 1200 || len(in.Reviewer) == 0 || len(in.Reviewer) > 200 || len(e.WorkIDs) > 100 || len(e.Evidence) > 100 || len(e.Steps) > 100 {
		return ErrInvalid
	}
	if !oneOf(e.Kind, "upstream_contribution", "adoption", "collaboration", "citation", "repost", "exposure", "unclassified") || !oneOf(e.Direction, "inbound", "outbound", "unknown") || !oneOf(e.State, "pending", "verified", "rejected") || !oneOf(e.Outcome, "not_confirmed", "in_progress", "achieved") {
		return ErrInvalid
	}
	if e.TargetURL != "" {
		if _, err := NormalizeURL(e.TargetURL); err != nil {
			return err
		}
	}
	seen := map[int64]bool{}
	for _, id := range e.WorkIDs {
		if id <= 0 || seen[id] {
			return ErrInvalid
		}
		seen[id] = true
	}
	mentions := map[int64]bool{}
	for _, m := range e.Evidence {
		if m.MentionID <= 0 || mentions[m.MentionID] || !oneOf(m.Role, "source", "context", "outcome") {
			return ErrInvalid
		}
		mentions[m.MentionID] = true
	}
	for _, step := range e.Steps {
		if !mentions[step.MentionID] || len(step.Label) == 0 || len(step.Label) > 240 || !oneOf(step.Kind, "submitted", "approved", "merged", "released", "retained", "cited", "adopted", "fixed", "validated", "closed", "other") {
			return ErrInvalid
		}
		if step.At != "" {
			if _, err := time.Parse(time.RFC3339, step.At); err != nil {
				return ErrInvalid
			}
		}
	}
	return nil
}
func (s *Store) SaveImpact(ctx context.Context, in SaveImpactInput) (ImpactEvent, error) {
	if err := validateImpact(in); err != nil {
		return ImpactEvent{}, err
	}
	return command(ctx, s, "save_impact_event", in.RequestID, in, func(tx *sql.Tx) (ImpactEvent, error) {
		e := in.Event
		if e.ID == 0 {
			if in.ExpectedVersion != 0 {
				return e, ErrConflict
			}
			var existing int64
			err := tx.QueryRowContext(ctx, "SELECT id FROM impact_events WHERE event_key=?", e.Key).Scan(&existing)
			if err == nil {
				return e, ErrConflict
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return e, err
			}
			res, err := tx.ExecContext(ctx, "INSERT INTO impact_events(event_key,version,payload) VALUES(?,1,'{}')", e.Key)
			if err != nil {
				return e, err
			}
			e.ID, err = res.LastInsertId()
			if err != nil {
				return e, err
			}
			e.Version = 1
		} else {
			var version int64
			var key string
			err := tx.QueryRowContext(ctx, "SELECT version,event_key FROM impact_events WHERE id=?", e.ID).Scan(&version, &key)
			if errors.Is(err, sql.ErrNoRows) {
				return e, ErrNotFound
			}
			if err != nil {
				return e, err
			}
			if version != in.ExpectedVersion || key != e.Key {
				return e, ErrConflict
			}
			e.Version = version + 1
		}
		for _, table := range []string{"impact_event_works", "impact_event_mentions", "impact_event_steps"} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE event_id=?", e.ID); err != nil {
				return e, err
			}
		}
		for _, id := range e.WorkIDs {
			if err := existsWork(ctx, tx, id); err != nil {
				return e, err
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO impact_event_works VALUES(?,?)", e.ID, id); err != nil {
				return e, err
			}
		}
		allVerified := len(e.Evidence) > 0
		outcomeVerified := false
		for _, m := range e.Evidence {
			var state string
			err := tx.QueryRowContext(ctx, "SELECT state FROM mentions WHERE id=?", m.MentionID).Scan(&state)
			if errors.Is(err, sql.ErrNoRows) {
				return e, ErrNotFound
			}
			if err != nil {
				return e, err
			}
			if state != "verified" {
				allVerified = false
			}
			if state == "verified" && m.Role == "outcome" {
				outcomeVerified = true
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO impact_event_mentions VALUES(?,?,?)", e.ID, m.MentionID, m.Role); err != nil {
				return e, err
			}
		}
		if e.State == "verified" && (!allVerified || e.Kind == "unclassified" || e.Direction == "unknown") {
			return e, ErrOriginalSource
		}
		if e.Outcome == "achieved" && (e.State != "verified" || !outcomeVerified) {
			return e, ErrOriginalSource
		}
		for pos, step := range e.Steps {
			payload, err := json.Marshal(step)
			if err != nil {
				return e, err
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO impact_event_steps VALUES(?,?,?,?)", e.ID, pos, step.MentionID, string(payload)); err != nil {
				return e, err
			}
		}
		e.EffectiveState = e.State
		e.History = nil
		payload, err := json.Marshal(e)
		if err != nil {
			return e, err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE impact_events SET version=?,payload=? WHERE id=?", e.Version, string(payload), e.ID); err != nil {
			return e, err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO impact_event_history VALUES(?,?,?,?,?,?)", e.ID, e.Version, string(payload), in.Reason, in.Reviewer, stamp(time.Now()))
		return e, err
	})
}
func (s *Store) ImpactEvents(ctx context.Context, workID int64) ([]ImpactEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT e.id,e.version,e.payload,
 (SELECT COUNT(*) FROM impact_event_mentions em JOIN mentions m ON m.id=em.mention_id WHERE em.event_id=e.id AND m.state!='verified')
 FROM impact_events e WHERE (?=0 OR EXISTS(SELECT 1 FROM impact_event_works w WHERE w.event_id=e.id AND w.work_id=?)) ORDER BY e.id DESC`, workID, workID)
	if err != nil {
		return nil, err
	}
	out := []ImpactEvent{}
	for rows.Next() {
		var e ImpactEvent
		var id, v int64
		var raw string
		var invalid int
		if err = rows.Scan(&id, &v, &raw, &invalid); err != nil {
			rows.Close()
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &e); err != nil {
			rows.Close()
			return nil, err
		}
		e.ID = id
		e.Version = v
		e.EffectiveState = e.State
		if e.State == "verified" && invalid > 0 {
			e.EffectiveState = "needs_review"
		}
		e.WorkIDs = []int64{}
		e.Evidence = []ImpactEvidence{}
		e.Steps = []ImpactStep{}
		out = append(out, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		e := &out[i]
		e.History = []ImpactReview{}
		history, err := s.db.QueryContext(ctx, "SELECT version,reason,reviewer,changed_at FROM impact_event_history WHERE event_id=? ORDER BY version", e.ID)
		if err != nil {
			return nil, err
		}
		for history.Next() {
			var h ImpactReview
			if err = history.Scan(&h.Version, &h.Reason, &h.Reviewer, &h.At); err != nil {
				history.Close()
				return nil, err
			}
			e.History = append(e.History, h)
		}
		err = history.Err()
		history.Close()
		if err != nil {
			return nil, err
		}

		rows, err = s.db.QueryContext(ctx, "SELECT work_id FROM impact_event_works WHERE event_id=? ORDER BY work_id", e.ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id int64
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			e.WorkIDs = append(e.WorkIDs, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		rows, err = s.db.QueryContext(ctx, "SELECT mention_id,role FROM impact_event_mentions WHERE event_id=? ORDER BY mention_id", e.ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var m ImpactEvidence
			if err = rows.Scan(&m.MentionID, &m.Role); err != nil {
				rows.Close()
				return nil, err
			}
			e.Evidence = append(e.Evidence, m)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		rows, err = s.db.QueryContext(ctx, "SELECT payload FROM impact_event_steps WHERE event_id=? ORDER BY position", e.ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var raw string
			var step ImpactStep
			if err = rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			if err = json.Unmarshal([]byte(raw), &step); err != nil {
				rows.Close()
				return nil, err
			}
			e.Steps = append(e.Steps, step)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
func (s Service) registerImpact(mux *http.ServeMux) {
	actionRoute(mux, "save_impact_event", func(ctx context.Context, in SaveImpactInput) (any, error) { return s.Store.SaveImpact(ctx, in) })
	mux.HandleFunc("GET /api/impact-events", func(w http.ResponseWriter, r *http.Request) {
		var id int64
		var err error
		if raw := r.URL.Query().Get("work_id"); raw != "" {
			id, err = strconv.ParseInt(raw, 10, 64)
			if err != nil || id < 0 {
				http.Error(w, "invalid work", 400)
				return
			}
		}
		events, err := s.Store.ImpactEvents(r.Context(), id)
		if err != nil {
			http.Error(w, "events unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	})
}
