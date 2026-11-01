# Reverfolio

[中文](README.zh-CN.md) · [Architecture](docs/architecture.zh-CN.md) · [Roadmap](docs/implementation-plan.md)

Local-first observability for open-source and content impact.

Reverfolio connects projects, multilingual articles, syndication channels, metric history, and reviewed evidence of external citations or adoption. AI clients can search for public evidence and submit it to the local MCP server for explicit review.

## Install and try it

**[Download the latest release](https://github.com/ThinkerQAQ/reverfolio/releases/latest)** · **[English quick start](docs/quickstart.md)** · **[中文快速入门](docs/quickstart.zh-CN.md)** · [Security](SECURITY.md)

- **Windows (x64 / ARM64):** unzip the portable release, run `Install-Reverfolio.ps1`, then `Start-Reverfolio.cmd`, and visit <http://127.0.0.1:19452>. An optional Edge/Chrome extension ships in the Windows archive.
- **Linux / macOS (x64 / ARM64):** extract the correct archive, copy `config.example.toml` to `config.toml`, run `./bin/reverfolio -config ./config.toml`, and visit <http://127.0.0.1:18440>. Browser-extension native-host auto-install is currently Windows-only.
- **No publishing dependencies:** new users can add a website through RSS/Atom/Sitemap, register GitHub repositories, then optionally authorize their own provider APIs. Personal data and API keys remain in local `Data/`; **there is no public ChatGPT plugin one-click installer**. The source is [MIT-licensed](LICENSE), with original [third-party license texts](THIRD_PARTY_NOTICES.md) in the release archives.

**Available now:** GitHub repository traffic/release-download observations; Umami, GSC, and Bing site metrics when authorized; independent website article discovery; direct DEV.to account/article API metrics; reviewed public citations; sourced history, charts, export, and 39 MCP tools. **Not yet available:** automatic scheduling of all platform accounts, universal creator-analytics APIs, verified counts of distinct human users, or public directory installation of the private ChatGPT app.

## Recent release history

**v0.9.2** completes public distribution documentation: clear English/Chinese quick starts, current repository governance, verified Windows/native and macOS/Linux CLI installation scope, and real bundled upstream license texts in every release archive. No analytics or schema changes. [Release notes](docs/release-notes-v0.9.2.md).

**v0.9.1** improves DEV.to canonical identity: recognize already-registered article channels, suggest (never silently merge) matching paths across historical website domains, and require an explicit selected Work or explicit new-Work confirmation. Available in Web UI, extension and MCP. [Guide](docs/platform-api.md).

**v0.9.0** adds [API-only DEV.to creator analytics](docs/platform-api.md): connect via private API key, preview and link published posts, collect official followers/views/reactions/comments and review trends in the Web UI, extension and 39-tool MCP. No SyndiRelay dependency or browser scraping; unsupported platforms remain unavailable.

**v0.8.0** introduces independent [RSS/Atom/Sitemap website discovery](docs/website-discovery.md): preview public article candidates, explicitly select URLs to register, deduplicate existing works, save source provenance and rescan existing sources. Three new MCP tools, native Web UI and Edge/Chrome extension are synchronized; no Inkerio or SyndiRelay is required. Platform account/article metrics remain planned separately.

**v0.7.1** validates the [31-tool MCP contract](docs/mcp-setup.md) end to end, fixes stale ChatGPT discovery by documenting **Refresh tools → New chat**, and packages an isolated, cross-platform MCP acceptance checker. Existing data and credential storage are unchanged.

**v0.7.0** also fixes Windows console-window startup: the extension probes and silently starts the local service through a separately registered Native Messaging host, with manual recovery controls. Existing Windows login tasks become silent.

**v0.7.0** adds the [Propagation paths](docs/impact-dashboard.md) view: canonical work → first-party language links → external syndication → reviewed third-party evidence, plus a genuine observed/verified timeline. Channel URLs have no inferred publication dates. SyndiRelay may optionally hand off confirmed public URL metadata without becoming a runtime dependency.

**v0.6.1** refines the impact charts: source-scoped article drilldown, calendar-day trend filters, actual interval download changes, unclipped SVG/PNG exports, and identical Web/extension UI.

**v0.6.0** adds [impact charts and usage guidance](docs/impact-dashboard.md): GitHub daily traffic, release/platform breakdowns, site trends, exact article URL attribution, and PNG/SVG/CSV exports. Web, extension and MCP share the same observations.

Platform API coverage: **DEV.to account and article metrics are implemented directly**, without a SyndiRelay runtime. Other platforms require documented, authorized official read APIs and are not claimed supported by default. The older [social impact proposal](docs/social-impact-architecture.md) is marked historical and superseded by the [API-only implementation](docs/platform-api.md).

## Architecture and data policy

One local Go service, SQLite, embedded Web UI and an MCP endpoint share the same business operations. An AI client performs searches using its own authorized tools; Reverfolio records the task, candidate link and review decision. A `secondary_report` mention cannot be verified without original `web` or `api` evidence and explicit review. Reference pages, verified identity URLs, grouped propagation events, GitHub downloads, visitors and actual software users are different metrics.

All live observations and credentials belong in a private runtime directory, not git. Do not commit imported reports, browsing data, access tokens, databases, or personal metrics. See [architecture](docs/architecture.zh-CN.md) and [implementation plan](docs/implementation-plan.md).

## Development and governance

Use feature branches and pull requests. Squash merge only; delete merged branches. Main must not be force-pushed or deleted. Repository settings and enforcement limitations are documented in [governance](docs/governance.md).

## License

[MIT](LICENSE). Third-party code must retain its applicable notices. No third-party application source has been imported during bootstrap.

## Deployment example: a single native Windows data source

For an existing Windows installation, the Web UI, Edge/Chrome extension and local/remote MCP are intended to share **one** native Windows `Data/state.sqlite`; do not run a second writable copy in WSL. Each other self-hosting user chooses their own local data directory and credentials. See [Windows authority and MCP connection](docs/windows-authority.md).

## Browser extension (optional; Windows portable distribution)

The optional Windows-packaged Reverfolio browser extension is in `extension/`. It loads the full native dashboard and its controls in an Edge/Chrome side panel or a browser tab, using exactly the same app source and styles as the local Web UI. It connects to the single Windows local origin 127.0.0.1:19452; there is no service-priority selection or launcher popup. Windows files go in C:\software\Coding\reverfolio\extension. To upgrade from v0.1 unpacked, remove the old extension and load the directory again using Edge edge://extensions > Developer mode > Load unpacked. See [installation steps](extension/README.md). It is independent of BlogCTL and does not read or store GitHub/MCP credentials.

## Configuration and Windows layout

Windows example (choose your own permanent folder): `C:\software\Coding\reverfolio\`. The executable is in `bin/reverfolio.exe`, SQLite state in Data/state.sqlite and local settings in config.toml. See [Windows layout](docs/windows-layout.md). The former WSL copy is backed up and inactive; only Windows receives new data.

## Run locally

Requires Go >= 1.26. From the repository checkout:

```bash
go run ./cmd/reverfolio -db data/reverfolio.sqlite
```

Open <http://127.0.0.1:18440> in a browser. **No manual token entry is required for the local UI.** It receives a process-scoped, HttpOnly, SameSite=Strict session cookie; unsafe requests require a same-origin UI header and cross-origin/cross-site access is rejected. The programmatic HTTP API and `/mcp` still accept the private `Authorization: Bearer <token>` credential from `data/api-token` (never exposed to page JavaScript). Alternatively set a 32+ character `REVERFOLIO_API_TOKEN` server-side. The HTTP server binds loopback only.

Supply `GITHUB_TOKEN` to collect repository traffic with appropriate GitHub permissions. When explicitly authorized to use the signed-in CLI, run with `-github-auth gh` instead. A missing permission or endpoint is recorded as partial/unavailable, not zero. Public release downloads are cumulative asset counters; the first observation is a baseline, not an interval increment.

The standalone bilingual UI defaults to Simplified Chinese; language and other editable preferences persist in machine-local TOML. Its sidebar provides overview, projects, channels, evidence review, discovery tasks, historical metrics and settings. It supports registering repositories/articles/sites, language/channel pages, submitting candidate mentions, human review and grouping, discovery/search run tracking, numeric history dry-run import and SQLite backup. **Reverfolio has no runtime dependency on BlogCTL, any browser extension or other publishing application.** All write commands require unique request IDs. The MCP tools expose the same operations, and `-mcp-stdio` provides a local stdio transport. The default `-collect-interval 24h` runs idempotent GitHub snapshots; set `-collect-interval 0` to disable scheduling.

Restore **only to a new database path**, never over the current database:

```bash
go run ./cmd/reverfolio -restore /private/backup.sqlite -db /private/new-state.sqlite
```

Historical numeric observations use `schema_version: 1`, each with source, unit, scope, timezone, start/end time and observation time. Dry-run validation precedes import. Source URLs for translations and distribution must be linked to the correct work; unknown distribution links should be left missing, not guessed. Backups contain private data and must not be committed.

## Interfaces

- MCP: `/mcp` over authenticated streamable HTTP (or `-mcp-stdio`); supports registering works/channels, search tasks, candidate mention submission/review, GitHub collection, historic observation import, report and audit.
- HTTP: `/api/works`, `/api/works/{id}/collect`, `/api/works/{id}/snapshots`, `/api/works/{id}/changes`, `/api/works/{id}/traffic/{views|clones}`; `POST /api/actions/{upsert_work|add_channel|submit_mention|review_mention|link_propagation_event|create_discovery_task|claim_discovery_task|complete_discovery_task|import_history}`; `GET /api/report`, `/api/mentions`, `/api/tasks`, `/api/audit`; `POST /api/backup`.
- Web: `/` (same local service and application state).

## Validation

```bash
go test -race ./...
go vet ./...
python3 scripts/check_docs.py
```

See [workflow acceptance](docs/workflow-validation.md) for test scope and platform caveats. Hosted CI runs Linux/Windows/macOS suites; cross-compilation alone does not prove native runtime behavior. No claim is made of accurately identified human readers or verified end-users.

## Rename compatibility

Formerly OpenEcho. Existing Windows data and configuration are preserved; the previous installation path remains a compatibility junction. See [migration notes](docs/rename-migration.md).

## Release download breakdown

Web UI, MCP and exported reports share the backend `filename-v1` classification: packages, source archives, auxiliary files (checksums/signatures/licenses/SBOM), and other or ambiguous files. Every asset exposes its category and classification reason. Ambiguous archives stay unclassified; package downloads are not verified installations or users. Historical snapshots are classified on read without rewriting their original observations.

Interval totals include only known asset deltas and separately count unknown deltas (late baselines, missing assets or decreased counters). They refer to the two observation timestamps, not calendar-day growth. Missing release permissions show unavailable rather than zero.

## Automatic collection

Settings → Data sources → GitHub → Automatic collection controls the collection interval, saved in `config.toml`. The running HTTP service reloads the schedule within one minute; disabling it retains existing data. The computer and service must be running. Explicit `-collect-interval` CLI overrides remain fixed until restart. Stdio MCP does not run a duplicate scheduler.

Configure `github.auth = "gh"` to use a valid native GitHub CLI login, or provide `GITHUB_TOKEN` in the service environment. GitHub secrets are not edited in the UI; optional site-provider credentials use a separate write-only control. Runtime/MCP status indicates whether a credential was loaded; snapshots show actual permissions and errors.

## Site providers and identity groups

See [provider setup and authorization](docs/site-sources.md). Each provider retains its own observation window, timezone and failure state. New sources are disabled by default. Operators supply credentials locally after explicitly authorizing access; an AI client must not generate, download or inspect credentials merely because a browser is signed in. Umami browser mode uses the signed-in local browser for restricted read-only JSON statistics without exporting cookies or API keys.

Identity links require explicit same-author evidence, reviewer and reason. Names alone do not merge identities; unlinking retains audit history. Schema 4 adds identity and provider tables without rewriting original observations. Back up the database before upgrading.

## Works and publishing integration

Choose which works to monitor, preview and import public publication links, and inspect their recorded origin. No publishing tool is required. See the [portable catalog and usage guide](docs/portable-catalog.md). Imports preserve existing titles and monitoring choices; first-run databases are empty.
