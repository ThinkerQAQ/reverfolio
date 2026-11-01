package echo

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

// Details are independent of the property summary: a failed drilldown never
// converts a successfully collected property observation into a failure.
func (p ProviderClient) collectDetails(ctx context.Context, c SiteSource, secret string, now time.Time) ([]ProviderMetric, string) {
	if c.Provider == "umami" {
		return p.umamiDetails(ctx, c, secret, now)
	}
	start, end := window(c, now)
	base := c.Endpoint
	if p.BaseURL != "" {
		base = p.BaseURL
	}
	out := []ProviderMetric{}
	if c.Provider == "gsc" {
		token, problem := p.googleToken(ctx, secret)
		if problem != "" {
			return out, problem
		}
		for _, dimensions := range [][]string{{"date"}, {"date", "page"}} {
			body := map[string]any{"startDate": start.Format(time.DateOnly), "endDate": end.AddDate(0, 0, -1).Format(time.DateOnly), "type": "web", "dataState": "final", "dimensions": dimensions, "aggregationType": "auto", "rowLimit": 25000}
			var data struct {
				Error json.RawMessage `json:"error"`
				Rows  []struct {
					Keys        []string `json:"keys"`
					Clicks      *float64 `json:"clicks"`
					Impressions *float64 `json:"impressions"`
				} `json:"rows"`
			}
			problem = p.request(ctx, "POST", base+"/sites/"+url.PathEscape(c.Property)+"/searchAnalytics/query", body, map[string]string{"Authorization": "Bearer " + token}, &data)
			if problem != "" {
				return out, problem
			}
			if len(data.Error) > 0 {
				return out, "provider_error_response"
			}
			for _, r := range data.Rows {
				if len(r.Keys) != len(dimensions) || r.Clicks == nil || r.Impressions == nil || *r.Clicks < 0 || *r.Impressions < 0 {
					return out, "detail_response_invalid"
				}
				date, err := time.ParseInLocation(time.DateOnly, r.Keys[0], start.Location())
				if err != nil || date.Before(start) || !date.Before(end) {
					return out, "detail_date_invalid"
				}
				page := ""
				if len(dimensions) > 1 {
					page = r.Keys[1]
					if _, err = NormalizeURL(page); err != nil {
						return out, "detail_page_invalid"
					}
				}
				scope := "gsc property=" + c.Property + "; web search; final; Google PT dates; dimensions=date"
				if page != "" {
					scope += ",page; top 25000 rows; incomplete page coverage; absent pages are unknown"
				}
				for name, value := range map[string]float64{"search_clicks": *r.Clicks, "search_impressions": *r.Impressions} {
					m := metric(c, name, value, "count", date, date.AddDate(0, 0, 1), scope)
					m.PageURL = page
					out = append(out, m)
				}
			}
		}
	} else if c.Provider == "bing" {
		q := url.Values{"siteUrl": {c.Property}, "apikey": {secret}}
		var data struct {
			Rows []struct {
				Clicks      *float64 `json:"Clicks"`
				Impressions *float64 `json:"Impressions"`
				Date        string   `json:"Date"`
				Query       string   `json:"Query"`
			} `json:"d"`
		}
		if problem := p.request(ctx, "GET", base+"/GetPageStats?"+q.Encode(), nil, nil, &data); problem != "" {
			return out, problem
		}
		if data.Rows == nil {
			return out, "detail_response_invalid"
		}
		for _, r := range data.Rows {
			date, ok := parseBingDate(r.Date)
			if !ok || r.Clicks == nil || r.Impressions == nil || *r.Clicks < 0 || *r.Impressions < 0 {
				return out, "detail_response_invalid"
			}
			if date.Format(time.DateOnly) < start.Format(time.DateOnly) || date.Format(time.DateOnly) >= end.Format(time.DateOnly) {
				continue
			}
			if _, err := NormalizeURL(r.Query); err != nil {
				return out, "detail_page_invalid"
			}
			for name, value := range map[string]float64{"search_clicks": *r.Clicks, "search_impressions": *r.Impressions} {
				m := metric(c, name, value, "count", date, date.AddDate(0, 0, 1), "bing property="+c.Property+"; provider top pages; weekly updates; incomplete page coverage; absent pages are unknown")
				m.PageURL = r.Query
				m.SourceDate = r.Date
				m.Timezone = date.Format("-07:00")
				out = append(out, m)
			}
		}
	}
	return out, ""
}

// Exact registered URL matching only; titles and search queries never establish ownership.
func (s *Store) attributeDetails(ctx context.Context, metrics []ProviderMetric, sourceWorkID int64) error {
	works, err := s.Works(ctx)
	if err != nil {
		return err
	}
	ids := map[string]int64{}
	ambiguous := map[string]bool{}
	add := func(raw string, id int64) {
		n, e := NormalizeURL(raw)
		if e != nil {
			return
		}
		if old, ok := ids[n]; ok && old != id {
			ambiguous[n] = true
		}
		ids[n] = id
	}
	for _, w := range works {
		add(w.CanonicalURL, w.ID)
		channels, e := s.Channels(ctx, w.ID)
		if e != nil {
			return e
		}
		for _, ch := range channels {
			add(ch.URL, w.ID)
		}
	}
	base, err := s.Work(ctx, sourceWorkID)
	if err != nil {
		return err
	}
	root, _ := url.Parse(base.CanonicalURL)
	for i := range metrics {
		if strings.HasPrefix(metrics[i].PageURL, "/") && root != nil && root.Host != "" {
			ref, e := url.Parse(metrics[i].PageURL)
			if e == nil && ref.Host == "" {
				metrics[i].PageURL = root.ResolveReference(ref).String()
			}
		}
		n, e := NormalizeURL(metrics[i].PageURL)
		if e == nil && !ambiguous[n] {
			metrics[i].WorkID = ids[n]
		}
	}
	return nil
}
