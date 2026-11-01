package echo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EmptyInput struct{}
type WorkIDInput struct {
	WorkID int64 `json:"work_id"`
}
type CollectInput struct {
	WorkID    int64  `json:"work_id"`
	RequestID string `json:"request_id"`
}
type Overview struct {
	Work      Work       `json:"work"`
	Channels  []Channel  `json:"channels"`
	Snapshots []Snapshot `json:"snapshots"`
	Changes   Interval   `json:"changes"`
}

func (s *Store) Overview(ctx context.Context, id int64) (Overview, error) {
	var v Overview
	var err error
	v.Work, err = s.Work(ctx, id)
	if err != nil {
		return v, err
	}
	v.Channels, err = s.Channels(ctx, id)
	if err != nil {
		return v, err
	}
	v.Snapshots, err = s.Snapshots(ctx, id)
	if err != nil {
		return v, err
	}
	v.Changes = Interval{State: "baseline", Assets: []Change{}}
	valid := []Snapshot{}
	for _, snap := range v.Snapshots {
		if snap.ReleasesOK {
			valid = append(valid, snap)
		}
	}
	if len(valid) > 1 {
		v.Changes = Compare(valid[1], valid[0])
	}
	return v, nil
}

type GroupInput struct {
	RequestID       string `json:"request_id"`
	MentionID       int64  `json:"mention_id"`
	ExpectedVersion int64  `json:"expected_version"`
	EventKey        string `json:"event_key"`
	Reason          string `json:"reason"`
	Reviewer        string `json:"reviewer"`
}

