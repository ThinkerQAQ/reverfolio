package echo

// Website discovery is deliberately separate from publishing integrations.
// Only explicitly selected articles are registered; HTML pages and sitemap
// paths are candidates, never evidence of publication/reads.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const maxWebCandidates = 100
const maxWebFetches = 12
const maxWebBody = 2 << 20

type WebsiteSource struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	AddedAt       string `json:"added_at"`
	LastScannedAt string `json:"last_scanned_at"`
	LastItemCount int    `json:"last_item_count"`
	ImportedWorks int    `json:"imported_works"`
}

type WebsiteCandidate struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	Format         string `json:"format"`
	TitleSource    string `json:"title_source"`
	ExistingWorkID int64  `json:"existing_work_id,omitempty"`
}

type WebsitePreview struct {
	WebsiteURL string             `json:"website_url"`
	SourceID   string             `json:"source_id"`
	Candidates []WebsiteCandidate `json:"candidates"`
	Checks     []string           `json:"checks"`
	Truncated  bool               `json:"truncated"`
}

type WebsitePreviewInput struct {
	WebsiteURL string `json:"website_url"`
}

type WebsiteImportInput struct {
	RequestID    string   `json:"request_id"`
	WebsiteURL   string   `json:"website_url"`
	SelectedURLs []string `json:"selected_urls"`
}
type WebsiteImportResult struct {
	Source   WebsiteSource `json:"source"`
	Imported int           `json:"imported"`
	Existing int           `json:"existing"`
	Works    []Work        `json:"works"`
}

func (s *Store) migrateWebSources() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= 6 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`CREATE TABLE website_sources(
  id TEXT PRIMARY KEY,url TEXT NOT NULL UNIQUE,added_at TEXT NOT NULL,
  last_scanned_at TEXT NOT NULL,last_item_count INTEGER NOT NULL DEFAULT 0);
  CREATE INDEX website_sources_url_idx ON website_sources(url);
  PRAGMA user_version=6;`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func normalizeWebsiteURL(raw string) (string, error) {
	v, err := NormalizeURL(raw)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(v)
	if err != nil {
		return "", ErrInvalid
	}
	if u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", ErrInvalid
	}
	if u.Port() != "" && u.Port() != "443" && u.Port() != "80" {
		return "", ErrInvalid
	}
	// Restrict ports even if the user supplied an explicit feed URL.
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String(), nil
}

func websiteSourceID(site string) string {
	h := sha256.Sum256([]byte(site))
	return fmt.Sprintf("website_%x", h[:12])
}

var blockedWebsitePrefixes = func() []netip.Prefix {
	// Global-unicast alone still permits documentation, CGNAT and benchmark
	// networks. Never let URL discovery access special-purpose routing space.
	values := []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24",
		"192.0.2.0/24", "192.88.99.0/24", "198.18.0.0/15",
		"198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4",
		"2001::/32", "2001:db8::/32", "2002::/16"}
	out := make([]netip.Prefix, 0, len(values))
	for _, v := range values {
		out = append(out, netip.MustParsePrefix(v))
	}
	return out
}()

func safeWebsiteIP(address net.IP) bool {
	ip, ok := netip.AddrFromSlice(address)
	if !ok {
		return false
	}
	ip = ip.Unmap() // Block IPv4 loopback/private inside IPv6-mapped addresses.
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	for _, prefix := range blockedWebsitePrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func realWebsiteClient() *http.Client {
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, ErrInvalid
			}
			if port != "80" && port != "443" {
				return nil, ErrInvalid
			}
			addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(addresses) == 0 {
				return nil, fmt.Errorf("public hostname DNS unavailable")
			}
			for _, a := range addresses {
				if !safeWebsiteIP(a.IP) {
					return nil, ErrInvalid
				}
			}
			dialer := &net.Dialer{Timeout: 2 * time.Second}
			// Dial verified numeric addresses only. A public IPv6 record should not
			// prevent using a reachable public IPv4 record (or vice versa).
			var dialErr error
			for _, candidate := range addresses {
				conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(candidate.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				dialErr = err
			}
			return nil, dialErr
		},
		DisableKeepAlives:      true,
		MaxResponseHeaderBytes: 16 << 10,
		TLSHandshakeTimeout:    5 * time.Second,
	}
	return &http.Client{Timeout: 5 * time.Second, Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 3 {
				return http.ErrUseLastResponse
			}
			initial := via[0].URL
			if req.URL.Hostname() != initial.Hostname() || req.URL.Scheme != "https" && req.URL.Scheme != "http" ||
				req.URL.Port() != "" && req.URL.Port() != "80" && req.URL.Port() != "443" {
				return http.ErrUseLastResponse
			}
			return nil
		}}
}

