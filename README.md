# OpenEcho

[中文](README.zh-CN.md) · [Architecture](docs/architecture.zh-CN.md) · [Roadmap](docs/implementation-plan.md)

Local-first observability for open-source and content impact.

OpenEcho will connect projects, multilingual articles, distribution channels, metric history, and evidence of external citations or adoption. AI clients can discover evidence through their search tools and submit it through MCP.

**Status: planning and repository bootstrap. No runnable application or release is available yet.**

## Planned architecture

One Go service, SQLite, a local web interface, and MCP. The business core is implemented for this domain; established libraries provide infrastructure. Metrics, evidence, and analytical claims remain distinct and traceable.

## First milestone

Register works and channel pages, persist GitHub snapshots, import historical observations, review external evidence, and expose the same application services through Web UI and MCP.

See the implementation plan for sequencing and acceptance gates. Real credentials, personal analytics, and imported history stay outside this repository.

## Development and governance

Use feature branches and pull requests. Squash merge only; delete merged branches. Main must not be force-pushed or deleted. Repository settings and enforcement limitations are documented in [governance](docs/governance.md).

## License

[MIT](LICENSE). Third-party code must retain its applicable notices. No third-party application source has been imported during bootstrap.
