package echo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Store) migrateCatalog() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= 5 {
		return nil
	}
	_, err = tx.Exec(`CREATE TABLE work_monitoring(work_id INTEGER PRIMARY KEY REFERENCES works(id),enabled INTEGER NOT NULL,version INTEGER NOT NULL);
 CREATE TABLE catalog_origins(source_id TEXT NOT NULL,external_id TEXT NOT NULL,work_id INTEGER NOT NULL REFERENCES works(id),producer TEXT NOT NULL,imported_at TEXT NOT NULL,PRIMARY KEY(source_id,external_id));
 PRAGMA user_version=5;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

type CatalogOrigin struct {
	SourceID   string `json:"source_id"`
	ExternalID string `json:"external_id"`
	Producer   string `json:"producer"`
	ImportedAt string `json:"imported_at"`
}
type CatalogEntry struct {
	Work    Work            `json:"work"`
	Enabled bool            `json:"enabled"`
	Version int64           `json:"version"`
	Origins []CatalogOrigin `json:"origins"`
}
type MonitoringInput struct {
	RequestID       string `json:"request_id"`
	WorkID          int64  `json:"work_id"`
	Enabled         bool   `json:"enabled"`
	ExpectedVersion int64  `json:"expected_version"`
}

func monitoring(ctx context.Context, tx *sql.Tx, id int64) (bool, int64, error) {
	var enabled bool
	var version int64
	err := tx.QueryRowContext(ctx, "SELECT enabled,version FROM work_monitoring WHERE work_id=?", id).Scan(&enabled, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return true, 0, nil
	}
	return enabled, version, err
}
func (s *Store) MonitoringEnabled(ctx context.Context, id int64) (bool, error) {
	var enabled bool
	err := s.db.QueryRowContext(ctx, "SELECT enabled FROM work_monitoring WHERE work_id=?", id).Scan(&enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	return enabled, err
}
func (s *Store) SetMonitoring(ctx context.Context, in MonitoringInput) (CatalogEntry, error) {
	return command(ctx, s, "set_monitoring", in.RequestID, in, func(tx *sql.Tx) (CatalogEntry, error) {
		if err := existsWork(ctx, tx, in.WorkID); err != nil {
			return CatalogEntry{}, err
		}
		_, version, err := monitoring(ctx, tx, in.WorkID)
		if err != nil {
			return CatalogEntry{}, err
		}
		if version != in.ExpectedVersion {
			return CatalogEntry{}, ErrConflict
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO work_monitoring VALUES(?,?,?) ON CONFLICT(work_id) DO UPDATE SET enabled=excluded.enabled,version=excluded.version", in.WorkID, in.Enabled, version+1)
		return CatalogEntry{Work: Work{ID: in.WorkID}, Enabled: in.Enabled, Version: version + 1}, err
	})
}
func (s *Store) Catalog(ctx context.Context) ([]CatalogEntry, error) {
	works, err := s.Works(ctx)
	if err != nil {
		return nil, err
	}
	out := []CatalogEntry{}
	for _, w := range works {
		entry := CatalogEntry{Work: w, Enabled: true, Origins: []CatalogOrigin{}}
		err = s.db.QueryRowContext(ctx, "SELECT enabled,version FROM work_monitoring WHERE work_id=?", w.ID).Scan(&entry.Enabled, &entry.Version)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		rows, e := s.db.QueryContext(ctx, "SELECT source_id,external_id,producer,imported_at FROM catalog_origins WHERE work_id=? ORDER BY source_id,external_id", w.ID)
		if e != nil {
			return nil, e
		}
		for rows.Next() {
			var o CatalogOrigin
			if e = rows.Scan(&o.SourceID, &o.ExternalID, &o.Producer, &o.ImportedAt); e != nil {
				rows.Close()
				return nil, e
			}
			entry.Origins = append(entry.Origins, o)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return nil, e
		}
		out = append(out, entry)
	}
	return out, nil
}

type CatalogWork struct {
	ExternalID   string           `json:"external_id"`
	Kind         string           `json:"kind"`
	Title        string           `json:"title"`
	Repository   string           `json:"repository,omitempty"`
	CanonicalURL string           `json:"canonical_url,omitempty"`
	Enabled      bool             `json:"enabled"`
	Channels     []CatalogChannel `json:"channels"`
}
type CatalogChannel struct {
	Platform string `json:"platform"`
	Language string `json:"language"`
	URL      string `json:"url"`
}
type CatalogManifest struct {
	SchemaVersion int           `json:"schema_version"`
	SourceID      string        `json:"source_id"`
	Producer      string        `json:"producer"`
	Works         []CatalogWork `json:"works"`
}
type CatalogImportInput struct {
	RequestID   string          `json:"request_id"`
	DryRun      bool            `json:"dry_run"`
	SelectedIDs []string        `json:"selected_ids"`
	Manifest    CatalogManifest `json:"manifest"`
}
type CatalogImportRow struct {
	ExternalID   string `json:"external_id"`
	WorkID       int64  `json:"work_id,omitempty"`
	Title        string `json:"title"`
	State        string `json:"state"`
	Enabled      bool   `json:"enabled"`
	ChannelCount int    `json:"channel_count"`
}
type CatalogImportResult struct {
	DryRun bool               `json:"dry_run"`
	Rows   []CatalogImportRow `json:"rows"`
}

func normalizeManifest(m CatalogManifest) (CatalogManifest, error) {
	if m.SchemaVersion != 1 || !sourceIDPattern.MatchString(m.SourceID) || strings.TrimSpace(m.Producer) == "" || len(m.Producer) > 80 || len(m.Works) < 1 || len(m.Works) > 500 {
		return m, ErrInvalid
	}
	seen := map[string]bool{}
	for i, w := range m.Works {
		if strings.TrimSpace(w.ExternalID) == "" || len(w.ExternalID) > 200 || seen[w.ExternalID] || len(w.Channels) > 50 {
			return m, ErrInvalid
		}
		seen[w.ExternalID] = true
		in, err := normalizeWorkInput(WorkInput{Kind: w.Kind, Title: w.Title, Repository: w.Repository, CanonicalURL: w.CanonicalURL})
		if err != nil {
			return m, err
		}
		m.Works[i].Kind = in.Kind
		m.Works[i].Title = in.Title
		m.Works[i].Repository = in.Repository
		m.Works[i].CanonicalURL = in.CanonicalURL
		for j, c := range w.Channels {
			v, err := normalizeChannelInput(ChannelInput{Platform: c.Platform, Language: c.Language, URL: c.URL})
			if err != nil {
				return m, err
			}
			m.Works[i].Channels[j] = CatalogChannel{Platform: v.Platform, Language: v.Language, URL: v.URL}
		}
	}
	return m, nil
}
func applyCatalog(ctx context.Context, tx *sql.Tx, in CatalogImportInput) (CatalogImportResult, error) {
	out := CatalogImportResult{DryRun: in.DryRun, Rows: []CatalogImportRow{}}
	selected := map[string]bool{}
	known := map[string]bool{}
	for _, w := range in.Manifest.Works {
		known[w.ExternalID] = true
	}
	for _, id := range in.SelectedIDs {
		if !known[id] || selected[id] {
			return out, ErrInvalid
		}
		selected[id] = true
	}
	if !in.DryRun && len(selected) == 0 {
		return out, ErrInvalid
	}
	for _, item := range in.Manifest.Works {
		if len(selected) > 0 && !selected[item.ExternalID] {
			continue
		}
		var previousID int64
		err := tx.QueryRowContext(ctx, "SELECT work_id FROM catalog_origins WHERE source_id=? AND external_id=?", in.Manifest.SourceID, item.ExternalID).Scan(&previousID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return out, err
		}
		var existingID int64
		if item.Kind == "project" {
			err = tx.QueryRowContext(ctx, "SELECT id FROM works WHERE repository=?", item.Repository).Scan(&existingID)
		} else {
			err = tx.QueryRowContext(ctx, "SELECT id FROM works WHERE canonical_url=?", item.CanonicalURL).Scan(&existingID)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return out, err
		}
		if previousID != 0 && previousID != existingID {
			return out, ErrConflict
		}
		w, err := registerWorkTx(ctx, tx, WorkInput{Kind: item.Kind, Title: item.Title, Repository: item.Repository, CanonicalURL: item.CanonicalURL})
		if err != nil {
			return out, err
		}
		state := "existing"
		if existingID == 0 {
			state = "new"
			if _, err = tx.ExecContext(ctx, "INSERT INTO work_monitoring VALUES(?,?,1)", w.ID, item.Enabled); err != nil {
				return out, err
			}
		}
		for _, channel := range item.Channels {
			if _, err = addChannelTx(ctx, tx, ChannelInput{WorkID: w.ID, Platform: channel.Platform, Language: channel.Language, URL: channel.URL}); err != nil {
				return out, err
			}
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO catalog_origins VALUES(?,?,?,?,?) ON CONFLICT(source_id,external_id) DO NOTHING", in.Manifest.SourceID, item.ExternalID, w.ID, in.Manifest.Producer, stamp(time.Now())); err != nil {
			return out, err
		}
		enabled, _, err := monitoring(ctx, tx, w.ID)
		if err != nil {
			return out, err
		}
		rowID := w.ID
		if in.DryRun && existingID == 0 {
			rowID = 0
		}
		out.Rows = append(out.Rows, CatalogImportRow{ExternalID: item.ExternalID, WorkID: rowID, Title: w.Title, State: state, Enabled: enabled, ChannelCount: len(item.Channels)})
	}
	return out, nil
}
func (s *Store) ImportCatalog(ctx context.Context, in CatalogImportInput) (CatalogImportResult, error) {
	var err error
	in.Manifest, err = normalizeManifest(in.Manifest)
	if err != nil {
		return CatalogImportResult{}, err
	}
	if in.DryRun {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return CatalogImportResult{}, err
		}
		defer tx.Rollback()
		return applyCatalog(ctx, tx, in)
	}
	return command(ctx, s, "import_catalog", in.RequestID, in, func(tx *sql.Tx) (CatalogImportResult, error) { return applyCatalog(ctx, tx, in) })
}
func (s *Store) ExportCatalog(ctx context.Context, sourceID string) (CatalogManifest, error) {
	m := CatalogManifest{SchemaVersion: 1, SourceID: sourceID, Producer: "Reverfolio", Works: []CatalogWork{}}
	if !sourceIDPattern.MatchString(sourceID) {
		return m, ErrInvalid
	}
	entries, err := s.Catalog(ctx)
	if err != nil {
		return m, err
	}
	for _, e := range entries {
		w := e.Work
		identity := w.CanonicalURL
		if w.Kind == "project" {
			identity = strings.ToLower(w.Repository)
		}
		channels, err := s.Channels(ctx, w.ID)
		if err != nil {
			return m, err
		}
		item := CatalogWork{ExternalID: fmt.Sprintf("%x", sha256.Sum256([]byte(w.Kind+":"+identity))), Kind: w.Kind, Title: w.Title, Repository: w.Repository, CanonicalURL: w.CanonicalURL, Enabled: e.Enabled, Channels: []CatalogChannel{}}
		for _, c := range channels {
			item.Channels = append(item.Channels, CatalogChannel{Platform: c.Platform, Language: c.Language, URL: c.URL})
		}
		m.Works = append(m.Works, item)
	}
	return m, nil
}
