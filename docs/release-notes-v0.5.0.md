# Reverfolio v0.5.0

Adds a portable works catalog for independently deployed impact tracking.
No publishing application is required, and fresh installations contain no preselected works.

- Preview a catalog, explicitly select works, and import their public original/edition/publication URLs.
- Show source ID, external work ID and producer without claiming verified adoption.
- Pause or enable automatic collection per work; history and explicit manual collection remain available.
- Preserve existing titles and monitoring preferences during imports; reject identity/channel conflicts atomically.
- Export public mappings without provider settings, credentials, account details or article bodies.
- Convert BlogCTL/SyndiRelay publication-record exports through an optional standalone Python adapter.
- Share catalog and monitoring operations between Web, extension, HTTP and MCP.
- Upgrade to schema 5 by adding tables while retaining existing observations and enabled legacy works.

See [portable catalog and usage](portable-catalog.md).
This release provides file handoff and a receiving API, not automatic publisher delivery or an outbox.
GitHub Traffic charts and per-article analytics remain separate follow-up work.

Validation includes race/vet and workflow tests, synthetic import conflict/replay/rollback cases,
converter tests, and an isolated real-browser preview → selected import → monitoring toggle flow.
Back up before migration. Preserve installed Data, config.toml and extension connection files.
