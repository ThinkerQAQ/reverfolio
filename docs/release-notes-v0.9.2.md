# Reverfolio v0.9.2 — open-source release readiness

Documentation, provenance and packaging-only maintenance release. **No changes to live metrics, data collection logic, platform credentials, SQLite schema, or MCP's 39 business tools.**

- Bring the English and Chinese READMEs, CONTRIBUTING, SECURITY and repository governance guide up to the **actual public MIT GitHub state**. Remove historical references to private repo restrictions and nonexistent CI; clarify that the **ChatGPT Development plugin remains private**.
- Add independent bilingual [quick start](quickstart.md) / [中文快速入门](quickstart.zh-CN.md) for all six available release targets. Explain the Windows native installer and optional Edge/Chrome extension, plus **CLI-only** Linux/macOS setup without promising an unsupported extension native host.
- Package actual license/notice texts for all resolved Go modules into `THIRD_PARTY_LICENSES.txt` in every binary archive, using deterministic module inventory and automatic failure if a module lacks an upstream notice.
- Validate all six release archives, checksums, required license texts and absence of bundled local state/credentials before publication; run the license availability gate in platform CI.
- Preserve the original locally installed SQLite, TOML, token files, DEV.to private API Key and all metrics on upgrade; only executable/extension/public documentation are replaced.

For first-time users: **download a portable archive, initialize your own local configuration and register your own works**. The maintainer's personal account/tunnel/configuration is not part of the distribution.
