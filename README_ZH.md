# Reverfolio

[English](README.md)

- [1. Introduction](#1-introduction)
- [2. Quick Start](#2-quick-start)
- [3. Configuration](#3-configuration)
- [4. Optional Features](#4-optional-features)
- [5. Development](#5-development)
- [6. License](#6-license)

## 1. Introduction

开源项目和文章发布后，反馈往往散落在不同地方：GitHub 有下载和仓库流量，网站有访问与搜索数据，其他作者可能引用文章，也可能将其中的方法用于自己的项目。只看下载量，无法说明谁使用了作品；只保存链接，又很难说明一次贡献最终产生了什么结果。

Reverfolio 在本机整理这些记录：登记作品和发布链接，采集已有接口支持的指标，通过 AI 客户端查找外部线索，再保留原始证据、审核并归并为影响事件。它也能记录上游 PR 被接纳、排障协作、转载和目录收录，并保留它们各自的性质。

> 请使用同一 Release 提供的安装器、便携包、启动脚本与扩展，不要混用不同版本的文件。

## 2. Quick Start

以一个公开 GitHub 项目为例，先完成安装、登记和第一次采集。想让 ChatGPT 协助检索，再继续完成连接；浏览器扩展可稍后安装。

### 2.1. 下载适合自己的构建

打开 [Releases](https://github.com/ThinkerQAQ/reverfolio/releases)，根据操作系统和 CPU 架构选择文件。同一次安装使用同一个发布或预览构建中的程序、启动脚本与扩展。

| 系统 | 选择与运行方式 |
| --- | --- |
| Windows | 优先选择对应架构的原生安装器 `.exe`；便携版为 `windows-amd64.zip` 或 `windows-arm64.zip` |
| Linux | 选择 `linux-amd64.tar.gz` 或 `linux-arm64.tar.gz`，使用解压后的启动脚本 |
| macOS | Apple Silicon 选择 `darwin-arm64.tar.gz`，Intel 选择 `darwin-amd64.tar.gz`，使用解压后的启动脚本 |

Windows 在「设置 → 系统 → 关于」查看架构，macOS 查看「关于本机」，Linux 执行 `uname -m`：`x86_64` 对应 amd64，`aarch64` 对应 arm64。直接使用发行包无需安装 Go、Node.js 或文章分发工具。

下载该构建附带的 SHA-256 清单，计算所选文件的哈希并与清单中**同一文件名**的记录比较：Windows 使用 `Get-FileHash`，Linux 使用 `sha256sum`，macOS 使用 `shasum -a 256`。结果不一致时重新下载。

### 2.2. 安装并打开工作台

**Windows 原生安装器：**

1. 运行下载的安装器，选择固定安装目录。
2. 阅读并确认许可证，然后开始安装。组件与快捷方式采用默认安排，无需逐项选择。
3. 完成页保留 **Open Reverfolio**，点击完成。需要扩展时才勾选 **Set up the Chrome / Edge extension**，操作见 [4.1](#41-安装浏览器扩展windows)。
4. 浏览器打开工作台后，确认「本地服务」显示「已连接」，左侧能看到「我的作品」。以后从开始菜单打开 Reverfolio 即可。

初始化会创建缺失的 `config.toml` 和本机数据目录，并准备扩展所需的本地授权与启动器；已有配置和数据会保留。普通使用无需手动填写本地 Token、配对码或注册 Native Messaging Host。

**Windows 便携版：**完整解压到固定目录，不要在 ZIP 内运行。进入直接包含 `bin/`、`extension/`、`config.example.toml` 的安装根目录，打开 PowerShell 并运行：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Open-Reverfolio.ps1
```

启动器会初始化缺失配置、启动服务并打开实际配置的地址。

**Linux/macOS 便携版：**完整解压，在解压后的 `Reverfolio/` 目录运行：

```bash
bash ./Open-Reverfolio.sh
```

脚本创建缺失配置、在后台启动本地服务并打开浏览器。Linux 需要 `curl` 和 `xdg-open`；macOS 使用系统的 `open`。未签名程序可能需要按系统提示确认来源并允许运行。Unix 便携版使用 Web UI，不要求 Windows 浏览器扩展。

新配置的默认地址为 [http://127.0.0.1:19452](http://127.0.0.1:19452)。已有安装以 `config.toml` 的 `server.listen` 为准；启动器会打开该地址，不要另开一份配置或空数据库来解决连接问题。

### 2.3. 登记并采集第一个项目

1. 打开「我的作品 → 作品列表」，展开顶部的「新增项目或文章」。没有作品时，表单会自动展开。
2. 类型选择「GitHub 项目」，填写标题，仓库路径填写自己的 `owner/repository`。也可用公开仓库 `ThinkerQAQ/reverfolio` 验证流程。
3. 点击「新建」，确认作品卡片出现。
4. 在卡片上点击「更新 GitHub」，再到「查看详情」或「查看影响 → 影响总览」检查指标、采集时间和失败原因。

公共指标可能无需凭据；仓库流量通常需要相应读取权限。需要授权时，先安装 [GitHub CLI](https://cli.github.com/)（已安装可跳过）并在本机完成 `gh auth login`，再到「设置 → 数据采集 → GitHub → GitHub 授权方式」选择「本机 GitHub CLI」并重启服务。也可选择环境变量方式，让服务进程读取自己的 `GITHUB_TOKEN`。

第一次下载观测是累计基线，不代表这次新增下载。部分指标失败不等于全部失败，缺失也不等于零。

### 2.4. 连接 ChatGPT（需要 AI 协助时）

这一步把本机 Reverfolio 提供给自己的 ChatGPT 连接。需要账号或工作区允许创建自定义 MCP/Tunnel；没有该入口时仍可使用工作台，或使用 [4.5](#45-连接本地-mcp-客户端) 的本地客户端。

1. 打开 [OpenAI Tunnel 管理](https://platform.openai.com/settings/organization/tunnels)，选对自己的组织，创建 Reverfolio Tunnel 或复用已有的专用 Tunnel，记下 Tunnel ID。创建需要 Tunnels Read + Manage 权限；Tunnel 与之后使用的 ChatGPT 工作区应匹配。复用旧 Tunnel 前先停止原安装的 MCP 进程，同一 Tunnel 不同时运行两份客户端。
2. 打开同一组织的 [Runtime API keys](https://platform.openai.com/settings/organization/api-keys)，创建供该 Tunnel 运行的受限密钥，确保具有 Tunnels Read + Use 权限。已有合适的运行密钥可复用；不要使用 Admin API Key 运行服务。
3. 回到 Reverfolio，打开「设置 → 运行环境 → MCP 连接」，填写 Tunnel ID 和 Runtime API Key，点击「保存」。缺少官方客户端时会自动下载并校验；校验失败保留之前的配置。
4. 点击「启动服务」，等待连接状态显示 **Ready**。保存只更新配置；服务运行时修改连接信息，需要点击「重启服务」生效。「断开 Tunnel」保留本地 Tunnel 服务运行，「停止服务」关闭该进程。
5. 点击「打开 ChatGPT 插件设置」。新连接选择新增自定义 MCP，名称填写 Reverfolio、连接方式选择 **Tunnel**，选择同一个 Tunnel 并完成安装/授权；已有连接则打开该条目启用或刷新。
6. 在 ChatGPT 中选择该连接，发送下面的只读验收请求：

```text
使用 Reverfolio，只读调用 list_works 和 get_collection_status。
确认能看到刚添加的项目，并报告服务版本。不要采集或修改数据，不要输出凭据。
```

能读到刚登记的项目，才说明 ChatGPT 已连接到这份本机数据。「本机连接已就绪」只表示本地 Tunnel 健康，不代表 ChatGPT 已完成授权。运行密钥的创建与权限说明见 [官方 tunnel-client](https://github.com/openai/tunnel-client)。

「本地服务开机自启」控制下次登录时启动 Tunnel 服务；「Tunnel 自动连接」控制该服务启动后是否连接 Tunnel，两项独立设置。密钥只在本机填写，Windows 使用 DPAPI 保存；已保存的密钥不会回显。新 MCP 实现不读取旧脚本方案的配置，需要重新填写连接信息。

命令行与界面使用同一套管理逻辑。在安装根目录运行（Windows 使用 `bin\reverfolio.exe`）：

```bash
./bin/reverfolio mcp status
./bin/reverfolio mcp start
./bin/reverfolio mcp restart
./bin/reverfolio mcp connect
./bin/reverfolio mcp disconnect
./bin/reverfolio mcp stop
```

`mcp configure` 从标准输入读取包含 `tunnel_id`、`runtime_api_key` 的 JSON；使用 `reuse_key: true` 可复用已保存的密钥。不要把密钥写进命令行参数。`mcp auto-enable` / `auto-disable` 管理登录自启，`mcp auto-connect-enable` / `auto-connect-disable` 管理自动连接。自定义配置路径使用 `--config /path/to/config.toml`。

### 2.5. 完成第一次使用

以后按左侧顺序操作：

| 入口 | 要做的事 | 完成后去哪里 |
| --- | --- | --- |
| 我的作品 | 登记项目和文章，导入网站文章，关联公开发布链接 | 发现线索 |
| 发现线索 | 为作品建立任务，让已连接的 AI 客户端实际搜索并记录检查范围 | 审核证据 |
| 审核证据 | 核对原页，审核引用与采用，将相关证据关联为影响事件 | 查看影响 |
| 查看影响 | 查看总览、传播路径、历史统计，导出报告 | 按新线索继续检索 |

作品采集正常后，可以到「设置 → 数据采集 → GitHub → 自动采集计划」选择间隔。电脑和服务需要保持运行。设置位于侧栏底部。

再次从安装入口打开应用，确认项目仍在；连接 ChatGPT 的用户再运行一次只读请求。这就完成了本机保存和远程读取的验收。

## 3. Configuration

### 3.1. 在设置中调整

配置保存在安装根目录的 `config.toml`，数据库通常为 `Data/state.sqlite`。完整字段见 [config.example.toml](config.example.toml)。

| 需求 | 设置入口 | 生效方式 |
| --- | --- | --- |
| 连接 ChatGPT、本地客户端配置 | 运行环境 → MCP 连接 | 保存后启动或重启 Tunnel 服务 |
| 语言、布局、首页模块 | 界面与语言 | 保存后生效 |
| GitHub 凭据来源 | 数据采集 → GitHub → GitHub 授权方式 | 重启服务 |
| 采集周期、下载统计口径 | 数据采集 → GitHub | 周期一分钟内应用，展示口径保存后应用 |
| GSC、Bing、Umami 的地址、站点与授权 | 数据采集 → 搜索与站点统计 | 保存来源后显式采集并检查结果 |
| 审核人、默认筛选、任务认领名称 | 引用与审核、AI 检索任务 | 保存后用于后续操作 |
| 本机监听地址、数据库路径 | 运行环境 → 高级启动配置 | 重启服务；改地址后重新加载扩展 |

监听地址仅允许回环 IP。修改端口会同步扩展的本地地址、权限与连接策略。数据库路径只接受已有的 Reverfolio SQLite 文件；保存设置不会移动或复制旧数据。切换前先备份，确认选中的是目标数据库。

### 3.2. 升级、备份和恢复

1. 在「设置 → 统计和备份 → 数据库备份」导出 SQLite 备份，另行保留 `config.toml` 与私有平台凭据。
2. 停止这份安装的服务和 MCP 连接。Windows 可在任务管理器中按可执行文件路径确认对应的 `reverfolio.exe`；前台终端启动的服务使用 `Ctrl+C`。
3. 原生安装器选择原安装目录；便携版替换程序文件，保留自己的配置、数据库及私有凭据。
4. 从原安装入口重新打开，扩展用户在浏览器中重新加载扩展。确认作品、历史证据与设置仍在，再恢复 MCP 连接。

恢复备份时，先在服务停止后将备份放到一个新的文件路径，再将 `storage.database` 指向该文件。也可使用 `reverfolio -restore BACKUP -db NEW_PATH`（Windows 为 `reverfolio.exe`），恢复目标必须不存在。不要把正在使用的 SQLite 文件直接覆盖。

### 3.3. 遇到问题时检查哪里

| 现象 | 检查与恢复 |
| --- | --- |
| 工作台打不开 | 从同一安装目录重新启动，核对 `server.listen` 和端口占用。Unix 查看 `Data/service.log` |
| 扩展未连接或未授权 | 重新运行安装目录的启动入口，再重新加载该目录的扩展；需要时点击「启动本地服务」「重试连接」 |
| 本机已就绪，ChatGPT 不能调用 | 在插件设置中核对同一个 Tunnel，完成启用/授权，再用只读请求验收 |
| GitHub 或站点采集失败 | 看最新尝试的错误与最近成功观测，检查账号权限、凭据与网络。不要将失败值当作零 |
| 升级后看不到原数据 | 核对安装目录、`config.toml` 和数据库路径，先找回原库，不要重新导入造成重复 |

仍有问题时，可在 [Issues](https://github.com/ThinkerQAQ/reverfolio/issues) 提供系统、复现步骤和脱敏错误。

## 4. Optional Features

### 4.1. 安装浏览器扩展（Windows）

安装器完成页可打开扩展安装指引；也可以完成主流程后手动安装：

1. 先从 Reverfolio 的安装入口打开一次工作台，完成本地初始化。
2. Edge 打开 `edge://extensions`，Chrome 打开 `chrome://extensions`，启用「开发者模式」。
3. 点击「加载解压缩的扩展程序」，选择**当前安装目录的 `extension/`**，该目录应直接包含 `manifest.json`。
4. 点击扩展图标打开侧边栏，确认「本地服务 → 已连接」及刚登记的项目。

扩展与网页使用相同数据。浏览器要求用户手动加载未上架扩展；安装器负责本地授权与 Host 注册。移动安装目录后，重新运行启动入口并重新加载扩展。

### 4.2. 导入网站文章与 DEV.to

- 网站：进入「我的作品 → 从网站导入」，填写自己管理的公开网站或 Feed 地址，预览 RSS、Atom、Sitemap 候选项，再选择需要的文章导入。
- DEV.to：进入「我的作品 → 平台账号」，在本地填写自己的 API Key，连接后预览已发布文章，勾选并关联或导入，再执行统计采集。
- 导入后：在「作品列表」确认作品，在「发布链接」确认原文、译文和公开渠道地址。未选中的文章不导入，标题相同不会自动合并。

目前创作者平台统计实现的是 DEV.to 官方 API；其他平台不能视为已支持。它们也不要求安装 Inkerio 或 SyndiRelay。

### 4.3. 接入网站与搜索统计

先在「我的作品」登记网站，再进入「设置 → 数据采集 → 搜索与站点统计」。选择 Umami、GSC 或 Bing，填写来源标识、关联作品、站点/属性、时区和统计窗口，按来源要求在本机配置凭据。保存并启用后，显式采集一次，再检查来源状态与结果。

GSC 需要可读取的属性，Bing 需要已验证站点；Umami 支持 Token 或单独授权的本机浏览器会话。保存成功不代表采集成功，浏览器登录也不代表允许导出 Cookie。各来源的点击、曝光、访问标识和阅读量保留各自口径。

### 4.4. 检索、审核并归并影响事件

先为作品补齐「发布链接」；旧名称、旧 URL、译名或特征句可在「发现线索」的高级区域补充。然后向已连接的 AI 客户端提出明确请求，例如：

```text
使用 Reverfolio，为我选择的作品查找外部影响线索。
先读取作品、发布链接和 get_discovery_coverage，按该作品的检索计划搜索。
核对原始页面，区分本人分发、第三方引用、转载、目录收录、AI 生成内容与实际采用。
把候选证据保存为待审核，记录实际查询、来源、日期及未完成范围；不要自动审核通过。
```

AI 需要实际执行搜索、认领并按需续租任务，再关联候选证据记录结果。任务和检索计划本身不会启动全网爬取。

在「审核证据 → 引用与采用」打开原页、核对相关性与来源，再填写理由并审核。在「影响事件」把文章、Issue、PR 等相关页面归为一次事件，关联作品与证据，并记录提交、审核、合并或验证等阶段。没有结果证明时，保留「结果未确认」。最后到「查看影响」核对事件总数并导出报告。

上游合并证明贡献被接纳；转载或目录收录证明传播或曝光；它们都不能换算为真实使用人数。同一事件的多个页面不应重复计为多个采用者。未在本次范围找到结果，也不代表全网不存在。

### 4.5. 连接本地 MCP 客户端

在「设置 → 运行环境 → MCP 连接 → 本地 MCP 客户端（STDIO）」复制应用生成的 JSON，粘贴到自己本机 AI 客户端的 MCP 配置并重新连接。生成内容使用当前安装的绝对路径，无需照抄其他人的目录。

调用 `list_works`、`get_collection_status` 检查是否读到同一份数据。本地 stdio 连接不要求 ChatGPT Tunnel，也不会启动第二个定时采集器。写操作需要唯一 `request_id`；审核和事件更新等操作还需要预期版本号。

## 5. Development

### 5.1. 编译并运行

开发需要 Git、符合 [go.mod](go.mod) 的 Go 工具链；完整仓库检查还需要 Python 3 和 Node.js。

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go test -race ./...
go vet ./...
go build -o bin/reverfolio ./cmd/reverfolio
# 仅首次创建配置，保留已有文件。
test -f config.toml || cp config.example.toml config.toml
./bin/reverfolio --config ./config.toml --collect-interval 0
```

Windows PowerShell 使用以下对应命令完成编译、配置和启动：

```powershell
go build -o bin/reverfolio.exe ./cmd/reverfolio
if (!(Test-Path config.toml)) { Copy-Item config.example.toml config.toml }
.\bin\reverfolio.exe --config .\config.toml --collect-interval 0
```

浏览器打开配置的监听地址，按 Quick Start 添加测试作品即可。源码 Web UI 无需安装扩展；开发 Windows 原生启动器时另行编译 `./cmd/reverfolio-host`。

### 5.2. 检查与打包

```sh
python3 scripts/check_docs.py
python3 scripts/check_mcp_contract.py --fixture
node scripts/check_extension.mjs
node scripts/check_native_host.mjs
```

MCP 模拟验收使用隔离数据库；真实平台授权与原生安装需要单独验收。CI 覆盖 Go 测试、race/vet 和 Windows、Linux、macOS 构建。

便携包由 `python3 scripts/package_release.py vX.Y.Z` 生成，版本参数替换为准备发布的版本。原生安装器在目标操作系统安装 Qt Installer Framework 后，通过 `python3 scripts/package_installer.py --help` 查看参数，将解压后的干净便携包作为 stage。打包时不得放入运行中的配置、SQLite 或私有访问文件。

## 6. License

原创代码和项目视觉资源采用 [MIT License](LICENSE)，第三方代码与依赖保留其许可证；发行包附带依赖许可原文。用户文章、摄影和平台数据不属于项目代码许可范围。

保护本机账户、配置、`Data/` 和凭据，不要将个人数据提交到仓库或打入安装包。明确通过 MCP 返回的内容可能传给所连接的 AI 服务。安全问题请私下联系 **contact@mail.thinkerqaq.com**，不要在公开 Issue 上传有效凭据。
