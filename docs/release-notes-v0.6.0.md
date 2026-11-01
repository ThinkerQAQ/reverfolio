# Reverfolio v0.6.0

Adds the shared impact dashboard and optional daily/page detail collection.

- Plot retained GitHub UTC views/clones and daily unique identifiers, cumulative stars/forks/download observations, release/platform asset downloads, and popular repository paths/referrers.
- Show latest collection attempts separately from successful fallback data, including failures and timestamps.
- Export charts as PNG/SVG and exact plotted values as CSV; inspect accessible value tables.
- Collect GSC date/page rows, Bing top-page statistics, and Umami daily series/top paths when explicitly enabled per source.
- Attribute article rows only by exact registered normalized canonical/channel URLs; ambiguous ownership remains unknown.
- Keep timezone, provider aggregation and page coverage limits visible. Window totals never become daily values or verified users.
- Add a first-use guide and read-only HTTP/MCP dashboard data. Fresh installs still contain no preselected works.
- Preserve schema 5 and all existing data/configuration. New detail collection defaults off until enabled.

The optional publisher-side automatic delivery/outbox belongs to SyndiRelay; Reverfolio remains independently usable.

Validation: Go race tests and vet, synthetic provider contracts and exact URL attribution, failure/fallback retention, chart missing/zero/CSV checks, extension parity and documentation checks. Upgrade by preserving Data, config.toml and the extension's local connection file.

See [dashboard usage and metric definitions](impact-dashboard.md).
