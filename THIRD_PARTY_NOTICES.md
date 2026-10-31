# Third-party dependencies

OpenEcho's business-domain source is independently implemented. No source from git-stats or Mentions United is copied or vendored.

One direct runtime dependency is modernc.org/sqlite v1.61.0 (BSD-3-Clause), a CGo-free SQLite driver. Source and license: https://pkg.go.dev/modernc.org/sqlite@v1.61.0 and https://gitlab.com/cznic/sqlite . SQLite itself is public domain; its Go port and supporting modules retain their individual licenses.

The TOML configuration parser github.com/BurntSushi/toml v1.5.0 is a direct dependency under the MIT license. Its COPYRIGHT and MIT license text are in the upstream COPYING file. Source: https://github.com/BurntSushi/toml/tree/v1.5.0 .\n\nThe official MCP Go SDK `github.com/modelcontextprotocol/go-sdk v1.8.0` is a direct runtime dependency. Its upstream `LICENSE` describes a transition from MIT to Apache-2.0: contributions with relicensing consent use Apache-2.0; previously contributed code without consent remains under MIT; documentation outside the specification is CC-BY-4.0. Retain the upstream notices and applicable license text when redistributing. Source: https://github.com/modelcontextprotocol/go-sdk/tree/v1.8.0 .

Go module dependencies and checksums are recorded in go.mod and go.sum. No dependency source is vendored. Before distributing binaries, assemble the exact dependency license texts (including transitive modules) into the distribution and review the platform support matrix. This document is not a substitute for those notices.

Pinned GitHub Actions: actions/checkout and actions/setup-go (MIT). They are CI tooling, not linked runtime code. Respective repositories: https://github.com/actions/checkout and https://github.com/actions/setup-go .
