# Reverfolio

[简体中文](README_ZH.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

After publishing a project or an article, its feedback is often scattered across several places. GitHub shows stars, release downloads, and repository traffic; a personal website has search and visitor statistics; publishing platforms report their own views and reactions. Following the progress of a work over time means checking and piecing together all those records.

Reverfolio brings them into one place. You can register your projects and articles, see available metrics and how they change over time, record mentions from other sites or projects, and review the supporting evidence. You can start with a single GitHub repository and add your website or publishing platforms later.

## 2. Quick Start

Register a public GitHub repository and collect its available metrics, from installation to your first result. On Windows, this walkthrough also sets up the browser extension.

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

Use the extension side panel on Windows or the Web UI on Linux/macOS:

1. Open **Projects** from the left navigation and scroll down to the **Add a project or article** form.
2. Select **GitHub project** under **Type**, enter a **Title**, and set **Repository** to `ThinkerQAQ/reverfolio` (a public example). You can use your own `owner/repo` instead.
3. Select **Create** and confirm the project card appears.
4. Select **Collect GitHub** on that card, then check **Details** or **Impact overview** for the results and collection time.

Public metrics may be available without credentials; some repository traffic requires GitHub authorization. The first cumulative download count is a baseline, not newly gained downloads.

### 2.5. Verify the result

Check the corresponding health endpoint from another terminal:

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```bash
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

The response should contain `"status":"ok"`. Refresh **Projects** and confirm the registered project is still present.

To check persistence, restart the service:

- **Windows:** in Task Manager → **Details**, end the `reverfolio.exe` process started by this installation. Run `Start-Reverfolio.cmd` again from the installation directory, then reopen the workspace.
- **Linux/macOS:** press `Ctrl+C` in the service terminal, then rerun `./bin/reverfolio -config ./config.toml`.

If the project survives restart, it has been saved to the local database. Collection can be successful, partial, or failed: for example, public downloads may be available while repository traffic requires additional authorization. Missing values are not zeros.

### 2.6. Troubleshooting

- **Workspace or extension cannot connect:** check the [Windows address](http://127.0.0.1:19452) or [Linux/macOS address](http://127.0.0.1:18440). On Windows, run `Start-Reverfolio.cmd` from the installation root and reload the browser extension.
- **Extension reports unauthorized:** make sure you loaded the `extension/` folder containing `manifest.json` from the current installation, rerun setup, and reload the extension.
- **GitHub collection fails:** check repository permissions, API rate limits, and local authentication. Unavailable metrics are not measured zeros.
- **Projects disappear after an upgrade:** verify `config.toml` and `storage.database`; the service may be pointing to a different database.

If needed, [open an Issue](https://github.com/ThinkerQAQ/reverfolio/issues) with your OS, steps to reproduce, and redacted error messages. Do not upload `Data/` or credentials.

## 3. Configuration

### 3.1. Common settings

Private configuration is stored in `config.toml` at the installation root, with local data under `Data/`. See [config.example.toml](config.example.toml) for the complete configuration. Nonsecret collection options can also be changed in **Settings → Data sources**.

- `server.listen`: loopback address. The Windows extension uses `127.0.0.1:19452`; Linux/macOS default to `127.0.0.1:18440`.
- `storage.database`: SQLite path, normally `Data/state.sqlite`.
- `collection.interval`: collection schedule; `0` disables it. The service must be running to collect.
- `github.auth`: `env` reads `GITHUB_TOKEN`; `gh` uses the authorized GitHub CLI session of the service account.
- `github.download_scope`: all release assets or installable packages.
- `interface.language`: `zh-CN` or `en`.

The local Web UI and installed extension do not require you to enter an API token. Enter third-party credentials in their respective local settings.

### 3.2. Upgrade, backup, and restore

1. Stop Reverfolio and back up `config.toml`, `Data/`, and any private provider credentials.
2. Download the new release and replace the application files in the installation root while **preserving your configuration and data**.
3. On Windows, rerun `Start-Reverfolio.cmd` and reload the extension. On Linux/macOS, start the new binary with the existing `config.toml`.
4. Open **Projects** and confirm your works and history are present.

To restore a SQLite backup, stop the service and run `reverfolio -restore BACKUP -db NEW_PATH` (`reverfolio.exe` on Windows). Restore to a **new database path**, never over the live file.

## 4. Optional Features

These cases are optional and independent. Start with the Quick Start before connecting external data.

### 4.1. Case 1: Import articles from your own website

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

### 4.4. Case 4: Review external citations

1. Register the work's published original, translation, or syndication links under **Channels**.
2. Submit a possible external citation in **Evidence & adoption**.
3. Check the original webpage or API evidence, then explicitly approve or reject the citation.
4. Inspect verified records in the propagation and history views.

**Discovery tasks** can be handled by an external AI client. Creating a task does not itself search the web.

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

### 5.1. Build and run from source

Development requires Git and a Go toolchain satisfying [go.mod](go.mod); the complete repository checks also use Python 3 and Node.js.

**Linux/macOS:**

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
mkdir -p Data
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

**Windows PowerShell:**

```powershell
git clone https://github.com/ThinkerQAQ/reverfolio.git
Set-Location reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio.exe ./cmd/reverfolio
New-Item -ItemType Directory -Force Data | Out-Null
.\reverfolio.exe -db .\Data\state.sqlite -collect-interval 0
```

After starting from source, open [http://127.0.0.1:18440](http://127.0.0.1:18440) and register a test repository as in [2.4](#24-register-and-collect-your-first-project). The source Web UI does not require the extension.

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