type WebsiteDiscoverer struct{ Client *http.Client }

func (d WebsiteDiscoverer) fetch(ctx context.Context, source string) ([]byte, string, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, "", ErrInvalid
	}
	if _, err = normalizeWebsiteURL(source); err != nil {
		return nil, "", err
	}
	client := d.Client
	if client == nil {
		client = realWebsiteClient()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", ErrInvalid
	}
	req.Header.Set("User-Agent", "ReverfolioWebsiteDiscovery/1.0 (+https://github.com/ThinkerQAQ/reverfolio)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, text/html;q=0.9")
	response, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("source temporarily unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, "", fmt.Errorf("source returned HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxWebBody {
		return nil, "", fmt.Errorf("source body too large")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxWebBody+1))
	if err != nil || len(body) > maxWebBody {
		return nil, "", fmt.Errorf("source body too large or unreadable")
	}
	final := response.Request.URL
	if final.Hostname() != u.Hostname() {
		return nil, "", ErrInvalid
	}
	return body, final.String(), nil
}

type webDocument struct {
	XMLName xml.Name `xml:""`
	Channel struct {
		Items []struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
		} `xml:"item"`
	} `xml:"channel"`
	Items []struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
	} `xml:"item"`
	Entries []struct {
		Title string `xml:"title"`
		Links []struct {
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

func webDocumentCandidates(body []byte) (string, []WebsiteCandidate, []string) {
	var doc webDocument
	if xml.Unmarshal(body, &doc) != nil {
		return "", nil, nil
	}
	items := []WebsiteCandidate{}
	nested := []string{}
	switch strings.ToLower(doc.XMLName.Local) {
	case "rss", "rdf":
		rows := doc.Channel.Items
		if len(rows) == 0 {
			rows = doc.Items
		}
		for _, e := range rows {
			items = append(items, WebsiteCandidate{Title: strings.TrimSpace(e.Title), URL: strings.TrimSpace(e.Link), Format: "rss", TitleSource: "feed"})
		}
		return "rss", items, nested
	case "feed":
		for _, e := range doc.Entries {
			for _, l := range e.Links {
				if l.Rel == "" || l.Rel == "alternate" {
					items = append(items, WebsiteCandidate{Title: strings.TrimSpace(e.Title), URL: strings.TrimSpace(l.Href), Format: "atom", TitleSource: "feed"})
					break
				}
			}
		}
		return "atom", items, nested
	case "urlset":
		for _, e := range doc.URLs {
			v := strings.TrimSpace(e.Loc)
			items = append(items, WebsiteCandidate{URL: v, Format: "sitemap", TitleSource: "url_path"})
		}
		return "sitemap", items, nested
	case "sitemapindex":
		for _, e := range doc.Sitemaps {
			nested = append(nested, strings.TrimSpace(e.Loc))
		}
		return "sitemap_index", items, nested
	}
	return "", nil, nil
}

var websiteLinkTag = regexp.MustCompile(`(?is)<link\b[^>]*>`)
var websiteAttr = regexp.MustCompile(`(?is)([a-z_:][-a-z0-9_:]*)\s*=\s*("[^"]*"|'[^']*')`)

func websiteHTMLFeeds(body []byte) []string {
	result := []string{}
	for _, tag := range websiteLinkTag.FindAll(body, 30) {
		attrs := map[string]string{}
		for _, match := range websiteAttr.FindAllSubmatch(tag, -1) {
			attrs[strings.ToLower(string(match[1]))] = string(match[2][1 : len(match[2])-1])
		}
		if !strings.Contains(strings.ToLower(attrs["rel"]), "alternate") {
			continue
		}
		typ := strings.ToLower(attrs["type"])
		if typ != "application/rss+xml" && typ != "application/atom+xml" {
			continue
		}
		result = append(result, attrs["href"])
	}
	return result
}
func candidateFromSitemap(u string) (string, bool) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", false
	}
	if parsed.RawQuery != "" {
		return "", false
	}
	p := strings.Trim(parsed.Path, "/")
	if p == "" {
		return "", false
	}
	lower := strings.ToLower(p)
	if lower == "articles" || lower == "blog" || lower == "posts" {
		return "", false
	}
	for _, seg := range strings.Split(lower, "/") {
		switch seg {
		case "tags", "tag", "categories", "category", "page", "search", "author", "authors", "about", "contact", "projects", "sitemap", "assets", "images", "static", "feed", "rss", "wp-json", "wp-content":
			return "", false
		}
	}
	if strings.HasSuffix(lower, ".xml") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".pdf") {
		return "", false
	}
	label := path.Base(p)
	if label == "index.html" || label == "index" || label == "" {
		return "", false
	}
	// Not a verified article title: show its URL slug and require selection.
	label, _ = url.PathUnescape(label)
	return strings.ReplaceAll(label, "-", " "), true
}

