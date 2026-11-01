# Contributing to Reverfolio

Reverfolio is **public MIT-licensed software**, not a planning-only private repo. Contributions are welcome through GitHub Issues and pull requests. See [Getting started](docs/quickstart.md), [中文快速入门](docs/quickstart.zh-CN.md) and [governance](docs/governance.md).

## Development workflow

Fetch `main`, create one focused branch per change, open a PR, and use Squash Merge after the required Linux, Windows, macOS and docs CI checks pass. The maintainer auto-deletes merged head branches. Do not force-push, delete or directly modify protected `main`.

Describe why the change is needed, user-facing behavior, relevant API scope, validation steps, and any remaining gaps. Keep the public English and Chinese instructions synchronized. Existing data and public interfaces should be backwards-compatible unless an explicit migration and safe rollback are provided.

## Checks

```bash
go test -race ./...
go vet ./...
python3 scripts/check_docs.py
python3 scripts/bundle_dependency_licenses.py --check
node scripts/check_extension.mjs
python3 scripts/check_mcp_contract.py --fixture
```

The CI matrix also validates native Go builds on Windows/macOS/Linux. Test platform API integrations against **synthetic responses**, not other users' account tokens. Never commit credentials, native-browser cookies, profile files, screenshots containing private identities, local `Data/` directories, SQLite backups, reports of personal activity, or full archived third-party articles.

## Architecture and platform policy

Core logic lives in the Go service. Embedded Web UI, Edge/Chrome extension and MCP must share the **same backend and SQLite authority**, not introduce shadow databases. The project does not require Inkerio or SyndiRelay at runtime.

A creator platform adapter may read the **documented, authorized official API only**. If a platform has no such API for a desired private statistic, label that metric unsupported or unavailable; do not scrape logged-in pages or reverse engineer private endpoints. Identity, follower counts, page views, downloads, and verified human adoption are distinct metrics. Preserve sources, observation times, missing values and review decisions.

Add dependency licenses and upstream notices when importing code. The packaging script now checks every resolved Go module and bundles original notices into the release. For bugs exposing tokens or user data, follow [private vulnerability reporting](SECURITY.md), not a public Issue.
