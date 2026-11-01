package echo

// DEV.to API-only integration; never imports SyndiRelay cookies or calls
// publishing endpoints. Credential files remain private on the local machine.
import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrPlatformAuth        = errors.New("DEV.to official API authorization unavailable")
	ErrPlatformRateLimited = errors.New("DEV.to official API rate limited")
	ErrPlatformUnavailable = errors.New("DEV.to official API temporarily unavailable")
)

const (
	devtoOrigin           = "https://dev.to"
	foremAccept           = "application/vnd.forem.api-v1+json"
	devtoPageSize         = 100
	devtoMaxPages         = 5
	devtoFollowerMaxPages = 30
)

var devtoNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,60}$`)

type DEVToClient struct {
	Client  *http.Client
	BaseURL string
} // BaseURL test-only; not configurable
func (d DEVToClient) get(ctx context.Context, key, endpoint string, out any) error {
	origin := d.BaseURL
	if origin == "" {
		origin = devtoOrigin
	}
	if !strings.HasPrefix(endpoint, "/api/") || strings.ContainsAny(endpoint, "\r\n#") || len(endpoint) > 512 {
		return ErrInvalid
	}
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(origin, "/")+endpoint, nil)
	if err != nil {
		return ErrInvalid
	}
	req.Header.Set("Accept", foremAccept)
	req.Header.Set("api-key", key)
	req.Header.Set("User-Agent", "Reverfolio/0.9.0 (+https://github.com/ThinkerQAQ/reverfolio)")
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{Proxy: nil, TLSHandshakeTimeout: 5 * time.Second, MaxResponseHeaderBytes: 16 << 10}}
	}
	response, err := client.Do(req)
	if err != nil {
		return ErrPlatformUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		switch response.StatusCode {
		case 401, 403:
			return ErrPlatformAuth
		case 429:
			return ErrPlatformRateLimited
		}
		return fmt.Errorf("DEV.to API returned HTTP %d", response.StatusCode)
	}
	if response.ContentLength > 2<<20 {
		return errors.New("DEV.to API response exceeds limit")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (2<<20)+1))
	if err != nil || len(raw) > 2<<20 {
		return errors.New("DEV.to API response unreadable or exceeds limit")
	}
	if json.Unmarshal(raw, out) != nil {
		return errors.New("DEV.to API response shape unavailable")
	}
	return nil
}

type devtoUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func (d DEVToClient) user(ctx context.Context, key string) (devtoUser, error) {
	var user devtoUser
	err := d.get(ctx, key, "/api/users/me", &user)
	if err != nil {
		return user, err
	}
	if user.ID <= 0 || !devtoNamePattern.MatchString(user.Username) {
		return user, errors.New("DEV.to authenticated user identity unavailable")
	}
	return user, nil
}

type devtoArticle struct {
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	URL               string `json:"url"`
	CanonicalURL      string `json:"canonical_url"`
	PublishedAt       string `json:"published_at"`
	Views             *int64 `json:"page_views_count"`
	Reactions         *int64 `json:"public_reactions_count"`
	PositiveReactions *int64 `json:"positive_reactions_count"`
	Comments          *int64 `json:"comments_count"`
}
type DEVToArticle struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	URL             string `json:"url"`
	CanonicalURL    string `json:"canonical_url,omitempty"`
	ExistingWorkID  int64  `json:"existing_work_id,omitempty"`
	SuggestedWorkID int64  `json:"suggested_work_id,omitempty"`
	ImportedWorkID  int64  `json:"imported_work_id,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
	Views           *int64 `json:"views,omitempty"`
	Reactions       *int64 `json:"reactions,omitempty"`
	Comments        *int64 `json:"comments,omitempty"`
}