func (d WebsiteDiscoverer) Discover(ctx context.Context, source string) (WebsitePreview, error) {
	website, err := normalizeWebsiteURL(source)
	if err != nil {
		return WebsitePreview{}, err
	}
	base, _ := url.Parse(website)
	out := WebsitePreview{WebsiteURL: website, SourceID: websiteSourceID(website), Candidates: []WebsiteCandidate{}, Checks: []string{}}
	feeds := []string{}
	sitemaps := []string{}
	visited := map[string]bool{}
	fetched := 0
	addURL := func(raw, from string) (string, bool) {
		candidate, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || candidate.Opaque != "" || candidate.User != nil {
			return "", false
		}
		absolute := base.ResolveReference(candidate)
		if absolute.Hostname() != base.Hostname() || (absolute.Port() != "" && absolute.Port() != "443" && absolute.Port() != "80") {
			return "", false
		}
		normalized, e := NormalizeURL(absolute.String())
		if e != nil {
			return "", false
		}
		return normalized, true
	}
	rows := map[string]WebsiteCandidate{}
	addCandidate := func(c WebsiteCandidate) {
		v, ok := addURL(c.URL, website)
		if !ok {
			return
		}
		c.URL = v
		if c.Format == "sitemap" {
			label, valid := candidateFromSitemap(v)
			if !valid {
				return
			}
			c.Title = label
		} else if c.Title == "" {
			return
		}
		if len(c.Title) > 200 {
			// Work titles are bounded by UTF-8 bytes. Do not split a rune.
			end := 0
			for _, r := range c.Title {
				width := utf8.RuneLen(r)
				if width < 1 || end+width > 200 {
					break
				}
				end += width
			}
			c.Title = c.Title[:end]
		}
		if old, ok := rows[v]; ok {
			if old.TitleSource == "feed" {
				return
			}
			if c.TitleSource != "feed" {
				return
			}
		}
		if len(rows) < maxWebCandidates+1 {
			rows[v] = c
		}
	}
	visit := func(raw string) ([]byte, string, bool) {
		u, ok := addURL(raw, website)
		if !ok || visited[u] || fetched >= maxWebFetches {
			return nil, "", false
		}
		visited[u] = true
		fetched++
		body, final, e := d.fetch(ctx, u)
		if e != nil {
			out.Checks = append(out.Checks, "unavailable "+u)
			return nil, "", false
		}
		return body, final, true
	}
	parse := func(body []byte) (string, []string) {
		kind, candidates, sub := webDocumentCandidates(body)
		for _, c := range candidates {
			addCandidate(c)
		}
		return kind, sub
	}
	body, final, ok := visit(website)
	if ok {
		kind, _ := parse(body)
		if kind == "rss" || kind == "atom" {
			feeds = append(feeds, final)
		}
		if kind == "sitemap" || kind == "sitemap_index" {
			sitemaps = append(sitemaps, final)
		}
		feeds = append(feeds, websiteHTMLFeeds(body)...)
	}
	// Prefer declared feeds. Avoid probing four conventional feed locations
	// after we already have a valid RSS/Atom document.
	foundFeed := false
	for _, p := range feeds {
		if len(rows) >= maxWebCandidates || ctx.Err() != nil {
			break
		}
		data, _, ok := visit(p)
		if !ok {
			continue
		}
		kind, _ := parse(data)
		if kind == "rss" || kind == "atom" {
			foundFeed = true
		}
	}
	// A directly provided feed is also considered successful.
	if len(rows) > 0 {
		for _, c := range rows {
			if c.TitleSource == "feed" {
				foundFeed = true
				break
			}
		}
	}
	if !foundFeed {
		for _, p := range []string{"/feed", "/rss.xml", "/feed.xml", "/atom.xml"} {
			if len(rows) >= maxWebCandidates || ctx.Err() != nil {
				break
			}
			data, _, ok := visit(p)
			if !ok {
				continue
			}
			kind, _ := parse(data)
			if kind == "rss" || kind == "atom" {
				foundFeed = true
				break
			}
		}
	}
	// Feed titles are authoritative. A sitemap can supplement short feeds,
	// but must not hold an already useful preview hostage.
	if len(rows) < 40 {
		sitemaps = append(sitemaps, "/sitemap-index.xml", "/sitemap.xml", "/sitemap_index.xml")
		for _, p := range sitemaps {
			if len(rows) >= maxWebCandidates || ctx.Err() != nil {
				break
			}
			data, _, ok := visit(p)
			if !ok {
				continue
			}
			kind, nested := parse(data)
			if kind == "sitemap_index" {
				for i, sub := range nested {
					if i >= 3 || len(rows) >= maxWebCandidates || ctx.Err() != nil {
						break
					}
					child, _, ok := visit(sub)
					if ok {
						parse(child)
					}
				}
			}
			if kind == "sitemap" || kind == "sitemap_index" {
				break
			}
		}
	}
	for _, c := range rows {
		out.Candidates = append(out.Candidates, c)
	}
	sort.Slice(out.Candidates, func(i, j int) bool { return out.Candidates[i].URL < out.Candidates[j].URL })
	if len(out.Candidates) > maxWebCandidates {
		out.Candidates = out.Candidates[:maxWebCandidates]
		out.Truncated = true
	}
	if len(rows) > maxWebCandidates {
		out.Truncated = true
	}
	if ctx.Err() != nil {
		out.Truncated = true
		out.Checks = append(out.Checks, "scan_time_limit_reached")
	}
	return out, nil
}

