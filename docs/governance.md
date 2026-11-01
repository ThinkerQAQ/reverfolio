# Repository governance

Verified using the live GitHub repository settings on **2026-11-01**. For the current repository, see [ThinkerQAQ/reverfolio](https://github.com/ThinkerQAQ/reverfolio). The older private-repository planning constraints no longer apply.

## Current enforced settings

- **Visibility: PUBLIC**; **MIT** open-source license; default branch `main`.
- Changes land via pull requests and **Squash Merge** only. Merge commits and rebase merges are disabled; completed head branches are deleted automatically.
- `main` branch protection rejects force pushes and branch deletion, and requires four GitHub Actions status checks: `docs`, `test (ubuntu-latest)`, `test (windows-latest)`, and `test (macos-latest)`.
- Required external approval count: **zero** for a solo maintainer. Review requirements can be increased when there are active collaborators. No bypass of mandated status checks is assumed.
- Documentation and Go matrix CI execute on pull requests and `main` pushes. CI is a repeatable validation gate, not a substitute for testing an actual installed release.

## Contributor workflow

1. Fetch latest `main`; create a focused feature or fix branch.
2. Describe the user-facing change, tests and security/privacy effects in the PR. Update English and Chinese docs together.
3. Wait for required CI checks before squashing to `main` and deleting the branch. Do not push straight to the protected branch or rewrite history.
4. Do not check in a live `Data/` directory, screenshots containing private details, API keys, tunnel details, real accounts' follower lists, database backups, or any personal traffic logs. Synthetic fixtures belong in tests.
5. Keep original third-party license notices; the release pipeline bundles the resolved module-graph texts into `THIRD_PARTY_LICENSES.txt` in every portable archive.

## What is not yet provided

The GitHub repo is public, but the **ChatGPT Reverfolio app/plugin itself remains a private Development plugin**. Running the MIT code locally or connecting your own secured MCP server does not require installing that private app. Publishing a plugin to the ChatGPT directory is a separate registration/review process, not an effect of making the GitHub repository public.

A successful CI check does not prove that every user has a working native extension installer on macOS or Linux. Windows bundles offer the built-in Native Messaging setup; the others provide CLI binaries and a Web UI.

## License, maintenance and privacy

Project source: [MIT](../LICENSE). Binary archives additionally include [upstream notices](../THIRD_PARTY_NOTICES.md) and their complete `THIRD_PARTY_LICENSES.txt`. Report security concerns using the private contact in [SECURITY.md](../SECURITY.md) rather than a public issue. Request changes via Issues/PRs; do not upload real secret values or private user data.
