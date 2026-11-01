# Getting started (Reverfolio v0.9.x)

Reverfolio is **public, MIT-licensed, local-first software**. You can run it without Inkerio, SyndiRelay, a hosted account, or an API key. Metrics from individual providers require your own authorization. Your SQLite database and credentials stay on your computer.

## Download a portable release

Download the matching archive from the [latest GitHub release](https://github.com/ThinkerQAQ/reverfolio/releases/latest):

| Operating system | File | Startup |
|---|---|---|
| Windows x64 | `Reverfolio-v*-windows-amd64.zip` | Extract; run `Install-Reverfolio.ps1`, then `Start-Reverfolio.cmd` |
| Windows ARM64 | `Reverfolio-v*-windows-arm64.zip` | Same steps |
| Linux x64 / ARM64 | `Reverfolio-v*-linux-{amd64,arm64}.tar.gz` | Extract; copy the example TOML; run `./bin/reverfolio -config ./config.toml` |
| macOS Intel / Apple Silicon | `Reverfolio-v*-darwin-{amd64,arm64}.tar.gz` | Same CLI steps; macOS may require approving an unsigned downloaded binary |

Verify downloads against the `SHA256SUMS` file published with the assets before extracting. **No API keys, private data, or predefined content are bundled**. `THIRD_PARTY_LICENSES.txt` contains original dependency license notices.

### Windows (native)

Extract the archive into a permanent folder you control, such as `C:\software\Coding\reverfolio`. In PowerShell opened at that extracted `Reverfolio` directory:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Install-Reverfolio.ps1
.\Start-Reverfolio.cmd
```

Open <http://127.0.0.1:19452>. On a fresh install, the installer creates `config.toml`, `Data/` and extension authorization. On upgrades, it **does not overwrite existing TOML, SQLite, or API tokens**. Keep the directory; it is your local data home.

**Optional Edge/Chrome extension:** enable Developer mode at `edge://extensions` or `chrome://extensions`, choose **Load unpacked** → the extracted `extension/` folder. It provides the same local UI in a side panel and uses an installation-specific local authorization key. The Windows installer registers the native host for silent startup. See [extension setup](../extension/README.md).

### Linux and macOS (standalone web service)

From the directory containing the downloaded archive:

```bash
# Replace the archive name with the exact downloaded filename.
tar -xzf Reverfolio-v*-linux-amd64.tar.gz
cd Reverfolio
cp -n config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

For macOS use the `darwin-amd64` or `darwin-arm64` archive instead of `linux-amd64`; for ARM64 Linux choose the `linux-arm64` archive. Open <http://127.0.0.1:18440>. This command runs the web service in the foreground; exiting the process stops the service. For ongoing background use, configure your own systemd/launchd service (not supplied by the portable archives).

The Windows Native Messaging extension wake-up workflow **is not provided as a one-click macOS/Linux installer**. The web service and local stdio MCP work without it. Do not assume Chrome will automatically launch the service on those operating systems.

### From source (contributors)

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go run ./cmd/reverfolio -db data/reverfolio.sqlite
```

Requires Go 1.26+; open <http://127.0.0.1:18440>. See [CONTRIBUTING](../CONTRIBUTING.md).

## Your first works and statistics

1. **Add a work** in Projects, or open **Website discovery** and preview a public RSS, Atom, or Sitemap. Confirm which article URLs you actually own. No publishing tool is required.
2. **GitHub:** add repositories. Private traffic metrics need an authorized `GITHUB_TOKEN` with appropriate permissions or, if you choose, local `gh` CLI authorization. Public release downloads are not proof of actual users.
3. **DEV.to:** in **Creator platforms**, enter a personal API key locally. Preview your *own* published articles, explicitly link them to works, then collect sourced, timestamped followers/views/reactions/comments. Keys are not transmitted to MCP and are never packaged.
4. **Website analytics:** configure the supported Umami, Google Search Console, or Bing Webmaster providers only if they apply to your own properties. Unavailable metrics remain missing, not zero.
5. **ChatGPT/MCP (optional):** your local process supports `-mcp-stdio`; remote ChatGPT access requires **your own secured MCP connection/tunnel**. The developer's private ChatGPT plugin is not a public one-click installer. See [MCP setup](mcp-setup.md).

Reverfolio **does not** currently support automatic platform-account polling, reliable identification of real people from followers or downloads, or private analytics for platforms that offer no authorized official API. Supported account metrics currently use DEV.to's official API. Other platforms must be evaluated individually, not scraped via undocumented endpoints.

## Upgrade and recovery

Back up `Data/state.sqlite` (or use the app's SQLite backup action) and private credentials before replacing executable/extension files. On Windows the installation script is idempotent but **does not migrate or overwrite private data files**. Reload an unpacked browser extension after updating files. A process restart may be required to pick up the new binary. For rollback, restore **a compatible executable and its backed-up SQLite together**, because schema upgrades may not be readable by older versions.

See [security](../SECURITY.md), [licensing](../THIRD_PARTY_NOTICES.md), and [current release notes](https://github.com/ThinkerQAQ/reverfolio/releases/latest).
