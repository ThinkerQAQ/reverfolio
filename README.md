# OpenEcho

[中文](README.zh-CN.md) · [Architecture](docs/architecture.zh-CN.md) · [Roadmap](docs/implementation-plan.md)

Local-first observability for open-source and content impact.

OpenEcho will connect projects, multilingual articles, distribution channels, metric history, and evidence of external citations or adoption. AI clients can discover evidence through their search tools and submit it through MCP.

**Status: v0.1 development preview on a feature branch; no release published.** GitHub snapshots, article/site registry, bilingual channel links, candidate evidence and review, search tasks, historical numeric observations, MCP, standalone bilingual Web UI, scheduled snapshots, raw-response archiving and backup/restore are implemented. Additional provider connectors and AI search execution are not included.

## Architecture and data policy

One local Go service, SQLite, embedded Web UI and an MCP endpoint share the same business operations. An AI client performs searches using its own authorized tools; OpenEcho records the task, candidate link and review decision. A `secondary_report` mention cannot be verified without original `web` or `api` evidence and explicit review. Reference pages, verified identity URLs, grouped propagation events, GitHub downloads, visitors and actual software users are different metrics.

All live observations and credentials belong in a private runtime directory, not git. Do not commit imported reports, browsing data, access tokens, databases, or personal metrics. See [architecture](docs/architecture.zh-CN.md) and [implementation plan](docs/implementation-plan.md).

## Development and governance

Use feature branches and pull requests. Squash merge only; delete merged branches. Main must not be force-pushed or deleted. Repository settings and enforcement limitations are documented in [governance](docs/governance.md).

## License

[MIT](LICENSE). Third-party code must retain its applicable notices. No third-party application source has been imported during bootstrap.

## Configuration and Windows layout

Windows: C:\software\Coding\opencho\. The executable is in bin/opencho.exe, SQLite state in Data/state.sqlite and local settings in config.toml. See [Windows layout](docs/windows-layout.md). This local installation is independent of the Ubuntu/WSL database and does not automatically synchronize.

## Run locally

Requires Go >= 1.26. From the repository checkout:

```bash
go run ./cmd/openecho -db data/openecho.sqlite
```

Open <http://127.0.0.1:18440> in a browser. **No manual token entry is required for the local UI.** It receives a process-scoped, HttpOnly, SameSite=Strict session cookie; unsafe requests require a same-origin UI header and cross-origin/cross-site access is rejected. The programmatic HTTP API and `/mcp` still accept the private `Authorization: Bearer <token>` credential from `data/api-token` (never exposed to page JavaScript). Alternatively set a 32+ character `OPENECHO_API_TOKEN` server-side. The HTTP server binds loopback only.

Supply `GITHUB_TOKEN` to collect repository traffic with appropriate GitHub permissions. When explicitly authorized to use the signed-in CLI, run with `-github-auth gh` instead. A missing permission or endpoint is recorded as partial/unavailable, not zero. Public release downloads are cumulative asset counters; the first observation is a baseline, not an interval increment.

The standalone bilingual UI defaults to Simplified Chinese; language and other editable preferences persist in machine-local TOML. Its sidebar provides overview, projects, channels, evidence review, discovery tasks, historical metrics and settings. It supports registering repositories/articles/sites, language/channel pages, submitting candidate mentions, human review and grouping, discovery/search run tracking, numeric history dry-run import and SQLite backup. **OpenEcho has no runtime dependency on BlogCTL, any browser extension or other publishing application.** All write commands require unique request IDs. The MCP tools expose the same operations, and `-mcp-stdio` provides a local stdio transport. The default `-collect-interval 24h` runs idempotent GitHub snapshots; set `-collect-interval 0` to disable scheduling.

Restore **only to a new database path**, never over the current database:

```bash
go run ./cmd/openecho -restore /private/backup.sqlite -db /private/new-state.sqlite
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

See [workflow acceptance](docs/workflow-validation.md) for the test scope, caveats and the current CI billing limitation. Linux tests do not prove Windows/macOS runtime behavior. No claim is made of accurately identified human readers or verified end-users.
