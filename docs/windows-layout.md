# Windows portable layout

Install OpenEcho in C:\software\Coding\opencho\.

The folder spelling "opencho" is intentional. The product display name remains OpenEcho.

Folder structure:

    opencho/
      bin/
        opencho.exe
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
      Initialize-OpenEchoExtension.ps1
      config.toml
      Start-OpenEcho.cmd
      Run-OpenEcho.ps1

- bin/: Windows executable with the OpenEcho icon, not personal data.
- extension/: optional unpacked Edge/Chrome extension running the full native OpenEcho dashboard as a browser side panel or tab, sharing the same UI bundle with the local Web page. Load the extension folder from edge://extensions (Developer mode). It does not read SQLite or TOML.
- Data/: private local SQLite database and automatically generated server API token. Never overwrite on upgrade.
- config.toml: local TOML configuration for startup and nonsecret UI preferences. It is NOT a SQLite table; never commit it.
- config.example.toml: tracked reference template; safe to share.
- Start with Start-OpenEcho.cmd. Browser UI requires no token, while MCP/API remain authenticated.
- The Windows config uses 127.0.0.1:19452 to avoid collision with Ubuntu/WSL on port 19450. They are separate local instances.
- Windows is the ONLY active business data authority. Legacy WSL rows were checked (identical), backed up and disabled. The MCP tunnel launches Windows OpenEcho via Windows interop. See [Windows authority](windows-authority.md).
- Secrets are not placed in TOML; Data/api-token is generated when the Windows app starts.
- github.auth = "env" is the safe no-token default on this Windows host. The signed-in Windows GitHub CLI has no available OAuth token; use Windows gh auth login and switch to gh when authorized. Until then, collection.interval = "0" avoids pretending repository traffic has been collected. No credentials are copied.
- Restart the app to apply startup settings such as listener or interval. UI preferences update TOML and take effect without restart.
- Cross-compile with GOOS=windows GOARCH=amd64 go build -o bin/opencho.exe ./cmd/openecho.
- OpenEcho has no dependency on any publishing tool or browser extension.

The extension uses exactly one local Windows origin (127.0.0.1:19452) with a pinned Chrome extension identity. It does not auto-fallback to WSL or connect to MCP. v0.1 unpacked users must remove and re-add the directory to register the pinned v0.2 identity; later updates only need Reload. An installation folder is NOT an installed browser extension until the user loads it via edge://extensions.

The full management extension uses a separate, randomly generated local-only capability stored in Data/extension-access.key and mirrored into extension/ui/local-auth.js. Start-OpenEcho.cmd auto-provisions this through Initialize-OpenEchoExtension.ps1 before the server launches. Neither this key nor the general API/MCP token is committed. The public repository has only an empty local-auth.js template; never share a populated installation folder.
