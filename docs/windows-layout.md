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
        popup/
        _locales/
      Data/
        state.sqlite
        api-token
      config.toml
      Start-OpenEcho.cmd
      Run-OpenEcho.ps1

- bin/: Windows executable with the OpenEcho icon, not personal data.
- extension/: optional unpacked Edge/Chrome extension with icon, local status and quick links. Load the extension folder from edge://extensions (Developer mode). It does not read SQLite or TOML.
- Data/: private local SQLite database and automatically generated server API token. Never overwrite on upgrade.
- config.toml: local TOML configuration for startup and nonsecret UI preferences. It is NOT a SQLite table; never commit it.
- config.example.toml: tracked reference template; safe to share.
- Start with Start-OpenEcho.cmd. Browser UI requires no token, while MCP/API remain authenticated.
- The Windows config uses 127.0.0.1:19452 to avoid collision with Ubuntu/WSL on port 19450. They are separate local instances.
- The Windows database is initially copied from a consistent Ubuntu backup. Subsequent writes DO NOT automatically synchronize. Never access the same SQLite file from Windows and WSL concurrently.
- Secrets are not placed in TOML; Data/api-token is generated when the Windows app starts.
- github.auth = "env" is the safe no-token default on this Windows host. The signed-in Windows GitHub CLI has no available OAuth token; use Windows gh auth login and switch to gh when authorized. Until then, collection.interval = "0" avoids pretending repository traffic has been collected. No credentials are copied.
- Restart the app to apply startup settings such as listener or interval. UI preferences update TOML and take effect without restart.
- Cross-compile with GOOS=windows GOARCH=amd64 go build -o bin/opencho.exe ./cmd/openecho.
- OpenEcho has no dependency on any publishing tool or browser extension.
