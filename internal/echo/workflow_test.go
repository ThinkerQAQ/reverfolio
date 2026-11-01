package echo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type authTransport struct{ token string }

func (a authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header = req.Header.Clone()
	r.Header.Set("Authorization", "Bearer "+a.token)
	return http.DefaultTransport.RoundTrip(r)
}
func callMCP[T any](t *testing.T, session *mcp.ClientSession, name string, input any) T {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: input})
	if err != nil {
		t.Fatal(name, err)
	}
	if res.IsError {
		t.Fatalf("%s: %+v", name, res.Content)
	}
	body, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Data T `json:"data"`
	}
	if err = json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	return result.Data
}
func TestMCPWorkflowAndSharedHTTPState(t *testing.T) {
	store := openTest(t)
	server := httptest.NewServer(Handler(Service{Store: store}, "fixture-token-for-mcp"))
	defer server.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "Reverfolio-integration", Version: "test"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: authTransport{"fixture-token-for-mcp"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("tools not discoverable: %v", err)
	}
	// This inventory represents the published MCP contract. Prior versions
	// registered 31 tools while a stale ChatGPT app snapshot showed only 21.
	// Fail CI if future tool additions/removals are not reviewed here.
	const wantTools = "add_channel,claim_discovery_task,collect_site_source,collect_source,complete_discovery_task,create_discovery_task,export_catalog,generate_report,get_audit,get_collection_status,get_dashboard,get_trends,get_work,import_catalog,import_history,import_website,link_identity,link_propagation_event,list_catalog,list_discovery_tasks,list_mentions,list_settings,list_site_sources,list_website_sources,list_works,preview_website,review_mention,save_site_source,set_monitoring,submit_mention,unlink_identity,update_setting,update_settings,upsert_work"
	gotTools := make([]string, 0, len(tools.Tools))
	for _, tool := range tools.Tools {
		gotTools = append(gotTools, tool.Name)
	}
	sort.Strings(gotTools)
	if strings.Join(gotTools, ",") != wantTools {
		t.Fatalf("MCP tools/list differs from the versioned tool contract: got %v", gotTools)
	}
	work := callMCP[Work](t, session, "upsert_work", WorkInput{RequestID: "article", Kind: "article", Title: "Concurrency sample", CanonicalURL: "https://example.org/article?utm_source=fixture#part"})
	if work.Kind != "article" || work.CanonicalURL != "https://example.org/article" {
		t.Fatalf("bad work %+v", work)
	}
	for _, language := range []string{"en", "zh-CN"} {
		callMCP[Channel](t, session, "add_channel", ChannelInput{RequestID: language, WorkID: work.ID, Platform: "blog", Language: language, URL: "https://example.org/" + language + "/article"})
	}
	task := callMCP[Task](t, session, "create_discovery_task", TaskInput{RequestID: "task", WorkID: work.ID, Query: "exact article URL citations"})
	claim := callMCP[Task](t, session, "claim_discovery_task", ClaimInput{RequestID: "claim", TaskID: task.ID, Worker: "fixture-agent"})
	if claim.LeaseToken == "" {
		t.Fatal("missing lease")
	}
	input := MentionInput{RequestID: "mention", WorkID: work.ID, URL: "https://reader.example.org/post?utm_source=fixture", Type: "citation", Origin: "third_party", Generation: "human", ActorKey: "https://reader.example.org/about", ActorName: "Sample reader", EventKey: "sample-adoption", Quote: "The article provides a useful framework.", SourceKind: "web", ObservedAt: time.Now().UTC()}
	mention := callMCP[Mention](t, session, "submit_mention", input)
	if mention.State != "pending" {
		t.Fatal("AI submission should be pending")
	}
	report := callMCP[ReportOutput](t, session, "generate_report", WorkIDInput{})
	if report.Report.Pending != 1 || report.Report.VerifiedPages != 0 {
		t.Fatal("candidate evidence counted as verified")
	}
	callMCP[Mention](t, session, "review_mention", ReviewInput{RequestID: "review", MentionID: mention.ID, ExpectedVersion: 1, State: "verified", Reason: "Inspected original page and matching source link", Reviewer: "fixture-reviewer"})
	callMCP[Task](t, session, "complete_discovery_task", CompleteInput{RequestID: "complete", TaskID: task.ID, LeaseToken: claim.LeaseToken, Engine: "fixture search", Scope: "one synthetic result page", Status: "complete", ResultCount: 1, Notes: "Synthetic end-to-end fixture, not a real web search."})
	report = callMCP[ReportOutput](t, session, "generate_report", WorkIDInput{})
	if report.Report.VerifiedPages != 1 || report.Report.ThirdPartyEvents != 1 || report.Report.VerifiedActors != 1 {
		t.Fatalf("bad report %+v", report.Report)
	}
	req, _ := http.NewRequest("GET", server.URL+"/api/mentions", nil)
	req.Header.Set("Authorization", "Bearer fixture-token-for-mcp")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var mentions []Mention
	if err = json.NewDecoder(res.Body).Decode(&mentions); err != nil {
		t.Fatal(err)
	}
	if len(mentions) != 1 || mentions[0].State != "verified" {
		t.Fatal("HTTP and MCP are not sharing state")
	}
	unauth, err := http.Post(server.URL+"/mcp", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	unauth.Body.Close()
	if unauth.StatusCode != 401 {
		t.Fatal("MCP is not authenticated")
	}
}

func TestEvidenceAuditGroupingAndVersionConflicts(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	w := mustWork(t, s, "sample/alpha")
	in := MentionInput{RequestID: "old-report", WorkID: w.ID, URL: "https://reader.example.org/post", Type: "citation", Origin: "third_party", Generation: "unknown", ActorKey: "https://reader.example.org/about", ActorName: "Reader", EventKey: "adoption", Quote: "Historical report claim", SourceKind: "secondary_report", ObservedAt: time.Now()}
	first, err := s.SubmitMention(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ReviewMention(ctx, ReviewInput{RequestID: "premature", MentionID: first.ID, ExpectedVersion: 1, State: "verified", Reason: "Historical only", Reviewer: "test"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("secondary evidence was promoted")
	}
	in.RequestID = "original"
	in.Quote = "Inspected original paragraph"
	in.SourceKind = "web"
	first, err = s.SubmitMention(ctx, in)
	if err != nil || len(first.Evidence) != 2 {
		t.Fatalf("original evidence missing: %+v %v", first, err)
	}
	reviewed, err := s.ReviewMention(ctx, ReviewInput{RequestID: "review-a", MentionID: first.ID, ExpectedVersion: 1, State: "verified", Reason: "Original page inspected", Reviewer: "test"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.ReviewMention(ctx, ReviewInput{RequestID: "stale", MentionID: first.ID, ExpectedVersion: 1, State: "rejected", Reason: "Stale state", Reviewer: "test"})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("stale review accepted")
	}
	in.RequestID = "second-page"
	in.URL = "https://reader.example.org/issue/1"
	second, err := s.SubmitMention(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReviewMention(ctx, ReviewInput{RequestID: "review-b", MentionID: second.ID, ExpectedVersion: 1, State: "verified", Reason: "Original issue inspected", Reviewer: "test"}); err != nil {
		t.Fatal(err)
	}
	report, err := s.Report(ctx, 0)
	if err != nil || report.VerifiedPages != 2 || report.ThirdPartyEvents != 1 || report.VerifiedActors != 1 {
		t.Fatalf("double counted %+v %v", report, err)
	}
	if _, err = s.ReviewMention(ctx, ReviewInput{RequestID: "reopen", MentionID: first.ID, ExpectedVersion: reviewed.Version, State: "pending", Reason: "Needs recheck", Reviewer: "test"}); err != nil {
		t.Fatal(err)
	}
	audit, err := s.Audit(ctx)
	if err != nil || len(audit.Reviews) != 3 {
		t.Fatalf("audit missing %+v %v", audit, err)
	}
}

func TestTaskLeaseAndImportAreAtomic(t *testing.T) {
	ctx := context.Background()
	s := openTest(t)
	w := mustWork(t, s, "sample/alpha")
	task, err := s.CreateTask(ctx, TaskInput{"create", w.ID, "Search for exact URL"})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := s.ClaimTask(ctx, ClaimInput{"claim", task.ID, "agent-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimTask(ctx, ClaimInput{"competing", task.ID, "agent-b"}); !errors.Is(err, ErrConflict) {
		t.Fatal("two agents claimed an active task")
	}
	if _, err = s.CompleteTask(ctx, CompleteInput{RequestID: "bad-token", TaskID: task.ID, LeaseToken: "wrong", Engine: "fixture", Scope: "fixture", Status: "complete"}); !errors.Is(err, ErrConflict) {
		t.Fatal("bad lease accepted")
	}
	if _, err = s.db.Exec(`UPDATE discovery_tasks SET lease_until='2000-01-01T00:00:00.000000000Z' WHERE id=?`, task.ID); err != nil {
		t.Fatal(err)
	}
	reclaimed, err := s.ClaimTask(ctx, ClaimInput{"reclaim", task.ID, "agent-b"})
	if err != nil || reclaimed.LeaseToken == lease.LeaseToken {
		t.Fatal("expired task not reclaimed")
	}
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	row := HistoricalObservation{WorkID: w.ID, ExternalID: "history-1", Metric: "search_clicks", Value: 5, Unit: "clicks", PeriodStart: now.Add(-24 * time.Hour), PeriodEnd: now, Timezone: "Asia/Shanghai", Scope: "property", Source: "fixture-report", ObservedAt: now}
	input := ImportInput{RequestID: "preview", SchemaVersion: 1, DryRun: true, Rows: []HistoricalObservation{row}}
	if result, err := s.ImportHistory(ctx, input); err != nil || result.Inserted != 1 {
		t.Fatalf("preview failed %+v %v", result, err)
	}
	audit, _ := s.Audit(ctx)
	if len(audit.History) != 0 {
		t.Fatal("preview mutated database")
	}
	bad := row
	bad.ExternalID = "bad"
	bad.WorkID = 999
	input.DryRun = false
	input.RequestID = "bad-import"
	input.Rows = append(input.Rows, bad)
	if _, err = s.ImportHistory(ctx, input); err == nil {
		t.Fatal("invalid batch accepted")
	}
	audit, _ = s.Audit(ctx)
	if len(audit.History) != 0 {
		t.Fatal("partial import committed")
	}
	input.Rows = []HistoricalObservation{row}
	input.RequestID = "commit"
	if _, err = s.ImportHistory(ctx, input); err != nil {
		t.Fatal(err)
	}
	input.RequestID = "repeat"
	if result, err := s.ImportHistory(ctx, input); err != nil || result.Existing != 1 || result.Inserted != 0 {
		t.Fatalf("duplicate imported %+v %v", result, err)
	}
	input.Rows[0].Value = 999
	input.RequestID = "overwrite"
	if _, err = s.ImportHistory(ctx, input); !errors.Is(err, ErrConflict) {
		t.Fatal("historical observation was silently changed")
	}
}

func TestMigrationPreservesOldSnapshotsAndBackupRestores(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.sqlite")
	db, err := sql.Open("sqlite", old)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE works(id INTEGER PRIMARY KEY,title TEXT NOT NULL,repository TEXT NOT NULL COLLATE NOCASE UNIQUE,github_id INTEGER UNIQUE);
 CREATE TABLE snapshots(id INTEGER PRIMARY KEY,work_id INTEGER NOT NULL REFERENCES works(id),request_key TEXT NOT NULL,captured_at TEXT NOT NULL,payload TEXT NOT NULL,UNIQUE(work_id,request_key));
 CREATE TABLE traffic_days(work_id INTEGER NOT NULL REFERENCES works(id),metric TEXT NOT NULL,day TEXT NOT NULL,count INTEGER NOT NULL,uniques INTEGER NOT NULL,snapshot_id INTEGER NOT NULL REFERENCES snapshots(id),PRIMARY KEY(work_id,metric,day));
 INSERT INTO works VALUES(1,'Old project','sample/alpha',42);
 INSERT INTO snapshots VALUES(1,1,'old','2026-10-20T00:00:00.000000000Z','{"work_id":1,"request_id":"old","captured_at":"2026-10-20T00:00:00Z","releases_available":true}');
 PRAGMA user_version=1;`)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(old)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	work, err := s.Work(ctx, 1)
	if err != nil || work.Kind != "project" || work.RepositoryID == nil || *work.RepositoryID != 42 {
		t.Fatalf("work lost %+v %v", work, err)
	}
	snaps, err := s.Snapshots(ctx, 1)
	if err != nil || len(snaps) != 1 {
		t.Fatal("snapshots lost")
	}
	if _, err = s.RegisterWork(ctx, WorkInput{RequestID: "article", Kind: "article", Title: "Article", CanonicalURL: "https://example.org/article"}); err != nil {
		t.Fatal(err)
	}
	backup, err := s.Backup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backup)
	dest := filepath.Join(dir, "restored.sqlite")
	if err = RestoreDatabase(ctx, backup, dest); err != nil {
		t.Fatal(err)
	}
	restored, err := Open(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	works, err := restored.Works(ctx)
	if err != nil || len(works) != 2 {
		t.Fatal("backup did not restore data")
	}
	if err = RestoreDatabase(ctx, backup, dest); !errors.Is(err, ErrConflict) {
		t.Fatal("restore overwrote existing file")
	}
}

func TestURLSanitizationPrivacy(t *testing.T) {
	for _, raw := range []string{
		"http://127.0.0.1/admin", "http://192.168.1.2/", "http://localhost/",
		"https://private.internal/path", "https://example.org/secret?access_token=abc",
		"https://user:password@example.org/", "file:///etc/passwd",
	} {
		if _, err := NormalizeURL(raw); err == nil {
			t.Fatalf("accepted unsafe URL: %q", raw)
		}
	}
	u, err := NormalizeURL("https://EXAMPLE.org/article/?utm_source=dev&view=full#part")
	if err != nil || u != "https://example.org/article/?view=full" {
		t.Fatalf("normalized %q: %v", u, err)
	}
}
