# Windows portable layout

Install Reverfolio in C:\software\Coding\reverfolio\.

The folder spelling "reverfolio" is intentional. The product display name remains Reverfolio.

Folder structure:

    reverfolio/
      bin/
        reverfolio.exe
      extension/
        manifest.json
        icons/
        background.js
        ui/
        _locales/
      Data/
        state.sqlite
        api-token
        extension-access.key
      Initialize-ReverfolioExtension.ps1
      config.toml
      Start-Reverfolio.cmd
      Run-Reverfolio.ps1

- bin/: Windows executable with the Reverfolio icon, not personal data.
- extension/: optional unpacked Edge/Chrome extension running the full native Reverfolio dashboard as a browser side panel or tab, sharing the same UI bundle with the local Web page. Load the extension folder from edge://extensions (Developer mode). It does not read SQLite or TOML.
- Data/: private local SQLite database and automatically generated server API token. Never overwrite on upgrade.
- config.toml: local TOML configuration for startup and nonsecret UI preferences. It is NOT a SQLite table; never commit it.
- config.example.toml: tracked reference template; safe to share.
- Start with Start-Reverfolio.cmd. Browser UI requires no token, while MCP/API remain authenticated.
- The Windows config uses 127.0.0.1:19452 to avoid collision with Ubuntu/WSL on port 19450. The old WSL application is inactive.
- Windows is the ONLY active business data authority. Legacy WSL rows were checked (identical), backed up and disabled. The native Windows tunnel launches Windows Reverfolio directly. See [Windows authority](windows-authority.md).
- Secrets are not placed in TOML; Data/api-token is generated when the Windows app starts.
- github.auth = "env" uses an explicitly supplied GITHUB_TOKEN; "gh" uses the authorized native GitHub CLI. Never copy credentials into source or installation templates.
- The service reloads collection.interval within one minute; listener changes require restart. UI preferences take effect without restart.
- Provider configuration is nonsecret TOML. Provider secret files remain private under Data/provider-credentials; source status checks report presence without reading those files. See [provider setup](site-sources.md).
- Cross-compile with GOOS=windows GOARCH=amd64 go build -o bin/reverfolio.exe ./cmd/reverfolio.
- Reverfolio has no dependency on any publishing tool or browser extension.

The extension uses exactly one local Windows origin (127.0.0.1:19452) with a pinned Chrome extension identity. It does not auto-fallback to WSL or connect to MCP. v0.1 unpacked users must remove and re-add the directory to register the pinned v0.2 identity; later updates only need Reload. An installation folder is NOT an installed browser extension until the user loads it via edge://extensions.

The full management extension uses a separate, randomly generated local-only capability stored in Data/extension-access.key and mirrored into extension/ui/local-auth.js. Start-Reverfolio.cmd auto-provisions this through Initialize-ReverfolioExtension.ps1 before the server launches. Neither this key nor the general API/MCP token is committed. The public repository has only an empty local-auth.js template; never share a populated installation folder.
