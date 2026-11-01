# Reverfolio v0.7.0

Adds a dedicated, bilingual Propagation paths workbench for works, channels and verified evidence.

- Four-stage visual map: canonical article/project → first-party translations → external distribution URLs → deduplicated, reviewed third-party evidence.
- Work selector isolates every link and evidence event to the selected work.
- Genuine evidence timeline uses first observed/discovered and review timestamps only. Channel publish dates are not available, so none are inferred.
- Fully shared Web UI and Chromium side-panel extension.
- Works with the optional, independently maintained SyndiRelay local catalog handoff, with no new dependency or data migration in Reverfolio.

No schema changes, API key changes or personal data migration. On Windows, retain `Data/`, `config.toml`, extension `ui/local-auth.js`, existing extension ID, and the local MCP tunnel.

## Quiet on-demand startup (Windows)

- The browser extension probes the local API first. When down, a pinned and registered Native Messaging host silently starts the Windows service. An offline page offers **Start local service** and **Retry connection**.
- Windows installer registers the native host for Edge and Chrome, and rewrites only legacy direct-console scheduled task actions to a silent `wscript.exe` launcher. Existing triggers and user data remain intact.
- A standalone Windows GUI-subsystem native host cannot create a visible console and never reads or returns local bearer tokens or SQLite data.
- Existing Linux/macOS services and local Web UI remain independently runnable. No remote service or WSL fallback is introduced.
