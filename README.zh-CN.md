# Reverfolio

[English](README.md) · [架构设计](docs/architecture.zh-CN.md) · [实施计划](docs/implementation-plan.md)

本地运行的开源与内容影响力观测工具。关联开源项目、文章中英文版本、多平台分发页面、指标历史，以及第三方引用和使用证据。

## 下载与安装

**[下载最新正式版](https://github.com/ThinkerQAQ/reverfolio/releases/latest)** · **[中文快速入门](docs/quickstart.zh-CN.md)** · **[English quick start](docs/quickstart.md)** · [安全说明](SECURITY.md)

- **Windows（x64 / ARM64）：**解压对应发行包，先运行 `Install-Reverfolio.ps1`，再运行 `Start-Reverfolio.cmd`，打开 <http://127.0.0.1:19452>。Windows 包自带可选 Edge／Chrome 扩展。
- **Linux / macOS（x64 / ARM64）：**解压对应压缩包，复制 `config.example.toml` 为 `config.toml`，运行 `./bin/reverfolio -config ./config.toml`，打开 <http://127.0.0.1:18440>。这些平台目前**未提供浏览器原生宿主的一键自动安装**。
- **真正独立：**无需 SyndiRelay 或特定博客引擎。新用户从 RSS／Atom／Sitemap 发现文章，或直接登记 GitHub 仓库；拥有对应平台账号时再自行授权。SQLite、凭据都保存在本人本机 `Data/`，发行包不含他人统计数据。**ChatGPT 私有开发版插件并不是已上架的一键安装插件**。

**已实现：**GitHub 访问／Release 下载、授权的 Umami／GSC／Bing 网站数据、通用博客文章发现、DEV.to 官方账号／文章统计、引用审核、历史趋势图／导出、39 个 MCP 工具。**未实现：**全部内容平台的定时采集、无官方 API 的平台私有统计、去重真人粉丝识别、公开插件目录安装。

## 近期版本更新

**v0.9.2** 完成开源交付收尾：双语快速安装、更新真实公开仓库治理信息、明确三平台安装边界，并在所有发行包中附带实际依赖许可证全文。**不增加采集平台、不变更 SQLite schema 或现有数据。** [发行说明](docs/release-notes-v0.9.2.md)。

**v0.9.1** 修复 DEV.to 文章跨域 canonical URL 关联：可识别现有英文渠道；旧博客域名路径一致时只建议作品，需明确确认，避免生成重复 Work。支持按文章指定归属的 Web UI、扩展与 MCP 导入，保留 v0.9.0 及以前的数据。详见 [DEV.to 官方 API 说明](docs/platform-api.zh-CN.md)。

**v0.9.0** 新增 [DEV.to 官方 API 账号与文章统计](docs/platform-api.zh-CN.md)：本地配置个人 API Key，关联已发布文章与原文，采集粉丝、浏览、点赞、评论的历史快照；网页、扩展和 39 个 MCP 工具共用 SQLite。无需 SyndiRelay，不抓取登录页面。其他平台必须先确认有合法官方接口。

**v0.8.0** 新增独立的 [RSS／Atom／Sitemap 网站作品发现](docs/website-discovery.zh-CN.md)：输入博客网址、预览候选、勾选导入、复用已有作品与来源记录，并可手动重新扫描。新增 3 个 MCP 工具，网页与扩展使用同一 UI；**无需 Inkerio 或 SyndiRelay**。平台粉丝／阅读统计仍属于后续阶段。

**v0.7.1** 完整验收 [31 个 MCP 工具](docs/mcp-setup.zh-CN.md)，补齐「刷新工具 → 新建对话」同步步骤，以及隔离 SQLite 的跨平台 MCP 端到端测试脚本。不会迁移或覆盖原有数据和密钥。

**v0.7.0** 同时修复 Windows 黑色控制台窗口与扩展自动启动：首次安装注册 Native Messaging，打开扩展时检查并静默启动服务，失败可手动重试；既有计划任务改为静默启动。

**v0.7.0** 新增[传播路径可视化](docs/impact-dashboard.md)：原文／项目 → 同站语言版本 → 外部分发链接 → 已审核的第三方传播，并显示真实的观测及审核时间线。不推断渠道发布时间；SyndiRelay 可通过本地 API 可选登记成功发布的 URL，Reverfolio 本身不依赖发布工具。

**v0.6.1** 完善统计图：精确归属文章、按日历天数筛选、实际观测区间下载增量、避免图片导出裁剪，并保持网页与扩展一致。

**v0.6.0** 新增[统计图与使用说明](docs/impact-dashboard.md)：GitHub 每日流量、版本／平台下载分布、网站趋势、按网址精确归属文章，以及 PNG／SVG／CSV 导出。网页、扩展与 MCP 共用观测数据。

平台支持现状：**DEV.to 已使用官方 API 直连实现账号、文章和互动指标**，不依赖 SyndiRelay；其他内容平台必须先核实稳定的官方授权 API。之前的[跨平台统计方案](docs/social-impact-architecture.zh-CN.md)已标记为历史提案，以[最新 API 实现说明](docs/platform-api.zh-CN.md)为准。

架构采用 Go 单服务＋SQLite＋内嵌 Web UI＋MCP。AI 客户端负责用自身工具搜索，Reverfolio 负责保存任务、线索、审核与统计；**不包含自动 AI 搜索执行器**。历史二手报告只能生成待审核线索，必须补充原网页或原始 API 证据并明确人工审核，才能计入已确认传播。GitHub 下载、浏览、去重访客与实际使用者分别计量，不相互推断。

通过功能分支和 PR 开发，仅使用 Squash Merge，合并后删除分支。规则执行情况见[仓库治理](docs/governance.md)。真实账号配置、凭据、个人统计和历史调查资料保存在运行数据目录中，不提交到代码库。

采用 [MIT](LICENSE) 许可证。引入第三方代码时保留适用的版权和许可声明。

## Windows 原生部署示例

在 Windows 安装中，Web UI、Edge／Chrome 扩展及 MCP 应访问同一个原生 `Data/state.sqlite`，不要让 WSL 再启动第二份可写数据库。其他用户自行选择本机安装目录及凭据，**不需要复制开发者的数据、隧道或配置**。详见 [Windows 单一数据源与 MCP 架构](docs/windows-authority.md)。

## 浏览器扩展（可选）

Reverfolio 独立扩展位于 extension/，**直接在 Edge/Chrome 侧边栏运行完整管理工作台**，与主 Web 页面复用同一份前端代码，可在扩展标签页全宽打开；已移除跳转按钮拼凑的旧弹窗和「Windows 优先/WSL 自动检测」选择器。扩展只连接 Windows 本地 127.0.0.1:19452。文件放在 C:\software\Coding\reverfolio\extension。旧 v0.1 解压缩扩展需在 edge://extensions 删除后重新加载此目录以注册固定 ID。详见 [安装说明](extension/README.zh-CN.md)。不依赖 BlogCTL，也不保存 GitHub/MCP 凭据。

## Windows 目录约定

Windows 安装目录示例：`C:\software\Coding\reverfolio\`（用户可自行选择固定目录）。可执行文件位于 bin/reverfolio.exe，SQLite 数据位于 Data/state.sqlite，非敏感运行配置和界面设置保存在 config.toml。详情参见 [Windows 目录结构](docs/windows-layout.md)。此前 WSL 数据库已校验一致并停用；后续统一使用 Windows 数据。

## 本地运行

Go 1.26 或以上：

```bash
go run ./cmd/reverfolio -db data/reverfolio.sqlite
```

浏览器直接打开 <http://127.0.0.1:18440>，**本地 Web 页面不需要手动填写 Token**。页面使用仅在当前服务进程有效的 HttpOnly、SameSite=Strict 会话 Cookie；写操作还需同源请求头，拒绝跨站请求。程序 HTTP API 和 `/mcp` 仍支持独立的 Bearer 鉴权，`data/api-token` 中的密钥不会暴露到页面脚本；也可在服务端设置不少于 32 字符的 `REVERFOLIO_API_TOKEN`。HTTP 服务只允许回环地址监听。

GitHub 指标采集优先使用服务端环境变量 `GITHUB_TOKEN`；明确允许复用已登录的 GitHub CLI 时可以加 `-github-auth gh`。权限不足标记为部分成功或不可用，不能当作零值。首次观察的 Release 下载是累计基线，而非本期新增。

独立 Web 页面默认中文，可切换英文并保存在 TOML 中；侧栏包含影响力总览、项目、公开分发渠道、传播路径、引用审核、检索任务、历史统计和设置。支持登记作品／公开链接、证据审核、任务管理、指标导入预览和数据库备份。**仅借鉴成熟管理面板的交互，不依赖 BlogCTL、浏览器扩展或其他发布工具运行，也不会直接连接它们的数据。** MCP 对外暴露同一业务逻辑；`-mcp-stdio` 支持本机客户端。默认 `-collect-interval 24h` 定时采集 GitHub，可设为 `-collect-interval 0` 禁用。

一致性备份通过 Web 的“Download SQLite backup”；恢复只能写入**不存在的新数据库**，不能覆盖现有数据：

```bash
go run ./cmd/reverfolio -restore /private/backup.sqlite -db /private/new-state.sqlite
```

历史数值导入采用 `schema_version: 1`，记录时间范围、时区、单位、范围和来源，先 dry-run 再确认；未核实的第三方分发链接保持为空，不编造。所有凭据、SQLite 数据、个人指标和历史资料不得提交 Git 仓库。

验证命令：`go test -race ./...`、`go vet ./...`、`python3 scripts/check_docs.py`。平台 CI 执行 Linux、Windows 和 macOS 测试，交叉编译不等于原生运行验收；详见[验收说明](docs/workflow-validation.md)。

## 名称迁移

原名 OpenEcho，现统一为 Reverfolio（回响集）。保留原有 Windows SQLite 数据、TOML 配置、扩展 ID 与端口；旧目录保留兼容入口。详见[迁移说明](docs/rename-migration.md)。

## 下载分类与区间变化

网页、扩展、MCP 和导出报告共用后端文件名分类：程序包、源码、辅助文件、其他／未分类。逐附件显示类别和分类理由；含义不明的压缩包不会直接计为程序包，下载量仍不代表安装或真实用户。历史快照读取时补充分类，不改写原始观测。

区间变化只合计已知增量，另列新增基线、计数下降、附件消失等未知增量；保留两个快照的时间范围，不冒充每日增长。没有 Release 访问权限时显示不可用，而不是 0。

## 自动采集

在「设置 → 数据采集 → GitHub → 自动采集计划」调整频率，配置保存到 `config.toml`，本地 HTTP 服务在一分钟内生效；关闭计划不会删除已有数据。需要电脑与服务保持运行。显式 `-collect-interval` 命令行参数优先，修改它需要重启；stdio MCP 不启动重复采集。

`github.auth = "gh"` 使用 Windows 原生 GitHub CLI 的有效登录；也支持服务环境中的 `GITHUB_TOKEN`。GitHub 密钥不在界面编辑；站点数据源另有只写凭据输入。运行状态及 MCP 只显示凭据是否已加载，实际权限和错误保留在采集快照。

## 站点数据源和身份归并

参见[配置和授权说明](docs/site-sources.md)。三个数据源分别保留统计窗口、时区和失败状态，不混加指标。新数据源默认关闭。凭据由操作者在明确授权后于本机提供；浏览器已登录不代表允许 AI 生成、下载或读取凭据。Umami 浏览器会话模式只请求受限的只读 JSON 统计接口，不导出 Cookie，不需要 API Key。

身份归并必须有明确的同一作者证据、理由和审核人，不按同名合并；撤销保留审计历史。schema 4 增加身份和数据源表，不重写原始观测；升级前备份数据库。

## 作品选择与发布工具集成

逐作品启停监测，先预览再勾选导入原文及公开分发链接，并查看导入来源。无需安装发布工具。参见[目录协议与使用说明](docs/portable-catalog.md)。导入不会覆盖已有标题和监测选择；首次启动没有预置作品。
