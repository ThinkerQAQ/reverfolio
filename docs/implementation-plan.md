# Reverfolio implementation plan

Date: 2026-10-31 (Asia/Shanghai)
Status: foundation implementation in progress on a feature branch; no release yet.

## Delivery approach

Build a thin vertical slice first: register a repository, collect metrics, persist a baseline, show a second observation as an interval change. Then add evidence discovery and review. Each stage is delivered through a small PR with its relevant checks. Documentation and UI distinguish implemented features from plans.

## P0 — Repository foundation (this bootstrap)

- Private repository, MIT license, English default README and Chinese README.
- Architecture, implementation plan, contribution and security guidance.
- Squash-only merges, merged-branch deletion, branch protection where GitHub permits it.
- Documentation CI, PR template, data exclusion rules.
- Acceptance: repository privacy and settings read back; local links pass; documented enforcement limits.

## P1 — Application foundation and work registry

- Go module, standard net/http routing, embedded static assets, SQLite migrations.
- Select SQLite driver and official MCP Go SDK after checking platform support, maintenance and licenses; pin versions.
- Shared application services for works, language editions, channel pages, source accounts and aliases.
- Separate configuration, runtime data and repository source; bind to loopback by default.
- Acceptance: start from an empty directory, register a work, restart without data loss; migration failure is recoverable.
- CI: Go formatting, vet and meaningful domain/storage tests on Linux, Windows and macOS once Go code exists.

## P2 — GitHub metrics vertical slice

- Repository identity by stable platform ID, paginated releases/assets, repository counters, daily traffic and window totals.
- Archive filtered raw responses and record partial failures, API permissions, time coverage and capture time.
- Idempotent snapshots, bounded retries, daily refresh; manual trigger and scheduled collection.
- Preserve revisions to traffic days. Baselines, late-discovered old assets, replacements, deletions and missing permissions have explicit behavior.
- Acceptance: two repositories do not collide; repeated collection does not double-count; missing data stays missing; interval deltas handle asset lifecycle cases.

## P3 — Historical import and evidence domain

- Versioned JSON import schema and dry-run preview; per-record results and stable import IDs.
- Separate observation, discovery, import, and verification dates.
- Mentions, actors, evidence, propagation groups, review audit and optimistic concurrency.
- Old reports are secondary evidence pending original-source verification. Real historical records are imported locally, never seeded into source.
- Acceptance: repeat import is idempotent; three pages from one author can form one evidence group; pending evidence is excluded from verified counts.

## P4 — MCP and discovery workflow

- Tools for work lookup, discovery task listing/claiming, search-run recording, mention submission/review, collection status and trends.
- Request IDs for writes, task leases, source URL validation, authenticated network access.
- ChatGPT uses its own search/browser tools; Reverfolio supplies tasks and receives evidence. Optional autonomous execution is a later adapter.
- Acceptance: a client retrieves a task, submits evidence, reviews it and retrieves a traceable report. Conflicting reviews and expired leases are handled explicitly.

## P5 — Web UI and reporting

- Overview, work details, source/channel trends, evidence review, tasks, settings and import preview.
- Show coverage, source timezone, latest successful collection, unknown/missing states and verified versus candidate evidence.
- Minimal embedded frontend; choose a framework only if required by actual interactions.
- Markdown/CSV exports and consistent SQLite backups with referenced evidence files.
- Acceptance: complete the same vertical flow in UI and MCP; export and restore preserve scope and review history.

## P6 — Additional sources

- Umami, Search Console, distribution-platform metrics and optional Webmention.
- Expose an independent, manually managed public-channel registry. The application has no dependency on a blog publisher or its database.
- Track technical events (release, distribution, migration, incidents) on the timeline.
- Validate provider permissions and capabilities before marking an adapter supported.

## v0.1.0 release gate

P1–P5 accepted; no claim that P6 providers are implemented. Verify Linux, Windows and macOS binaries; SQLite backup/restore; authenticated MCP round trip; private-data exclusions; third-party notices; English and Chinese setup documentation. Release notes list tested configurations and remaining gaps.

## Immediate next PR

Implement P1 plus a GitHub fixture-backed storage slice: work creation, migration, two-repository snapshot isolation, and restart persistence. Do not import all historical reports until import validation and review states exist.

## Foundation slice progress (2026-10-31)

Implemented: repository work registration, private SQLite files, transactional schema initialization, stable GitHub ID guard, idempotent selected-field snapshots, paginated releases/assets, partial endpoint errors, daily traffic revisions, interval asset lifecycle calculations, loopback-only authenticated HTTP, and a minimal embedded UI.

Pending before P1/P2 are complete: generic articles/editions/channels, full raw-response archives, referrers/popular paths, collection task persistence and scheduling, bounded retry/backoff, backup/export and the next-stage MCP/evidence domain. GitHub calls currently have per-request and overall collection timeouts and no automatic retry.

The implementation PR is stacked on the repository-bootstrap PR until that PR is merged. Hosted CI availability is reported separately from local test results.

## Workflow implementation checkpoint — 2026-10-31

The development branch extends the GitHub foundation with article/site registry, channel pages, reviewed mentions and auditable grouping, lease-based AI discovery tasks, historical numerical metric import, GitHub endpoint retry and filtered raw-response archival, streamable HTTP/stdio MCP, standalone bilingual sidebar UI with local browser session, scheduled collection, SQLite backup/restore, and Markdown/JSON report exports.

Outstanding: independent external-provider adapters (Umami, GSC, Bing Webmaster), actual automatic AI search, definitive independent usage attribution, full user-configurable source/accounts, packaged releases, and tested Windows/macOS runtime behavior. These are **not** represented as completed. CI remains blocked by GitHub account billing/spending settings. For detailed acceptance evidence see [workflow-validation.md](workflow-validation.md).
