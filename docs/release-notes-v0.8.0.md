# Reverfolio v0.8.0

**Independent RSS / Atom / Sitemap website article discovery**, removing the need for a particular blog engine or publisher as a prerequisite for registering works.

- Add a dedicated **Website discovery** view in the embedded Web UI and Chromium extension, with source URL form, article preview, candidate selection, provenance/status and manual rescan.
- Discover RSS 2.0, Atom and Sitemap/Sitemap Index, preferring real feed titles. Sitemap URL-derived titles remain explicitly unverified until user confirmation.
- Preserve existing work IDs, titles, channels, monitoring preferences and historical observations by normalized canonical URL; register only selected candidates with durable source provenance.
- Add SQLite schema v6 (`website_sources`) with migration from v5 and no rewriting of existing work data.
- Add three shared MCP tools: `list_website_sources`, `preview_website`, `import_website` (34 total), plus corresponding local HTTP endpoints. Browser and MCP use one database.
- Enforce public-only DNS-rebinding-safe HTTP(S), no local network endpoints, bounded fetches, sizes, redirects and candidate counts. Never export cookies, API tokens or fetched content.
- Tests use synthetic sites and an isolated SQLite database; previewed a real independent blog and the user's own public website **without importing or modifying personal data**.

**Not yet supported:** automatic scheduled website polling; official platform account/article discovery; cross-platform follower/read counts. These are subsequent independent milestones, not implied by this release.

Keep private `Data/` and `config.toml` unchanged on upgrade. Windows Native Messaging silent startup and six-platform builds remain intact.
