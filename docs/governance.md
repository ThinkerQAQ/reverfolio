# Repository governance

Verified on 2026-10-31 (Asia/Shanghai).

## Enforced repository settings

- Private repository; default branch main.
- Squash merge enabled; merge commits and rebase merges disabled.
- Delete head branches after merge enabled.
- Squash commit title from PR title; message from PR body.

## Required working practice

Use feature branches and PRs. Main must not be force-pushed or deleted. Resolve discussions and pass relevant checks before merging. Initial required check is docs; add real Go and platform checks as implementation lands. No mandatory third-party approval count is imposed, matching the owner's solo-maintainer workflow.

## GitHub enforcement limitation

The authenticated account is GitHub Free. GitHub rejected the request to enable protection on this private repository with HTTP 403: Upgrade to GitHub Pro or make this repository public to enable this feature.

Therefore PR-required merges, administrator enforcement, required status checks, conversation resolution, linear history protection, and force-push/deletion prevention are documented practices, **not server-enforced branch protection**. Squash-only settings do not block direct pushes. Documentation CI is configured but cannot block direct pushes under this limitation. Its first GitHub-hosted run did not start: GitHub reported recent account payment failures or an insufficient spending limit. No workflow steps executed. Local documentation and whitespace validation passed; hosted CI remains unverified until the account restriction is resolved.

Keep the repository private. If the account later supports private branch protection, configure main with strict status checks, docs plus actual application CI contexts, required PRs (zero mandatory external approvals), stale-review dismissal, resolved conversations, linear history, enforced administrators, no force pushes and no deletions. Re-read settings after applying.

## Licensing and privacy

MIT license, with independent business-domain implementation and maintained infrastructure dependencies. Preserve applicable notices for any imported source. Personal history, raw analytics, credentials and local state stay outside source control. No runnable release or logo is claimed during bootstrap.
