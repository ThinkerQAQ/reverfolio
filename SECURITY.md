# Reverfolio Security

Reverfolio is a loopback-only, **self-hosted** project and content-impact service. Users keep their SQLite database, screenshots, imported reports, browser profiles, and credential files on their own machines. Report vulnerabilities privately to **contact@mail.thinkerqaq.com**; never publish working token values, private databases, or reproduction steps that expose another user's data in an Issue.

## Local web app

- The local web UI uses a process-scoped HttpOnly SameSite=Strict session cookie and an explicit header for write requests.
- The general HTTP API and MCP accept a separate Bearer credential from Data/api-token. The MCP tunnel has its own transport configuration. Do not expose either publicly.
- The server binds only to a loopback IP and validates the incoming Host to reduce DNS rebinding risks.

## Edge/Chrome extension (optional, Windows portable setup)

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
- Release builds include the original upstream dependency license texts in `THIRD_PARTY_LICENSES.txt`; their presence is checked during packaging. Security review and platform-specific smoke tests remain separate from license compliance.
- Linux, Windows and macOS GitHub Actions tests plus documentation checks gate protected `main`. Passing CI is not equivalent to proving runtime behavior on every supported machine.
- DEV.to API keys are saved in a per-installation private file in `Data/platform-credentials`, and are entered only through the local UI; the MCP tool set cannot accept a raw enrollment key. Use Windows directory ACLs or POSIX owner-only file permissions to protect local data.
