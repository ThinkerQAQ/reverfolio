# Reverfolio

[English](README.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

开源项目和文章发布后，相关反馈往往散落在不同平台：GitHub 有 Star、Release 下载和仓库流量，个人网站有访问和搜索数据，内容平台也有各自的阅读和互动统计。想持续了解作品的传播情况，就得反复查找和整理这些信息。

Reverfolio 将这些记录集中到一个地方。你可以登记自己的开源项目和文章，查看已接入平台的统计数据及历史变化，记录其他网站或项目对作品的引用，并核实相关证据。它可以独立使用：先添加一个 GitHub 仓库就能开始，之后再按需接入网站或内容平台。

## 2. Quick Start

以统计一个公开 GitHub 仓库为例，从下载、安装到查看结果，完成第一次使用。Windows 用户会同时安装浏览器扩展。

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

Windows 在扩展侧边栏继续；Linux/macOS 在 Web UI 继续：

1. 打开左侧「项目管理」，向下找到页面底部的「新增项目或文章」表单。
2. 「作品类型」选择「GitHub 项目」，填写标题；「仓库路径」输入 `ThinkerQAQ/reverfolio`（公开测试仓库），也可以换成自己的 `owner/repo`。
3. 点击「新建」，确认项目卡片出现在列表中。
4. 点击该卡片上的「更新 GitHub」，随后打开「查看详情」或「影响力总览」，检查采集结果和时间。

公开数据可能无需凭据即可读取；部分仓库流量需要 GitHub 授权。第一次读到的累计下载量是统计基线，不是本次新增量。

### 2.5. 检查结果

在另一个终端执行与系统对应的健康检查：

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```bash
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

返回结果应包含 `"status":"ok"`。然后刷新「项目管理」，检查刚添加的项目是否还在。

还可以重启服务，确认数据确实写入了本地数据库：

- **Windows：**在任务管理器「详细信息」中找到并结束本次安装启动的 `reverfolio.exe`，回到安装目录执行 `Start-Reverfolio.cmd`，重新打开工作台。
- **Linux/macOS：**在运行服务的终端按 `Ctrl+C`，再执行 `./bin/reverfolio -config ./config.toml`。

重启后项目仍存在，说明登记数据已持久化。采集可能显示成功、部分成功或失败；例如公开下载量已获取，但仓库流量缺少授权时，不能把未获取的数据当成零。

### 2.6. 常见问题

- **工作台或扩展无法连接：**先检查 [Windows 本地地址](http://127.0.0.1:19452) 或 [Linux/macOS 本地地址](http://127.0.0.1:18440)。Windows 在安装目录重新运行 `Start-Reverfolio.cmd`，然后重新加载浏览器扩展。
- **扩展提示未授权：**确认加载的是当前安装目录下包含 `manifest.json` 的 `extension/` 文件夹，重新执行安装脚本并加载扩展。
- **GitHub 采集失败：**检查仓库访问权限、API 限额和本地授权配置；未获取到的数据不代表零。
- **升级后找不到项目：**检查当前使用的 `config.toml` 和 `storage.database`，确保没有切换到另一份数据库。

仍无法解决时，可[提交 Issue](https://github.com/ThinkerQAQ/reverfolio/issues)，附上系统、复现步骤和脱敏错误信息，不要上传完整 `Data/` 或任何凭据。

## 3. Configuration

### 3.1. 常用配置

私有配置位于安装目录的 `config.toml`，数据保存在 `Data/`。完整字段以 [config.example.toml](config.example.toml) 为准，也可以在「设置 → 数据采集」中调整相关选项。

- `server.listen`：本地服务地址。Windows 扩展使用 `127.0.0.1:19452`；Linux/macOS 默认使用 `127.0.0.1:18440`。
- `storage.database`：SQLite 文件路径，通常是 `Data/state.sqlite`。
- `collection.interval`：自动采集周期；`0` 表示关闭。采集期间服务需要保持运行。
- `github.auth`：`env` 使用 `GITHUB_TOKEN`；`gh` 使用当前服务账户已授权的 GitHub CLI。
- `github.download_scope`：选择全部 Release 资产或常见安装包的下载统计。
- `interface.language`：界面语言 `zh-CN` 或 `en`。

本地 Web UI 和已安装的扩展不需要手动输入 API Token。连接第三方统计平台时，在对应设置页面填写自己的授权信息。

### 3.2. 升级、备份与恢复

1. 退出 Reverfolio，备份 `config.toml`、`Data/` 和私有平台凭据。
2. 下载新版发行包，更新原安装目录中的程序文件，**保留原有配置和数据**。
3. Windows 重新运行 `Start-Reverfolio.cmd` 并在浏览器中重新加载扩展；Linux/macOS 使用原 `config.toml` 启动。
4. 打开「项目管理」，确认已有作品和历史统计仍在。

如需恢复 SQLite 备份，先停止服务，再运行 `reverfolio -restore BACKUP -db NEW_PATH`（Windows 为 `reverfolio.exe`）；恢复目标必须是**新的数据库路径**，不能直接覆盖现有数据库。

## 4. Optional Features

以下 Case 均为可选能力，可以独立使用。第一次接入建议先完成 Quick Start。

### 4.1. Case 1：从自己的网站导入文章

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

### 4.4. Case 4：审核外部引用

1. 在「分发渠道」登记作品已经公开的原文、译文或分发链接。
2. 在「引用与采用」提交发现的外部引用页面。
3. 查看原始页面或 API 证据，确认是否与自己的作品有关，再选择审核通过或拒绝。
4. 到传播路径或历史统计中查看已确认的记录。

「发现任务」用于把检索工作交给外部 AI 客户端；Reverfolio 不会仅因创建任务就自动搜索互联网。

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

### 5.1. 从源码编译与运行

开发需要 Git、符合 [go.mod](go.mod) 声明的 Go 工具链；运行完整仓库检查还需要 Python 3、Node.js。

**Linux/macOS：**

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
mkdir -p Data
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

**Windows PowerShell：**

```powershell
git clone https://github.com/ThinkerQAQ/reverfolio.git
Set-Location reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio.exe ./cmd/reverfolio
New-Item -ItemType Directory -Force Data | Out-Null
.\reverfolio.exe -db .\Data\state.sqlite -collect-interval 0
```

源码启动后打开 [http://127.0.0.1:18440](http://127.0.0.1:18440)，按照 [2.4](#24-登记并采集第一个项目) 登记测试仓库即可。源码 Web UI 不要求安装扩展。

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
