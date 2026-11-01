# Reverfolio v0.7.1

MCP completeness and developer/self-hosted onboarding patch (no SQLite schema changes).

- Verified the real Windows stdio MCP handshake: **31 tools**, including `get_dashboard`, `list_site_sources`, `collect_site_source`, directory import/export, evidence, identity, settings and audit.
- Refreshed the developer ChatGPT app's tool registry and confirmed `get_dashboard` and `list_site_sources` execute successfully in a **new ChatGPT conversation**; existing conversations may retain their previous tool snapshot.
- Added `scripts/check_mcp_contract.py`, a no-secrets checker with an isolated SQLite end-to-end fixture and a read-only installed-instance mode. Included in release archives.
- CI now checks the exact 31-tool inventory, a real MCP subprocess handshake and a fixture evidence-to-report workflow on Windows, Linux and macOS.
- Confirmed GSC and Bing credentials are currently loaded in the author's private installation. Umami has historical successful observations but **no loaded credential now**; freshness requires a new successful authorized collection.
- Added bilingual [MCP guide](mcp-setup.md) and [中文指南](mcp-setup.zh-CN.md), including first-run setup, scope, source attribution, plugin tool refresh, per-user credentials and security.

The author's personal ChatGPT app remains Private/Development; release archives are self-hostable software, not a public one-click ChatGPT app install. Do not ship or restore anyone's `Data/`, `config.toml`, extension local access key, tunnel credentials or private provider files.

This patch retains Windows Native Messaging silent startup, the propagation dashboard, and all v0.7.0 behavior.
