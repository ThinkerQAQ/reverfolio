# Reverfolio v0.4.0

Adds independent Umami, Google Search Console and Bing observation sources, explicit audited
identity groups, and shared Web, extension and MCP controls. Provider windows, missing values,
errors and source timezone are retained. Schema 4 preserves existing observations.

- Umami browser-session mode requests restricted read-only statistics without exporting cookies
  or requiring an API key. Credential inputs are hidden in browser mode.
- Provider credentials are write-only; status checks use file metadata rather than file contents.
  New sources are disabled by default. Explicit authorization is required before obtaining credentials.
- Identity links and unlink operations require evidence/reason and reviewer; names are not merged.
- Configuration changes stop subsequent items in the scheduled batch; credential replacement
  errors restore the prior configuration where possible.
- Partial/failed discovery tasks can be reclaimed with a fresh lease while preserving search history.
- Extension and application versions are 0.4.0; the extension uses the shared embedded UI.

Back up Data/state.sqlite before upgrading. Preserve Data, config.toml and installed extension
credentials. Upgrading does not authorize or enable new provider access.
See [provider setup](site-sources.md) for permissions and storage behavior.

Validation uses synthetic data for Linux race tests, Windows native tests, provider HTTP/MCP
round trips, browser configuration forms, and extension/document checks. Six release targets are
cross-built; macOS runtime validation requires a native runner. Live Bing/GSC activation remains
a deployment-specific, explicitly authorized step.