func (s *Store) LinkEvent(ctx context.Context, in GroupInput) (Mention, error) {
	if in.EventKey == "" || len(in.EventKey) > 200 || in.Reason == "" || in.Reviewer == "" {
		return Mention{}, ErrInvalid
	}
	return command(ctx, s, "link_event", in.RequestID, in, func(tx *sql.Tx) (Mention, error) {
		m, err := scanMention(tx.QueryRowContext(ctx, `SELECT `+mentionColumns+` FROM mentions WHERE id=?`, in.MentionID))
		if err != nil {
			return Mention{}, err
		}
		if m.Version != in.ExpectedVersion {
			return Mention{}, ErrConflict
		}
		if _, err = tx.ExecContext(ctx, `UPDATE mentions SET event_key=?,version=version+1 WHERE id=?`, in.EventKey, m.ID); err != nil {
			return Mention{}, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO reviews(mention_id,from_state,to_state,reason,reviewer,reviewed_at,version) VALUES(?,?,?,?,?,?,?)`, m.ID, m.State, m.State, "Group "+m.EventKey+" -> "+in.EventKey+": "+in.Reason, in.Reviewer, stamp(time.Now()), m.Version+1); err != nil {
			return Mention{}, err
		}
		m.EventKey = in.EventKey
		m.Version++
		m.Evidence, err = readEvidence(ctx, tx, m.ID)
		return m, err
	})
}

type ReportOutput struct {
	Report   Report `json:"report"`
	Markdown string `json:"markdown"`
}

func (s *Store) GenerateReport(ctx context.Context, id int64) (ReportOutput, error) {
	r, err := s.Report(ctx, id)
	if err != nil {
		return ReportOutput{}, err
	}
	text := fmt.Sprintf("# Reverfolio impact report\n\nGenerated: %s\n\nVerified pages: %d\n\nVerified third-party propagation groups: %d\n\nVerified identity groups: %d\n\nPending evidence pages: %d\n\nHistorical observations: %d\n\n", r.GeneratedAt, r.VerifiedPages, r.ThirdPartyEvents, r.VerifiedActors, r.Pending, r.HistoricalObservations)
	text += fmt.Sprintf("Identity URLs before explicit grouping: %d\n\n", r.VerifiedActorURLs)
	text += "## GitHub observations\n\n"
	for _, source := range r.Sources {
		text += "- " + source.Repository + ": "
		if source.Latest == nil {
			text += "not collected\n"
			continue
		}
		latest := source.Latest
		text += latest.State + " at " + stamp(latest.CapturedAt) + "\n"
		if latest.Stars != nil {
			text += fmt.Sprintf("  - Stars: %d\n", *latest.Stars)
		}
		if latest.Forks != nil {
			text += fmt.Sprintf("  - Forks: %d\n", *latest.Forks)
		}
		if latest.ReleasesOK {
			annotateSnapshot(latest)
			d := latest.Downloads
			text += fmt.Sprintf("  - Release downloads: %d total; %d packages; %d source; %d auxiliary; %d other/unclassified (filename heuristic %s, not users)\n", d.Total, d.Packages, d.Source, d.Auxiliary, d.Other, d.ClassificationVersion)
			if changes := source.Changes; changes.KnownDownloads != nil {
				text += fmt.Sprintf("  - Observed interval downloads: %d total; %d packages; %d assets with unknown delta; interval=%s to %s (not a daily rate)\n", changes.KnownDownloads.Total, changes.KnownDownloads.Packages, changes.UnknownAssets, stamp(*changes.From), stamp(*changes.To))
			}
		}
		for _, metric := range []string{"views", "clones"} {
			if window, ok := latest.Traffic[metric]; ok {
				text += fmt.Sprintf("  - %s: %d; API window uniques: %d (rolling window, not verified people)\n", metric, window.Count, window.Uniques)
			}
		}
	}
	text += "\n## Site analytics observations\n\n"
	for _, run := range r.ProviderRuns {
		text += fmt.Sprintf("- %s (%s): %s; captured=%s; problem=%s\n", run.SourceID, run.Provider, run.State, stamp(run.CapturedAt), run.Problem)
		for _, m := range run.Metrics {
			text += fmt.Sprintf("  - %s: %g %s; %s to %s; timezone=%s; scope=%s\n", m.Metric, m.Value, m.Unit, stamp(m.PeriodStart), stamp(m.PeriodEnd), m.Timezone, m.Scope)
		}
	}
	text += "\n## Historical observations\n\n"
	for _, row := range r.History {
		text += fmt.Sprintf("- %s: %g %s; scope=%s; source=%s; observed=%s\n", row.Metric, row.Value, row.Unit, row.Scope, row.Source, stamp(row.ObservedAt))
	}
	text += "\n## Verified evidence\n\n"
	for _, m := range r.Mentions {
		if m.State == "verified" {
			text += fmt.Sprintf("- %s | %s | %s\n", m.Type, m.Origin, m.URL)
		}
	}
	for _, n := range r.Notes {
		text += "\n" + n + "\n"
	}
	return ReportOutput{r, text}, nil
}
func publicError(err error) error {
	switch {
	case errors.Is(err, ErrOriginalSource):
		return errors.New("original-page or API evidence is required before verification")
	case errors.Is(err, ErrInvalid):
		return ErrInvalid
	case errors.Is(err, ErrNotFound):
		return ErrNotFound
	case errors.Is(err, ErrConflict):
		return ErrConflict
	case errors.Is(err, context.Canceled):
		return errors.New("request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return errors.New("request timed out")
	default:
		return errors.New("internal operation failed")
	}
}

type ToolOutput struct {
	Data any `json:"data"`
}

func addTool[In any](server *mcp.Server, name, description string, readonly bool, fn func(context.Context, In) (any, error)) {
	additive := name == "review_mention" || name == "link_propagation_event"
	closed := name == "collect_source"
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readonly, DestructiveHint: &additive, OpenWorldHint: &closed, IdempotentHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, ToolOutput, error) {
		value, err := fn(ctx, in)
		if err != nil {
			return nil, ToolOutput{}, publicError(err)
		}
		return nil, ToolOutput{value}, nil
	})
}
func (s Service) MCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "Reverfolio", Version: "0.5.0"}, nil)
	addTool(server, "list_works", "Read registered repositories and articles.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.Store.Works(ctx) })
	addTool(server, "get_work", "Read one work, channel pages, recent snapshots and interval changes.", true, func(ctx context.Context, in WorkIDInput) (any, error) { return s.Store.Overview(ctx, in.WorkID) })
	addTool(server, "upsert_work", "Register a project or article. Requires request_id; duplicate identity returns the existing work without changing its title.", false, func(ctx context.Context, in WorkInput) (any, error) { return s.Store.RegisterWork(ctx, in) })
	addTool(server, "add_channel", "Associate a language or platform URL with a work. Writes locally; requires request_id.", false, func(ctx context.Context, in ChannelInput) (any, error) { return s.Store.AddChannel(ctx, in) })
	addTool(server, "collect_source", "Fetch GitHub metrics for a project and persist a snapshot. Requires request_id; no downloads are fetched.", false, func(ctx context.Context, in CollectInput) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		return s.Collect(ctx, in.WorkID, in.RequestID)
	})
	addTool(server, "get_trends", "Read the latest 100 observations and current interval per-asset changes; no user-count inference.", true, func(ctx context.Context, in WorkIDInput) (any, error) { return s.Store.Overview(ctx, in.WorkID) })
	addTool(server, "submit_mention", "Persist candidate evidence. source_kind web/api declares original-source inspection; secondary_report records historical claims. New evidence remains pending until review.", false, func(ctx context.Context, in MentionInput) (any, error) { return s.Store.SubmitMention(ctx, in) })
	addTool(server, "list_mentions", "Read mentions and evidence. work_id=0 returns all works.", true, func(ctx context.Context, in WorkIDInput) (any, error) { return s.Store.Mentions(ctx, in.WorkID) })
	addTool(server, "review_mention", "Change evidence review state with expected_version, reason and reviewer. Verification requires original-source evidence; this is an explicit audit action.", false, func(ctx context.Context, in ReviewInput) (any, error) { return s.Store.ReviewMention(ctx, in) })
	addTool(server, "link_propagation_event", "Explicitly group related pages for the same work. Requires expected_version and an audit reason.", false, func(ctx context.Context, in GroupInput) (any, error) { return s.Store.LinkEvent(ctx, in) })
	addTool(server, "create_discovery_task", "Write a search task for an AI client. This tool does not execute a search.", false, func(ctx context.Context, in TaskInput) (any, error) { return s.Store.CreateTask(ctx, in) })
	addTool(server, "list_discovery_tasks", "Read search tasks. Lease secrets are omitted.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.Store.Tasks(ctx) })
	addTool(server, "claim_discovery_task", "Claim a pending task or expired lease for ten minutes. Keep returned lease_token private; use a fresh request_id to reclaim.", false, func(ctx context.Context, in ClaimInput) (any, error) { return s.Store.ClaimTask(ctx, in) })
	addTool(server, "complete_discovery_task", "Record actual search engine, scope, results and completion status with a valid lease token.", false, func(ctx context.Context, in CompleteInput) (any, error) { return s.Store.CompleteTask(ctx, in) })
	addTool(server, "import_history", "Validate or atomically import schema_version=1 historical observations. Dry run never writes; conflicts do not overwrite prior observations.", false, func(ctx context.Context, in ImportInput) (any, error) { return s.Store.ImportHistory(ctx, in) })
	addTool(server, "generate_report", "Read evidence-backed counts and a Markdown report. work_id=0 includes all works. No files are written.", true, func(ctx context.Context, in WorkIDInput) (any, error) { return s.Store.GenerateReport(ctx, in.WorkID) })
	addTool(server, "list_settings", "Read nonsecret Reverfolio presentation settings and their versions.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.Config.Settings(ctx) })
	addTool(server, "update_settings", "Atomically save allowlisted nonsecret preferences with expected versions and audit request ID.", false, func(ctx context.Context, in UpdateSettingsInput) (any, error) {
		return s.Config.UpdateSettings(ctx, in)
	})
	addTool(server, "update_setting", "Change an allowlisted nonsecret Reverfolio preference with optimistic version and request ID.", false, func(ctx context.Context, in UpdateSettingInput) (any, error) {
		return s.Config.UpdateSettings(ctx, UpdateSettingsInput{RequestID: in.RequestID, Updates: []SettingChange{{Key: in.Key, Value: in.Value, ExpectedVersion: in.ExpectedVersion}}})
	})
	addTool(server, "list_catalog", "Read works, monitoring preferences and explicit import provenance.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.Store.Catalog(ctx) })
	addTool(server, "set_monitoring", "Enable or pause automatic collection for a work without deleting history. Requires expected_version.", false, func(ctx context.Context, in MonitoringInput) (any, error) { return s.Store.SetMonitoring(ctx, in) })
	addTool(server, "import_catalog", "Preview or atomically import selected works and public publication links. No network fetch or credentials. Applied imports require explicit selected_ids.", false, func(ctx context.Context, in CatalogImportInput) (any, error) { return s.Store.ImportCatalog(ctx, in) })
	addTool(server, "export_catalog", "Export portable works and channel URLs only, with an operator-chosen source_id. No account configuration or credentials.", true, func(ctx context.Context, in struct {
		SourceID string `json:"source_id"`
	}) (any, error) {
		return s.Store.ExportCatalog(ctx, in.SourceID)
	})
	addTool(server, "get_collection_status", "Read configured collection interval and credential presence; actual endpoint permission or collection errors remain in snapshots. No secrets returned.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.RuntimeStatus(ctx) })
	addTool(server, "link_identity", "Explicitly group known actor identity URLs with a reason and reviewer; no name-based inference.", false, func(ctx context.Context, in IdentityInput) (any, error) { return s.Store.LinkIdentity(ctx, in) })
	addTool(server, "unlink_identity", "Detach an identity alias with expected group and audited reason.", false, func(ctx context.Context, in UnlinkIdentityInput) (any, error) { return s.Store.UnlinkIdentity(ctx, in) })
	addTool(server, "list_site_sources", "Read configured independent Umami/GSC/Bing sources, credential presence and last collection state. No secrets.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.SiteSources(ctx) })
	addTool(server, "save_site_source", "Save a site source with expected config version. Optional credential is write-only and stored locally, never returned.", false, func(ctx context.Context, in SaveSiteSourceInput) (any, error) { return s.SaveSiteSource(ctx, in) })
	addTool(server, "collect_site_source", "Fetch one configured site analytics provider and persist an idempotent, scoped observation run. Missing credentials are blocked, not zero.", false, func(ctx context.Context, in CollectSiteSourceInput) (any, error) { return s.CollectSiteSource(ctx, in) })
	addTool(server, "get_audit", "Read historical observations, review transitions and recorded search scopes. Lease secrets are omitted.", true, func(ctx context.Context, _ EmptyInput) (any, error) { return s.Store.Audit(ctx) })
	return server
}

func actionRoute[In any](mux *http.ServeMux, name string, fn func(context.Context, In) (any, error)) {
	mux.HandleFunc("POST /api/actions/"+name, func(w http.ResponseWriter, r *http.Request) {
		var input In
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2*1024*1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(new(any)) != io.EOF {
			http.Error(w, "invalid input", 400)
			return
		}
		output, err := fn(r.Context(), input)
		if err != nil {
			code := 500
			if errors.Is(err, ErrInvalid) {
				code = 400
			}
			if errors.Is(err, ErrNotFound) {
				code = 404
			}
			if errors.Is(err, ErrConflict) {
				code = 409
			}
			http.Error(w, publicError(err).Error(), code)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(output)
	})
}
func (s Service) registerWorkflow(mux *http.ServeMux) {
	actionRoute(mux, "set_monitoring", func(ctx context.Context, in MonitoringInput) (any, error) { return s.Store.SetMonitoring(ctx, in) })
	actionRoute(mux, "import_catalog", func(ctx context.Context, in CatalogImportInput) (any, error) { return s.Store.ImportCatalog(ctx, in) })
	mux.HandleFunc("GET /api/catalog", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.Catalog(r.Context())
		if err != nil {
			http.Error(w, "catalog unavailable or invalid source ID", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("GET /api/catalog/export", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.ExportCatalog(r.Context(), r.URL.Query().Get("source_id"))
		if err != nil {
			http.Error(w, "catalog unavailable or invalid source ID", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(v)
	})

	s.registerOperations(mux)
	actionRoute(mux, "save_site_source", func(ctx context.Context, in SaveSiteSourceInput) (any, error) { return s.SaveSiteSource(ctx, in) })
	actionRoute(mux, "collect_site_source", func(ctx context.Context, in CollectSiteSourceInput) (any, error) { return s.CollectSiteSource(ctx, in) })
	mux.HandleFunc("GET /api/site-sources", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.SiteSources(r.Context())
		if err != nil {
			http.Error(w, "source status unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	actionRoute(mux, "link_identity", func(ctx context.Context, in IdentityInput) (any, error) { return s.Store.LinkIdentity(ctx, in) })
	actionRoute(mux, "unlink_identity", func(ctx context.Context, in UnlinkIdentityInput) (any, error) { return s.Store.UnlinkIdentity(ctx, in) })
	actionRoute(mux, "upsert_work", func(ctx context.Context, in WorkInput) (any, error) { return s.Store.RegisterWork(ctx, in) })
	actionRoute(mux, "add_channel", func(ctx context.Context, in ChannelInput) (any, error) { return s.Store.AddChannel(ctx, in) })
	actionRoute(mux, "submit_mention", func(ctx context.Context, in MentionInput) (any, error) { return s.Store.SubmitMention(ctx, in) })
	actionRoute(mux, "review_mention", func(ctx context.Context, in ReviewInput) (any, error) { return s.Store.ReviewMention(ctx, in) })
	actionRoute(mux, "link_propagation_event", func(ctx context.Context, in GroupInput) (any, error) { return s.Store.LinkEvent(ctx, in) })
	actionRoute(mux, "create_discovery_task", func(ctx context.Context, in TaskInput) (any, error) { return s.Store.CreateTask(ctx, in) })
	actionRoute(mux, "claim_discovery_task", func(ctx context.Context, in ClaimInput) (any, error) { return s.Store.ClaimTask(ctx, in) })
	actionRoute(mux, "complete_discovery_task", func(ctx context.Context, in CompleteInput) (any, error) { return s.Store.CompleteTask(ctx, in) })
	mux.HandleFunc("GET /api/runtime", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.RuntimeStatus(r.Context())
		if err != nil {
			http.Error(w, "invalid local config", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("GET /api/settings", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Config.Settings(r.Context())
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	actionRoute(mux, "update_settings", func(ctx context.Context, in UpdateSettingsInput) (any, error) {
		return s.Config.UpdateSettings(ctx, in)
	})
	actionRoute(mux, "update_setting", func(ctx context.Context, in UpdateSettingInput) (any, error) {
		return s.Config.UpdateSettings(ctx, UpdateSettingsInput{RequestID: in.RequestID, Updates: []SettingChange{{Key: in.Key, Value: in.Value, ExpectedVersion: in.ExpectedVersion}}})
	})
	actionRoute(mux, "import_history", func(ctx context.Context, in ImportInput) (any, error) { return s.Store.ImportHistory(ctx, in) })
	mux.HandleFunc("GET /api/mentions", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.Mentions(r.Context(), 0)
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("GET /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.Tasks(r.Context())
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("GET /api/report", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.Store.GenerateReport(r.Context(), 0)
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("GET /api/works/{id}/channels", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.Error(w, "invalid input", 400)
			return
		}
		v, err := s.Store.Channels(r.Context(), id)
		if err != nil {
			http.Error(w, "internal error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	})
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s.MCPServer() }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}))
}
