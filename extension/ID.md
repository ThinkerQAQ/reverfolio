# Reverfolio extension identity

The Manifest V3 public key is pinned so the extension keeps a stable ID across Chromium installations. The corresponding local API origin is chrome-extension://khdciakflcmilakmaiejamcgmggnkego. It is built into the Reverfolio server as a trusted local UI origin. The public manifest key is not a secret. A separate local API capability is generated per installation; its populated extension/ui/local-auth.js must never be distributed or committed.

If migrating from the old unpacked v0.1 popup, remove and re-add the extension in edge://extensions so Chromium uses the new identity. The extension never stores GitHub, SQLite or MCP credentials, but its own installation-specific local API capability is stored in the local extension folder. Treat a populated installation as private.
