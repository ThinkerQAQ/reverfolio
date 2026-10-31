# Reverfolio v0.3.0

- Classify release downloads once in the backend: packages, source, auxiliary and other/ambiguous assets. The dashboard, browser extension, MCP and exports share the same classification. Historical observations are enriched without schema changes or raw-data rewrites.
- Show per-asset breakdowns and known interval changes, including unknown asset changes. Missing metrics display unavailable rather than zero. Filename heuristics do not imply installations or users.
- Add editable GitHub collection schedules in hierarchical bilingual settings. Save to TOML; the HTTP service picks up changes within one minute, including enabling a disabled schedule. Explicit CLI interval overrides remain fixed; stdio MCP does not start a scheduler.
- Expose credential presence without secrets and add the read-only MCP `get_collection_status` tool. Snapshot errors remain authoritative for endpoint permissions.
- Replace the obsolete About preview label with the runtime version.

Validation includes Go race tests, vet, historical snapshot/report tests, schedule enable/disable/restart/CLI override integration tests, extension resources and six cross-platform builds. Native platform validation is recorded separately; builds alone are not native execution tests.
