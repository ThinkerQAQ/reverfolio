# Reverfolio

[English](README.md)

## 项目简介

面向开源项目和公开文章的本地影响力统计工具。登记作品及发布链接，采集 GitHub、网站及 DEV.to 指标，查看图表并审核引用证据。一个 Go 服务通过 SQLite 为 Web 与 MCP 提供相同数据。Inkerio、SyndiRelay 和浏览器扩展均为可选，全新安装不预置作者账户或作品。下载数、粉丝数不等于经过核实的人数或用户数。

## 快速开始

### 1. 前置条件

Windows、Linux 或 macOS（x64/ARM64）、浏览器和对应发行包。使用发行包不需要 Go 或文章分发应用。登记作品无需第三方凭据，读取私有统计才需要自己的服务授权。

### 2. 下载与安装

从 [Releases](https://github.com/ThinkerQAQ/reverfolio/releases/latest) 下载压缩包和 `SHA256SUMS`，核对 SHA-256 后解压到固定私有目录。压缩包命名为 `Reverfolio-vVERSION-OS-ARCH`，OS 为 `windows`、`linux` 或 `darwin`，ARCH 为 `amd64` 或 `arm64`。

**Windows：**在解压后的 `Reverfolio` 目录打开 PowerShell：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Install-Reverfolio.ps1
.\Start-Reverfolio.cmd
```

访问 <http://127.0.0.1:19452>。安装程序创建配置与 Data，不覆盖已有设置或 SQLite。

**Linux/macOS：**在解压后的 `Reverfolio` 目录运行：

```sh
# 仅限全新安装，不要覆盖已有私有配置。
cp config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

访问 <http://127.0.0.1:18440>，保持前台进程运行。Unix 发行包不自动安装 systemd/launchd 服务。macOS 二进制未签名，可能需要系统授权。

### 3. 登记第一个项目

进入「项目」，新增一个项目，填写标题及自己拥有的 GitHub 仓库，格式为 `owner/name`，保存并刷新，确认项目仍在。使用同一配置重启服务，确认记录仍存在，即完成一次真实的持久化目录操作，无需 API Key。

需要首次数值统计时，使用该项目的采集操作。公开 Release 下载计数可能无需 Token；私有流量需要有权限的 `GITHUB_TOKEN`，或设置 `github.auth = "gh"` 并使用自己原生系统中的 GitHub CLI 登录。查看采集状态、时间和可用指标。仓库没有 Release 或缺少权限时，显示不可用是合理结果，不能当作零或声称已成功读取流量。首次累计快照只是基线。

### 4. 验证与排错

在另一个终端检查健康状态：

```powershell
# Windows
Invoke-RestMethod http://127.0.0.1:19452/healthz
```

```sh
# Linux/macOS
curl -fsS http://127.0.0.1:18440/healthz
```

完成标准：`status=ok`、项目在重启后仍存在、请求的采集有明确的成功/部分成功/失败状态。仅健康检查通过不算完整验收。

- **升级后工作区为空：**检查 `storage.database` 和配置路径，不要另建可写 WSL 数据库。
- **401/403 或缺少流量：**检查对应仓库的授权，不可用不等于零。
- **ChatGPT 仍是旧工具：**升级后刷新 MCP 连接并新建聊天。
- **Windows 扩展不能拉起 Core：**在原固定目录重跑安装，用 `Start-Reverfolio.cmd` 手动启动查看问题。
- **升级/回滚：**先备份 SQLite 与凭据，回滚时使用兼容版本的二进制和数据库。CLI 恢复使用 `-restore BACKUP -db NEW_PATH`，不要覆盖当前数据库。

## 配置说明

默认配置以 [config.example.toml](config.example.toml) 为准。私有 `config.toml` 位于 `bin/` 旁，可用 `-config` 指定其他文件。常用设置：

```toml
version = 1
[server]
listen = "127.0.0.1:18440"
[storage]
database = "Data/state.sqlite"
[collection]
interval = "24h"
[github]
auth = "env"
```

Windows 安装程序使用扩展要求的 `19452`，应保留生成值。路径由配置加载器解析。`env` 模式从 `GITHUB_TOKEN` 读取自己的 GitHub 凭据；`gh` 模式要求运行服务的账户已有 CLI 授权。可用至少 32 字符的服务端 `REVERFOLIO_API_TOKEN`，否则程序生成私有本地 API Token，不要将它贴入页面脚本或聊天。

「设置 → 数据源」控制采集周期和来源开关。`-collect-interval 0` 关闭定时采集，显式 CLI 参数优先；机器和服务必须保持运行。DEV.to 账户目前不自动定时轮询。本地 UI 使用独立会话，无需手动输入 Bearer Token。

## 可选功能

### 本地 MCP

在本地 stdio 客户端的服务设置中，填写安装后二进制和配置的绝对路径：

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

Windows 使用 `bin/reverfolio.exe`，可用正斜杠路径。重连后调用 `get_dashboard`、`list_works`、`get_collection_status`，确认能读到同一项目和服务版本。stdio 不重复运行定时采集。写工具需要唯一 `request_id` 和适用的版本检查。远程 ChatGPT 还需要自己的认证 MCP Tunnel / 连接；开发者私有应用不是公开的一键安装入口。不要公开暴露本机 HTTP 服务。

### 网站与创作者指标

- **网站发现：**在「网站发现」预览自己的 RSS/Atom/Sitemap，勾选拥有的文章 URL 后导入。保留既有作品与明确的发布映射，不按同名标题合并身份。
- **DEV.to：**在本地「创作平台」输入个人 API Key，连接、预览已发布文章，明确选择并关联作品，再采集粉丝/阅读/互动/评论。凭据只写不读，MCP 不接收原始密钥。历史域名匹配建议必须人工确认。其他创作者平台集成尚未实现。
- **Umami/GSC/Bing：**先登记网站，再进入「设置 → 数据源 → 搜索与网站统计」，选择作品、属性 ID、时区、回看窗口和延迟；新来源需明确启用。GSC 需要已授权的只读 Search Console 凭据与属性，例如 `sc-domain:example.org`；Bing 需要已验证站点 URL 与授权 Webmaster API Key。Umami 支持授权 Token，或另行授权的本地浏览器登录会话。保存不等于采集，不要导出 Cookie，也不要因浏览器已登录就自行生成凭据。
- **目录交接：**「项目 → 导入作品与发布链接」接受目录 JSON，先预览再勾选。`scripts/publications_to_catalog.py --help` 说明可选分发导出转换器。不共享分发工具数据库或账户会话。
- **证据与图表：**附加确认过的语言版/分发 URL，提交外部引用并明确审核，再查看传播路径、历史图表或导出 SVG/PNG/CSV。保留来源时区、观测窗口、缺失值与证据来源，不能把不同平台计数相加当作独立用户数。

### Windows 扩展

安装后在 `edge://extensions` 或 `chrome://extensions` 开启开发者模式，加载安装目录的 `extension/`。点击图标进入同一工作台，通过已注册 Native Messaging 静默拉起服务。更新后重新加载；旧扩展 ID 与固定清单身份不同时移除并重加。生成的本地访问能力是私有信息。原生唤醒安装仅限 Windows，Unix 无需扩展即可使用 Web 与 stdio。

## 开发与验证

```sh
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o reverfolio ./cmd/reverfolio
./reverfolio -db ./Data/state.sqlite -collect-interval 0
```

按 `go.mod` 使用 Go **1.26+**。Windows 可构建 `reverfolio.exe` 并使用相同参数。源码默认端口为 18440。

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
python3 -m unittest discover -s scripts -p 'test_publications_to_catalog.py'
```

MCP 模拟验收使用隔离数据库，检查完整 39 工具契约及写入/读取/审核/导出流程，无需真实凭据。CI 还运行 Web/扩展检查、race/vet 及 Windows/Linux/macOS 检查。打包使用 `python3 scripts/package_release.py vX.Y.Z`，将版本替换为目标版本，生成六种系统/CPU 归档及依赖许可原文。原生图形界面、账户授权、真实重启不能由模拟测试或交叉编译代替。图标可通过 `scripts/build-icons.sh` 重建，原始 SVG 与派生资源随源码许可。使用功能分支、PR 和 Squash Merge。

## 许可证与合规

原创代码及项目视觉资源采用 [MIT](LICENSE)，依赖保留[各自声明](THIRD_PARTY_NOTICES.md)，发行包包含依赖许可原文。许可不覆盖用户文章、摄影、平台数据或第三方内容。SQLite、私有 TOML、Token、扩展访问能力与报告不得进入 Git/发行包。MCP 选择返回的数据会进入客户端或 AI 服务。凭据应在本地输入，账户已登录不代表授权获取密钥。详见[安全说明](SECURITY.md)。
