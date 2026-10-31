# Reverfolio v0.2.1

OpenEcho is now Reverfolio (回响集). The repository, CLI/Go module, dashboard, browser extension and portable release archives use the new name.

Existing Windows data, TOML settings and private keys are preserved. The local port and extension public key remain unchanged; the previous installation directory is a compatibility junction. The old API token environment variable and browser language key remain accepted. See [migration details](rename-migration.md).

Validated with Go race tests, vet, extension resource/identity checks, documentation validation and six cross-platform builds. Windows production data was verified against all 12 pre-migration tables. Other platforms are cross-compiled, not native runtime certified.
