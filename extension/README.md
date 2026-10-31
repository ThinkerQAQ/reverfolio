# Reverfolio browser extension (Edge / Chrome)

The browser extension and the local web app now use the SAME management UI. Clicking the Reverfolio toolbar icon opens the full dashboard in the browser side panel, not a separate launcher, services dropdown or shortcuts screen. The header button opens that same dashboard as a full browser tab.

## Installation

1. Start Reverfolio on Windows using C:\software\Coding\reverfolio\Start-Reverfolio.cmd, and verify http://127.0.0.1:19452/.
2. Run Initialize-ReverfolioExtension.ps1 once from the Reverfolio install folder, or start the app with Start-Reverfolio.cmd (which now runs provisioning automatically). This creates a random, installation-specific Data/extension-access.key and its extension/ui/local-auth.js copy. There is no manual token field.
3. In Edge visit edge://extensions (Chrome: chrome://extensions), enable Developer mode and click Load unpacked.
4. Select C:\software\Coding\reverfolio\extension (the folder containing manifest.json).
5. Pin Reverfolio and click its toolbar icon. The FULL dashboard opens in the side panel, with projects, channels, citations, search tasks, historical statistics and recursive settings.
6. To use a large screen, click the open-in-tab icon in the dashboard header.

UPGRADING from the original v0.1 popup: v0.2 has a pinned public key and a stable extension ID. Remove the previous unpacked extension, then load this directory again so the new identity is registered correctly. Later updates only require Reload on the extension card.

## Security and runtime

- One local service: http://127.0.0.1:19452. No service auto-detection or Windows/WSL priority selector.
- Browser management requires a unique random, install-specific local UI capability generated under Data/extension-access.key. The matching extension/ui/local-auth.js remains local and MUST NOT be committed or shared. It is independent of GitHub credentials, MCP tokens and the general API bearer token.
- The same native Reverfolio UI/CSS is mirrored into the extension by python3 scripts/build_extension_ui.py.
- sidePanel is the only browser API permission. Host access is limited to 127.0.0.1:19452.
- Unrelated extensions and websites cannot use the pinned browser origin. Programmatic API/MCP still require a bearer token.
- Windows config.toml remains the authoritative application configuration and Data/state.sqlite remains the authoritative local data store.
- If the service is not running, the dashboard shows a connection error instead of silently accessing another database.

Chinese instructions: README.zh-CN.md
