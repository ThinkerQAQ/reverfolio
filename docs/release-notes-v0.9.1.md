# Reverfolio v0.9.1 — DEV.to canonical article association

- Improve linking DEV.to published posts to a preexisting Work when the provider's `canonical_url` matches an English or other registered channel URL, not just the primary Work canonical URL.
- Detect a **suggested, unconfirmed** same-article-path association when historical canonical pages use a different hostname (e.g., a former GitHub Pages domain). A cross-host suggestion is **never silently merged**.
- In the Web UI and Edge/Chrome extension, let the operator choose the owning Work (or explicitly confirm a new independent Work) for each selected DEV.to post. Suggested Work IDs are visible and require review.
- Extend the existing `import_devto` MCP/HTTP input with optional `work_associations` (string article ID → Work ID, or 0 for explicit new Work). Do not allow a confirmed exact mapping to be forcibly overwritten. Preserve the original Work ID/title, language channels, sources and history.
- Add synthetic API + SQLite tests for legacy-domain matching, existing channel matching, explicit confirmation, wrong association rejection, and zero duplicate Works.
- No new API permissions, credentials, schema migration or provider. Cross-platform builds and 39 MCP tools remain unchanged.

Use the existing local private API Key and SQLite. Back up the production Data directory before replacing binaries, extension and docs.
