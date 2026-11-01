# Reverfolio

[English](README.md)

Reverfolio 是一款面向开源项目、网站和公开文章的本地影响力统计工具。它将作品、发布链接、平台指标以及引用证据关联起来，支持采集、审核与历史分析。Web UI 和 MCP 共用一个 Go 服务及本地 SQLite 数据库。

## 目录

- [1. 背景](#1-背景)
- [2. Quick Start](#2-quick-start)
- [3. Tutorial](#3-tutorial)
- [4. Document](#4-document)
- [5. Contributing](#5-contributing)
- [6. 安全与许可](#6-安全与许可)

## 1. 背景

同一个项目或文章可能分布在 GitHub、个人网站和多个内容平台。Release 下载量、仓库流量、搜索曝光、页面浏览量与互动量的统计范围和时间窗口并不相同，不能简单相加并当作真实人数。

Reverfolio 将统计数据与作品、来源和采集时间关联，主要解决四个问题：

- 独立管理项目、文章和网站，不要求使用特定博客或分发工具。
- 区分“没有数据”“采集失败”和实际测量结果为零。
- 保留统计时间窗口、来源、采集状态与历史基线。
- 对外部引用和采用证据进行人工审核，不把未经核实的线索当作成果。

Inkerio 和 SyndiRelay 均为可选组件。全新安装不会预置作者账户或作品。

## 2. Quick Start

### 2.1. 下载

1. 打开 [Releases](https://github.com/ThinkerQAQ/reverfolio/releases/latest)。
2. 根据操作系统和 CPU 下载对应压缩包（`windows`/`linux`/`darwin`、`amd64`/`arm64`），同时下载 `SHA256SUMS`。
3. 将压缩包的 SHA-256 与 `SHA256SUMS` 中对应记录核对，再解压到固定的私有目录；保持 `Reverfolio` 目录结构不变。

Windows 可使用 `Get-FileHash .\Reverfolio-*.zip -Algorithm SHA256`，Linux 使用 `sha256sum ./Reverfolio-*.tar.gz`，macOS 使用 `shasum -a 256 ./Reverfolio-*.tar.gz`。将结果与 `SHA256SUMS` 中**对应压缩包**的记录比较，不一致时不要继续安装。

发行包自带可执行文件和配置模板，直接使用不需要安装 Go。

### 2.2. 安装与启动

**Windows：**在解压后的 `Reverfolio` 目录打开 PowerShell：

```powershell
.\Start-Reverfolio.cmd
```

启动脚本会按需调用安装程序，创建 `config.toml` 和 `Data/`，且不会覆盖已有数据；随后启动本地服务。安装程序也会初始化扩展凭据并注册 Native Messaging Host，**不需要手动注册 Host**。打开 [http://127.0.0.1:19452](http://127.0.0.1:19452)。

**Linux / macOS：**在解压后的 `Reverfolio` 目录打开终端：

```sh
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

**只有首次安装时才复制配置模板**，升级不要覆盖已有私有配置。保持终端进程运行，打开 [http://127.0.0.1:18440](http://127.0.0.1:18440)。Unix 发行包不自动安装后台服务；macOS 未签名程序可能需要系统授权。

### 2.3. 登记第一个作品

1. 打开「项目管理」，选择「新增项目或文章」。
2. 选择「GitHub 项目」，填写标题以及自己拥有的仓库路径（`owner/name`），保存。
3. 刷新「项目管理」，确认记录仍然存在。
4. 对这个项目执行「更新 GitHub」，检查采集状态和时间。公开 Release 下载计数可能不需要凭据；私有流量需要拥有权限的 GitHub 凭据。
5. 使用相同的 `config.toml` 重启 Reverfolio，再次确认项目存在，完成 SQLite 持久化验收。

首次累计统计是基线，不能当作新增量。没有 Release、缺少权限或接口请求失败时，部分字段可能不可用，不能记为零。

### 2.4. 验收

在另一个终端调用对应的健康检查：

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```sh
# Linux / macOS
curl -fsS http://127.0.0.1:18440/healthz
```

验收标准：`status=ok`、重启后作品仍存在、每次尝试采集均能区分成功/部分成功/失败。**仅健康检查通过，不代表统计采集已经成功。**

## 3. Tutorial

以下 Case 均为可选能力，可以独立使用。第一次接入建议先完成 Quick Start。

### 3.1. Case 1：从自己的网站导入文章

1. 打开「网站发现」，输入自己管理的公开网站地址或 Feed 地址。
2. 预览 RSS、Atom 或 Sitemap 中发现的候选文章，核对标题和 URL。发现候选文章并不等于已登记发布。
3. 勾选需要的文章，执行导入。
4. 在「项目管理」和「分发渠道」中检查导入的作品与发布链接。

只有明确选择的 URL 才会导入。Feed 和 Sitemap 不保证覆盖所有文章；标题相同也不会自动合并不同作品。

### 3.2. Case 2：接入 DEV.to

1. 获取自己的 DEV.to API Key，在本地「平台账号」界面填写。
2. 连接账号并预览已经发布的文章。
3. 勾选需要导入的文章，关联已有作品或创建对应的作品记录。
4. 执行 DEV.to 采集，在「平台账号」及历史统计中检查阅读、粉丝和互动数据。

凭据仅保存在本地，MCP 不提供接受原始 API Key 的账号接入工具。目前已经实现的创作者平台 API 集成是 DEV.to；其他创作者平台不能直接视为已支持。

### 3.3. Case 3：接入网站与搜索统计

1. 先在「项目管理」登记网站作品。
2. 打开「设置 → 数据采集 → 搜索与站点统计」，选择 Umami、Google Search Console（GSC）或 Bing Webmaster。
3. 填写来源要求的站点/属性标识、时区、统计窗口和授权凭据。GSC 需要自己有读取权限的属性；Bing 需要已验证站点及 API Key；Umami 支持 Token 或另外明确授权的本地浏览器会话。
4. 保存并明确启用数据源，然后执行采集。
5. 检查来源状态和统计结果。**保存成功不等于采集成功。**

各平台的指标口径不同，应保留各自统计范围。浏览器已经登录不代表可以直接导出 Cookie 或提取凭据。

### 3.4. Case 4：审核引用与传播

先在「分发渠道」登记确认公开的发布链接，再到「引用与采用」提交引用线索，查看原始网页或 API 证据，并明确审核通过或拒绝。仅有二手报告不能证明原始引用。「发现任务」用于记录交由 AI 客户端执行的检索任务，创建任务本身不会执行网络搜索。历史统计和传播路径会区分待审核线索与已确认结果。

## 4. Document

### 4.1. 架构与存储

Reverfolio 由 Go 服务、嵌入式 Web UI、本地 SQLite 和 MCP 服务组成。Web 与 MCP 使用同一套业务操作；HTTP 仅监听本机回环地址，MCP 也支持本地 stdio。Windows 浏览器扩展复用相同管理界面，通过已安装的 Native Messaging Host 按需拉起本地服务。

作品、配置、指标、引用证据和审核记录存放在本地。访问第三方 API 时，只能使用用户明确配置且拥有相应权限的凭据。

### 4.2. 配置

以 [config.example.toml](config.example.toml) 为配置字段依据。安装程序在 `bin/` 同级创建私有 `config.toml`，SQLite 默认存放在 `Data/`。Windows 为扩展使用 `19452` 端口，配置模板和 Unix 安装使用 `18440`。

```toml
[server]
listen = "127.0.0.1:18440"

[storage]
database = "Data/state.sqlite"

[collection]
interval = "24h"

[github]
auth = "env"
```

`env` 模式通过 `GITHUB_TOKEN` 授权 GitHub；若启动服务的账户已经通过 GitHub CLI 登录，也可以设置 `github.auth = "gh"`。`-collect-interval 0` 禁用定时采集，命令行明确指定的参数优先于配置文件。DEV.to 账号不会自动定时轮询；需要自动采集 GitHub 时，本地服务必须保持运行。

HTTP API 使用独立的私有凭据，默认保存在 `Data/api-token`；也可以提供不少于 32 字符的 `REVERFOLIO_API_TOKEN`。本地 Web UI 使用自己的会话，不需要手动粘贴 API Token。升级时应沿用原来的 `config.toml` 与 `Data/`。

### 4.3. MCP

**本地 stdio MCP 客户端**填写已安装二进制和配置文件的绝对路径：

```json
{
  "mcpServers": {
    "reverfolio": {
      "command": "/absolute/path/Reverfolio/bin/reverfolio",
      "args": ["-mcp-stdio", "-config", "/absolute/path/Reverfolio/config.toml"]
    }
  }
}
```

Windows 改为 `bin/reverfolio.exe`。重新连接 MCP 后，调用 `get_dashboard`、`list_works`、`get_collection_status`，核对读取的是同一份数据及当前服务信息。stdio 进程不会启动第二个定时采集器。

MCP 既有读取总览、作品、数据源和证据的只读工具，也有导入、采集和审核等写操作。写操作需要唯一的 `request_id`，部分操作还需要预期版本号。远程 ChatGPT 使用场景还需要自行配置经过认证的 MCP Tunnel；作者私有连接不是公开的一键安装入口。不要把本地 HTTP 端口直接暴露到公网。

### 4.4. Windows 浏览器扩展

Windows 安装完成后，打开 `edge://extensions` 或 `chrome://extensions`，启用「开发者模式」，选择「加载解压缩的扩展程序」，指向安装目录内的 `extension/`。点击扩展图标进入同一个工作台，Native Messaging Host 可以在服务未启动时将它拉起。

浏览器安全限制使用户仍需加载扩展，但**不需要额外手动注册 Native Messaging Host**。升级后重新加载扩展即可。Linux/macOS 可以直接使用 Web UI 和本地 stdio MCP，不依赖扩展。

### 4.5. 备份、升级与排错

- **升级前：**备份 `config.toml`、`Data/` 和私有平台凭据；不要用新模板覆盖旧配置。
- **升级后数据为空：**检查当前配置路径和 `storage.database`，不要在 WSL 再创建第二份可写数据库。
- **GitHub 流量缺失或出现 401/403：**检查仓库授权和凭据；不可用不代表零。
- **扩展无法启动服务：**在安装目录重跑 `Start-Reverfolio.cmd`，并检查本地健康接口。
- **MCP 仍显示旧工具：**更新后重连客户端并新建会话。
- **离线恢复：**使用 `reverfolio -restore BACKUP -db NEW_PATH`，恢复到**新数据库路径**，不能直接覆盖正在使用的数据库。

## 5. Contributing

### 5.1. 从源码编译和测试

安装 [go.mod](go.mod) 要求的 Go 工具链；运行仓库检查还需要 Python 3 和 Node.js。克隆源码后执行：

```sh
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

Windows 可以编译为 `reverfolio.exe`，并在 PowerShell 使用相同参数执行。

### 5.2. 扩展检查与打包

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
python3 -m unittest discover -s scripts -p 'test_publications_to_catalog.py'
node scripts/check_extension.mjs
node scripts/check_dashboard.mjs
```

MCP 模拟验收使用隔离数据库，不要求真实平台凭据。CI 还包括其他扩展检查、Go race/vet 和跨平台构建；模拟测试不能代替真实账号授权和原生图形界面验收。

生成发行包可执行 `python3 scripts/package_release.py vX.Y.Z`，将占位符换成目标版本。打包流程验证六种系统/架构组合、SHA-256 和第三方依赖许可原文。代码变更采用功能分支、PR Review 和 Squash Merge。

## 6. 安全与许可

Reverfolio 是本地运行的工具，**不是**对公网开放的 API。HTTP API 与 MCP 使用独立凭据；Windows 扩展使用安装时生成的本地访问能力。应保护操作系统账户、`Data/`、SQLite、私有配置和 Token，不要将它们提交到 Git 或打进发行包；文章内容、浏览器资料和统计报告也应保持私有。通过 MCP 明确返回给客户端的数据可能被传输到该客户端或其 AI 服务。

安全问题请通过 **contact@mail.thinkerqaq.com** 私下报告，不要在公开 Issue 中发布个人数据或有效凭据。

原创代码及项目视觉资源采用 [MIT License](LICENSE)。用户文章、摄影、平台数据和第三方素材不包含在此许可内。依赖保留各自许可证：`go.mod`/`go.sum` 记录依赖信息，发行包中的 `THIRD_PARTY_LICENSES.txt` 包含上游许可和声明原文，打包时自动生成并验证。
