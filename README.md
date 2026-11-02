# Reverfolio

[简体中文](README_ZH.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

After publishing a project or article, its feedback is scattered across different places: GitHub has downloads and repository traffic, websites have search and visitor statistics, and other authors may cite an article or apply its methods. Download counts alone do not establish who used a work; a list of links does not explain what a contribution achieved.

Reverfolio organizes these records locally. Register works and publication links, collect metrics from supported APIs, use an AI client to find external leads, then retain original evidence, review it and group related pages into impact events. It also records accepted upstream PRs, troubleshooting collaboration, reposts and directory listings, preserving what each record actually proves.

> Use files from the same release: do not mix the installer, portable package, launch scripts and extension from different versions.

## 2. Quick Start

Start with one public GitHub project: install Reverfolio, register the project and collect its first observation. Continue to ChatGPT setup if you want AI-assisted discovery. The browser extension can be installed later.

### 2.1. Choose a build

Open [Releases](https://github.com/ThinkerQAQ/reverfolio/releases) and choose your operating system and CPU architecture. Use the executable, launchers and extension from the same release or preview build.

| System | Package and entry point |
| --- | --- |
| Windows | Prefer the native `.exe` installer for your architecture; portable alternatives end in `windows-amd64.zip` or `windows-arm64.zip` |
| Linux | Extract `linux-amd64.tar.gz` or `linux-arm64.tar.gz` and use the included launcher |
| macOS | Use `darwin-arm64.tar.gz` for Apple Silicon or `darwin-amd64.tar.gz` for Intel, then use the included launcher |

Check Windows **Settings → System → About**, macOS **About This Mac**, or run `uname -m` on Linux: `x86_64` means amd64 and `aarch64` means arm64. Release packages require no Go, Node.js or publishing application.

Download the build's SHA-256 checksum file and compare your selected file against the entry for the **same filename**. Use `Get-FileHash` on Windows, `sha256sum` on Linux or `shasum -a 256` on macOS. Download again if the values differ.

### 2.2. Install and open the workspace

**Windows native installer:**

1. Run the installer and choose a permanent installation directory.
2. Read and accept the license, then install. Components and shortcuts use the default arrangement without separate selection pages.
3. Leave **Open Reverfolio** selected on the finish page and finish. Select **Set up the Chrome / Edge extension** only if you want the optional extension; see [4.1](#41-install-the-browser-extension-windows).
4. When the workspace opens in your browser, confirm **Local service → Connected** and **My works** in the sidebar. Use the Start menu entry to open Reverfolio again.

Initialization creates missing `config.toml` and local data files and prepares extension authorization and the native launcher. Existing configuration and data are retained. Normal local use requires no manual API token, pairing code or Native Messaging Host registration.

**Windows portable package:** extract the entire ZIP to a permanent directory. Open PowerShell in the installation root containing `bin/`, `extension/` and `config.example.toml`, then run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Open-Reverfolio.ps1
```

The launcher initializes missing configuration, starts the service and opens its configured address. Do not run the application inside the ZIP.

**Linux/macOS portable package:** extract the complete archive and run this in its `Reverfolio/` directory:

```bash
bash ./Open-Reverfolio.sh
```

The launcher creates missing configuration, starts the local service in the background and opens your browser. Linux needs `curl` and `xdg-open`; macOS uses its system `open` command. An unsigned application may require approval through the operating system after checking its source. Unix portable packages use the Web UI without requiring the Windows extension.

New configurations default to [http://127.0.0.1:19452](http://127.0.0.1:19452). Existing installations use `server.listen` from their own `config.toml`; the launcher opens that address. Do not create a separate configuration or empty database to recover a connection.

### 2.3. Register and collect your first project

1. Open **My works → Works**, then expand **Add a project or article** at the top. The form opens automatically when no works exist.
2. Select **GitHub project**, enter a title and your `owner/repository`. The public repository `ThinkerQAQ/reverfolio` can also be used to check the flow.
3. Select **Create** and confirm the work card appears.
4. Select **Collect GitHub** on the card, then inspect **Details** or **View impact → Overview** for metrics, collection times and errors.

Public metrics may be available without credentials; repository traffic normally needs appropriate read permissions. For authorization, install [GitHub CLI](https://cli.github.com/) if needed and run `gh auth login` locally, then choose **Local GitHub CLI** under **Settings → Data sources → GitHub → GitHub authentication** and restart the service. Alternatively, select the environment option and provide your own `GITHUB_TOKEN` to the service process.

The first cumulative download observation is a baseline, not newly gained downloads. A failed metric does not mean every metric failed, and missing values are not zeros.

### 2.4. Connect ChatGPT when you need AI assistance

This connects your local installation to your own ChatGPT connection. Your account or workspace must allow custom MCP/Tunnel connections. Otherwise, continue with the workspace or a [local MCP client](#45-connect-a-local-mcp-client).

1. Open [OpenAI Tunnel management](https://platform.openai.com/settings/organization/tunnels), select your organization, and create a Reverfolio Tunnel or reuse a dedicated one. Keep its Tunnel ID. Creation requires Tunnels Read + Manage; the Tunnel must match the intended ChatGPT workspace. Stop the previous installation's MCP runtime before reusing its Tunnel; do not run two clients for the same Tunnel.
2. In that organization's [Runtime API keys](https://platform.openai.com/settings/organization/api-keys), create a restricted runtime key with Tunnels Read + Use, or reuse a suitable key. Do not run the service with an Admin API key.
3. In Reverfolio, open **Settings → MCP connection**, paste the Tunnel ID and select **Connect ChatGPT**. The guide downloads and installs the official client if missing.
4. Enter the prepared runtime key when prompted locally. Do not paste it into chat. Saved credentials are reused. Windows uses a local setup window; Linux/macOS use the terminal prompt for initial credentials.
5. Wait for **Your local connection is ready**, then select **Open ChatGPT plugin settings**. For a new connection, add a custom MCP named Reverfolio, choose **Tunnel**, select the same Tunnel, and finish installation/authorization. For an existing connection, open its entry to enable or refresh it. See [ChatGPT plugin documentation](https://learn.chatgpt.com/docs/plugins) for interface and account restrictions.
6. Select that connection in ChatGPT and send this read-only check:

```text
Use Reverfolio to call list_works and get_collection_status read-only.
Confirm that the project I just registered is present and report the service version.
Do not collect or change data, and do not output credentials.
```

Seeing the registered project confirms that ChatGPT reads this installation. Local readiness confirms tunnel health; it does not confirm ChatGPT authorization. Runtime key creation and permissions are described by the [official tunnel-client](https://github.com/openai/tunnel-client).

### 2.5. Complete the first-use workflow

Follow the sidebar in order:

| Entry | What to do | Continue to |
| --- | --- | --- |
| My works | Register projects and articles, import website articles and link public publications | Find evidence |
| Find evidence | Create tasks and have the connected AI client execute searches and record checked scopes | Review evidence |
| Review evidence | Inspect originals, review citations and adoption, then associate evidence with impact events | View impact |
| View impact | Inspect overview, propagation and historical metrics; export the report | Follow new leads |

After a successful project collection, set a schedule under **Settings → Data sources → GitHub → Automatic collection** if needed. The computer and service must remain running. Settings stays at the bottom of the sidebar.

Open the application again from its installation entry point and confirm your project remains. If using ChatGPT, repeat the read-only request. This completes the local persistence and remote-read checks.

## 3. Configuration

### 3.1. Change settings in the workspace

Configuration is stored in `config.toml` at the installation root; the database is normally `Data/state.sqlite`. See [config.example.toml](config.example.toml) for all fields.

| Need | Settings location | When it applies |
| --- | --- | --- |
| ChatGPT connection or local client configuration | MCP connection | Follow the connection check |
| Language, layout and overview sections | Interface & language | After saving |
| GitHub credential source | Data sources → GitHub → GitHub authentication | Service restart |
| Collection interval and download scope | Data sources → GitHub | Interval within one minute; display scope after saving |
| GSC, Bing and Umami endpoints, properties and authorization | Data sources → Search and site analytics | Save, explicitly collect, then check results |
| Reviewer, initial filter and task worker name | Evidence & review; AI discovery | Subsequent operations after saving |
| Local listener and database path | Runtime → Advanced startup configuration | Service restart; reload extension after address changes |

Only loopback IPs are accepted for the listener. A port change synchronizes the extension's endpoint, permissions and connection policy. Database switching accepts an existing Reverfolio SQLite file; saving settings does not move or copy data. Back up first and check that you selected the intended database.

### 3.2. Upgrade, back up and restore

1. Export SQLite through **Settings → Statistics & backup → SQLite backup**. Separately retain `config.toml` and private provider credentials.
2. Stop this installation's service and MCP connection. On Windows, use the executable path in Task Manager to identify its `reverfolio.exe`; for a foreground terminal process, use `Ctrl+C`.
3. Point the native installer at the same installation directory. For a portable upgrade, replace application files while preserving your configuration, database and private credentials.
4. Reopen through the same launcher and reload the browser extension if installed. Confirm works, evidence history and settings before reconnecting MCP.

Restore while the service is stopped: place the backup at a new file path and point `storage.database` at that file. Alternatively, run `reverfolio -restore BACKUP -db NEW_PATH` (`reverfolio.exe` on Windows); the restore destination must not already exist. Do not overwrite the live SQLite file.

### 3.3. Troubleshooting

| Symptom | Check and recover |
| --- | --- |
| Workspace does not open | Launch again from the same installation; check `server.listen` and port conflicts. On Unix, inspect `Data/service.log` |
| Extension disconnected or unauthorized | Rerun the installation launcher and reload that installation's extension; use **Start local service** and **Retry connection** if needed |
| Local connection ready but ChatGPT cannot call tools | Check the same Tunnel in plugin settings, complete enablement/authorization and repeat the read-only check |
| GitHub or site collection fails | Inspect the latest attempt and last successful observation, then check permissions, credentials and network. Failed values are not zeros |
| Data seems missing after an upgrade | Check installation, configuration and database paths; recover the original database before importing again |

For further help, [open an Issue](https://github.com/ThinkerQAQ/reverfolio/issues) with your system, reproduction steps and redacted errors.

## 4. Optional Features

### 4.1. Install the browser extension (Windows)

The installer finish page can open the extension guide. To add it after the main walkthrough:

1. Open Reverfolio once through its installed launcher to initialize local access.
2. Open `edge://extensions` in Edge or `chrome://extensions` in Chrome and enable **Developer mode**.
3. Choose **Load unpacked** and select **the current installation's `extension/` folder**, directly containing `manifest.json`.
4. Click its icon to open the side panel. Confirm **Local service → Connected** and your registered project.

The extension and webpage share data. The browser requires you to load an unpacked extension manually; the installer prepares local authorization and host registration. After moving the installation, rerun its launcher and reload the extension.

### 4.2. Import website articles and DEV.to publications

- Website: open **My works → Import website**, enter your public website or feed URL, preview RSS/Atom/sitemap candidates and select the articles to import.
- DEV.to: open **My works → Creator accounts**, enter your own API key locally, connect, preview published articles, select and associate or import them, then collect metrics.
- After importing: confirm works under **Works** and originals, translations and public channel URLs under **Published links**. Unselected articles are not imported; matching titles do not automatically merge works.

DEV.to is the implemented creator-platform statistics API. Other platforms are not implicitly supported. Neither import path requires Inkerio or SyndiRelay.

### 4.3. Connect website and search analytics

First register the website under **My works**, then open **Settings → Data sources → Search and site analytics**. Choose Umami, GSC or Bing and configure the source ID, associated work, site/property, timezone and observation window. Configure the provider's required credentials locally. Save and enable the source, explicitly collect once, then inspect its status and results.

GSC needs a readable property; Bing needs a verified site. Umami supports a token or a separately authorized local browser session. Saving is not collection success, and browser sign-in is not permission to export cookies. Preserve each source's own click, impression, visitor-identifier and view scopes.

### 4.4. Discover, review and group impact events

Register the work's **Published links** first. Historical names, old URLs, translated titles and distinctive phrases can be added in the advanced area of **Find evidence**. Then give your connected AI client a specific request, for example:

```text
Use Reverfolio to find external impact leads for the work I select.
Read the work, publication links and get_discovery_coverage, then execute its search plan.
Inspect original pages. Distinguish self-publication, third-party citation, reposts,
directory listings, AI-generated content and actual adoption.
Save candidates as pending evidence and record actual queries, sources, dates and
unfinished scopes. Do not automatically approve the evidence.
```

The AI must execute the searches, claim and renew tasks as needed, then link candidate evidence and record results. Creating tasks or plans does not start an autonomous web crawl.

Under **Review evidence → Evidence**, inspect originals, check relevance and provenance, enter a reason and review the record. Under **Impact events**, group related articles, issues and PRs into one event, associate works and evidence, and record submitted, approved, merged or validated stages. Keep outcomes unconfirmed when result evidence is missing. Finally, check event totals and export under **View impact**.

An upstream merge proves contribution acceptance; reposts and directories prove propagation or exposure. None converts to a count of real users. Multiple pages for one event are not multiple adopters, and a scoped search with no result does not prove global absence.

### 4.5. Connect a local MCP client

Open **Settings → MCP connection → Advanced options and local clients**, copy the generated JSON into your local AI client's MCP configuration and reconnect. It uses absolute paths for this installation, avoiding another person's directory layout.

Call `list_works` and `get_collection_status` to confirm it reads the same data. Local stdio needs no ChatGPT Tunnel and does not start a second collection scheduler. Writes need a unique `request_id`; reviews and event updates also require the expected version.

## 5. Development

### 5.1. Build and run

Development needs Git and a Go toolchain satisfying [go.mod](go.mod). Complete repository checks also use Python 3 and Node.js.

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o bin/reverfolio ./cmd/reverfolio
# Initialize only when absent; preserve existing configuration.
test -f config.toml || cp config.example.toml config.toml
./bin/reverfolio --config ./config.toml --collect-interval 0
```

Use these corresponding build, configuration and startup commands in Windows PowerShell:

```powershell
go build -o bin/reverfolio.exe ./cmd/reverfolio
if (!(Test-Path config.toml)) { Copy-Item config.example.toml config.toml }
.\bin\reverfolio.exe --config .\config.toml --collect-interval 0
```

Open the configured listener in your browser and register a test work as in Quick Start. Source Web UI development needs no extension. Build `./cmd/reverfolio-host` separately when developing the Windows native launcher.

### 5.2. Check and package

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
node scripts/check_extension.mjs
node scripts/check_native_host.mjs
```

The MCP fixture uses isolated data; real provider authorization and native installation need separate acceptance. CI covers Go tests, race/vet and Windows, Linux and macOS builds.

Generate portable packages with `python3 scripts/package_release.py vX.Y.Z`, replacing the version argument with the intended release. For native installers, install Qt Installer Framework on the target operating system, inspect `python3 scripts/package_installer.py --help` and use a clean extracted portable package as the stage. Never include live configuration, SQLite or private access files.

## 6. License

Original code and project visual assets use the [MIT License](LICENSE). Third-party code and dependencies retain their licenses, whose text is included in release packages. User writing, photography and provider data are outside the code license.

Protect local accounts, configuration, `Data/` and credentials; keep personal data out of Git and installers. Content explicitly returned over MCP may reach the connected AI service. Report security issues privately to **contact@mail.thinkerqaq.com**, without posting working credentials in public Issues.