func truncateWorkTitle(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 200 {
		return s
	}
	end := 0
	for _, r := range s {
		w := utf8.RuneLen(r)
		if w < 1 || end+w > 200 {
			break
		}
		end += w
	}
	return s[:end]
}
func normalizeDevtoArticle(a devtoArticle) (DEVToArticle, error) {
	var out DEVToArticle
	if a.ID <= 0 || strings.TrimSpace(a.Title) == "" {
		return out, ErrInvalid
	}
	raw, err := NormalizeURL(a.URL)
	if err != nil {
		return out, ErrInvalid
	}
	u, _ := url.Parse(raw)
	if u.Hostname() != "dev.to" || u.Scheme != "https" {
		return out, ErrInvalid
	}
	out = DEVToArticle{ID: a.ID, Title: truncateWorkTitle(a.Title), URL: raw, PublishedAt: a.PublishedAt, Views: a.Views, Comments: a.Comments}
	if a.Reactions != nil {
		out.Reactions = a.Reactions
	} else {
		out.Reactions = a.PositiveReactions
	}
	if a.CanonicalURL != "" {
		if v, e := NormalizeURL(a.CanonicalURL); e == nil {
			out.CanonicalURL = v
		}
	}
	return out, nil
}

type DEVToPreviewInput struct {
	AccountID string `json:"account_id"`
}
type DEVToPreview struct {
	AccountID string         `json:"account_id"`
	Articles  []DEVToArticle `json:"articles"`
	Truncated bool           `json:"truncated"`
}

