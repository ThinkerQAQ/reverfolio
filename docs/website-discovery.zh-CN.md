# 通用网站作品发现：RSS、Atom、Sitemap

适用于 **Reverfolio v0.8.0**。本功能不需要 Inkerio、blog-content、SyndiRelay、WordPress 插件或另一份数据库。

## 直接使用

1. 安装 Reverfolio 后，在 Web 工作台或 Edge／Chrome 扩展里打开**「网站发现」**。
2. 输入公开网站主页，或直接输入公开的 RSS／Atom 订阅地址，例如 `https://go.dev/blog/`。
3. 点击「发现文章」。读取该域名的 RSS／Atom／Sitemap 及网站声明的订阅链接，在本机预览候选标题、真实链接、来源类型和是否已有同 URL 作品。**这一步不写数据**。
4. 仅勾选确定属于自己的文章，点击「导入勾选文章」。Reverfolio 重新读取来源验证候选，再原子性写入 SQLite。相同规范化 URL 不重复建立作品；旧标题、渠道、监测选择不被覆盖。
5. 「已添加的网站来源」展示历史导入数量与上次扫描时间。点击「重新发现新文章」可以再次预览，选择增量作品。

只有个人网站也可用；只有 Feed 地址也可用；没有 SyndiRelay 不影响导入和统计。没有公开订阅源的平台，仍可通过已有「新增项目或文章」手动登记 URL。

## 具体规则

| 规则 | 行为 |
|---|---|
| RSS、Atom | 使用订阅源提供的文章标题、链接；解析 `rel=alternate` 的 Atom 链接 |
| Sitemap / Sitemap Index | 根据网页路径生成**待确认的候选标签**；标签不代表实际文章标题，不自动入库。对分类、标签、图片和常见导航路径做基本过滤 |
| 去重 | 规范化 URL、移除 UTM 等追踪参数；优先保留 Feed 中的真实标题，不覆盖之前录入的标题 |
| 来源 | 在专用 `website_sources` 表保存网站来源与上次扫描，使用既有 `catalog_origins` 记录作品来源；不读取任何发布器数据库 |
| 安全 | 仅 HTTP(S) 公共地址；禁止 localhost、内网／保留地址、非常用端口、跨域重定向及跨域文章候选；真正建连时校验 DNS 解析出的 IP，避免回绑定访问内网 |
| 限制 | 最多 12 次来源请求、单个响应至多 2 MiB、最多返回 100 条候选、17 秒总体超时；缺失／失败不可伪装成空文章。过多的来源可按不同 Feed 分批处理 |
| 隐私 | 不读取私人登录信息，不运行远端脚本，不采集浏览器 Cookie，不把 HTML 内容存入 Git，不要求第三方平台 API Key |
| 定时 | **v0.8.0 仅提供手动发现和重新扫描**，尚未内置每日自动轮询或平台账号文章列表自动发现。后续阶段单独实现并做来源权限控制 |

Sitemap 不保证只包含文章，因此候选确认是必需的；即便网站提供了全部页面，也不能因此证明用户是这些页面的作者。首次使用请只选择自己拥有的作品。

## MCP 与 API

| MCP | 用途 |
|---|---|
| `list_website_sources` | 只读列出已登记来源、上次成功导入时间与作品数量 |
| `preview_website({website_url})` | 只读获取候选及现有作品归属，不写入 |
| `import_website({request_id, website_url, selected_urls})` | 根据再次抓取结果，原子导入用户勾选的候选；对重复 `request_id` 幂等 |

本地 HTTP 对应 `GET /api/website-sources`、`POST /api/actions/preview_website` 与 `POST /api/actions/import_website`。统一使用原有本地登录会话或 Bearer 鉴权；与工作台、扩展共享 SQLite。

## 验收

```bash
go test ./internal/echo -run TestWebsite -count=1 -v
python3 scripts/check_mcp_contract.py --fixture
```

使用独立数据库验证 RSS／Atom／Sitemap Index、前端候选筛选、跨域／私有地址拒绝、响应大小限制、重复请求幂等、旧 schema v5 → v6 升级与既有作品保留。不能将远端统计数据、个人凭据或缓存文件提交公开仓库。
