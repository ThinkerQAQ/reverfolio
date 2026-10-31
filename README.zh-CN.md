# OpenEcho

[English](README.md) · [架构设计](docs/architecture.zh-CN.md) · [实施计划](docs/implementation-plan.md)

本地运行的开源与内容影响力观测工具。关联开源项目、文章中英文版本、多平台分发页面、指标历史，以及第三方引用和使用证据。

**当前状态：v0.1 开发预览，尚未发布 Release。** 已实现 GitHub 指标快照、文章／网站与项目统一登记、中英文及分发渠道关联、传播候选证据与人工审核、搜索任务与审计、历史数值指标导入、MCP、本地 Web、定时采集、过滤后的 GitHub 原始响应归档和 SQLite 备份恢复。

架构采用 Go 单服务＋SQLite＋内嵌 Web UI＋MCP。AI 客户端负责用自身工具搜索，OpenEcho 负责保存任务、线索、审核与统计；**不包含自动 AI 搜索执行器**。历史二手报告只能生成待审核线索，必须补充原网页或原始 API 证据并明确人工审核，才能计入已确认传播。GitHub 下载、浏览、去重访客与实际使用者分别计量，不相互推断。

通过功能分支和 PR 开发，仅使用 Squash Merge，合并后删除分支。规则执行情况见[仓库治理](docs/governance.md)。真实账号配置、凭据、个人统计和历史调查资料保存在运行数据目录中，不提交到代码库。

采用 [MIT](LICENSE) 许可证。引入第三方代码时保留适用的版权和许可声明。

## 本地运行

Go 1.26 或以上：

```bash
go run ./cmd/openecho -db data/openecho.sqlite
```

浏览器直接打开 <http://127.0.0.1:18440>，**本地 Web 页面不需要手动填写 Token**。页面使用仅在当前服务进程有效的 HttpOnly、SameSite=Strict 会话 Cookie；写操作还需同源请求头，拒绝跨站请求。程序 HTTP API 和 `/mcp` 仍支持独立的 Bearer 鉴权，`data/api-token` 中的密钥不会暴露到页面脚本；也可在服务端设置不少于 32 字符的 `OPENECHO_API_TOKEN`。HTTP 服务只允许回环地址监听。

GitHub 指标采集优先使用服务端环境变量 `GITHUB_TOKEN`；明确允许复用已登录的 GitHub CLI 时可以加 `-github-auth gh`。权限不足标记为部分成功或不可用，不能当作零值。首次观察的 Release 下载是累计基线，而非本期新增。

独立 Web 页面默认中文，可切换并记住英文；侧栏包含影响力总览、项目、公开分发渠道、引用审核、检索任务、历史统计和设置。支持登记作品／公开链接、证据审核、任务管理、指标导入预览和数据库备份。**仅借鉴成熟管理面板的交互，不依赖 BlogCTL、浏览器扩展或其他发布工具运行，也不会直接连接它们的数据。** MCP 对外暴露同一业务逻辑；`-mcp-stdio` 支持本机客户端。默认 `-collect-interval 24h` 定时采集 GitHub，可设为 `-collect-interval 0` 禁用。

一致性备份通过 Web 的“Download SQLite backup”；恢复只能写入**不存在的新数据库**，不能覆盖现有数据：

```bash
go run ./cmd/openecho -restore /private/backup.sqlite -db /private/new-state.sqlite
```

历史数值导入采用 `schema_version: 1`，记录时间范围、时区、单位、范围和来源，先 dry-run 再确认；未核实的第三方分发链接保持为空，不编造。所有凭据、SQLite 数据、个人指标和历史资料不得提交 Git 仓库。

验证命令：`go test -race ./...`、`go vet ./...`、`python3 scripts/check_docs.py`。当前平台 CI 受 GitHub 账号账单／额度限制，不代表 Windows/macOS 已实机验证；详见[验收说明](docs/workflow-validation.md)。
