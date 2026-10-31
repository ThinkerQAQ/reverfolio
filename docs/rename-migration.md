# Reverfolio rename migration

OpenEcho is now Reverfolio. Repository: https://github.com/ThinkerQAQ/reverfolio.

- Windows authority: `C:\software\Coding\reverfolio\Data\state.sqlite`.
- Existing `config.toml`, SQLite rows, backups and local credentials are retained.
- The former `C:\software\Coding\opencho` path remains a junction for old browser extension registrations and shortcuts.
- The browser extension public key/ID and loopback port 19452 are unchanged.
- `REVERFOLIO_API_TOKEN` takes precedence; `OPENECHO_API_TOKEN` remains accepted.
- CLI installs without TOML reuse `data/openecho.sqlite` when it exists and the new default does not.
- Browser language preferences fall back to the previous storage key.
- The MCP server identifies as Reverfolio; existing tunnel ID is preserved.
- Historical release v0.2.0 remains an OpenEcho artifact; renamed packages start at v0.2.1.
