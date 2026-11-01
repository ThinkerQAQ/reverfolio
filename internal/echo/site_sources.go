package echo

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

var sourceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,39}$`)
var credentialEnvPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,79}$`)

type SiteSource struct {
	AuthMode        string `json:"auth_mode" toml:"auth_mode"`
	BrowserEndpoint string `json:"browser_endpoint" toml:"browser_endpoint"`
	ID              string `json:"id" toml:"id"`
	Provider        string `json:"provider" toml:"provider"`
	WorkID          int64  `json:"work_id" toml:"work_id"`
	Endpoint        string `json:"endpoint" toml:"endpoint"`
	Property        string `json:"property" toml:"property"`
	Timezone        string `json:"timezone" toml:"timezone"`
	CredentialEnv   string `json:"credential_env" toml:"credential_env"`
	LookbackDays    int    `json:"lookback_days" toml:"lookback_days"`
	LagDays         int    `json:"lag_days" toml:"lag_days"`
	Enabled         bool   `json:"enabled" toml:"enabled"`
}

func (c SiteSource) validate() error {
	if c.AuthMode != "" && !oneOf(c.AuthMode, "token", "browser") {
		return ErrInvalid
	}
	if c.AuthMode == "browser" && (c.Provider != "umami" || c.Endpoint != "https://cloud.umami.is/api" || (c.BrowserEndpoint != "" && !validBrowserEndpoint(c.BrowserEndpoint))) {
		return ErrInvalid
	}
	if !sourceIDPattern.MatchString(c.ID) || !oneOf(c.Provider, "umami", "gsc", "bing") || c.WorkID < 1 || c.LookbackDays < 1 || c.LookbackDays > 90 || c.LagDays < 1 || c.LagDays > 30 || c.Property == "" || len(c.Property) > 1000 {
		return ErrInvalid
	}
	if c.CredentialEnv != "" && !credentialEnvPattern.MatchString(c.CredentialEnv) {
		return ErrInvalid
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return ErrInvalid
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return ErrInvalid
	}
	if c.Provider == "gsc" && (c.Endpoint != "https://www.googleapis.com/webmasters/v3" || c.Timezone != "America/Los_Angeles") {
		return ErrInvalid
	}
	if c.Provider == "bing" && c.Endpoint != "https://ssl.bing.com/webmaster/api.svc/json" {
		return ErrInvalid
	}
	if c.Provider == "umami" && !regexp.MustCompile(`^[A-Za-z0-9-]+$`).MatchString(c.Property) {
		return ErrInvalid
	}
	if c.Provider == "gsc" && !strings.HasPrefix(c.Property, "sc-domain:") {
		if _, err = NormalizeURL(c.Property); err != nil {
			return err
		}
	}
	if c.Provider == "bing" {
		if _, err = NormalizeURL(c.Property); err != nil {
			return err
		}
	}
	return nil
}

type SaveSiteSourceInput struct {
	RequestID       string     `json:"request_id"`
	ExpectedVersion int64      `json:"expected_version"`
	Source          SiteSource `json:"source"`
	Credential      string     `json:"credential,omitempty"` // Write-only, never returned or persisted in config/database.
}
type CollectSiteSourceInput struct {
	RequestID string `json:"request_id"`
	SourceID  string `json:"source_id"`
}
type ProviderMetric struct {
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Timezone    string    `json:"timezone"`
	Scope       string    `json:"scope"`
	SourceDate  string    `json:"source_date,omitempty"`
}
type ProviderRun struct {
	ID         int64            `json:"id"`
	SourceID   string           `json:"source_id"`
	Provider   string           `json:"provider"`
	WorkID     int64            `json:"work_id"`
	RequestID  string           `json:"request_id"`
	ConfigHash string           `json:"config_hash"`
	CapturedAt time.Time        `json:"captured_at"`
	State      string           `json:"state"`
	Problem    string           `json:"problem,omitempty"`
	Metrics    []ProviderMetric `json:"metrics"`
}
type SiteSourceStatus struct {
	Source           SiteSource   `json:"source"`
	CredentialLoaded bool         `json:"credential_loaded"`
	LastRun          *ProviderRun `json:"last_run"`
}
type SiteSourcesStatus struct {
	Version int64              `json:"version"`
	Sources []SiteSourceStatus `json:"sources"`
}
type ProviderClient struct {
	Client  *http.Client
	BaseURL string // Test-only endpoint override; not configurable through API/TOML.
}

