# OpenEcho v0.2.0

First portable, cross-platform release of OpenEcho: a personal, evidence-backed
open-source and writing impact observatory.

## Included

- GitHub repository registry, stars/forks, release asset download counters,
  limited traffic snapshots, per-item data source notes and historical import.
- Project/article/site works and bilingual distribution channel links.
- Third-party mention and adoption candidate review, source attribution,
  deduplication, audit history, and AI-assisted discovery-task tracking.
- Native SQLite persistence, consistent backups, and idempotent collection.
- Full Simplified Chinese/English local Web interface, hierarchical TOML
  settings, project metric cards and an inspectable raw-statistics view.
- An independent Edge/Chrome side-panel extension that opens the same full UI
  and works with the Windows local service. No BlogCTL dependency.
- Original logo assets and Windows executable icon.
- Windows portable first-run setup, automatic local extension authorization
  and data preservation on upgrades.
- Linux, Windows and macOS binaries for amd64 and arm64.
- Go MCP SDK integration, HTTP and stdio MCP transports, 20 available tools.

## Windows single source of truth

The personal deployment uses Windows at C:\software\Coding\opencho,
with a single Data/state.sqlite database shared by native Windows Web and
MCP processes. The legacy WSL application was stopped and disabled after
a row-by-row comparison of both databases (12 identical tables). WSL is
used ONLY for the existing OpenAI Tunnel transport, which launches the
native Windows executable using Windows interop. No cross-OS SQLite file
access and no automatic data fallback.

## Download and start

For Windows: download OpenEcho-v0.2.0-windows-amd64.zip (most common x64)
or the arm64 variant. Extract into your own dedicated folder. Run
Start-OpenEcho.cmd to create missing config.toml and Data safely; it will
not overwrite existing configuration or SQLite. The default local URL is
http://127.0.0.1:19452/. Load the extension folder through Edge
edge://extensions > Developer mode > Load unpacked. If an old OpenEcho
unpacked extension is present, remove it and re-add the folder once because
v0.2 pins a stable extension ID.

For macOS or Linux: extract the matching tar.gz archive, copy
config.example.toml to config.toml, adjust listen and database for your
machine, and run bin/openecho -config config.toml. No extension is required
to use the native local Web interface or stdio MCP.

SHA256SUMS accompanies all six archives. Windows packages include the
extension, a private-key-free configuration template and PowerShell scripts.

## Known limitations

- GitHub asset download counts are cumulative. Downloads and traffic identifiers
  do not prove unique users, installs, article readers or software adoption.
- Live Windows GitHub collection is initially disabled until an authorized local
  GitHub token is configured; the present dataset contains verified historical
  snapshots from a prior collection.
- GSC, Bing and Umami historical records are not real-time adapters yet.
- AI-assisted discovery is a review workflow, not autonomous full-web coverage.
  Most generated search tasks are pending.
- Windows/macOS binaries were cross-compiled. Windows x64 startup, private
  SQLite, extension and browser interactions were smoke-tested on native
  Windows; macOS runtime and Windows ARM64 runtime were not directly tested.
- GitHub-hosted CI jobs were blocked before executing their steps. Independent
  Go tests/race/vet, frontend checks, six cross-builds, native Windows browser
  tests and package checksum verification passed locally.

## Safety

No user SQLite files, historical reports, account credentials, extension keys,
browser profiles or TOML machine settings are included in released archives.
The local service binds to loopback, while MCP and programmatic HTTP access
remain separately authenticated. See SECURITY.md.
