# Third-party dependencies

OpenEcho's business-domain source is independently implemented. No source from git-stats or Mentions United is copied or vendored.

The direct runtime dependency is modernc.org/sqlite v1.61.0 (BSD-3-Clause), a CGo-free SQLite driver. Source and license: https://pkg.go.dev/modernc.org/sqlite@v1.61.0 and https://gitlab.com/cznic/sqlite . SQLite itself is public domain; its Go port and supporting modules retain their individual licenses.

Go module dependencies and checksums are recorded in go.mod and go.sum. No dependency source is vendored. Before distributing binaries, assemble the exact dependency license texts (including transitive modules) into the distribution and review the platform support matrix. This document is not a substitute for those notices.

Pinned GitHub Actions: actions/checkout and actions/setup-go (MIT). They are CI tooling, not linked runtime code. Respective repositories: https://github.com/actions/checkout and https://github.com/actions/setup-go .
