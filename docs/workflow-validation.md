# Reverfolio v0.4.0 acceptance

## Automated coverage

Linux race tests and vet cover authenticated MCP/HTTP shared state, work/channel registration,
evidence review, discovery leases and retries, identity link/unlink and conflicts, historical
imports, schema migration, backup/restore, GitHub idempotency and provider failure states.

Provider fixtures cover Umami/GSC/Bing protocols, missing/zero/empty responses, dates and timezones,
write-only credentials, configuration rollback, and collection/report round trips.
A schedule-change regression verifies that disabling a running batch stops subsequent works.

A source-only temporary Windows checkout passes native tests and vet. Hosted CI runs race tests,
vet and builds on Linux, Windows and macOS. Refer to the PR checks for each exact commit result;
cross-compilation alone is not native runtime acceptance.

Six portable archives cover Windows/Linux/macOS on amd64 and arm64. Archive checks verify
checksums, empty extension credential templates, included setup documentation, and exclusion of
runtime databases, provider secret files and installed config.toml.

## Browser coverage

An isolated browser context and synthetic database exercise settings navigation, the Umami
browser-mode source form, save/readback with automatic collection disabled, credential-input
visibility, and authentication-mode switching. Web and extension share application source;
extension identity, icon assets, permissions and shared bundle checks are automated.

No browser screenshots or real credential values are committed. Schema migrations and automated
tests use synthetic fixtures; deployed data preservation is checked separately during upgrade.

## Limits

Live provider authorization is deployment-specific. Signed-in browsers do not authorize creating,
downloading or inspecting credentials. New provider sources default to disabled. Missing access
is blocked or failed, never an invented zero. Umami browser mode needs a running signed-in browser.

Provider observations, verified identity groups, propagation events and downloads use different
scopes. They do not establish unique readers or software users. AI clients perform searches;
the application records their scope, evidence and review, without claiming autonomous full-web coverage.

A passing native macOS CI suite is not a manual macOS desktop or signed-installer test.
Portable packages are unsigned. Back up the database before upgrading; preserve installed
configuration and extension credentials.
