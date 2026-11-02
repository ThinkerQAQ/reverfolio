# Repository instructions

- Fetch latest main before starting changes. Use small feature branches and PRs.
- Squash merge only and delete merged branches; never force-push or delete main.
- Do not merge changes unless the task authorizes merging.
- Default documentation is English, with a Chinese README. Keep status claims accurate.
- Architecture is a single Go service, SQLite, embedded Web UI and MCP with shared application services.
- Implement business logic independently; reuse maintained libraries. Track upstream versions and license notices for imported code.
- Keep runtime data, credentials, personal metrics and historical investigations out of git. Use synthetic fixtures only.
- Preserve source timezones, aggregation scopes, missing values and evidence provenance. Do not convert downloads, clones or visitor identifiers into verified users.
- Run appropriate checks for each change. Add Linux/Windows/macOS checks when application code exists.

## Accepted implementation direction (2026-11-02)

1. Deliver native Qt IFW installers, preserving SQLite/configuration on upgrades and uninstall. Configure optional ChatGPT Tunnel in the application; verify control-plane polling before reporting Ready. Follow VidOtter installation patterns.
2. Separate collection attempts from the last valid observation, including per-metric timestamps and missing/error states.
3. Extend existing evidence and event_key grouping into reviewed impact events, work/evidence links and outcome timelines. Count independent events, not repeated pages. Historical groups migrate as pending candidates.
4. Complete existing discovery leases/search_runs with renewal, candidate linkage and checked-scope coverage. An empty search is not proof of global absence. AI clients execute searches; no mandatory model API key.
5. Keep Go + SQLite + shared HTTP/MCP services and existing identity audit. No SyndiRelay/Inkerio dependency. Keep personal investigations out of git.
6. Ship changes through reviewable PRs; update README only when separately authorized after functional validation.
