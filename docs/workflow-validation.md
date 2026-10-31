# OpenEcho workflow acceptance and limitations

Date: 2026-10-31. Branch: `feat/evidence-mcp-workflow` (stacked on `feat/github-foundation`).

## Verified implementation and test boundaries

- Six CGo-disabled cross-compiles succeeded: Linux, Windows and macOS, each amd64 and arm64. These are build checks, not runtime smoke tests.
- Linux `go test -race ./...`: exercises the MCP-to-HTTP shared-state round trip, work/channel registration, evidence submission, verification and grouping, optimistic review version conflicts, discovery leases, atomic history validation and imports, legacy schema migration, consistent backup/restore, repository isolation, snapshot idempotency and partial-source failures.
- `go vet ./...`, docs link/fence validation and embedded Web UI JavaScript parsing.
- The previous implementation workflow exercised an isolated synthetic browser database. The current running loopback instance was also inspected through Remote Browser Tool in a background isolated browser context: bearer-authenticated connection succeeded, and the overview displayed actual private GitHub snapshots and evidence counters. The current live re-check was read-only; mutations remain covered by automated browser/workflow integration fixtures.
- HTTP bearer authentication covers `/api/` and `/mcp`; only loopback listening is accepted; cookies and API credentials are never sent to the UI or stored in source.
- GitHub direct API: stars/forks, release-asset lifetime download counts, GitHub rolling views/clones, popular referrers/paths; missing permissions and failed endpoints stay unavailable. Filtered raw responses are persisted separately from snapshots.
- Data source import: historical numerical metrics have explicit time windows, scopes, units, timezones and source metadata. A `secondary_report` mention remains pending without original-source evidence and review.

## Gaps and release gate

GitHub-hosted runner jobs are blocked by account payment/spending limits. Cross-compilation for Windows/macOS is not runtime verification. Integration with Umami, Google Search Console, content distribution sites and autonomous search has not been implemented. No verified software-user count can be inferred from downloads, views, identities or page references. A v0.1 release still needs final platform packaging, runtime smoke testing and dependency license bundling.

Do not commit real observation databases, local API tokens, GitHub credentials, imported reports, personal traffic or historical evidence.

## Private installation acceptance (not committed to source)

The owner-specific runtime database is outside the checkout. It was initialized from the current private content source with 20 bilingual article pairs and currently public GitHub repositories. The first GitHub collection observed ten registered project repositories with no missing endpoint fields in that run. These are cumulative snapshots, not inferred unique users or period increments. All user-specific identities, credential material and exact numbers remain outside repository documentation.
