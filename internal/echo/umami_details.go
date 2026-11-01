package echo

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

func (p ProviderClient) umamiDetails(ctx context.Context, c SiteSource, secret string, now time.Time) ([]ProviderMetric, string) {
	start, end := window(c, now)
	base := c.Endpoint
	if p.BaseURL != "" {
		base = p.BaseURL
	}
	q := url.Values{"startAt": {strconv.FormatInt(start.UnixMilli(), 10)}, "endAt": {strconv.FormatInt(end.UnixMilli()-1, 10)}, "timezone": {c.Timezone}, "unit": {"day"}}
	request := func(route string, params url.Values, out any) string {
		endpoint := base + "/websites/" + url.PathEscape(c.Property) + "/" + route + "?" + params.Encode()
		if c.AuthMode == "browser" {
			return p.browserRequest(ctx, c, endpoint, out)
		}
		return p.request(ctx, "GET", endpoint, nil, map[string]string{"Authorization": "Bearer " + secret}, out)
	}
	type point struct {
		X string   `json:"x"`
		Y *float64 `json:"y"`
	}
	var series struct {
		Pageviews []point `json:"pageviews"`
		Sessions  []point `json:"sessions"`
	}
	out := []ProviderMetric{}
	if problem := request("pageviews", q, &series); problem != "" {
		return out, problem
	}
	if series.Pageviews == nil || series.Sessions == nil {
		return out, "detail_response_invalid"
	}
	for name, rows := range map[string][]point{"pageviews": series.Pageviews, "sessions": series.Sessions} {
		for _, row := range rows {
			if len(row.X) < 10 || row.Y == nil || *row.Y < 0 {
				return out, "detail_response_invalid"
			}
			day, err := time.ParseInLocation(time.DateOnly, row.X[:10], start.Location())
			if err != nil || day.Before(start) || !day.Before(end) {
				return out, "detail_date_invalid"
			}
			out = append(out, metric(c, name, *row.Y, "count", day, day.AddDate(0, 0, 1), "umami website="+c.Property+"; daily provider buckets; owner/testing/automation not excluded"))
		}
	}
	q.Del("unit")
	q.Set("type", "path")
	q.Set("limit", "500")
	var pages []point
	if problem := request("metrics", q, &pages); problem != "" {
		return out, problem
	}
	if pages == nil {
		return out, "detail_response_invalid"
	}
	for _, page := range pages {
		if page.Y == nil || *page.Y < 0 || len(page.X) == 0 || page.X[0] != '/' {
			return out, "detail_page_invalid"
		}
		m := metric(c, "pageviews", *page.Y, "count", start, end, "umami website="+c.Property+"; top 500 paths over entire window; absent paths unknown; host resolved from explicitly linked site")
		m.PageURL = page.X
		out = append(out, m)
	}
	return out, ""
}
