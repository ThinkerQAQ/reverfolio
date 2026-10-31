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
