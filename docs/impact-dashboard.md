# Impact dashboard

The Overview is the shared Web/extension dashboard. Every install starts with its own catalog and local configuration; no author's works or credentials are built into the product.

1. Register a repository, site or article under **Works**, or preview and select a portable catalog import.
2. Add the article's published URLs under **Channels**. Import provenance is visible on each work. Unknown legacy origins remain unknown.
3. Configure sources under **Settings**. Choose the work, provider property or website ID, timezone, lookback and lag. Enable **daily and page details** for drilldown. Credentials are write-only and saved locally; browser mode uses an existing signed-in local browser.
4. Collect, then inspect the latest attempt and successful observation timestamps. A failed attempt does not erase the previous successful observation.
5. Select a repository or article. Export exact plotted data as CSV, a vector SVG, or a 2x PNG.

## Chart semantics

- GitHub views/clones: retained daily UTC buckets and the most recent API window total and unique identifier counts. Missing days are absent, never invented zeros. Daily uniques cannot be summed into cross-day people.
- Stars/forks/release downloads: cumulative observations at collection time, up to the latest 100 snapshots. The 7/14 selector means observations for these series and calendar days relative to the last observed day for traffic.
- Release/platform breakdown: all release assets. Platform categories are filename heuristics. Legacy snapshots without release tags are labelled unknown; collect again to obtain tags. Downloads do not prove installations.
- Popular paths and referrers refer to the repository, not the blog.
- GSC: final web-search data, Google Pacific dates. Summary uses property aggregation. Daily/page detail uses its own query with automatic aggregation; page detail is the top 25,000 rows and may be incomplete.
- Bing: source day offset is retained. Page statistics are top pages updated weekly, not a complete inventory.
- Umami: daily pageviews/sessions and top 500 paths over the entire configured window. Page rankings are window totals, never presented as daily counts.
- Article attribution uses exact normalized registered canonical/channel URLs. Ambiguous URL ownership remains unattributed. Umami relative paths resolve only against the explicitly linked site's canonical URL. A title match is never ownership evidence.
- Sources and timezones stay separate. Missing page rows mean unknown, not zero. Site totals and article rows can use different aggregation rules and should not be added.

The read-only HTTP endpoint is `GET /api/dashboard`; MCP exposes the same data as `get_dashboard`. Existing local UI/API access protection applies. No keys or raw authentication responses are included.
