# OpenEcho Security

OpenEcho is a loopback-only, personal-use influence-monitoring service. Keep the database, screenshots, imported reports, browser profiles, and credential files private. Report issues privately to contact@mail.thinkerqaq.com.

## Local web app

- The local web UI uses a process-scoped HttpOnly SameSite=Strict session cookie and an explicit header for write requests.
- The general HTTP API and MCP accept a separate Bearer credential from Data/api-token. The MCP tunnel has its own transport configuration. Do not expose either publicly.
- The server binds only to a loopback IP and validates the incoming Host to reduce DNS rebinding risks.

## Browser extension v0.2

- The extension renders the same management UI code as the local Web app, but its own chrome-extension:// origin cannot reuse same-site browser cookies.
- Chromium may omit the Origin header even for legitimate extension requests. The server therefore REQUIRES a separate random, installation-specific key for EVERY extension API request. It verifies a suitable Fetch Metadata mode and either the pinned origin or extension-host-permission transport.
- Windows installation creates Data/extension-access.key locally and mirrors that key ONLY to extension/ui/local-auth.js; neither key belongs in Git. The public repository contains only an empty, nonworking template.
- The extension-specific key has access ONLY to local management routes, not /mcp. It is not the GitHub credential, the general API Bearer key, or the MCP tunnel credential.
- Extension permissions are restricted to sidePanel and a single 127.0.0.1 host permission. There are no content scripts, cookies permissions, third-party website permissions, remote CDN scripts or data uploads.
- Do not share a populated extension folder or commit the generated local-auth.js.
- Any attacker with access to your Windows account and private application files can potentially impersonate a local UI. Protect that account with normal OS access controls and do not move this service off loopback.

## Release and maintenance

- Make SQLite backups before upgrades. Credentials, private database files, profiles, logs and user-provided evidence must not be committed to GitHub.
- Keep binary dependency licenses and notices for every packaged release.
- CI currently has account-level constraints; successful local checks are not equivalent to successful hosted CI.
