# Reverfolio v0.9.0 — DEV.to API-only creator metrics

- Add official **DEV.to/Forem API V1** client, independent of SyndiRelay, browser sessions, or any blog engine.
- Verify owned accounts via `GET /api/users/me`, preview/select personally published articles, associate existing canonical Works, and persist selected publication URLs with origin provenance. Drafts and unselected posts are never imported.
- Collect followers (only after exact pagination), article views (including official analytics totals fallback), reactions, and comments. Source endpoint, UTC observation time, platform account and publication are preserved. Missing is not zero; rate limits and auth failures retain explicit states.
- SQLite **schema v7** preserves existing observations and introduces account/publication/run/metric history. Credentials are private local files under `Data/platform-credentials`; the public release does not include them.
- Web UI + Edge/Chrome extension include **Creator platforms** and follower history mini-chart with article ranking; MCP adds five tools (39 total). API key enrollment is a local UI action and **not** an MCP tool.
- Bilingual docs and synthetic end-to-end HTTP/MCP integration tests. Windows/Linux/macOS build targets unchanged.

Only DEV.to is implemented. Hashnode official API analysis, recurring scheduled polling and other verified platform API adapters are future work. No unofficial website scraping.
