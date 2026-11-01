# Reverfolio

[简体中文](README_ZH.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

Reverfolio links open-source projects, published articles, provider metrics and citation evidence in a local SQLite workspace accessible through Web UI and MCP.

GitHub projects, personal websites and articles published to multiple platforms expose different metrics. Release downloads, repository traffic, search impressions, pageviews and reactions have different scopes and collection windows. They cannot be summed into a count of real people.

Reverfolio keeps these observations alongside the work and source they belong to. Its goals are to:

- Register projects, articles and websites independently of any publishing system.
- Distinguish a missing metric or failed request from a measured zero.
- Preserve the observation window, provider, collection status and historical baseline.
- Review the original evidence before treating a citation or adoption claim as verified.

Inkerio and SyndiRelay are optional. A fresh installation contains no preconfigured author accounts or publications.

## 2. Quick Start

Follow one complete workflow: install Reverfolio, connect the browser extension on Windows, register a GitHub repository, and check that your work survives a restart.

### 2.1. Prepare the environment

Use Windows, Linux, or macOS (x64/ARM64). The Windows browser extension supports Edge and Chrome; Linux/macOS portable packages use the Web UI **without an extension**. A release archive needs no Go, Node.js, publishing tool, or API key for the initial project.

For your CPU architecture, check Windows **Settings → System → About**, macOS **About This Mac**, or run `uname -m` on Linux (`x86_64` = amd64, `aarch64` = arm64).

### 2.2. Download and install

1. Open the [latest Release](https://github.com/ThinkerQAQ/reverfolio/releases/latest). Choose the archive whose name ends in `windows-amd64.zip`, `windows-arm64.zip`, `linux-amd64.tar.gz`, `linux-arm64.tar.gz`, `darwin-amd64.tar.gz`, or `darwin-arm64.tar.gz` for your platform.
2. Also download `SHA256SUMS`. In the download directory, calculate the archive hash: Windows `Get-FileHash .\Reverfolio-*.zip -Algorithm SHA256`, Linux `sha256sum ./Reverfolio-*.tar.gz`, or macOS `shasum -a 256 ./Reverfolio-*.tar.gz`. Compare with the entry for the **same archive filename** in `SHA256SUMS`; do not install a mismatched archive.
3. Extract the **entire** archive into a permanent private directory. The resulting `Reverfolio/` installation root directly contains `bin/` and `config.example.toml`. On Windows it also contains `extension/` and `Start-Reverfolio.cmd`. Do not run anything inside the compressed archive.

**Windows:** open PowerShell and replace the sample path with the actual extracted directory:

```powershell
Set-Location 'C:\path\to\Reverfolio'
.\Start-Reverfolio.cmd
```

This command runs installation setup as needed, creates `config.toml` and `Data/` without replacing existing files, prepares the extension credentials, **registers** the native messaging host for Edge/Chrome, and starts the local service hidden. No manual host registration or pairing code is required. Open [http://127.0.0.1:19452](http://127.0.0.1:19452); refresh if the process is still starting.

**Linux/macOS:** open a terminal in the extracted `Reverfolio/` directory:

```bash
# Copy only on a fresh installation; do not overwrite an existing config.
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

Keep the terminal running and open [http://127.0.0.1:18440](http://127.0.0.1:18440). Portable Unix archives do not install a background service. An unsigned macOS binary may require operating-system approval.

### 2.3. Install the browser extension (Windows)

Run [2.2](#22-download-and-install) **first**, so the Windows installer can initialize the extension.

1. Open `edge://extensions` in Edge or `chrome://extensions` in Chrome.
2. Enable **Developer mode**, then select **Load unpacked**.
3. Select the `extension/` directory **inside the extracted Reverfolio installation**. That directory must directly contain `manifest.json`. Do not select the ZIP, `bin/`, or the repository root.
4. Find Reverfolio in the browser's Extensions menu, optionally pin it, and click its icon. It opens a **side panel**, not a popup.
5. Confirm **Local service → Connected**, then open **Projects** and verify that the page loads. The side panel and [local Web UI](http://127.0.0.1:19452) share the same data.

If the panel says **Service disconnected**, click **Start local service**, then **Retry connection**. If native messaging is unavailable or the extension is unauthorized, run `.\Start-Reverfolio.cmd` again from the **same installation root**, then **Reload** the extension from the extensions page. Moving the installation requires running the launcher again from its new location. **Loading the extension is the only manual browser installation step**; the installer registers the native host automatically.

Linux/macOS users **skip this section** and continue through the Web UI.

### 2.4. Register and collect your first project

In the Windows side panel or the Unix Web UI:

1. Open **Projects** from the left navigation.
2. Scroll to the **Add a project or article** form **at the bottom** of the page. This is a section heading, not a separate button.
3. Set **Type** to **GitHub project**, enter a **Title**, and fill **Repository** with `owner/repo` for a public repository you want to track.
4. Click **Create** and confirm that its project card appears.
5. Click **Collect GitHub** on that card. Inspect its collection status, timestamp, and available metrics under **Details** or the overview.

Public repository metrics may work without a token. Private traffic needs authorized GitHub credentials. No releases, missing permissions, and provider errors can produce unavailable values; these must not be treated as zero. The first cumulative observation is a baseline rather than growth.

### 2.5. Verify the result

Run the relevant health check from another terminal:

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```bash
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

The response should contain `"status":"ok"`. Refresh **Projects** and confirm that your project remains. Restart the same installation with the **same `config.toml`**, then confirm it is still present. On Windows, also check that the extension can still read the project.

**Acceptance:** extension connected (Windows), project persisted after restart (all platforms), and collection has an explicit success/partial/failure result when attempted. A healthy HTTP endpoint alone does **not** prove a successful metrics collection.

### 2.6. Troubleshooting

- **Dashboard will not open on Windows:** rerun `Start-Reverfolio.cmd` from its installation root, check `http://127.0.0.1:19452/healthz`, and verify the listen port in `config.toml` is `19452` and not occupied.
- **Extension is missing or cannot connect:** load the `extension/` folder containing `manifest.json`, not the ZIP. Reload it, then rerun the installer to repair native messaging. Check that `Data/extension-access.key` and `extension/ui/local-auth.js` exist, but **never reveal their contents**.
- **Extension reports unauthorized:** rerun the launcher from the folder used to load the extension; then reload the extension. Do not paste the API bearer token into it.
- **No extension on Linux/macOS:** expected for these portable packages; use the Web UI or local MCP.
- **Projects disappear after upgrade:** inspect the original `config.toml` and `storage.database`. Do not start a second writable WSL database.
- **GitHub 401/403 or missing metrics:** check repository permissions, the provider rate limit, and authorized credentials. Unavailable is not zero.
- **Old UI or MCP tools remain:** update binaries, reload the browser extension and reconnect the MCP client.

If necessary, [open an Issue](https://github.com/ThinkerQAQ/reverfolio/issues) with OS, install method, reproduction steps, and sanitized error details. Do not upload `Data/` or private credentials.

## 3. Configuration

### 3.1. Common settings

The [configuration template](config.example.toml) is the reference for supported settings. The installer creates a private `config.toml` next to `bin/`, with SQLite under `Data/`. Windows uses port `19452` for its browser extension; the template and Unix installation use `18440`.

```toml
[server]
listen = "127.0.0.1:18440"

[storage]
database = "Data/state.sqlite"

[collection]
interval = "24h"

[github]
auth = "env"
```

Use `GITHUB_TOKEN` in `env` mode, or select `github.auth = "gh"` if the service account has an authorized GitHub CLI login. `-collect-interval 0` disables scheduling; command-line overrides take precedence over the configured interval. DEV.to account polling is not automatically scheduled. The application must be running for scheduled GitHub collection.

The HTTP API uses a separate private token in `Data/api-token` unless `REVERFOLIO_API_TOKEN` is provided (at least 32 characters). The Web UI uses its own local session; do not paste the API token into page scripts or chats. Upgrades must preserve the same `config.toml` and `Data/` directory.

### 3.2. Upgrade, backup and restore

- **Before upgrading:** back up `config.toml`, `Data/` and any private provider credentials. Reinstall into a stable directory without replacing these files.
- **Empty workspace after upgrading:** check the config path and `storage.database`. Do not start a second writable database in WSL.
- **Missing GitHub traffic / 401 / 403:** verify repository access and the configured credential. Unavailable data is not zero.
- **Extension cannot wake the server:** rerun `Start-Reverfolio.cmd` from the installed directory and inspect the local health endpoint.
- **MCP still shows old tools:** reconnect the client and start a new session after updating.
- **Offline restore:** `reverfolio -restore BACKUP -db NEW_PATH` restores into a **new** database path, not over the live database.

## 4. Optional Features

These cases are optional and independent. Start with the Quick Start before connecting external data.

### 3.1. Case 1: Import articles from your own website

1. Open **Discover websites** and enter the public website or feed URL you control.
2. Preview discovered RSS, Atom or Sitemap candidates. Check the titles and URLs; discovery does not itself register a publication.
3. Select the articles you want, then import them.
4. Open **Projects** and **Channels** to confirm the imported works and publication URLs.

Only explicitly selected URLs are imported. A feed or sitemap might not list every article. Matching titles do not automatically merge two works.

### 4.2. Case 2: Connect DEV.to

1. Obtain your own DEV.to API key and enter it in the local **Creator platforms** UI.
2. Connect the account and preview its published articles.
3. Select and link the articles to existing works, or import the selected publications.
4. Run the DEV.to collection action, then inspect **Creator platforms** and its metric history.

The key is stored locally and is not accepted by an MCP enrollment tool. DEV.to is the implemented creator-platform API integration; other creator-platform accounts are not automatically supported.

### 4.3. Case 3: Connect website and search analytics

1. Register the website as a work in **Projects**.
2. Open **Settings → Data sources → Search and site analytics** and choose Umami, Google Search Console (GSC) or Bing Webmaster.
3. Enter the provider-specific site/property ID, timezone, observation window and an authorized credential. GSC requires a property you can read; Bing requires a verified site and its API key. Umami supports a token or a separately authorized local browser session.
4. Save and explicitly enable the source, then start collection.
5. Review the source status and observations. Saving a source does not mean that any metrics were collected.

The providers return different measurements; keep their scopes separate. Browser sign-in alone is not permission to export cookies or extract credentials.

### 4.4. Case 4: Review citations and propagation

Register confirmed publication links under **Channels**. Use **Evidence & adoption** to submit a possible citation, inspect the original page/API evidence, and explicitly approve or reject it. A secondary report alone is not verified original evidence. **Discovery tasks** record searches for an AI client to perform; creating a task does not execute a web search. History and propagation views distinguish pending evidence from reviewed results.

### 4.5. Case 5: Local MCP integration

For a **local stdio MCP client**, configure the absolute paths to your installed executable and config:

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

Use `bin/reverfolio.exe` on Windows. Reconnect the MCP client and call `get_dashboard`, `list_works` and `get_collection_status` to verify that it reads the same installation. The stdio process does not start a second collection scheduler.

MCP includes read tools for dashboards, sources, works, evidence and collection status, plus explicit write tools for imports, observations and reviews. Writes require a unique `request_id` and, where applicable, an expected version. A remote ChatGPT connection requires a separately authenticated MCP tunnel; the author's private connector is not a public one-click installation. Never expose the loopback HTTP service directly to the Internet.

## 5. Development

### 5.1. Build and test from source

Install the Go toolchain required by [go.mod](go.mod), plus Python 3 and Node.js for the repository checks. From a source checkout:

```sh
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

On Windows, build `reverfolio.exe` instead and invoke it from PowerShell with the same flags.

### 5.2. Additional checks and packaging

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
python3 -m unittest discover -s scripts -p 'test_publications_to_catalog.py'
node scripts/check_extension.mjs
node scripts/check_dashboard.mjs
```

The MCP fixture uses an isolated database and does not require a live provider credential. CI also runs the remaining extension checks, Go race tests, vet and cross-platform builds. A passing fixture is not proof of a real account authorization or a native GUI install.

To generate portable archives for the intended release tag, run `python3 scripts/package_release.py vX.Y.Z` (replace the placeholder). Packaging verifies six OS/architecture archives, checksums and bundled upstream dependency licenses. Changes go through feature branches, PR review and Squash Merge.

## 6. License

Reverfolio is local-first, **not** a publicly hosted API. Its loopback HTTP API and MCP require their own credentials; the Windows extension uses a separate installation-specific local capability. Protect your OS account, `Data/`, SQLite, private configuration and tokens. Do not commit them or include user content, browser profiles or reports in release archives. MCP responses explicitly requested by a connected client may be transmitted to that client or its AI service.

Report security issues privately to **contact@mail.thinkerqaq.com**, without publishing private data or working credentials.

Original source code and project visual assets use the [MIT License](LICENSE). User writing, photographs, provider data and third-party content are not covered. Dependencies retain their upstream licenses: `go.mod`/`go.sum` record the module graph, and portable releases contain `THIRD_PARTY_LICENSES.txt` with upstream license/notice text. This file is generated and validated during packaging.
