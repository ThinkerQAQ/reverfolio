# Reverfolio browser extension (Edge / Chrome)

The browser extension and the local web app now use the SAME management UI. Clicking the Reverfolio toolbar icon opens the full dashboard in the browser side panel, not a separate launcher, services dropdown or shortcuts screen. The header button opens that same dashboard as a full browser tab.

## Installation and automatic startup

1. Extract the Windows release to a permanent folder such as `C:\software\Coding\reverfolio` and run `Install-Reverfolio.ps1` once in PowerShell. This initializes local configuration, creates a private extension access key, and registers the `com.reverfolio.bridge` Native Messaging host for **Edge and Chrome** under the current user. No API token is shared with the browser host.
2. In `edge://extensions` (or `chrome://extensions`), enable Developer mode, **Load unpacked** from the installation's `extension/` folder, and pin the extension.
3. Click the Reverfolio icon. The extension checks the local `127.0.0.1:19452` service, and if it is down asks the registered native host to start it **hidden**, then loads the full dashboard. The dashboard shows **Start local service** and **Retry connection** if startup fails. No permanently opened terminal window is required.
4. If automatic startup fails, use `Start-Reverfolio-Hidden.vbs` to start manually without a console, or rerun `Install-Reverfolio.ps1` to repair the native host registration. Upgrading an existing unpacked extension normally requires just **Reload**; its fixed ID must remain unchanged.

The native host is a separate minimal `bin/reverfolio-host.exe` built as a Windows GUI-subsystem executable. It accepts only `status` and `ensure_bridge` messages, checks the loopback health endpoint, and starts `bin/reverfolio.exe` with no console. It does **not** read or return any API token, SQLite data, cookies, or GitHub credential. The host's browser registry manifests are private runtime metadata under `Data/native-host/`.

Windows logon installations using the old direct-console scheduled task are migrated in place to a silent `wscript.exe` launcher; **existing triggers, user identity, and data are preserved**.

## Security and platform behavior

- The extension has `sidePanel` and `nativeMessaging` permissions and host access exclusively to `http://127.0.0.1:19452/*`. The native manifest is restricted to Reverfolio's pinned extension ID.
- An existing running service is usable even if the native host has not been registered yet. The native messaging flow is needed only to wake a stopped service.
- The web UI's local session, the extension's unique local capability in `Data/extension-access.key` and `extension/ui/local-auth.js`, and the MCP/API bearer token are independent. None is exposed through Native Messaging.
- Windows automatic hidden startup is included in the current portable installer. Linux and macOS bundles retain their existing startup behavior; they do not automatically acquire a registered native messaging host from a Windows installer.
- If a service cannot start, the extension shows a recoverable offline state and does not switch to WSL or another database.

Chinese instructions: README.zh-CN.md