func (s *Store) WebsiteSources(ctx context.Context) ([]WebsiteSource, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,url,added_at,last_scanned_at,last_item_count,
 (SELECT COUNT(*) FROM catalog_origins WHERE source_id=website_sources.id)
 FROM website_sources ORDER BY added_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []WebsiteSource{}
	for rows.Next() {
		var v WebsiteSource
		if err = rows.Scan(&v.ID, &v.URL, &v.AddedAt, &v.LastScannedAt, &v.LastItemCount, &v.ImportedWorks); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (s Service) PreviewWebsite(ctx context.Context, in WebsitePreviewInput) (WebsitePreview, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 17*time.Second)
	defer cancel()
	preview, err := s.Websites.Discover(probeCtx, in.WebsiteURL)
	if err != nil {
		return preview, err
	}
	for i := range preview.Candidates {
		row := &preview.Candidates[i]
		var id int64
		e := s.Store.db.QueryRowContext(ctx, `SELECT id FROM works WHERE canonical_url=?`, row.URL).Scan(&id)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return WebsitePreview{}, e
		}
		row.ExistingWorkID = id
	}
	return preview, nil
}

func (s Service) ImportWebsite(ctx context.Context, in WebsiteImportInput) (WebsiteImportResult, error) {
	// Refetch and verify the selected URLs. Clients cannot assert unobserved pages.
	if len(in.SelectedURLs) == 0 || len(in.SelectedURLs) > maxWebCandidates {
		return WebsiteImportResult{}, ErrInvalid
	}
	in.WebsiteURL, _ = normalizeWebsiteURL(in.WebsiteURL)
	if in.WebsiteURL == "" {
		return WebsiteImportResult{}, ErrInvalid
	}
	selected := map[string]bool{}
	for _, raw := range in.SelectedURLs {
		v, err := NormalizeURL(raw)
		if err != nil || selected[v] {
			return WebsiteImportResult{}, ErrInvalid
		}
		selected[v] = true
	}
	preview, err := s.PreviewWebsite(ctx, WebsitePreviewInput{WebsiteURL: in.WebsiteURL})
	if err != nil {
		return WebsiteImportResult{}, err
	}
	candidates := []WebsiteCandidate{}
	for _, c := range preview.Candidates {
		if selected[c.URL] {
			candidates = append(candidates, c)
			delete(selected, c.URL)
		}
	}
	if len(selected) != 0 {
		return WebsiteImportResult{}, ErrInvalid
	}
	return command(ctx, s.Store, "import_website", in.RequestID, in, func(tx *sql.Tx) (WebsiteImportResult, error) {
		now := stamp(time.Now())
		result := WebsiteImportResult{Works: []Work{}}
		err := tx.QueryRowContext(ctx, "SELECT id,url,added_at,last_scanned_at,last_item_count FROM website_sources WHERE id=?", preview.SourceID).
			Scan(&result.Source.ID, &result.Source.URL, &result.Source.AddedAt, &result.Source.LastScannedAt, &result.Source.LastItemCount)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return result, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			result.Source = WebsiteSource{ID: preview.SourceID, URL: preview.WebsiteURL, AddedAt: now}
		} else if result.Source.URL != preview.WebsiteURL {
			return result, ErrConflict
		}
		result.Source.LastScannedAt = now
		result.Source.LastItemCount = len(preview.Candidates)
		_, err = tx.ExecContext(ctx, `INSERT INTO website_sources(id,url,added_at,last_scanned_at,last_item_count) VALUES(?,?,?,?,?)
  ON CONFLICT(id) DO UPDATE SET last_scanned_at=excluded.last_scanned_at,last_item_count=excluded.last_item_count`,
			result.Source.ID, result.Source.URL, result.Source.AddedAt, now, len(preview.Candidates))
		if err != nil {
			return result, err
		}
		for _, c := range candidates {
			var existingID int64
			e := tx.QueryRowContext(ctx, "SELECT id FROM works WHERE canonical_url=?", c.URL).Scan(&existingID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return result, e
			}
			w, e := registerWorkTx(ctx, tx, WorkInput{Kind: "article", Title: c.Title, CanonicalURL: c.URL})
			if e != nil {
				return result, e
			}
			if existingID == 0 {
				result.Imported++
			} else {
				result.Existing++
			}
			var boundID int64
			e = tx.QueryRowContext(ctx, "SELECT work_id FROM channel_pages WHERE url=?", c.URL).Scan(&boundID)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return result, e
			}
			if e == nil && boundID != w.ID {
				return result, ErrConflict
			}
			if errors.Is(e, sql.ErrNoRows) {
				host, _ := url.Parse(c.URL)
				if _, e = addChannelTx(ctx, tx, ChannelInput{WorkID: w.ID, Platform: host.Hostname(), Language: "und", URL: c.URL}); e != nil {
					return result, e
				}
			}
			hash := sha256.Sum256([]byte(c.URL))
			externalID := fmt.Sprintf("%x", hash[:])
			var mapped int64
			e = tx.QueryRowContext(ctx, "SELECT work_id FROM catalog_origins WHERE source_id=? AND external_id=?", preview.SourceID, externalID).Scan(&mapped)
			if e != nil && !errors.Is(e, sql.ErrNoRows) {
				return result, e
			}
			if mapped != 0 && mapped != w.ID {
				return result, ErrConflict
			}
			_, e = tx.ExecContext(ctx, "INSERT INTO catalog_origins(source_id,external_id,work_id,producer,imported_at) VALUES(?,?,?,?,?) ON CONFLICT(source_id,external_id) DO NOTHING",
				preview.SourceID, externalID, w.ID, "Website Discovery", now)
			if e != nil {
				return result, e
			}
			result.Works = append(result.Works, w)
		}
		result.Source.ImportedWorks = 0
		if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM catalog_origins WHERE source_id=?", preview.SourceID).Scan(&result.Source.ImportedWorks); err != nil {
			return result, err
		}
		return result, nil
	})
}
