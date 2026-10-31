# Foundation validation

Date: 2026-10-31 (Asia/Shanghai). Scope: repository-registration and GitHub-snapshot foundation preview.

- Linux: go test -race ./..., go vet ./..., Go formatting and documentation checks passed.
- Native HTTP smoke: start server, serve embedded UI, register two public GitHub repositories, collect complete release/view/clone snapshots, repeat request IDs without creating new snapshots, stop/restart, and verify persistence passed. Real observations were stored only in a temporary runtime directory outside the repository.
- Embedded UI JavaScript passed Node syntax validation; interactive browser acceptance has not yet been performed.
- CGo-disabled builds passed for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 and windows/arm64. macOS and Windows binaries were cross-compiled, not executed on those operating systems.
- Domain tests cover repository isolation, restart persistence, source corrections (including fractional-second ordering), idempotency, stable platform identity, paginated GitHub responses, unavailable endpoints, secret-free error recording, asset baselines/decreases/deletions, API authentication, Host checks and cross-origin rejection.
- GitHub-hosted CI is configured but subject to the account payment/spending restriction documented in governance. Local success does not imply hosted CI passed.

Not in this preview: MCP, evidence review, article/channel editing, historical imports, full raw archives, referrers/paths, scheduling, retries, backup/export, release packaging or an always-running installed service.
