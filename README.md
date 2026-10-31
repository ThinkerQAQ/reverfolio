# OpenEcho

[中文](README.zh-CN.md) · [Architecture](docs/architecture.zh-CN.md) · [Roadmap](docs/implementation-plan.md)

Local-first observability for open-source and content impact.

OpenEcho will connect projects, multilingual articles, distribution channels, metric history, and evidence of external citations or adoption. AI clients can discover evidence through their search tools and submit it through MCP.

**Status: foundation preview. Repository registration, SQLite persistence, manual GitHub snapshots, per-asset interval changes and a minimal Web UI are implemented. No release is published yet.**

## Planned architecture

One Go service, SQLite, a local web interface, and MCP. The business core is implemented for this domain; established libraries provide infrastructure. Metrics, evidence, and analytical claims remain distinct and traceable.

## First milestone

Register works and channel pages, persist GitHub snapshots, import historical observations, review external evidence, and expose the same application services through Web UI and MCP.

See the implementation plan for sequencing and acceptance gates. Real credentials, personal analytics, and imported history stay outside this repository.

## Development and governance

Use feature branches and pull requests. Squash merge only; delete merged branches. Main must not be force-pushed or deleted. Repository settings and enforcement limitations are documented in [governance](docs/governance.md).

## License

[MIT](LICENSE). Third-party code must retain its applicable notices. No third-party application source has been imported during bootstrap.

## Run the foundation preview

Requires Go 1.26 or later. Run from a source checkout:

```bash
go run ./cmd/openecho
```

Open http://127.0.0.1:18440 and enter the token from `data/api-token`. This file is private runtime configuration; never paste it into issues or commit it. The page keeps the entered token in memory only. The service also accepts `OPENECHO_API_TOKEN` (at least 32 characters).

For authenticated GitHub data, set `GITHUB_TOKEN` in the server environment. Public release counters work without it; repository traffic requires write access to the target repository and appropriate token permissions. Credentials are not sent to the Web UI. GitHub API redirects are refused to avoid credential forwarding.

Optional flags: `-listen 127.0.0.1:18440` and `-db data/openecho.sqlite`. Listening outside loopback is intentionally rejected in this preview.

Register an `owner/repository`, then collect a snapshot. Collect again with a new request ID to observe interval changes. Reusing an `Idempotency-Key` returns the original snapshot. First observations and late-discovered old assets are baselines; failed endpoints remain unavailable. Daily traffic uses UTC, and daily unique counts must not be added as unique people.

Implemented HTTP endpoints: `GET/POST /api/works`, `POST /api/works/{id}/collect`, `GET /api/works/{id}/snapshots`, `GET /api/works/{id}/changes`, and `GET /api/works/{id}/traffic/{views|clones}`. All API requests require Bearer authentication; collection also requires `Idempotency-Key`. Snapshot listing currently returns the latest 100 observations.

This slice stores selected, sanitized observation fields as JSON in SQLite, including repository IDs, assets and daily/window traffic. Full raw-response archives, referrers/popular paths, scheduling, edition/channel editing, evidence/history imports, MCP and backup/export are not implemented yet. There is no automatic classification of human readers or active users.

## Validation

```bash
go test -race ./...
go vet ./...
python3 scripts/check_docs.py
```

Platform CI is configured for Linux, Windows and macOS. See governance for the current GitHub-hosted runner account restriction; configured checks are not evidence of completed platform tests.
