# Reverfolio MCP: setup, tools, acceptance and troubleshooting

Reverfolio's Web UI, browser extension and MCP share one Go service and one local SQLite database. MCP is the **business API**, not a second storage service or the Chromium Native Messaging launcher.

## 1. Connect a client

For a local MCP client, initialize your own configuration and use stdio:

```bash
./bin/reverfolio -mcp-stdio -config ./config.toml
```

On Windows use `bin\reverfolio.exe -mcp-stdio -config config.toml` with full local paths if necessary. The stdio process does not run a second scheduled collector; it opens the configured SQLite database. Requested write tools may still modify it.

ChatGPT cannot reach another person's `127.0.0.1` directly. A remote ChatGPT integration must use an **individually authenticated** MCP connection/tunnel to the user's own installation. Never reuse the author's tunnel credentials, IDs or SQLite files. The developer's [Reverfolio ChatGPT app](https://chatgpt.com/plugins/plugin_asdk_app_6ae5d8510c1c819187e96e9290753815) is currently **Private/Development**, *not* a public one-click plugin.

After a server upgrade, if the chat is still missing tools, open **Plugins → Reverfolio → Manage → Refresh tools**, then start a **new chat**. Old chats may retain their previous tool snapshot. Verify the read-only `get_dashboard` and `list_site_sources` tools there.

## 2. The complete 34-tool MCP contract

| Domain | Tools |
|---|---|
| Works and catalog (11) | `list_works`, `get_work`, `upsert_work`, `add_channel`, `list_catalog`, `set_monitoring`, `import_catalog`, `export_catalog`, `list_website_sources`, `preview_website`, `import_website` |
| GitHub and site analytics (7) | `collect_source`, `get_trends`, `get_dashboard`, `get_collection_status`, `list_site_sources`, `save_site_source`, `collect_site_source` |
| Citations and identity (6) | `submit_mention`, `list_mentions`, `review_mention`, `link_propagation_event`, `link_identity`, `unlink_identity` |
| AI discovery tasks (4) | `create_discovery_task`, `list_discovery_tasks`, `claim_discovery_task`, `complete_discovery_task` |
| Settings, history and reports (6) | `list_settings`, `update_setting`, `update_settings`, `import_history`, `get_audit`, `generate_report` |

All writes require unique `request_id` values; several also use `expected_version` for optimistic concurrency. Preview imports with `dry_run` and explicitly selected work IDs. Review candidate evidence against original web/API sources before marking it verified.

## 3. First-use workflow

1. Discover any public blog using RSS/Atom/Sitemap without a publisher dependency: [website discovery guide](website-discovery.md). Preview candidates and explicitly select works before import.
2. Register GitHub projects with `kind=project` and `repository=owner/name`; register articles with `kind=article` and exact public `canonical_url`.
3. Attach actual translations and syndication URLs using `add_channel`. Optional SyndiRelay integration transfers only confirmed public publication metadata; it is not a required runtime dependency.
4. Configure *your* GitHub CLI or service environment credentials. Add independent Umami, Google Search Console and Bing sources using [site-source instructions](site-sources.md), including property IDs, authorization and time scope.
5. Collect data through `collect_source` or `collect_site_source` and read `get_dashboard`/`get_trends`. Preserve provider outages, missing days and observation windows. Historical successful reads do not prove current credentials are still available.
6. An AI client may search the web with **its own approved tools**, then use discovery-task and mention tools to record evidence for review. Reverfolio does **not** run a built-in autonomous web-crawling agent.
7. Use `generate_report` and `get_audit` for a traceable Markdown report. Release downloads are not installations or unique users; uniqueness across providers must not be inferred.

## 4. Check the live tool contract

From source:

```bash
# Builds against an isolated, throwaway SQLite database. No network collection.
python3 scripts/check_mcp_contract.py --fixture
```

On an installed Windows instance (read-only):

```powershell
python scripts/check_mcp_contract.py --binary "C:\software\Coding\reverfolio\bin\reverfolio.exe" --config "C:\software\Coding\reverfolio\config.toml"
```

The checker performs a real MCP handshake and requires the **exact 34-tool inventory**. It reads dashboard/provider/runtime/report data without printing credentials or raw reports. Fixture mode exercises work/channel registration, task leasing, evidence submission/review, reporting, catalog export and import preview through MCP.

## 5. Security and portable deployment

- Keep `Data/`, `config.toml`, generated browser capabilities, credentials, personal reports and original source data **out of Git and release archives**.
- The developer’s Windows installation at `C:\software\Coding\reverfolio` uses one native `Data/state.sqlite`; there is no automatic WSL fallback. Other users can choose their own installation path and must initialize their own config and database.
- Native Messaging starts the browser's **local service** silently, but never returns API keys or SQLite content. It is not the MCP protocol.
- The local HTTP `/mcp` endpoint requires a bearer token and binds to loopback. Remote connections need their own secured bridge/tunnel—never publicly expose the loopback port directly.
- The public plugin-directory review process is separate from shipping self-hosted MCP source and installers.
