# Reverfolio MCP：安装、工具、验证与排障

Reverfolio 的 Web UI、浏览器扩展和 MCP 使用同一个 Go 后端与 SQLite 作品数据。MCP 是**业务操作接口**，不是第二份数据库，也不是浏览器扩展的 Native Messaging 启动器。

## 1. 怎么连接

### 本机 MCP 客户端

下载本平台发行包并初始化配置，启动本机 stdio MCP（Windows 示例）：

```powershell
& "C:\software\Coding\reverfolio\bin\reverfolio.exe" -mcp-stdio -config "C:\software\Coding\reverfolio\config.toml"
```

macOS／Linux 使用对应的 `bin/reverfolio`，传入本机 `config.toml`。stdio 模式不会额外启动 GitHub 定时采集器；它与本机 Web UI 读写同一数据源，但会执行实际请求的明确写操作。

### ChatGPT

ChatGPT 无法直接连接别人电脑上的 `127.0.0.1`。远程使用需要**自己**运行的、受鉴权保护的 MCP 连接／隧道；服务端仍在用户自己的机器上存储数据。绝不要复用开发者的隧道 ID、令牌或 SQLite。

现有个人开发者应用的入口是 [Reverfolio](https://chatgpt.com/plugins/plugin_asdk_app_6ae5d8510c1c819187e96e9290753815)，当前是 **Private／Development**，**不是**已公开上架、可供任意用户一键安装的插件。用户可通过符合 ChatGPT 支持方式的自部署 MCP 连接自己的服务；发布到插件目录是独立的审批流程。

若服务端已更新但 ChatGPT 看不到新工具：

1. 在 ChatGPT 打开「插件 → Reverfolio → 管理 → 刷新工具」。
2. **新建对话**以获得最新工具目录；旧对话可能保留连接时的工具快照。
3. 在新对话测试只读 `get_dashboard`、`list_site_sources`。不要删除重装数据，不要建立第二份 Windows／WSL 数据库。

## 2. 31 个 MCP 工具

| 领域 | 工具 |
|---|---|
| 作品／目录（8） | `list_works`、`get_work`、`upsert_work`、`add_channel`、`list_catalog`、`set_monitoring`、`import_catalog`、`export_catalog` |
| GitHub／网站指标（7） | `collect_source`、`get_trends`、`get_dashboard`、`get_collection_status`、`list_site_sources`、`save_site_source`、`collect_site_source` |
| 引用／身份（6） | `submit_mention`、`list_mentions`、`review_mention`、`link_propagation_event`、`link_identity`、`unlink_identity` |
| AI 搜索任务（4） | `create_discovery_task`、`list_discovery_tasks`、`claim_discovery_task`、`complete_discovery_task` |
| 设置／历史／报告（6） | `list_settings`、`update_setting`、`update_settings`、`import_history`、`get_audit`、`generate_report` |

所有写操作均采用 `request_id`，部分操作另要求 `expected_version`，防止重复与并发冲突。导入必须先 `dry_run` 并明确勾选 `selected_ids`；审核原始证据必须人工确认，AI 不能把未经核验的链接自动计为传播。

## 3. 第一次怎样用

1. **注册作品**：GitHub 仓库使用 `kind=project` 和 `repository=owner/name`，文章使用 `kind=article` 和准确的公开 `canonical_url`；也可以导入标准公开作品清单。
2. **登记分发**：用 `add_channel` 关联具体的中英文、Medium、DEV、掘金等实际发布 URL；SyndiRelay 的可选本地交接仅同步成功发布的公开元数据。
3. **配置指标**：GitHub 使用自己授权的 `gh` 登录或服务环境变量；Umami／GSC／Bing 各自独立配置属性、认证和采集范围，详见 [网站数据源](site-sources.md)。缺失凭据时保留历史，但新采集应显示未授权／失败，不能伪造为 0。
4. **采集与查询**：调用 `collect_source` 或 `collect_site_source`；用 `get_dashboard`、`get_trends` 查看实际观测区间、来源、缺失日、下载量和页面归属。
5. **发现传播**：AI 客户端用自己的浏览器或搜索工具执行检索，MCP 管理任务和记录候选，`review_mention` 通过后才进入已核实统计。Reverfolio **没有**自行执行全网搜索的内置 Agent。
6. **报告和审计**：`generate_report` 获取 Markdown，`get_audit` 查看审核及来源历史。Release 下载量不是安装用户数，跨渠道的去重访客不合并。

## 4. 验收命令

从仓库源码或发行包包含的 `scripts/check_mcp_contract.py` 执行：

```bash
# 在仓库源码下：编译一个独立服务，用临时 SQLite 全流程验证，无网络采集、不写真实数据。
python3 scripts/check_mcp_contract.py --fixture
```

```powershell
# Windows 已安装版本：只读、无凭据输出，不改真实 SQLite。
python scripts/check_mcp_contract.py --binary "C:\software\Coding\reverfolio\bin\reverfolio.exe" --config "C:\software\Coding\reverfolio\config.toml"
```

脚本真实握手并调用 `tools/list`，要求精确 31 个工具；检查统计、网站数据源、运行状态、报告。fixture 模式还通过 MCP 协议跑完整的合成作品、渠道、发现任务、证据、审核、报告与导入预览。

## 5. 数据安全与公开范围

- 开发者当前的 Windows 安装目录示例为 `C:\software\Coding\reverfolio`，其业务数据库仅位于 Windows 的 `Data/state.sqlite`（既有 WSL 副本已停用）。其他用户可选择自己的安装路径，必须独立配置 `Data/` 和 `config.toml`。
- 扩展的 Native Messaging 主机用于**静默启动服务**，与 MCP 是不同通道；不会把浏览器 Cookie、GitHub 密钥或 SQLite 经 Native Messaging 发送。
- HTTP `/mcp` 需要 Bearer 认证并仅监听回环地址；远程隧道必须单独受保护，不能把本地端口直接裸露到互联网。
- 发布包与 Git 仓库不得包含 `Data/`、`config.toml`、Token／API Key、真实个人报表与第三方服务的私有数据。