func (d DEVToClient) articles(ctx context.Context, key string) ([]DEVToArticle, bool, error) {
	result := []DEVToArticle{}
	seen := map[int64]bool{}
	for page := 1; page <= devtoMaxPages; page++ {
		var rows []devtoArticle
		endpoint := fmt.Sprintf("/api/articles/me/published?per_page=%d&page=%d", devtoPageSize, page)
		if err := d.get(ctx, key, endpoint, &rows); err != nil {
			return nil, false, err
		}
		for _, a := range rows {
			v, err := normalizeDevtoArticle(a)
			if err == nil && !seen[v.ID] {
				result = append(result, v)
				seen[v.ID] = true
			}
		}
		if len(rows) < devtoPageSize {
			sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
			return result, false, nil
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, true, nil
}
func (d DEVToClient) followersCount(ctx context.Context, key string) (*int64, error) {
	var count int64
	for page := 1; page <= devtoFollowerMaxPages; page++ {
		var rows []json.RawMessage
		if err := d.get(ctx, key, fmt.Sprintf("/api/followers/users?per_page=%d&page=%d", devtoPageSize, page), &rows); err != nil {
			return nil, err
		}
		count += int64(len(rows))
		if len(rows) < devtoPageSize {
			return &count, nil
		}
	}
	return nil, errors.New("follower list exceeds safe pagination limit; exact total unavailable")
}
func (d DEVToClient) analyticsTotals(ctx context.Context, key string, id int64) (map[string]*int64, error) {
	var raw struct {
		PageViews struct {
			Total *int64 `json:"total"`
		} `json:"page_views"`
		Reactions struct {
			Total *int64 `json:"total"`
		} `json:"reactions"`
		Comments struct {
			Total *int64 `json:"total"`
		} `json:"comments"`
	}
	if err := d.get(ctx, key, "/api/analytics/totals?article_id="+strconv.FormatInt(id, 10), &raw); err != nil {
		return nil, err
	}
	return map[string]*int64{"views": raw.PageViews.Total, "reactions": raw.Reactions.Total, "comments": raw.Comments.Total}, nil
}

type PlatformAccount struct {
	ID                string `json:"id"`
	Platform          string `json:"platform"`
	RemoteID          int64  `json:"remote_id"`
	Username          string `json:"username"`
	ProfileURL        string `json:"profile_url"`
	ConnectedAt       string `json:"connected_at"`
	CredentialPresent bool   `json:"credential_present"`
}
type PlatformPublication struct {
	ID              int64  `json:"id"`
	AccountID       string `json:"account_id"`
	RemoteArticleID int64  `json:"remote_article_id"`
	WorkID          int64  `json:"work_id"`
	URL             string `json:"url"`
	Title           string `json:"title"`
}
type PlatformPoint struct {
	AccountID     string `json:"account_id"`
	PublicationID int64  `json:"publication_id,omitempty"`
	WorkID        int64  `json:"work_id,omitempty"`
	Metric        string `json:"metric"`
	Value         int64  `json:"value"`
	Unit          string `json:"unit"`
	ObservedAt    string `json:"observed_at"`
	Source        string `json:"source"`
}
type PlatformRun struct {
	AccountID   string   `json:"account_id"`
	RequestID   string   `json:"request_id"`
	CapturedAt  string   `json:"captured_at"`
	State       string   `json:"state"`
	Problems    []string `json:"problems"`
	SavedPoints int      `json:"saved_points"`
}
type PlatformPublicationSummary struct {
	Publication PlatformPublication `json:"publication"`
	Latest      []PlatformPoint     `json:"latest"`
	History     []PlatformPoint     `json:"history"`
}
type PlatformAccountSummary struct {
	Account PlatformAccount `json:"account"`
	Latest  []PlatformPoint `json:"latest"`
	History []PlatformPoint `json:"history"`
	LastRun *PlatformRun    `json:"last_run,omitempty"`
}
type PlatformDashboard struct {
	Accounts     []PlatformAccountSummary     `json:"accounts"`
	Publications []PlatformPublicationSummary `json:"publications"`
}

func (s *Store) migratePlatformMetrics() error {
	var v int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return err
	}
	if v >= 7 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE platform_accounts(
 id TEXT PRIMARY KEY,platform TEXT NOT NULL,remote_id INTEGER NOT NULL,
 username TEXT NOT NULL,profile_url TEXT NOT NULL,connected_at TEXT NOT NULL,
 UNIQUE(platform,remote_id));
 CREATE TABLE platform_publications(
 id INTEGER PRIMARY KEY,account_id TEXT NOT NULL REFERENCES platform_accounts(id),
 remote_article_id INTEGER NOT NULL,work_id INTEGER NOT NULL REFERENCES works(id),
 url TEXT NOT NULL UNIQUE,title TEXT NOT NULL, UNIQUE(account_id,remote_article_id));
 CREATE TABLE platform_runs(
 id INTEGER PRIMARY KEY,account_id TEXT NOT NULL REFERENCES platform_accounts(id),
 request_id TEXT NOT NULL UNIQUE,captured_at TEXT NOT NULL,state TEXT NOT NULL,payload TEXT NOT NULL);
 CREATE TABLE platform_points(
 id INTEGER PRIMARY KEY,account_id TEXT NOT NULL REFERENCES platform_accounts(id),
 publication_id INTEGER REFERENCES platform_publications(id),metric TEXT NOT NULL,
 value INTEGER NOT NULL CHECK(value>=0),unit TEXT NOT NULL,observed_at TEXT NOT NULL,
 source TEXT NOT NULL,run_id INTEGER NOT NULL REFERENCES platform_runs(id),
 UNIQUE(run_id,publication_id,metric));
 CREATE INDEX platform_points_account_time ON platform_points(account_id,observed_at DESC);
 PRAGMA user_version=7;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

type ConnectDEVToInput struct {
	RequestID string `json:"request_id"`
	APIKey    string `json:"api_key"`
}
type ImportDEVToInput struct {
	RequestID          string           `json:"request_id"`
	AccountID          string           `json:"account_id"`
	SelectedArticleIDs []int64          `json:"selected_article_ids"`
	WorkAssociations   map[string]int64 `json:"work_associations,omitempty"`
}
type ImportDEVToResult struct {
	AccountID    string                `json:"account_id"`
	Created      int                   `json:"created"`
	Linked       int                   `json:"linked"`
	Publications []PlatformPublication `json:"publications"`
}
type CollectDEVToInput struct {
	RequestID string `json:"request_id"`
	AccountID string `json:"account_id"`
}

var devtoAccountIDPattern = regexp.MustCompile(`^devto_[1-9][0-9]*$`)

func (s Service) platformCredentialPath(id string) string {
	if s.Config == nil || s.Config.Path == "" || !devtoAccountIDPattern.MatchString(id) {
		return ""
	}
	return filepath.Join(filepath.Dir(s.Config.Path), "Data", "platform-credentials", id+".secret")
}
func (s Service) readPlatformCredential(id string) string {
	p := s.platformCredentialPath(id)
	if p == "" {
		return ""
	}
	raw, err := os.ReadFile(p)
	if err != nil || len(raw) > 4096 {
		return ""
	}
	return strings.TrimSpace(string(raw))
}
func (s Service) setPlatformCredential(id, key string) error {
	p := s.platformCredentialPath(id)
	if p == "" || key == "" || len(key) > 4096 {
		return ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(p), ".key-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.WriteString(key + "\n")
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), p)
}
func (s *Store) platformAccount(ctx context.Context, id string) (PlatformAccount, error) {
	var a PlatformAccount
	err := s.db.QueryRowContext(ctx, `SELECT id,platform,remote_id,username,profile_url,connected_at FROM platform_accounts WHERE id=?`, id).
		Scan(&a.ID, &a.Platform, &a.RemoteID, &a.Username, &a.ProfileURL, &a.ConnectedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}
func (s Service) PlatformAccounts(ctx context.Context) ([]PlatformAccount, error) {
	rows, err := s.Store.db.QueryContext(ctx, `SELECT id,platform,remote_id,username,profile_url,connected_at FROM platform_accounts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PlatformAccount{}
	for rows.Next() {
		var a PlatformAccount
		if err = rows.Scan(&a.ID, &a.Platform, &a.RemoteID, &a.Username, &a.ProfileURL, &a.ConnectedAt); err != nil {
			return nil, err
		}
		a.CredentialPresent = s.readPlatformCredential(a.ID) != ""
		out = append(out, a)
	}
	return out, rows.Err()
}
func (s Service) DEVToConnect(ctx context.Context, in ConnectDEVToInput) (PlatformAccount, error) {
	key := strings.TrimSpace(in.APIKey)
	if in.RequestID == "" || len(in.RequestID) > 128 || len(key) < 8 || len(key) > 4096 || s.Config == nil || s.Config.Path == "" {
		return PlatformAccount{}, ErrInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	user, err := s.DEVTo.user(ctx, key)
	if err != nil {
		return PlatformAccount{}, err
	}
	id := "devto_" + strconv.FormatInt(user.ID, 10)
	profile := "https://dev.to/" + user.Username
	return command(ctx, s.Store, "connect_devto", in.RequestID, in, func(tx *sql.Tx) (PlatformAccount, error) {
		a := PlatformAccount{ID: id, Platform: "devto", RemoteID: user.ID, Username: user.Username, ProfileURL: profile, ConnectedAt: stamp(time.Now()), CredentialPresent: true}
		var previous string
		e := tx.QueryRowContext(ctx, "SELECT connected_at FROM platform_accounts WHERE id=?", id).Scan(&previous)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return PlatformAccount{}, e
		}
		if e == nil {
			a.ConnectedAt = previous
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO platform_accounts(id,platform,remote_id,username,profile_url,connected_at) VALUES(?,?,?,?,?,?)
   ON CONFLICT(id) DO UPDATE SET username=excluded.username,profile_url=excluded.profile_url`,
			a.ID, a.Platform, a.RemoteID, a.Username, a.ProfileURL, a.ConnectedAt); e != nil {
			return PlatformAccount{}, e
		}
		if e = s.setPlatformCredential(id, key); e != nil {
			return PlatformAccount{}, e
		}
		return a, nil
	})
}

