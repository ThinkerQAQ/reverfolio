# Third-party dependencies

Reverfolio's business-domain source is independently implemented. No source from git-stats or Mentions United is copied or vendored.

One direct runtime dependency is modernc.org/sqlite v1.61.0 (BSD-3-Clause), a CGo-free SQLite driver. Source and license: https://pkg.go.dev/modernc.org/sqlite@v1.61.0 and https://gitlab.com/cznic/sqlite . SQLite itself is public domain; its Go port and supporting modules retain their individual licenses.

The TOML configuration parser github.com/BurntSushi/toml v1.5.0 is a direct dependency under the MIT license. Its COPYRIGHT and MIT license text are in the upstream COPYING file. Source: https://github.com/BurntSushi/toml/tree/v1.5.0 .

The official MCP Go SDK `github.com/modelcontextprotocol/go-sdk v1.8.0` is a direct runtime dependency. Its upstream `LICENSE` describes a transition from MIT to Apache-2.0: contributions with relicensing consent use Apache-2.0; previously contributed code without consent remains under MIT; documentation outside the specification is CC-BY-4.0. Retain the upstream notices and applicable license text when redistributing. Source: https://github.com/modelcontextprotocol/go-sdk/tree/v1.8.0 .

Go module dependencies and checksums are recorded in go.mod and go.sum. No dependency source is vendored. Every portable release archive now includes **`THIRD_PARTY_LICENSES.txt`**, generated directly from the resolved Go module graph by `scripts/bundle_dependency_licenses.py`. The file contains upstream license/notice texts for direct and transitive modules, including license variants for `modernc.org/sqlite`, `modernc.org/libc`, and `modernc.org/memory`. It also includes some build/test-only dependencies as a safe superset. The packaging process fails if a resolved dependency lacks a root-level license file. CI checks this inventory independently. This hand-written overview is not a substitute for the bundled original notices.

Pinned GitHub Actions: actions/checkout and actions/setup-go (MIT). They are CI tooling, not linked runtime code. Respective repositories: https://github.com/actions/checkout and https://github.com/actions/setup-go .


Windows icon resource generation uses the pinned build-time tool github.com/akavel/rsrc v0.10.2 (MIT, Copyright 2013-2020 The rsrc Authors). The committed COFF resource files contain the project's original icon artwork, not third-party assets. Source and license: https://github.com/akavel/rsrc/tree/v0.10.2 .

## Browser-session transport

- `github.com/coder/websocket` v1.8.15 — ISC license. Used for the narrow, read-only Chromium DevTools network transport. Upstream: https://github.com/coder/websocket. No arbitrary JavaScript evaluation, cookie export or existing-tab control is exposed by this integration.
