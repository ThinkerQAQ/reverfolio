# Windows as the only OpenEcho data authority

Source of truth: C:\software\Coding\opencho\Data\state.sqlite.
Configuration: C:\software\Coding\opencho\config.toml.

Browser UI, Chromium extension and ChatGPT MCP all use native Windows Go
processes against the same SQLite database.

## Verified migration from the former WSL copy

On 2026-10-31, all 12 application tables were checked row by row between
Windows and the former WSL SQLite. The 12 sets of rows are identical.
No SQL import or upsert was needed, avoiding duplicated evidence or metrics.
Both copies were backed up before changing services.

The legacy WSL service openecho.service was disabled and stopped. Its database
remains for rollback but is not used by an active service.

The existing openecho-tunnel.service retains WSL only as a network transport:
its MCP command launches native Windows opencho.exe via WSL Windows interop,
using Windows config.toml and the -mcp-stdio flag with collection disabled.
It never references the former WSL SQLite file. Verified with the official
Go MCP SDK (20 tools) and a live OpenEcho MCP list_works call (31 works).

This is NOT cross-OS access to a SQLite file: only native Windows processes
open the Windows database and SQLite locking is handled by Windows.

## Endpoints and limitations

- Windows Web UI: http://127.0.0.1:19452.
- Windows also serves authenticated /mcp for local programmatic access.
- WSL 19450 is intentionally inactive.
- No fallback to WSL data. If Windows is off, MCP cannot access OpenEcho data.
- Windows GitHub automatic collection remains disabled until authorized.

## Backups and credentials

- Windows backups: C:\software\Coding\opencho\Data\backups.
- The extension has a separate local access key under Data/extension-access.key,
  mirrored only into the local installed extension/ui/local-auth.js.
- GitHub and MCP/API tokens are separate. Never commit private runtime files.
- Release packages only include an EMPTY local-auth.js template; the Windows
  setup script provisions the installation-specific secret.
