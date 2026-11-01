# DEV.to 官方 API：创作者账号与文章影响力统计

Reverfolio v0.9.0。**不需要 Inkerio、SyndiRelay 或浏览器 Cookie**，只使用 [Forem 官方 API V1](https://developers.forem.com/api/v1)，且全部为 GET 请求。

## v0.9.1：同一作品跨平台／旧域名关联

如果文章的 canonical URL 精确匹配已登记的 Work 原文或任一语言渠道 URL，Reverfolio 直接关联现有作品。如果旧博客域名与新域名不一致、但完整文章路径与已登记的某条渠道一致，**只显示候选关联，不静默合并**。在「平台账号 → 读取已发布文章」中逐篇确认归属后再勾选导入。

MCP 的 `import_devto` 现接受可选 `work_associations`：JSON 对象键是 DEV.to article ID（字符串），值是已有 Work ID（整数）；将值显式设置为 `0` 表示确认是全新作品。遇到跨域同路径的候选时，这个字段**必须明确指定**，避免自动产生重复的原创作品。对于已经确定归属的作品，不允许改绑到不一致的 Work。作品原有的中文／英文标题与历史指标均不会改写。

## 五步使用

1. 在 DEV.to 账户设置中生成个人 API Key（[官方说明](https://developers.forem.com/api)）。**API Key 本身可能具有写权限**，Reverfolio 只使用只读请求；不要公开、不要提交 Git。
2. 在 Reverfolio「平台账号」页面输入 API Key，点击「验证并连接」。`GET /api/users/me` 验证该 Key 对应哪一个账号。密钥只写入本地私有文件，不显示在回复、报表或 MCP 中。
3. 点击「读取已发布文章」，通过官方 `GET /api/articles/me/published` 读取本人已发布作品。仅在勾选并确认后才登记。最多预览 500 篇，超限明确标记未完全覆盖。
4. 如果返回的 canonical_url 已经存在于 Reverfolio 的 Work，就关联原文；否则以 DEV.to 的实际发布 URL 建立独立作品。**不会因标题一致就自动合并作品，也不会导入草稿。**
5. 点击「采集粉丝与文章统计」，写入 SQLite 带 API 来源、采集时间和账号/作品归属的历史快照。点击两次以后才能比较累计数据变化。

## 数据口径

| 数据 | API | 注意 |
|---|---|---|
| 账号粉丝 | `GET /api/followers/users` | 完整分页计数；上限 30 × 100，任何页失败／数据超限都不伪造精确粉丝数；不持久化关注者身份 |
| 文章浏览 | `/api/articles/me/published` 的 `page_views_count` | 平台浏览事件，**不是独立读者** |
| 文章浏览补充 | `GET /api/analytics/totals?article_id=…` | 当列表缺少计数时调用，限量补充；无法读取时保持缺失 |
| 文章点赞/互动 | `public_reactions_count` 或 `positive_reactions_count` | 平台反应计数，不等于收藏或传播 |
| 文章评论 | `comments_count` | 平台评论总数 |
| 历史趋势 | 多次有效的统计快照 | 首次累计值为基线；后续才能比较增长，粉丝数可以下降 |

所有指标都附带来源和时间；**缺失、未授权、限流不会写成 0**。0 只有在 API 明确返回准确计数时才展示。多个平台的粉丝总和不等于去重后的读者数。

## 数据独立与权限

- SQLite v7 新增平台账号、发布文章、采集任务和指标观测表；旧作品、GSC/Bing/Umami 与 GitHub 历史都保留。
- 私有密钥：安装目录 `Data/platform-credentials/devto_<account_id>.secret`，POSIX 文件权限 0600；Windows 采用安装目录继承的本机 ACL，应将安装目录限制为仅授权用户可读。不进入公开 Release、JSON 报告、Git 或 SyndiRelay。
- **密钥连接只允许本地 Web UI 或扩展写入**，不会成为可由 ChatGPT 传入密钥的 MCP Tool。MCP 只提供 `list_platform_accounts`、`preview_devto`、`import_devto`、`collect_devto`、`get_platform_dashboard`。
- HTTP 为 `GET /api/platform-dashboard`、`POST /api/actions/connect_devto`、`preview_devto`、`import_devto`、`collect_devto`，全部继续遵循原本本地鉴权。
- 不读取其他用户的私人统计数据；不使用浏览器 Cookie、不逆向请求、不自动发布文章；可撤销 DEV.to API Key 后停止新的指标采集。

## 支持状态

| 平台 | 状态 |
|---|---|
| DEV.to / Forem | 已实现完整官方 API 连接、作者文章发现和指标采集 |
| Hashnode | [官方 GraphQL API](https://docs.hashnode.com/quickstart/introduction) 存在，但具体账号与统计字段仍需验证；下一适配阶段 |
| Medium | [不再签发新的集成 Token](https://help.medium.com/hc/en-us/articles/213480228-API-Importing)，不承诺普通新用户的私有统计自动获取 |
| 掘金、知乎、头条、CSDN 等 | 官方授权统计 API 未核实／未集成；不爬取私有后台来宣称支持 |

## 验收

`go test ./internal/echo -run TestDEVTo -count=1 -v`

使用模拟官方 API + 临时 SQLite 进行身份验证、已发布文章选择、原文关联、分页粉丝计数、缺失不为零、来源时间、重复请求幂等、密钥不泄露及 MCP／HTTP 数据一致性测试。**测试数据并非用户真实粉丝量或文章访问量。**