// suggestDEVToWork is a REVIEW HINT, not proof of publication ownership.
// The caller must explicitly confirm a cross-host association in import_devto.
// Do not match by title, or accept ambiguous matches.
func (s *Store) suggestDEVToWork(ctx context.Context, canonical string) (int64, error) {
	parsed, err := url.Parse(canonical)
	if err != nil || parsed.Path == "" {
		return 0, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT w.id FROM works w
  LEFT JOIN channel_pages c ON c.work_id=w.id
  WHERE w.kind='article' AND (w.canonical_url LIKE ? OR c.url LIKE ?)`, "%"+parsed.Path, "%"+parsed.Path)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var match int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return 0, err
		}
		if match != 0 && match != id {
			return 0, nil
		}
		match = id
	}
	return match, rows.Err()
}

func (s Service) DEVToPreview(ctx context.Context, in DEVToPreviewInput) (DEVToPreview, error) {
	a, err := s.Store.platformAccount(ctx, in.AccountID)
	if err != nil {
		return DEVToPreview{}, err
	}
	key := s.readPlatformCredential(a.ID)
	if key == "" {
		return DEVToPreview{}, ErrPlatformAuth
	}
	ctx, cancel := context.WithTimeout(ctx, 28*time.Second)
	defer cancel()
	u, err := s.DEVTo.user(ctx, key)
	if err != nil {
		return DEVToPreview{}, err
	}
	if u.ID != a.RemoteID {
		return DEVToPreview{}, ErrConflict
	}
	articles, truncated, err := s.DEVTo.articles(ctx, key)
	if err != nil {
		return DEVToPreview{}, err
	}
	for i := range articles {
		item := &articles[i]
		if item.CanonicalURL != "" {
			e := s.Store.db.QueryRowContext(ctx, `SELECT id FROM works WHERE kind='article' AND canonical_url=?`, item.CanonicalURL).Scan(&item.ExistingWorkID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return DEVToPreview{}, e
			}
		}
		if item.ExistingWorkID == 0 {
			e := s.Store.db.QueryRowContext(ctx, `SELECT id FROM works WHERE kind='article' AND canonical_url=?`, item.URL).Scan(&item.ExistingWorkID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return DEVToPreview{}, e
			}
		}
		if item.ExistingWorkID == 0 && item.CanonicalURL != "" {
			e := s.Store.db.QueryRowContext(ctx, `SELECT work_id FROM channel_pages WHERE url=?`, item.CanonicalURL).Scan(&item.ExistingWorkID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return DEVToPreview{}, e
			}
		}
		if item.ExistingWorkID == 0 && item.CanonicalURL != "" {
			id, e := s.Store.suggestDEVToWork(ctx, item.CanonicalURL)
			if e != nil {
				return DEVToPreview{}, e
			}
			item.SuggestedWorkID = id
		}
		e := s.Store.db.QueryRowContext(ctx, `SELECT work_id FROM platform_publications WHERE account_id=? AND remote_article_id=?`, a.ID, item.ID).Scan(&item.ImportedWorkID)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return DEVToPreview{}, e
		}
	}
	return DEVToPreview{AccountID: a.ID, Articles: articles, Truncated: truncated}, nil
}
func (s Service) DEVToImport(ctx context.Context, in ImportDEVToInput) (ImportDEVToResult, error) {
	if in.RequestID == "" || len(in.SelectedArticleIDs) == 0 || len(in.SelectedArticleIDs) > 200 {
		return ImportDEVToResult{}, ErrInvalid
	}
	selected := map[int64]bool{}
	for _, id := range in.SelectedArticleIDs {
		if id <= 0 || selected[id] {
			return ImportDEVToResult{}, ErrInvalid
		}
		selected[id] = true
	}
	preview, err := s.DEVToPreview(ctx, DEVToPreviewInput{AccountID: in.AccountID})
	if err != nil {
		return ImportDEVToResult{}, err
	}
	entries := []DEVToArticle{}
	for _, a := range preview.Articles {
		if selected[a.ID] {
			// Cross-host path equivalence is only a suggestion. A client must
			// explicitly select the candidate Work or acknowledge "new" (0).
			if a.SuggestedWorkID != 0 && a.ExistingWorkID == 0 {
				if _, confirmed := in.WorkAssociations[strconv.FormatInt(a.ID, 10)]; !confirmed {
					return ImportDEVToResult{}, ErrConflict
				}
			}
			entries = append(entries, a)
			delete(selected, a.ID)
		}
	}
	if len(selected) != 0 {
		return ImportDEVToResult{}, ErrInvalid
	}
	for raw, id := range in.WorkAssociations {
		parsed, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || id < 0 {
			return ImportDEVToResult{}, ErrInvalid
		}
		exists := false
		for _, item := range entries {
			if item.ID == parsed {
				exists = true
				if item.ExistingWorkID != 0 && item.ExistingWorkID != id {
					return ImportDEVToResult{}, ErrConflict
				}
				break
			}
		}
		if !exists {
			return ImportDEVToResult{}, ErrInvalid
		}
	}
	return command(ctx, s.Store, "import_devto", in.RequestID, in, func(tx *sql.Tx) (ImportDEVToResult, error) {
		out := ImportDEVToResult{AccountID: in.AccountID, Publications: []PlatformPublication{}}
		for _, article := range entries {
			var current PlatformPublication
			err := tx.QueryRowContext(ctx, `SELECT id,account_id,remote_article_id,work_id,url,title FROM platform_publications WHERE account_id=? AND remote_article_id=?`, in.AccountID, article.ID).
				Scan(&current.ID, &current.AccountID, &current.RemoteArticleID, &current.WorkID, &current.URL, &current.Title)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return out, err
			}
			if err == nil {
				if current.URL != article.URL {
					return out, ErrConflict
				}
				out.Linked++
				out.Publications = append(out.Publications, current)
				continue
			}
			var work Work
			targetID := article.ExistingWorkID
			if override, confirmed := in.WorkAssociations[strconv.FormatInt(article.ID, 10)]; confirmed {
				if article.ExistingWorkID != 0 && override == 0 {
					return out, ErrConflict
				}
				targetID = override
			}
			if targetID != 0 {
				err = tx.QueryRowContext(ctx, `SELECT id,title,COALESCE(repository,''),github_id,kind,COALESCE(canonical_url,'') FROM works WHERE id=?`, targetID).
					Scan(&work.ID, &work.Title, &work.Repository, &work.RepositoryID, &work.Kind, &work.CanonicalURL)
				if err != nil || work.Kind != "article" {
					return out, ErrConflict
				}
				out.Linked++
			} else {
				work, err = registerWorkTx(ctx, tx, WorkInput{Kind: "article", Title: article.Title, CanonicalURL: article.URL})
				if err != nil {
					return out, err
				}
				out.Created++
			}
			var channel Channel
			err = tx.QueryRowContext(ctx, `SELECT id,work_id,platform,language,url FROM channel_pages WHERE url=?`, article.URL).
				Scan(&channel.ID, &channel.WorkID, &channel.Platform, &channel.Language, &channel.URL)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return out, err
			}
			if err == nil {
				if channel.WorkID != work.ID {
					return out, ErrConflict
				}
			} else {
				_, err = addChannelTx(ctx, tx, ChannelInput{WorkID: work.ID, Platform: "devto", Language: "und", URL: article.URL})
				if err != nil {
					return out, err
				}
			}
			result, err := tx.ExecContext(ctx, `INSERT INTO platform_publications(account_id,remote_article_id,work_id,url,title) VALUES(?,?,?,?,?)`, in.AccountID, article.ID, work.ID, article.URL, article.Title)
			if err != nil {
				return out, err
			}
			id, err := result.LastInsertId()
			if err != nil {
				return out, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO catalog_origins(source_id,external_id,work_id,producer,imported_at) VALUES(?,?,?,?,?)
     ON CONFLICT(source_id,external_id) DO NOTHING`, in.AccountID, strconv.FormatInt(article.ID, 10), work.ID, "DEV.to Official API", stamp(time.Now()))
			if err != nil {
				return out, err
			}
			out.Publications = append(out.Publications, PlatformPublication{ID: id, AccountID: in.AccountID, RemoteArticleID: article.ID, WorkID: work.ID, URL: article.URL, Title: article.Title})
		}
		return out, nil
	})
}
func (s *Store) platformPublications(ctx context.Context, id string) ([]PlatformPublication, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,account_id,remote_article_id,work_id,url,title FROM platform_publications WHERE account_id=? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PlatformPublication{}
	for rows.Next() {
		var p PlatformPublication
		if err = rows.Scan(&p.ID, &p.AccountID, &p.RemoteArticleID, &p.WorkID, &p.URL, &p.Title); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s Service) DEVToCollect(ctx context.Context, in CollectDEVToInput) (PlatformRun, error) {
	if in.RequestID == "" || len(in.RequestID) > 128 {
		return PlatformRun{}, ErrInvalid
	}
	account, err := s.Store.platformAccount(ctx, in.AccountID)
	if err != nil {
		return PlatformRun{}, err
	}
	key := s.readPlatformCredential(account.ID)
	if key == "" {
		return PlatformRun{}, ErrPlatformAuth
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	user, err := s.DEVTo.user(ctx, key)
	if err != nil {
		return PlatformRun{}, err
	}
	if user.ID != account.RemoteID {
		return PlatformRun{}, ErrConflict
	}
	articles, truncated, err := s.DEVTo.articles(ctx, key)
	if err != nil {
		return PlatformRun{}, err
	}
	byID := map[int64]DEVToArticle{}
	for _, a := range articles {
		byID[a.ID] = a
	}
	pubs, err := s.Store.platformPublications(ctx, in.AccountID)
	if err != nil {
		return PlatformRun{}, err
	}
	run := PlatformRun{AccountID: in.AccountID, RequestID: in.RequestID, CapturedAt: stamp(time.Now()), State: "complete", Problems: []string{}}
	type sample struct {
		PublicationID int64
		Metric        string
		Value         int64
		Source        string
	}
	samples := []sample{}
	add := func(id int64, metric string, value *int64, source string) {
		if value == nil || *value < 0 {
			return
		}
		samples = append(samples, sample{id, metric, *value, source})
	}
	followers, e := s.DEVTo.followersCount(ctx, key)
	if e != nil {
		run.Problems = append(run.Problems, "followers unavailable: "+e.Error())
	} else {
		add(0, "followers", followers, "/api/followers/users")
	}
	fallbackBudget := 10
	for _, p := range pubs {
		article, ok := byID[p.RemoteArticleID]
		if !ok {
			if !truncated {
				run.Problems = append(run.Problems, "published article unavailable: "+strconv.FormatInt(p.RemoteArticleID, 10))
			}
			continue
		}
		if article.URL != p.URL {
			run.Problems = append(run.Problems, "published article URL changed: "+strconv.FormatInt(p.RemoteArticleID, 10))
			continue
		}
		add(p.ID, "views", article.Views, "/api/articles/me/published")
		add(p.ID, "reactions", article.Reactions, "/api/articles/me/published")
		add(p.ID, "comments", article.Comments, "/api/articles/me/published")
		if article.Views == nil && fallbackBudget > 0 {
			fallbackBudget--
			totals, e := s.DEVTo.analyticsTotals(ctx, key, p.RemoteArticleID)
			if e == nil {
				add(p.ID, "views", totals["views"], "/api/analytics/totals")
				if article.Reactions == nil {
					add(p.ID, "reactions", totals["reactions"], "/api/analytics/totals")
				}
				if article.Comments == nil {
					add(p.ID, "comments", totals["comments"], "/api/analytics/totals")
				}
			} else {
				run.Problems = append(run.Problems, "article analytics unavailable: "+strconv.FormatInt(p.RemoteArticleID, 10))
			}
		}
	}
	if truncated {
		run.Problems = append(run.Problems, "published article list truncated; data coverage partial")
	}
	if len(run.Problems) > 0 {
		run.State = "partial"
	}
	if len(samples) == 0 {
		run.State = "unavailable"
	}
	return command(ctx, s.Store, "collect_devto", in.RequestID, in, func(tx *sql.Tx) (PlatformRun, error) {
		payload, _ := json.Marshal(run)
		res, e := tx.ExecContext(ctx, `INSERT INTO platform_runs(account_id,request_id,captured_at,state,payload) VALUES(?,?,?,?,?)`, in.AccountID, in.RequestID, run.CapturedAt, run.State, string(payload))
		if e != nil {
			return PlatformRun{}, e
		}
		runID, e := res.LastInsertId()
		if e != nil {
			return PlatformRun{}, e
		}
		for _, m := range samples {
			var publication any
			if m.PublicationID != 0 {
				publication = m.PublicationID
			}
			_, e = tx.ExecContext(ctx, `INSERT INTO platform_points(account_id,publication_id,metric,value,unit,observed_at,source,run_id)
    VALUES(?,?,?,?,?,?,?,?)`, in.AccountID, publication, m.Metric, m.Value, "count", run.CapturedAt, m.Source, runID)
			if e != nil {
				return PlatformRun{}, e
			}
		}
		run.SavedPoints = len(samples)
		payload, _ = json.Marshal(run)
		if _, e = tx.ExecContext(ctx, `UPDATE platform_runs SET payload=? WHERE id=?`, string(payload), runID); e != nil {
			return PlatformRun{}, e
		}
		return run, nil
	})
}
func (s Service) PlatformDashboard(ctx context.Context) (PlatformDashboard, error) {
	out := PlatformDashboard{Accounts: []PlatformAccountSummary{}, Publications: []PlatformPublicationSummary{}}
	accounts, err := s.PlatformAccounts(ctx)
	if err != nil {
		return out, err
	}
	for _, account := range accounts {
		summary := PlatformAccountSummary{Account: account, Latest: []PlatformPoint{}, History: []PlatformPoint{}}
		rows, e := s.Store.db.QueryContext(ctx, `SELECT COALESCE(publication_id,0),metric,value,unit,observed_at,source
   FROM platform_points WHERE account_id=? AND publication_id IS NULL ORDER BY id DESC LIMIT 100`, account.ID)
		if e != nil {
			return out, e
		}
		seen := map[string]bool{}
		for rows.Next() {
			var p PlatformPoint
			p.AccountID = account.ID
			if e = rows.Scan(&p.PublicationID, &p.Metric, &p.Value, &p.Unit, &p.ObservedAt, &p.Source); e != nil {
				rows.Close()
				return out, e
			}
			summary.History = append(summary.History, p)
			if !seen[p.Metric] {
				summary.Latest = append(summary.Latest, p)
				seen[p.Metric] = true
			}
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return out, e
		}
		var payload string
		e = s.Store.db.QueryRowContext(ctx, `SELECT payload FROM platform_runs WHERE account_id=? ORDER BY id DESC LIMIT 1`, account.ID).Scan(&payload)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return out, e
		}
		if e == nil {
			var run PlatformRun
			if json.Unmarshal([]byte(payload), &run) == nil {
				summary.LastRun = &run
			}
		}
		out.Accounts = append(out.Accounts, summary)
		pubs, e := s.Store.platformPublications(ctx, account.ID)
		if e != nil {
			return out, e
		}
		for _, pub := range pubs {
			row := PlatformPublicationSummary{Publication: pub, Latest: []PlatformPoint{}, History: []PlatformPoint{}}
			pRows, e := s.Store.db.QueryContext(ctx, `SELECT metric,value,unit,observed_at,source FROM platform_points WHERE publication_id=? ORDER BY id DESC LIMIT 30`, pub.ID)
			if e != nil {
				return out, e
			}
			seen := map[string]bool{}
			for pRows.Next() {
				point := PlatformPoint{AccountID: account.ID, PublicationID: pub.ID, WorkID: pub.WorkID}
				if e = pRows.Scan(&point.Metric, &point.Value, &point.Unit, &point.ObservedAt, &point.Source); e != nil {
					pRows.Close()
					return out, e
				}
				row.History = append(row.History, point)
				if !seen[point.Metric] {
					row.Latest = append(row.Latest, point)
					seen[point.Metric] = true
				}
			}
			e = pRows.Err()
			pRows.Close()
			if e != nil {
				return out, e
			}
			out.Publications = append(out.Publications, row)
		}
	}
	return out, nil
}
