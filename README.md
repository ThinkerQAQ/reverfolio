# Reverfolio

[简体中文](README_ZH.md)

Reverfolio is a local impact tracker for open-source projects, websites and published articles. It records works and publication links, collects available metrics from supported providers, and keeps evidence of citations and adoption for review. One Go service provides a Web UI and MCP tools over the same local SQLite database.

## Contents

- [1. Background](#1-background)
- [2. Quick Start](#2-quick-start)
- [3. Tutorial](#3-tutorial)
- [4. Document](#4-document)
- [5. Contributing](#5-contributing)
- [6. Security and License](#6-security-and-license)

## 1. Background

GitHub projects, personal websites and articles published to multiple platforms expose different metrics. Release downloads, repository traffic, search impressions, pageviews and reactions have different scopes and collection windows. They cannot be summed into a count of real people.

Reverfolio keeps these observations alongside the work and source they belong to. Its goals are to:

- Register projects, articles and websites independently of any publishing system.
- Distinguish a missing metric or failed request from a measured zero.
- Preserve the observation window, provider, collection status and historical baseline.
- Review the original evidence before treating a citation or adoption claim as verified.

Inkerio and SyndiRelay are optional. A fresh installation contains no preconfigured author accounts or publications.

## 2. Quick Start

### 2.1. Download

1. Open [Releases](https://github.com/ThinkerQAQ/reverfolio/releases/latest).
2. Download the archive for your OS and CPU (`windows`/`linux`/`darwin`; `amd64`/`arm64`), together with `SHA256SUMS`.
3. Verify the archive's SHA-256 against the entry in `SHA256SUMS`, then extract it to a permanent, private directory. Keep the extracted `Reverfolio` directory intact.

For the hash, use `Get-FileHash .\Reverfolio-*.zip -Algorithm SHA256` on Windows, `sha256sum ./Reverfolio-*.tar.gz` on Linux, or `shasum -a 256 ./Reverfolio-*.tar.gz` on macOS. Compare the result with the **matching archive entry** in `SHA256SUMS`; stop if it differs.

Portable releases include the executable and configuration template. Go is not needed to run a release.

### 2.2. Install and run

**Windows (PowerShell in the extracted `Reverfolio` directory):**

```powershell
.\Start-Reverfolio.cmd
```

This launcher runs the installer when needed, initializes `config.toml` and `Data/` without overwriting existing data, and starts the local service. The installer also initializes the optional extension credentials and registers its Native Messaging host; **do not register the host manually**. Open [http://127.0.0.1:19452](http://127.0.0.1:19452).

**Linux / macOS (terminal in the extracted `Reverfolio` directory):**

```sh
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

Copy the template **only on a new installation**. Keep the terminal running and open [http://127.0.0.1:18440](http://127.0.0.1:18440). Unix packages do not install a background service. On macOS, the unsigned binary may require OS approval.

### 2.3. Register your first work

1. In **Projects**, select **Add a project or article**.
2. Choose **GitHub project**, enter a title and an owned repository in `owner/name` format, then save.
3. Refresh **Projects** and confirm the work is still present.
4. Select **Collect GitHub** for that work. Check the recorded collection status and timestamp. Public release counts may be available without credentials; private traffic requires an authorized GitHub credential.
5. Restart Reverfolio using the same `config.toml`. Confirm that the project is still present; this verifies SQLite persistence.

The first cumulative metrics snapshot is a baseline, not growth. No Releases, missing permissions and provider failures may produce unavailable fields; those are not zero measurements.

### 2.4. Verify installation

From a second terminal, check the corresponding endpoint:

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```sh
# Linux / macOS
curl -fsS http://127.0.0.1:18440/healthz
```

Acceptance requires `status=ok`, a persisted work after restart and an explicit success/partial/failure outcome for any collection attempted. The health endpoint alone does not verify collection or persistence.

## 3. Tutorial

These cases are optional and independent. Start with the Quick Start before connecting external data.

### 3.1. Case 1: Import articles from your own website

1. Open **Discover websites** and enter the public website or feed URL you control.
2. Preview discovered RSS, Atom or Sitemap candidates. Check the titles and URLs; discovery does not itself register a publication.
3. Select the articles you want, then import them.
4. Open **Projects** and **Channels** to confirm the imported works and publication URLs.

Only explicitly selected URLs are imported. A feed or sitemap might not list every article. Matching titles do not automatically merge two works.

### 3.2. Case 2: Connect DEV.to

1. Obtain your own DEV.to API key and enter it in the local **Creator platforms** UI.
2. Connect the account and preview its published articles.
3. Select and link the articles to existing works, or import the selected publications.
4. Run the DEV.to collection action, then inspect **Creator platforms** and its metric history.

The key is stored locally and is not accepted by an MCP enrollment tool. DEV.to is the implemented creator-platform API integration; other creator-platform accounts are not automatically supported.

### 3.3. Case 3: Connect website and search analytics

1. Register the website as a work in **Projects**.
2. Open **Settings → Data sources → Search and site analytics** and choose Umami, Google Search Console (GSC) or Bing Webmaster.
3. Enter the provider-specific site/property ID, timezone, observation window and an authorized credential. GSC requires a property you can read; Bing requires a verified site and its API key. Umami supports a token or a separately authorized local browser session.
4. Save and explicitly enable the source, then start collection.
5. Review the source status and observations. Saving a source does not mean that any metrics were collected.

The providers return different measurements; keep their scopes separate. Browser sign-in alone is not permission to export cookies or extract credentials.

### 3.4. Case 4: Review citations and propagation

Register confirmed publication links under **Channels**. Use **Evidence & adoption** to submit a possible citation, inspect the original page/API evidence, and explicitly approve or reject it. A secondary report alone is not verified original evidence. **Discovery tasks** record searches for an AI client to perform; creating a task does not execute a web search. History and propagation views distinguish pending evidence from reviewed results.

## 4. Document

### 4.1. Architecture and storage

The application comprises a Go service, an embedded Web UI, a local SQLite database and an MCP server using the same application operations. Its HTTP interface binds to loopback only; the MCP server is also available through local stdio. The Windows browser extension shows the same management interface and can wake the service through the installed Native Messaging host.

User works, settings, observations, evidence and audit records stay in the local database. External provider requests require the credentials and access that the user explicitly configures.

### 4.2. Configuration

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

### 4.3. MCP

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

### 4.4. Windows extension

After running the Windows installer, open `edge://extensions` or `chrome://extensions`, enable **Developer mode**, and **Load unpacked** from the installation's `extension/` directory. Clicking the icon opens the shared dashboard; the registered Native Messaging host can start the local service if needed.

Browser security requires the user to load the unpacked extension; it does **not** require a separate manual Native Messaging registration. Reload the extension after an update. Unix users can use the Web UI and stdio MCP without the extension.

### 4.5. Backup, upgrade and troubleshooting

- **Before upgrading:** back up `config.toml`, `Data/` and any private provider credentials. Reinstall into a stable directory without replacing these files.
- **Empty workspace after upgrading:** check the config path and `storage.database`. Do not start a second writable database in WSL.
- **Missing GitHub traffic / 401 / 403:** verify repository access and the configured credential. Unavailable data is not zero.
- **Extension cannot wake the server:** rerun `Start-Reverfolio.cmd` from the installed directory and inspect the local health endpoint.
- **MCP still shows old tools:** reconnect the client and start a new session after updating.
- **Offline restore:** `reverfolio -restore BACKUP -db NEW_PATH` restores into a **new** database path, not over the live database.

## 5. Contributing

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

## 6. Security and License

Reverfolio is local-first, **not** a publicly hosted API. Its loopback HTTP API and MCP require their own credentials; the Windows extension uses a separate installation-specific local capability. Protect your OS account, `Data/`, SQLite, private configuration and tokens. Do not commit them or include user content, browser profiles or reports in release archives. MCP responses explicitly requested by a connected client may be transmitted to that client or its AI service.

Report security issues privately to **contact@mail.thinkerqaq.com**, without publishing private data or working credentials.

Original source code and project visual assets use the [MIT License](LICENSE). User writing, photographs, provider data and third-party content are not covered. Dependencies retain their upstream licenses: `go.mod`/`go.sum` record the module graph, and portable releases contain `THIRD_PARTY_LICENSES.txt` with upstream license/notice text. This file is generated and validated during packaging.
