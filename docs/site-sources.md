# Site sources and explicit identity groups

## Authorization / 授权

A signed-in browser is not permission to create, download or inspect API credentials.
An operator must explicitly authorize provider access before an AI client obtains credentials.
Prefer entering credentials directly into the local application. Never send secret values in chat,
commit them, include them in release archives, or export browser cookies.

浏览器登录不代表授权生成、下载或读取密钥。由操作者明确授权后，在本地界面输入凭据；
不要把凭据发送到聊天、提交 Git 或打包发行。状态接口只报告凭据文件是否存在，不读取文件内容。
Bing/GSC 未配置或未授权时可保持关闭，其余功能继续工作。

## Setup

Settings → Data sources → Search and site analytics opens separate provider forms.
Register a site or article first, choose it as the work, and supply the property ID, timezone,
lookback and lag. New sources default to disabled. Saving and collecting are separate actions.
A blank credential retains the previous local value. Editing requires the current config version;
conflicting edits are rejected. If credential replacement fails, configuration is rolled back;
a rollback failure is reported explicitly. Credentials and TOML are separate files, so a system
crash between their writes is not an atomic transaction; verify source status after recovery.

- Umami: browser authentication uses `https://cloud.umami.is/api` and a local CDP endpoint.
  Keep the authorized browser signed in. Only allowlisted read-only stats, daily series and page-ranking requests are made in a
  temporary background target; it is closed afterwards. No Cookie export or arbitrary page scripts.
  Missing/expired sessions are recorded as failures. Token authentication is also available
  for operators who already have an authorized API credential.
- GSC: official Search Console API; Google property such as `sc-domain:example.org`.
  Dates use `America/Los_Angeles`. Supports an access token, authorized-user OAuth JSON,
  or a service account explicitly granted access. Read-only scope:
  `https://www.googleapis.com/auth/webmasters.readonly`. Credentials are supplied locally;
  this application does not create Google projects or grant account permissions.
- Bing: official Webmaster API; property is the verified site URL. Supply an explicitly authorized
  API key locally. Returned source date offsets are retained.

网页、扩展和 MCP 共用配置及采集逻辑。保存不等于授权 AI 获取密钥，也不等于成功采集。
自动采集遵循全局计划与数据源开关；配置变化会停止当前批次的后续采集，
已经发起的请求可能完成。失败记录保留，手动重试使用新的 request_id；
同一 request_id 返回原结果，不会覆盖失败历史。

## Metric semantics

Each run records provider, source, work, capture time, configuration hash, state and scoped metrics.
Missing credentials are blocked; request/permission errors are failed; empty responses differ
from explicit zero. GSC, Bing and Umami windows and units must not be added together.
Pageviews, session identifiers and download counters do not identify real people or software users.
Browser-mode source status reports no API credential, even when the browser is signed in.

## Identity groups

Evidence → Actor identity groups links public identity URLs only with explicit same-author evidence.
Choose an existing identity, supply alias URLs, reviewer and reason. Same-name matching is never
automatic. Conflicting groups are rejected; unlink requires the expected group and a reason.
Reports show both grouped verified identities and the original URL count; neither is a user count.
All links and unlink operations retain audit history.

## Upgrade and interfaces

Back up the live database before schema 4 migration. Keep Data and config.toml outside git and
release archives. Do not replace an installed config or extension credential with a template.

MCP adds `list_site_sources`, `save_site_source`, `collect_site_source`, `link_identity`,
and `unlink_identity`. HTTP adds `GET /api/site-sources` and matching
`POST /api/actions/{action}` operations. Writes use unique request IDs; source edits use
`expected_version`. Source credentials are write-only and are never returned by these interfaces.

## Daily and page detail

Enable `detail_enabled` per source to collect independent drilldowns. A failed drilldown retains the successful property summary and any completed detail responses, with a separate detail problem. See [dashboard semantics](impact-dashboard.md).