func (s Service) credentialPath(id string) string {
	if s.Config == nil || s.Config.Path == "" || !sourceIDPattern.MatchString(id) {
		return ""
	}
	return filepath.Join(filepath.Dir(s.Config.Path), "Data", "provider-credentials", id+".secret")
}

// credentialPresent reports presence without reading the credential file.
func (s Service) credentialPresent(c SiteSource) bool {
	if c.AuthMode == "browser" {
		return false
	}
	if c.CredentialEnv != "" {
		if value, ok := os.LookupEnv(c.CredentialEnv); ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	info, err := os.Stat(s.credentialPath(c.ID))
	return err == nil && info.Mode().IsRegular() && info.Size() > 0 && info.Size() <= 64*1024
}
func (s Service) sourceCredential(c SiteSource) string {
	if c.AuthMode == "browser" {
		return ""
	}
	if c.CredentialEnv != "" {
		if v := strings.TrimSpace(os.Getenv(c.CredentialEnv)); v != "" {
			return v
		}
	}
	p := s.credentialPath(c.ID)
	if p == "" {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil || len(b) > 64*1024 {
		return ""
	}
	return strings.TrimSpace(string(b))
}
func (s Service) SaveSiteSource(ctx context.Context, in SaveSiteSourceInput) (SiteSourcesStatus, error) {
	if s.Config == nil || s.Config.Path == "" || in.RequestID == "" || len(in.RequestID) > 128 || len(in.Credential) > 64*1024 || in.Source.validate() != nil {
		return SiteSourcesStatus{}, ErrInvalid
	}
	w, err := s.Store.Work(ctx, in.Source.WorkID)
	if err != nil {
		return SiteSourcesStatus{}, err
	}
	if w.Kind != "site" && w.Kind != "article" {
		return SiteSourcesStatus{}, ErrInvalid
	}
	err = func() error {
		s.Config.mu.Lock()
		defer s.Config.mu.Unlock()
		c, err := LoadConfig(s.Config.Path)
		if err != nil {
			return err
		}
		if c.Version != in.ExpectedVersion {
			return ErrConflict
		}
		original := c
		original.Sources = append([]SiteSource(nil), c.Sources...)
		index := -1
		for i, v := range c.Sources {
			if v.ID == in.Source.ID {
				index = i
				break
			}
		}
		if index < 0 {
			c.Sources = append(c.Sources, in.Source)
		} else {
			c.Sources[index] = in.Source
		}
		// A key belongs to its provider/endpoint. Retargeting requires a fresh credential.
		if index >= 0 && strings.TrimSpace(in.Credential) == "" {
			old, _ := LoadConfig(s.Config.Path)
			v := old.Sources[index]
			if v.Provider != in.Source.Provider || v.Endpoint != in.Source.Endpoint || v.CredentialEnv != in.Source.CredentialEnv {
				return ErrInvalid
			}
		}
		var secretPath, temp string
		if strings.TrimSpace(in.Credential) != "" {
			secretPath = s.credentialPath(in.Source.ID)
			if err = os.MkdirAll(filepath.Dir(secretPath), 0700); err != nil {
				return err
			}
			f, e := os.CreateTemp(filepath.Dir(secretPath), ".credential-*")
			if e != nil {
				return e
			}
			temp = f.Name()
			defer os.Remove(temp)
			if e = f.Chmod(0600); e == nil {
				_, e = f.WriteString(strings.TrimSpace(in.Credential))
			}
			closeErr := f.Close()
			if e != nil {
				return e
			}
			if closeErr != nil {
				return closeErr
			}
		}
		c.Version++
		if err = saveConfig(s.Config.Path, c); err != nil {
			return err
		}
		if temp != "" {
			if err = os.Rename(temp, secretPath); err != nil {
				if rollbackErr := saveConfig(s.Config.Path, original); rollbackErr != nil {
					return fmt.Errorf("credential save failed; configuration rollback failed")
				}
				return fmt.Errorf("credential save failed; configuration restored")
			}
		}
		return nil
	}()
	if err != nil {
		return SiteSourcesStatus{}, err
	}
	return s.SiteSources(ctx)
}
func (s *Store) ProviderRuns(ctx context.Context) ([]ProviderRun, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,payload FROM provider_runs ORDER BY id DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ProviderRun{}
	for rows.Next() {
		var id int64
		var raw string
		if err = rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		var run ProviderRun
		if err = json.Unmarshal([]byte(raw), &run); err != nil {
			return nil, err
		}
		run.ID = id
		out = append(out, run)
	}
	return out, rows.Err()
}
func (s Service) SiteSources(ctx context.Context) (SiteSourcesStatus, error) {
	c, err := s.Config.Read()
	if err != nil {
		return SiteSourcesStatus{}, err
	}
	runs, err := s.Store.ProviderRuns(ctx)
	if err != nil {
		return SiteSourcesStatus{}, err
	}
	out := SiteSourcesStatus{Version: c.Version, Sources: []SiteSourceStatus{}}
	for _, source := range c.Sources {
		status := SiteSourceStatus{Source: source, CredentialLoaded: s.credentialPresent(source)}
		for _, r := range runs {
			if r.SourceID == source.ID {
				copy := r
				status.LastRun = &copy
				break
			}
		}
		out.Sources = append(out.Sources, status)
	}
	return out, nil
}
func (s Service) CollectSiteSource(ctx context.Context, in CollectSiteSourceInput) (ProviderRun, error) {
	if !sourceIDPattern.MatchString(in.SourceID) || in.RequestID == "" || len(in.RequestID) > 128 {
		return ProviderRun{}, ErrInvalid
	}
	c, err := s.Config.Read()
	if err != nil {
		return ProviderRun{}, err
	}
	var source SiteSource
	found := false
	for _, v := range c.Sources {
		if v.ID == in.SourceID {
			source = v
			found = true
			break
		}
	}
	if !found {
		return ProviderRun{}, ErrNotFound
	}
	raw, _ := json.Marshal(source)
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))
	var previous ProviderRun
	var savedID int64
	var payload string
	err = s.Store.db.QueryRowContext(ctx, `SELECT id,payload FROM provider_runs WHERE source_id=? AND request_id=?`, source.ID, in.RequestID).Scan(&savedID, &payload)
	if err == nil {
		if json.Unmarshal([]byte(payload), &previous) != nil {
			return previous, ErrInvalid
		}
		previous.ID = savedID
		if previous.ConfigHash != hash {
			return previous, ErrConflict
		}
		return previous, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return previous, err
	}
	run := ProviderRun{SourceID: source.ID, Provider: source.Provider, WorkID: source.WorkID, RequestID: in.RequestID, ConfigHash: hash, CapturedAt: time.Now().UTC(), State: "complete", Metrics: []ProviderMetric{}}
	credential := s.sourceCredential(source)
	if credential == "" && source.AuthMode != "browser" {
		run.State = "blocked"
		run.Problem = "credential_missing"
	} else {
		metrics, problem := s.Providers.collect(ctx, source, credential, run.CapturedAt)
		run.Metrics = metrics
		if problem != "" {
			run.State = "failed"
			run.Problem = problem
		} else if len(metrics) == 0 {
			run.State = "empty"
		}
	}
	encoded, _ := json.Marshal(run)
	result, err := s.Store.db.ExecContext(ctx, `INSERT INTO provider_runs(source_id,request_id,config_hash,captured_at,state,payload) VALUES(?,?,?,?,?,?) ON CONFLICT(source_id,request_id) DO NOTHING`, source.ID, in.RequestID, hash, stamp(run.CapturedAt), run.State, string(encoded))
	if err != nil {
		return run, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		var saved string
		if err = s.Store.db.QueryRowContext(ctx, `SELECT id,payload FROM provider_runs WHERE source_id=? AND request_id=?`, source.ID, in.RequestID).Scan(&savedID, &saved); err != nil {
			return run, err
		}
		if err = json.Unmarshal([]byte(saved), &previous); err != nil {
			return run, err
		}
		previous.ID = savedID
		if previous.ConfigHash != hash {
			return run, ErrConflict
		}
		return previous, nil
	}
	run.ID, _ = result.LastInsertId()
	return run, nil
}
func (p ProviderClient) httpClient() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
}
func (p ProviderClient) request(ctx context.Context, method, endpoint string, body any, headers map[string]string, out any) string {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return "request_invalid"
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := p.httpClient().Do(req)
	if err != nil {
		return "request_failed"
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Sprintf("provider_http_%d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024+1))
	if err != nil || len(data) > 4*1024*1024 {
		return "response_too_large"
	}
	if json.Unmarshal(data, out) != nil {
		return "response_invalid"
	}
	return ""
}
func window(c SiteSource, now time.Time) (time.Time, time.Time) {
	loc, _ := time.LoadLocation(c.Timezone)
	local := now.In(loc)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	end := midnight.AddDate(0, 0, 1-c.LagDays)
	return end.AddDate(0, 0, -c.LookbackDays), end
}
func metric(c SiteSource, name string, value float64, unit string, start, end time.Time, scope string) ProviderMetric {
	return ProviderMetric{Metric: name, Value: value, Unit: unit, PeriodStart: start, PeriodEnd: end, Timezone: c.Timezone, Scope: scope}
}
func jsonCount(raw json.RawMessage) (float64, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false
	}
	var n float64
	if json.Unmarshal(raw, &n) == nil {
		return n, n >= 0
	}
	var v struct {
		Value *float64 `json:"value"`
	}
	if json.Unmarshal(raw, &v) != nil || v.Value == nil {
		return 0, false
	}
	return *v.Value, *v.Value >= 0
}
func (p ProviderClient) googleToken(ctx context.Context, secret string) (string, string) {
	if !strings.HasPrefix(strings.TrimSpace(secret), "{") {
		return secret, ""
	}
	var c struct {
		Type         string `json:"type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
		Email        string `json:"client_email"`
		PrivateKey   string `json:"private_key"`
	}
	if json.Unmarshal([]byte(secret), &c) != nil {
		return "", "google_credential_invalid"
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient())
	var source oauth2.TokenSource
	switch c.Type {
	case "authorized_user":
		if c.ClientID == "" || c.ClientSecret == "" || c.RefreshToken == "" {
			return "", "google_credential_invalid"
		}
		cfg := oauth2.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, Endpoint: oauth2.Endpoint{TokenURL: "https://oauth2.googleapis.com/token"}}
		source = cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: c.RefreshToken})
	case "service_account":
		cfg := jwt.Config{Email: c.Email, PrivateKey: []byte(c.PrivateKey), Scopes: []string{"https://www.googleapis.com/auth/webmasters.readonly"}, TokenURL: "https://oauth2.googleapis.com/token"}
		source = cfg.TokenSource(ctx)
	default:
		if c.AccessToken != "" {
			return c.AccessToken, ""
		}
		return "", "google_credential_invalid"
	}
	token, err := source.Token()
	if err != nil {
		return "", "google_token_refresh_failed"
	}
	return token.AccessToken, ""
}

var bingDate = regexp.MustCompile(`^/Date\((-?[0-9]+)([+-][0-9]{4})?\)/$`)

func parseBingDate(value string) (time.Time, bool) {
	m := bingDate.FindStringSubmatch(value)
	if m == nil {
		t, e := time.Parse(time.RFC3339, value)
		return t, e == nil
	}
	ms, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	offset := 0
	if m[2] != "" {
		hours, _ := strconv.Atoi(m[2][1:3])
		minutes, _ := strconv.Atoi(m[2][3:5])
		if hours > 14 || minutes > 59 {
			return time.Time{}, false
		}
		offset = (hours*60 + minutes) * 60
		if m[2][0] == '-' {
			offset = -offset
		}
	}
	return time.UnixMilli(ms).In(time.FixedZone(m[2], offset)), true
}
func (p ProviderClient) collect(ctx context.Context, c SiteSource, secret string, now time.Time) ([]ProviderMetric, string) {
	if c.validate() != nil {
		return []ProviderMetric{}, "source_config_invalid"
	}
	start, end := window(c, now)
	base := strings.TrimRight(c.Endpoint, "/")
	if p.BaseURL != "" {
		base = p.BaseURL
	}
	scope := c.Provider + " property=" + c.Property + "; unfiltered provider aggregate; owner/testing/automation not excluded"
	headers := map[string]string{"Authorization": "Bearer " + secret}
	switch c.Provider {
	case "umami":
		q := url.Values{"startAt": {strconv.FormatInt(start.UnixMilli(), 10)}, "endAt": {strconv.FormatInt(end.UnixMilli()-1, 10)}, "timezone": {c.Timezone}}
		var data map[string]json.RawMessage
		endpoint := base + "/websites/" + url.PathEscape(c.Property) + "/stats?" + q.Encode()
		problem := ""
		if c.AuthMode == "browser" {
			problem = p.browserRequest(ctx, c, endpoint, &data)
		} else {
			problem = p.request(ctx, "GET", endpoint, nil, headers, &data)
		}
		if problem != "" {
			return []ProviderMetric{}, problem
		}
		if c.AuthMode == "browser" {
			scope += "; authenticated browser-session API; requires running signed-in browser"
		}
		out := []ProviderMetric{}
		for key, name := range map[string]string{"pageviews": "pageviews", "visitors": "visitor_identifiers", "visits": "sessions"} {
			v, ok := jsonCount(data[key])
			if !ok {
				return []ProviderMetric{}, "required_metric_missing"
			}
			out = append(out, metric(c, name, v, "count", start, end, scope))
		}
		return out, ""
	case "gsc":
		token, problem := p.googleToken(ctx, secret)
		if problem != "" {
			return []ProviderMetric{}, problem
		}
		headers["Authorization"] = "Bearer " + token
		body := map[string]any{"startDate": start.Format(time.DateOnly), "endDate": end.AddDate(0, 0, -1).Format(time.DateOnly), "type": "web", "dataState": "final", "aggregationType": "byProperty"}
		var data struct {
			Error json.RawMessage `json:"error"`
			Rows  []struct {
				Clicks      *float64 `json:"clicks"`
				Impressions *float64 `json:"impressions"`
				CTR         *float64 `json:"ctr"`
				Position    *float64 `json:"position"`
			} `json:"rows"`
		}
		if problem = p.request(ctx, "POST", base+"/sites/"+url.PathEscape(c.Property)+"/searchAnalytics/query", body, headers, &data); problem != "" {
			return []ProviderMetric{}, problem
		}
		if len(data.Error) > 0 {
			return []ProviderMetric{}, "provider_error_response"
		}
		if len(data.Rows) == 0 {
			return []ProviderMetric{}, ""
		}
		if len(data.Rows) != 1 {
			return []ProviderMetric{}, "aggregate_response_invalid"
		}
		r := data.Rows[0]
		if r.Clicks == nil || r.Impressions == nil || r.CTR == nil || r.Position == nil || *r.Clicks < 0 || *r.Impressions < 0 || *r.CTR < 0 || *r.CTR > 1 || *r.Position < 0 {
			return []ProviderMetric{}, "required_metric_missing"
		}
		scope += "; web search; final data; byProperty; Google PT dates"
		return []ProviderMetric{metric(c, "search_clicks", *r.Clicks, "count", start, end, scope), metric(c, "search_impressions", *r.Impressions, "count", start, end, scope), metric(c, "ctr", *r.CTR, "ratio", start, end, scope), metric(c, "average_position", *r.Position, "position", start, end, scope)}, ""
	case "bing":
		q := url.Values{"siteUrl": {c.Property}, "apikey": {secret}}
		var data struct {
			Rows []struct {
				Clicks      *float64 `json:"Clicks"`
				Impressions *float64 `json:"Impressions"`
				Date        string   `json:"Date"`
			} `json:"d"`
		}
		if problem := p.request(ctx, "GET", base+"/GetRankAndTrafficStats?"+q.Encode(), nil, nil, &data); problem != "" {
			return []ProviderMetric{}, problem
		}
		if data.Rows == nil {
			return []ProviderMetric{}, "response_invalid"
		}
		out := []ProviderMetric{}
		seen := map[string]bool{}
		for _, r := range data.Rows {
			date, ok := parseBingDate(r.Date)
			if !ok || r.Clicks == nil || r.Impressions == nil || *r.Clicks < 0 || *r.Impressions < 0 {
				return []ProviderMetric{}, "required_metric_missing"
			}
			day := date.Format(time.DateOnly)
			if day < start.Format(time.DateOnly) || day >= end.Format(time.DateOnly) {
				continue
			}
			if seen[day] {
				return []ProviderMetric{}, "duplicate_provider_day"
			}
			seen[day] = true
			for name, value := range map[string]float64{"search_clicks": *r.Clicks, "search_impressions": *r.Impressions} {
				m := metric(c, name, value, "count", date, date.AddDate(0, 0, 1), scope+"; all verticals; original provider date offset retained")
				m.Timezone = date.Format("-07:00")
				m.SourceDate = r.Date
				out = append(out, m)
			}
		}
		return out, ""
	}
	return []ProviderMetric{}, "provider_unsupported"
}
