# DEV.to official API: creator accounts and article impact

Reverfolio **v0.9.0** supports the documented [Forem API V1](https://developers.forem.com/api/v1) independently of SyndiRelay, Inkerio or browser cookies. The integration makes **GET requests only**; it never publishes, edits, or deletes posts.

## v0.9.1: historical domains and confirmed work identity

A post whose canonical URL exactly matches an existing Work or a registered language/channel URL is linked to the existing Work. A canonical URL on an old website domain with **the same complete article path** as one catalog channel is only a **suggestion**—cross-host path similarity is not proof of ownership. Review each proposed Work in the Creator platforms preview, then select and import.

The `import_devto` MCP tool accepts an optional `work_associations` object, keyed by stringified DEV.to article ID. Values are existing Work IDs, or explicitly `0` to confirm a new independent Work. A path-only cross-domain suggestion **requires a deliberate mapping** to prevent accidental duplicates. An exact existing mapping cannot be reassigned to a different Work. Existing titles, channels and historical observations stay unchanged.

## First use

1. Generate a **personal DEV.to API key** from account settings ([API guide](https://developers.forem.com/api)). The token itself may allow publishing: keep it private even though Reverfolio only uses read endpoints.
2. Open **Creator platforms** in the local Web UI or extension, enter the key into the write-only password field, and connect. `GET /api/users/me` verifies the account identity.
3. Discover your own published articles with `GET /api/articles/me/published` (up to 500 candidates per scan); only explicitly selected articles are registered. A known canonical URL is associated with the existing Work; otherwise the DEV.to published URL is registered separately. Matching titles do not establish identity and drafts are not imported.
4. Collect sourced, timestamped account-level follower counts and publication-level views/reactions/comments. A first cumulative snapshot is a baseline, not a period increase.

| Metric | Official API | Semantics |
|---|---|---|
| Followers | `GET /api/followers/users` | Only report an exact count after complete pagination (up to 30 pages × 100). Follower identities are not persisted. |
| Views | Owned `/api/articles/me/published` `page_views_count` | Cumulative platform page views, not unique people. |
| Views fallback | `GET /api/analytics/totals?article_id=…` | Only if an owner's article list does not supply views; preserve missing fields. |
| Reactions | `public_reactions_count` or `positive_reactions_count` | Platform-defined reactions. |
| Comments | `comments_count` | Platform-defined comments. |
| History | Successive sourced observations | Missing/unavailable/unauthorized is not zero. Follower counts may fall. |

## Privacy and interfaces

- SQLite **schema v7** adds creator account identities, confirmed published articles, collection attempts, and metric history. Existing GitHub, Umami, GSC, Bing and work catalog data are unchanged.
- Keys are saved only in private `Data/platform-credentials/devto_<account_id>.secret` (mode 0600 on POSIX; on Windows the file inherits the installation directory ACL, which should be user-restricted), never echoed to MCP, dashboard, JSON exports, or source control.
- Enrollment intentionally requires the authenticated local UI or browser extension. **MCP does not accept a raw API key.** Five new tools: `list_platform_accounts`, `preview_devto`, `import_devto`, `collect_devto`, `get_platform_dashboard`.
- Local HTTP: `GET /api/platform-dashboard` and `POST /api/actions/{connect_devto,preview_devto,import_devto,collect_devto}`. All routes retain the existing local authorization policy.
- The application never reads authenticated website HTML, bypasses platform access controls, or treats summed follower relationships as unique persons.

## Coverage

**Implemented:** DEV.to/Forem. **Official API exists but not yet integrated/validated for available statistics:** [Hashnode GraphQL](https://docs.hashnode.com/quickstart/introduction). **No new API integration tokens:** [Medium](https://help.medium.com/hc/en-us/articles/213480228-API-Importing). Other platforms remain unsupported until verified, stable, authorized official metrics endpoints are established—no private scraping fallback.

Testing: `go test ./internal/echo -run TestDEVTo -count=1 -v`, with synthetic, isolated Forem API responses and SQLite; no private user account data enters the repository.
