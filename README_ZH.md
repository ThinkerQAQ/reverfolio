# Reverfolio

[English](README.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

Reverfolio 用本地 SQLite 关联开源项目、公开文章、平台统计与引用证据，通过 Web UI 和 MCP 统一管理。

同一个项目或文章可能分布在 GitHub、个人网站和多个内容平台。Release 下载量、仓库流量、搜索曝光、页面浏览量与互动量的统计范围和时间窗口并不相同，不能简单相加并当作真实人数。

Reverfolio 将统计数据与作品、来源和采集时间关联，主要解决四个问题：

- 独立管理项目、文章和网站，不要求使用特定博客或分发工具。
- 区分“没有数据”“采集失败”和实际测量结果为零。
- 保留统计时间窗口、来源、采集状态与历史基线。
- 对外部引用和采用证据进行人工审核，不把未经核实的线索当作成果。

Inkerio 和 SyndiRelay 均为可选组件。全新安装不会预置作者账户或作品。

## 2. Quick Start

下面以登记一个 GitHub 项目并采集公开指标为例，从安装、Windows 扩展连接，到重启验收，走完第一次使用流程。

### 2.1. 准备运行环境

使用 Windows、Linux 或 macOS（x64/ARM64）。Windows 扩展支持 Edge 和 Chrome；Linux/macOS 发行包通过 Web UI 使用，**不包含 Windows 浏览器扩展**。直接使用发行包无需 Go、Node.js、文章分发工具或 API Key。

查看 CPU 架构：Windows 打开「设置 → 系统 → 关于」，macOS 打开「关于本机」，Linux 执行 `uname -m`（`x86_64` 为 amd64，`aarch64` 为 arm64）。

### 2.2. 下载与安装

1. 打开[最新 Release](https://github.com/ThinkerQAQ/reverfolio/releases/latest)，根据操作系统和 CPU 架构选择文件名以 `windows-amd64.zip`、`windows-arm64.zip`、`linux-amd64.tar.gz`、`linux-arm64.tar.gz`、`darwin-amd64.tar.gz` 或 `darwin-arm64.tar.gz` 结尾的压缩包。
2. 同时下载 `SHA256SUMS`。在下载目录校验：Windows 执行 `Get-FileHash .\Reverfolio-*.zip -Algorithm SHA256`；Linux 执行 `sha256sum ./Reverfolio-*.tar.gz`；macOS 执行 `shasum -a 256 ./Reverfolio-*.tar.gz`。与 `SHA256SUMS` 中**相同文件名**的记录对比，不一致则停止安装。
3. 将**完整压缩包**解压到固定私有目录。解压后的 `Reverfolio/` 是安装根目录，应直接包含 `bin/` 和 `config.example.toml`；Windows 还包含 `extension/` 和 `Start-Reverfolio.cmd`。不要在压缩包内部直接运行程序。

**Windows：**打开 PowerShell，将示例路径替换为实际解压目录：

```powershell
Set-Location 'C:\path\to\Reverfolio'
.\Start-Reverfolio.cmd
```

启动脚本会按需完成安装，创建 `config.toml`、`Data/` 且不覆盖旧数据，初始化扩展访问凭据、自动注册 Edge/Chrome Native Messaging Host，并在后台启动服务。**无需手动注册 Host，也没有配对码。** 然后访问 [http://127.0.0.1:19452](http://127.0.0.1:19452)；首次启动未就绪时刷新页面即可。

**Linux/macOS：**在解压后的 `Reverfolio/` 目录打开终端：

```bash
# 仅首次安装时复制，升级时不能覆盖已有配置。
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

保持终端运行，打开 [http://127.0.0.1:18440](http://127.0.0.1:18440)。Unix 发行包不会自动安装后台服务；macOS 未签名程序可能需要系统授权。

### 2.3. 安装浏览器扩展（Windows）

**先完成 [2.2](#22-下载与安装)**，再加载扩展，保证安装程序已生成本机私有访问文件。

1. Edge 打开 `edge://extensions`，Chrome 打开 `chrome://extensions`。
2. 开启**开发者模式**，点击**加载解压缩的扩展程序**。
3. 选择已解压 Reverfolio 安装目录内的 **`extension/` 文件夹**，其中应直接包含 `manifest.json`。不要选择 ZIP、`bin/` 或仓库根目录。
4. 在浏览器工具栏的扩展菜单找到 Reverfolio，可以固定图标；点击图标打开的是**侧边栏**，不是弹窗。
5. 确认左侧显示「**本地服务 → 已连接**」，再点击「**项目管理**」确认页面可以正常读取。扩展侧边栏与[本地 Web UI](http://127.0.0.1:19452) 使用同一份数据。

如果出现「服务未连接」，点击**启动本地服务**，再点击**重试连接**。提示本机启动器不可用或未授权时，在**原安装目录**重新执行 `.\Start-Reverfolio.cmd`，回到扩展管理页点击**重新加载**。如果移动了安装目录，也要在新位置重新运行启动脚本以更新注册路径。**浏览器中手动加载扩展是唯一需要用户完成的扩展安装操作；Native Messaging 注册由安装脚本负责。**

Linux/macOS 用户**跳过本节**，直接使用 Web UI。

### 2.4. 登记并采集第一个项目

Windows 在扩展侧边栏继续，Linux/macOS 在 Web UI 继续：

1. 点击左侧「**项目管理**」。
2. 向下找到页面**底部**的「**新增项目或文章**」表单。它是表单标题，**不是独立按钮**。
3. 「**作品类型**」选择「**GitHub 项目**」，填写「**标题**」，在「**仓库路径**」输入要统计的公开仓库 `owner/repo`。
4. 点击「**新建**」，确认出现项目卡片。
5. 在该卡片点击「**更新 GitHub**」，通过「查看详情」或总览查看本次采集状态、时间和可用指标。

公开仓库指标可能无需 Token，但私有流量需要 GitHub 授权。仓库没有 Release、接口无权限或采集出错时，部分指标可能不可用，不能当作零。首次累计指标只是**基线**，不是新增量。

### 2.5. 检查结果

在另一个终端执行对应系统的健康检查：

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```bash
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

应返回包含 `"status":"ok"` 的 JSON。刷新「项目管理」确认作品仍在；用**相同的 `config.toml`** 重启服务，再检查作品是否保留。Windows 还应确认扩展可以继续读取项目。

**验收标准：**Windows 扩展已连接；所有平台的项目在重启后仍存在；发起采集后能明确区分成功、部分成功与失败。**HTTP 健康检查通过不代表指标采集成功。**

### 2.6. 常见问题

- **Windows 打不开工作台：**从安装根目录重新运行 `Start-Reverfolio.cmd`，检查 `http://127.0.0.1:19452/healthz`，核对 `config.toml` 仍监听 `19452` 且端口未被占用。
- **扩展缺失或无法连接：**确认加载的是直接包含 `manifest.json` 的 `extension/`，而不是 ZIP。重新加载扩展、重跑安装脚本修复 Native Messaging。检查 `Data/extension-access.key` 与 `extension/ui/local-auth.js` 是否存在，**不要分享文件内容**。
- **扩展提示未授权：**在加载扩展的安装目录重新执行启动脚本，再重新加载扩展；不需要手动粘贴 API Token。
- **Linux/macOS 看不到扩展：**当前发行包按设计使用 Web UI 或本地 MCP，不包含 Windows 扩展。
- **升级后项目消失：**检查原来的 `config.toml`、`storage.database`，不要在 WSL 中另建可写数据库。
- **GitHub 401/403 或缺少指标：**核对仓库权限、接口限额和授权凭据。不可用不等于零。
- **仍显示旧界面或旧 MCP 工具：**更新程序、重新加载扩展、重连 MCP。

仍未解决时，[提交 Issue](https://github.com/ThinkerQAQ/reverfolio/issues)，附上操作系统、安装方式、复现步骤和脱敏错误信息，不要上传 `Data/` 或私有凭据。

## 3. Configuration

### 3.1. 常用配置

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

### 3.2. 升级、备份与恢复

- **升级前：**备份 `config.toml`、`Data/` 和私有平台凭据；不要用新模板覆盖旧配置。
- **升级后数据为空：**检查当前配置路径和 `storage.database`，不要在 WSL 再创建第二份可写数据库。
- **GitHub 流量缺失或出现 401/403：**检查仓库授权和凭据；不可用不代表零。
- **扩展无法启动服务：**在安装目录重跑 `Start-Reverfolio.cmd`，并检查本地健康接口。
- **MCP 仍显示旧工具：**更新后重连客户端并新建会话。
- **离线恢复：**使用 `reverfolio -restore BACKUP -db NEW_PATH`，恢复到**新数据库路径**，不能直接覆盖正在使用的数据库。

## 4. Optional Features

以下 Case 均为可选能力，可以独立使用。第一次接入建议先完成 Quick Start。

### 3.1. Case 1：从自己的网站导入文章

1. 打开「网站发现」，输入自己管理的公开网站地址或 Feed 地址。
2. 预览 RSS、Atom 或 Sitemap 中发现的候选文章，核对标题和 URL。发现候选文章并不等于已登记发布。
3. 勾选需要的文章，执行导入。
4. 在「项目管理」和「分发渠道」中检查导入的作品与发布链接。

只有明确选择的 URL 才会导入。Feed 和 Sitemap 不保证覆盖所有文章；标题相同也不会自动合并不同作品。

### 4.2. Case 2：接入 DEV.to

1. 获取自己的 DEV.to API Key，在本地「平台账号」界面填写。
2. 连接账号并预览已经发布的文章。
3. 勾选需要导入的文章，关联已有作品或创建对应的作品记录。
4. 执行 DEV.to 采集，在「平台账号」及历史统计中检查阅读、粉丝和互动数据。

凭据仅保存在本地，MCP 不提供接受原始 API Key 的账号接入工具。目前已经实现的创作者平台 API 集成是 DEV.to；其他创作者平台不能直接视为已支持。

### 4.3. Case 3：接入网站与搜索统计

1. 先在「项目管理」登记网站作品。
2. 打开「设置 → 数据采集 → 搜索与站点统计」，选择 Umami、Google Search Console（GSC）或 Bing Webmaster。
3. 填写来源要求的站点/属性标识、时区、统计窗口和授权凭据。GSC 需要自己有读取权限的属性；Bing 需要已验证站点及 API Key；Umami 支持 Token 或另外明确授权的本地浏览器会话。
4. 保存并明确启用数据源，然后执行采集。
5. 检查来源状态和统计结果。**保存成功不等于采集成功。**

各平台的指标口径不同，应保留各自统计范围。浏览器已经登录不代表可以直接导出 Cookie 或提取凭据。

### 4.4. Case 4：审核引用与传播

先在「分发渠道」登记确认公开的发布链接，再到「引用与采用」提交引用线索，查看原始网页或 API 证据，并明确审核通过或拒绝。仅有二手报告不能证明原始引用。「发现任务」用于记录交由 AI 客户端执行的检索任务，创建任务本身不会执行网络搜索。历史统计和传播路径会区分待审核线索与已确认结果。

### 4.5. Case 5：本地 MCP 连接

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

## 5. Development

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

## 6. License

Reverfolio 是本地运行的工具，**不是**对公网开放的 API。HTTP API 与 MCP 使用独立凭据；Windows 扩展使用安装时生成的本地访问能力。应保护操作系统账户、`Data/`、SQLite、私有配置和 Token，不要将它们提交到 Git 或打进发行包；文章内容、浏览器资料和统计报告也应保持私有。通过 MCP 明确返回给客户端的数据可能被传输到该客户端或其 AI 服务。

安全问题请通过 **contact@mail.thinkerqaq.com** 私下报告，不要在公开 Issue 中发布个人数据或有效凭据。

原创代码及项目视觉资源采用 [MIT License](LICENSE)。用户文章、摄影、平台数据和第三方素材不包含在此许可内。依赖保留各自许可证：`go.mod`/`go.sum` 记录依赖信息，发行包中的 `THIRD_PARTY_LICENSES.txt` 包含上游许可和声明原文，打包时自动生成并验证。
