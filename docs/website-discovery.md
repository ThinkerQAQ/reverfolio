# Generic website discovery: RSS / Atom / Sitemap

**Reverfolio v0.8.0** adds website discovery without requiring Inkerio, SyndiRelay, WordPress plugins, or any publisher-side database.

Open **Website discovery** in the native Web UI or Chromium extension, enter a public blog URL or direct RSS/Atom feed URL, and preview detected article candidates. Nothing is written on preview. Select only your own articles and confirm import. The server refetches the source, verifies the selected URLs were actually discovered, and registers or reuses canonical works. Existing titles, channels and monitoring preferences are preserved; import provenance is recorded in the existing catalog alongside the new `website_sources` registry. Previously imported websites can be rescanned manually.

**RSS / Atom** supply titles and article URLs. **Sitemap / Sitemap Index** provide only page URLs: proposed titles are inferred from URL paths, so users must verify which entries really are articles. Tracking query parameters are removed for identity. No automatic guessing of common authorship or cross-platform publication equivalence.

### Safety

- Only public HTTP(S) hosts and ports 80/443. Resolve DNS before dialing, reject private/loopback/link-local addresses (including DNS rebinding), and disallow cross-host redirects and article candidates.
- At most 12 source fetches, 2 MiB per response, 100 candidate rows, 17 seconds per discovery. No arbitrary links, HTML scripts, browser cookies, API tokens, or remote code.
- No false claims of ownership: a public sitemap can contain category pages, profiles and other non-articles; every imported candidate requires user selection.
- **v0.8.0 supports manually triggered discovery/rescan only**. Automatic polling and platform account article discovery are separate future stages. If a site has no exposed RSS/Atom/Sitemap, manually register individual article URLs using the existing work registry.

### Interfaces

| MCP tool | Behavior |
|---|---|
| `list_website_sources` | Read-only list of previously imported website sources |
| `preview_website({website_url})` | Read-only public RSS/Atom/sitemap preview, with existing-work indicators |
| `import_website({request_id,website_url,selected_urls})` | Re-fetch, validate and atomically register selected article URLs (idempotent) |

Local HTTP: `GET /api/website-sources`, `POST /api/actions/preview_website`, `POST /api/actions/import_website`. All endpoints use the existing local browser session or authorized Bearer credentials and the same SQLite authority.

Automated tests cover feed formats, Sitemap Index, privacy guards, an isolated MCP and HTTP integration, deduplication, provenance and schema-v5-to-v6 migration.
