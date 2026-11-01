package echo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

func (s *Store) migrateIntegrations() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var v int
	if err = tx.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		return err
	}
	if v >= 4 {
		return nil
	}
	_, err = tx.Exec(`CREATE TABLE actor_aliases(alias TEXT PRIMARY KEY,identity_url TEXT NOT NULL,reason TEXT NOT NULL,reviewer TEXT NOT NULL,linked_at TEXT NOT NULL);
 CREATE TABLE actor_link_history(id INTEGER PRIMARY KEY,alias TEXT NOT NULL,from_identity TEXT NOT NULL,to_identity TEXT NOT NULL,reason TEXT NOT NULL,reviewer TEXT NOT NULL,changed_at TEXT NOT NULL);
 CREATE TABLE provider_runs(id INTEGER PRIMARY KEY,source_id TEXT NOT NULL,request_id TEXT NOT NULL,config_hash TEXT NOT NULL,captured_at TEXT NOT NULL,state TEXT NOT NULL,payload TEXT NOT NULL,UNIQUE(source_id,request_id));
 PRAGMA user_version=4;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

type ActorAlias struct {
	Alias       string `json:"alias"`
	IdentityURL string `json:"identity_url"`
	Reason      string `json:"reason"`
	Reviewer    string `json:"reviewer"`
	LinkedAt    string `json:"linked_at"`
}
type ActorLinkHistory struct {
	ID           int64  `json:"id"`
	Alias        string `json:"alias"`
	FromIdentity string `json:"from_identity"`
	ToIdentity   string `json:"to_identity"`
	Reason       string `json:"reason"`
	Reviewer     string `json:"reviewer"`
	ChangedAt    string `json:"changed_at"`
}
type IdentityInput struct {
	RequestID   string   `json:"request_id"`
	IdentityURL string   `json:"identity_url"`
	Aliases     []string `json:"aliases"`
	Reason      string   `json:"reason"`
	Reviewer    string   `json:"reviewer"`
}
type UnlinkIdentityInput struct {
	RequestID        string `json:"request_id"`
	Alias            string `json:"alias"`
	ExpectedIdentity string `json:"expected_identity"`
	Reason           string `json:"reason"`
	Reviewer         string `json:"reviewer"`
}

func validIdentityAudit(reason, reviewer string) bool {
	return strings.TrimSpace(reason) != "" && len(reason) <= 1200 && strings.TrimSpace(reviewer) != "" && len(reviewer) <= 80
}
func (s *Store) LinkIdentity(ctx context.Context, in IdentityInput) ([]ActorAlias, error) {
	var err error
	in.IdentityURL, err = NormalizeURL(in.IdentityURL)
	if err != nil || !validIdentityAudit(in.Reason, in.Reviewer) || len(in.Aliases) == 0 || len(in.Aliases) > 50 {
		return nil, ErrInvalid
	}
	seen := map[string]bool{in.IdentityURL: true}
	normalized := []string{in.IdentityURL}
	for _, alias := range in.Aliases {
		u, e := NormalizeURL(alias)
		if e != nil {
			return nil, e
		}
		if !seen[u] {
			seen[u] = true
			normalized = append(normalized, u)
		}
	}
	in.Aliases = normalized
	return command(ctx, s, "link_identity", in.RequestID, in, func(tx *sql.Tx) ([]ActorAlias, error) {
		now := stamp(time.Now())
		out := []ActorAlias{}
		for _, alias := range in.Aliases {
			var exists int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM mentions WHERE actor_key=?`, alias).Scan(&exists); err != nil {
				return nil, err
			}
			if exists == 0 {
				return nil, ErrNotFound
			}
			var previous string
			err := tx.QueryRowContext(ctx, `SELECT identity_url FROM actor_aliases WHERE alias=?`, alias).Scan(&previous)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if previous != "" && previous != in.IdentityURL {
				return nil, ErrConflict
			}
			// A canonical key cannot already be the target of another group being relabelled.
			var targets int
			if alias != in.IdentityURL {
				if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM actor_aliases WHERE identity_url=? AND identity_url<>?`, alias, in.IdentityURL).Scan(&targets); err != nil {
					return nil, err
				}
				if targets > 0 {
					return nil, ErrConflict
				}
			}
			if previous == "" {
				if _, err = tx.ExecContext(ctx, `INSERT INTO actor_aliases VALUES(?,?,?,?,?)`, alias, in.IdentityURL, in.Reason, in.Reviewer, now); err != nil {
					return nil, err
				}
				if _, err = tx.ExecContext(ctx, `INSERT INTO actor_link_history(alias,from_identity,to_identity,reason,reviewer,changed_at) VALUES(?,?,?,?,?,?)`, alias, "", in.IdentityURL, in.Reason, in.Reviewer, now); err != nil {
					return nil, err
				}
			}
			out = append(out, ActorAlias{Alias: alias, IdentityURL: in.IdentityURL, Reason: in.Reason, Reviewer: in.Reviewer, LinkedAt: now})
		}
		return out, nil
	})
}
func (s *Store) UnlinkIdentity(ctx context.Context, in UnlinkIdentityInput) (ActorAlias, error) {
	if !validIdentityAudit(in.Reason, in.Reviewer) {
		return ActorAlias{}, ErrInvalid
	}
	return command(ctx, s, "unlink_identity", in.RequestID, in, func(tx *sql.Tx) (ActorAlias, error) {
		var a ActorAlias
		err := tx.QueryRowContext(ctx, `SELECT alias,identity_url,reason,reviewer,linked_at FROM actor_aliases WHERE alias=?`, in.Alias).Scan(&a.Alias, &a.IdentityURL, &a.Reason, &a.Reviewer, &a.LinkedAt)
		if errors.Is(err, sql.ErrNoRows) {
			return a, ErrNotFound
		}
		if err != nil {
			return a, err
		}
		if a.IdentityURL != in.ExpectedIdentity {
			return a, ErrConflict
		}
		// Canonical self-link is retained until other aliases are detached.
		if a.Alias == a.IdentityURL {
			var count int
			if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM actor_aliases WHERE identity_url=? AND alias<>?`, a.IdentityURL, a.Alias).Scan(&count); err != nil {
				return a, err
			}
			if count > 0 {
				return a, ErrConflict
			}
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM actor_aliases WHERE alias=?`, in.Alias); err != nil {
			return a, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO actor_link_history(alias,from_identity,to_identity,reason,reviewer,changed_at) VALUES(?,?,?,?,?,?)`, in.Alias, a.IdentityURL, "", in.Reason, in.Reviewer, stamp(time.Now()))
		return a, err
	})
}
func (s *Store) IdentityAudit(ctx context.Context) ([]ActorAlias, []ActorLinkHistory, error) {
	aliases := []ActorAlias{}
	history := []ActorLinkHistory{}
	rows, err := s.db.QueryContext(ctx, `SELECT alias,identity_url,reason,reviewer,linked_at FROM actor_aliases ORDER BY identity_url,alias`)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var a ActorAlias
		if err = rows.Scan(&a.Alias, &a.IdentityURL, &a.Reason, &a.Reviewer, &a.LinkedAt); err != nil {
			rows.Close()
			return nil, nil, err
		}
		aliases = append(aliases, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT id,alias,from_identity,to_identity,reason,reviewer,changed_at FROM actor_link_history ORDER BY id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var h ActorLinkHistory
		if err = rows.Scan(&h.ID, &h.Alias, &h.FromIdentity, &h.ToIdentity, &h.Reason, &h.Reviewer, &h.ChangedAt); err != nil {
			return nil, nil, err
		}
		history = append(history, h)
	}
	return aliases, history, rows.Err()
}
