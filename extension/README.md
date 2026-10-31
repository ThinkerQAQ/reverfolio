# OpenEcho Extension (Edge / Chrome)

OpenEcho's optional Manifest V3 browser companion is an independent launcher, not a BlogCTL integration.

## Install locally

1. Ensure the OpenEcho application is running.
   - Windows: run C:\software\Coding\opencho\Start-OpenEcho.cmd; then check http://127.0.0.1:19452/.
   - Ubuntu/WSL: the configured service is typically http://127.0.0.1:19450/.
2. In Microsoft Edge, open edge://extensions; in Chrome, open chrome://extensions.
3. Enable Developer mode.
4. Click Load unpacked and select **the extension folder containing manifest.json**:
   - Windows: C:\software\Coding\opencho\extension
   - Source checkout: (repository)/extension
5. Pin OpenEcho to the toolbar. Click its icon to see service status and quick links.
6. Select Auto-detect (Windows -> WSL -> generic), or select a specific local instance.

To update, replace the extension files, open the extensions management page, and click Reload on the OpenEcho card.

## Features

- Dedicated blue echo-wave icon for browser toolbar, website favicon and Windows executable.
- Read-only online/offline check of the local unauthenticated GET /healthz endpoint.
- Links to the seven management pages and optional browser side panel of shortcuts.
- Simplified Chinese and English, with extension-specific browser-local preferences.
- Detects three loopback ports: Windows 19452, WSL 19450, generic 18440.
- No account Cookie access, third-party site access, page injection, nativeMessaging, GitHub tokens, MCP credentials, SQLite access or data uploads.

The extension does NOT replace the full management UI or grant access to protected APIs. Work with data in the local page it opens.

## Permissions

- storage: remembers only extension language and chosen local instance.
- sidePanel: opens its own shortcut UI in a side panel.
- host_permissions: only the three explicit loopback HTTP origins for health checks.

OpenEcho runs independently of the extension. Never expose the local server publicly just to install it.

Chinese instructions: README.zh-CN.md
