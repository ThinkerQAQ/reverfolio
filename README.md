# Reverfolio

[简体中文](README_ZH.md)

## Introduction

Local impact tracking for open-source projects and published writing. Register works and their publication URLs, collect GitHub/site/DEV.to observations, inspect charts and review citation evidence. One Go service shares SQLite across Web UI and MCP. Inkerio, SyndiRelay and browser extensions are optional; an empty installation contains no author accounts or content. Downloads and followers are observations, not verified people or users.

## Quick Start

### 1. Prerequisites

Windows, Linux or macOS (x64/ARM64), a browser and a matching portable release. No Go or publishing application is required for packaged use. Registering works needs no provider credential; private provider metrics require your own authorization.

### 2. Download and install

Download the matching archive and `SHA256SUMS` from [Releases](https://github.com/ThinkerQAQ/reverfolio/releases/latest), verify its SHA-256, and extract it into a permanent private directory. Archive names use `Reverfolio-vVERSION-OS-ARCH`, with OS `windows`, `linux` or `darwin` and ARCH `amd64` or `arm64`.

**Windows:** open PowerShell in the extracted `Reverfolio` directory:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Install-Reverfolio.ps1
.\Start-Reverfolio.cmd
```

Open <http://127.0.0.1:19452>. The installer creates config and Data without replacing existing settings or SQLite.

**Linux/macOS:** from the extracted `Reverfolio` directory:

```sh
# Fresh installation only; do not overwrite an existing private config.
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

Open <http://127.0.0.1:18440>. Keep the foreground process running. Portable Unix archives do not install a systemd/launchd service. macOS binaries are unsigned and may require operating-system approval.

### 3. Register your first project

Open **Projects**, add a project with a title and a GitHub repository you own in `owner/name` format, then save. Refresh the page: the same project must remain. Restart the service with the same config and verify it still exists. This completes a real persistent catalog operation without an API key.

For a first metrics observation, use that project's collection action. Public release-download counters can be available without a token. Private traffic requires an authorized `GITHUB_TOKEN`, or `github.auth = "gh"` with your own valid native GitHub CLI login. Inspect the collection state, timestamp and available metrics. A repository without releases or missing permissions may legitimately show unavailable data; it must not be reported as zero or as a successful traffic read. The first cumulative snapshot is a baseline.

### 4. Verify and troubleshoot

Health checks from a second terminal:

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```sh
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

Require `status=ok`, a saved project that survives restart, and an explicitly recorded success/partial/failure state for any requested collection. Health alone is not the full acceptance criterion.

- **Empty workspace after upgrade:** confirm `storage.database` and config path; do not create a second writable WSL database.
- **401/403 or missing traffic:** check authorization for that repository; unavailable does not mean zero.
- **Old tools in ChatGPT:** refresh the registered MCP connection and start a new chat after upgrading.
- **Windows extension cannot start Core:** rerun the installer in the same permanent directory; use `Start-Reverfolio.cmd` to inspect a manual start.
- **Upgrade/rollback:** back up SQLite and credentials first. Restore a compatible binary and database together. CLI restore uses `-restore BACKUP -db NEW_PATH`, never the live database path.

## Configuration

[config.example.toml](config.example.toml) is the source of defaults. The private `config.toml` sits beside `bin/`; `-config` selects another file. Common settings:

```toml
version = 1
[server]
listen = "127.0.0.1:18440"
[storage]
database = "Data/state.sqlite"
[collection]
interval = "24h"
[github]
auth = "env"
```

Windows setup uses port `19452` for its extension; keep that installer-generated value. Paths are resolved by the application config loader. `GITHUB_TOKEN` supplies your GitHub credential in `env` mode; `gh` mode requires an authorized CLI in the service account. `REVERFOLIO_API_TOKEN` can supply a server-side token of at least 32 characters. Otherwise a private local API token is generated; do not paste it into page JavaScript or chat.

Settings → Data sources controls the collection interval and individual sources. `-collect-interval 0` disables scheduling; explicit CLI overrides take precedence. The machine/service must remain on. DEV.to account polling is not automatically scheduled. The local UI uses its own session and requires no manual bearer-token entry.

## Optional Features

### Local MCP

In your local stdio client's server settings, use absolute paths to the installed binary and config:

```json
{
  "mcpServers": {
    "reverfolio": {
      "command": "/absolute/path/Reverfolio/bin/reverfolio",
      "args": ["-mcp-stdio", "-config", "/absolute/path/Reverfolio/config.toml"]
    }
  }
}
```

Windows uses `bin/reverfolio.exe`; forward-slash paths are accepted. Reconnect and call `get_dashboard`, `list_works`, then `get_collection_status`. Verify the same registered project and service version. Stdio does not run a duplicate scheduler. Write tools require unique `request_id` values and applicable version checks. Remote ChatGPT additionally needs your own authenticated MCP tunnel/connection; the developer's private app is not a public one-click installer. Do not expose the loopback HTTP service publicly.

### Website and creator metrics

- **Website discovery:** preview your RSS/Atom/Sitemap in Website discovery, select owned article URLs, then import. Existing works and explicit publication mappings are preserved; titles alone never merge identities.
- **DEV.to:** enter your personal API key in the local **Creator platforms** page, connect, preview your published articles, deliberately select/link works, then collect followers/views/reactions/comments. Credentials are write-only; MCP does not accept the raw key. Review historical-domain suggestions explicitly. Other creator-platform integrations are not implemented.
- **Umami/GSC/Bing:** first register your site, then use **Settings → Data sources → Search and site analytics**. Select the work, property ID, timezone, lookback and lag; new sources are disabled until explicitly enabled. GSC needs an authorized read-only Search Console credential and property (for example `sc-domain:example.org`); Bing needs a verified site URL and authorized Webmaster API key. Umami supports an authorized token or a separately authorized signed-in local browser session. Saving is separate from collecting. Never export browser cookies or generate credentials merely because a browser is signed in.
- **Catalog handoff:** Projects → Import works & publication links accepts catalog JSON; preview and select works before import. `scripts/publications_to_catalog.py --help` describes the optional publishing-export converter. No shared publisher database or account session is needed.
- **Evidence/charts:** attach confirmed language/publication URLs, submit external citations for explicit review and inspect propagation/history charts or SVG/PNG/CSV exports. Keep source timezone, observation window, missing values and evidence provenance. Different provider counts must not be added into a unique-user count.

### Windows extension

After installation, enable Developer mode at `edge://extensions` or `chrome://extensions` and load the installation's `extension/` folder unpacked. Click its icon: it opens the same dashboard and can silently wake the service through the installed Native Messaging host. Reload after updates; re-add an older unpacked extension if its ID differs from the pinned manifest identity. Its generated local capability is private. This native wake-up installation is Windows-only; Unix users can use Web and stdio without it.

## Development

```sh
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

Use Go **1.26+** as declared by `go.mod`. Windows may build `reverfolio.exe` and run it with the same flags. Source defaults to port 18440.

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
python3 -m unittest discover -s scripts -p 'test_publications_to_catalog.py'
```

The MCP fixture builds an isolated database and validates the exact 39-tool contract and a write/read/review/export workflow without live credentials. CI additionally runs Web/extension checks, race/vet tests and Windows/Linux/macOS checks. Packaging uses `python3 scripts/package_release.py vX.Y.Z` (replace with the intended version); it builds six OS/CPU archives with dependency license texts. Native GUI/account/reboot checks remain separate from fixtures and cross-compilation. Regenerate icons with `scripts/build-icons.sh`; the original SVG and derived assets are source-licensed. Use feature branches, PRs and Squash Merge.

## License

Original code and project visual assets are [MIT](LICENSE); dependencies retain their [own notices](THIRD_PARTY_NOTICES.md), including full license texts in portable archives. This does not license user writing, photography, provider data or third-party content. Keep SQLite, private TOML, tokens, extension capabilities and reports out of Git/releases. Selected MCP data reaches the connected client/AI service. Credentials are entered locally, never obtained merely from a signed-in account. See [security policy](SECURITY.md).
