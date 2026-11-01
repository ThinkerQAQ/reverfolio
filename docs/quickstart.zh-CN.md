# Reverfolio 快速入门（v0.9.x）

Reverfolio 已公开开源，采用 **MIT 协议、本地优先、独立部署**。不依赖 Inkerio、SyndiRelay 或特定博客引擎；不连接任何平台也能先登记作品。SQLite 与 API Key 都留在用户自己的电脑，发行包不带开发者的个人数据。

## 下载与安装

从 [GitHub 最新 Release](https://github.com/ThinkerQAQ/reverfolio/releases/latest) 选择系统对应的安装包，按照同一页面的 `SHA256SUMS` 校验文件。

| 系统 | 对应文件 | 安装方式 |
|---|---|---|
| Windows x64 | `Reverfolio-v*-windows-amd64.zip` | 解压后运行安装脚本和启动脚本 |
| Windows ARM64 | `Reverfolio-v*-windows-arm64.zip` | 同上 |
| Linux x64 / ARM64 | `Reverfolio-v*-linux-{amd64,arm64}.tar.gz` | 解压、复制 TOML、运行 Go 程序 |
| macOS Intel / Apple Silicon | `Reverfolio-v*-darwin-{amd64,arm64}.tar.gz` | 同上；未签名程序可能需要在系统设置中单独批准 |

### Windows

将压缩包解压到固定目录，例如 `C:\software\Coding\reverfolio`。在解压出的 `Reverfolio` 目录打开 PowerShell，执行：

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\Install-Reverfolio.ps1
.\Start-Reverfolio.cmd
```

浏览器打开 <http://127.0.0.1:19452>。首次安装会生成本地 TOML、`Data/`、扩展鉴权文件；重复安装**不覆盖原有 SQLite、TOML 或 Token**。

**可选浏览器扩展：**打开 Edge 的 `edge://extensions` 或 Chrome 的 `chrome://extensions`，启用开发者模式 →「加载已解压的扩展」→ 选择本地 `extension/` 目录。扩展使用同一工作台；Windows 安装脚本会注册 Native Messaging，实现停止服务时的静默启动。详情见[扩展安装](../extension/README.zh-CN.md)。

### Linux / macOS

以 Linux amd64 为例，切换到刚刚下载发行包的目录，运行：

```bash
tar -xzf Reverfolio-v*-linux-amd64.tar.gz
cd Reverfolio
cp -n config.example.toml config.toml
./bin/reverfolio -config ./config.toml
```

macOS 把压缩包文件名替换为 `darwin-amd64` 或 `darwin-arm64`；Linux ARM64 替换为 `linux-arm64`。打开 <http://127.0.0.1:18440>。这是前台运行方式；要持续后台运行需自行配置 systemd 或 launchd。**Windows 扩展自动唤醒功能不等于 macOS/Linux 已提供一键原生宿主安装**，不要把这些平台的支持混淆。

### 开发者从源码启动

```bash
git clone https://github.com/ThinkerQAQ/reverfolio.git
cd reverfolio
go run ./cmd/reverfolio -db data/reverfolio.sqlite
```

需要 Go 1.26 或以上。详见[贡献指南](../CONTRIBUTING.md)。

## 第一次使用

1. **登记作品：**在「项目管理」新增作品，或「网站发现」输入公开博客网址，预览 RSS、Atom、Sitemap，再勾选属于自己的文章。不需要博客分发工具。
2. **GitHub：**登记仓库。私有流量统计需要用户自己授权的 `GITHUB_TOKEN`，或明确选择本机 `gh` CLI 登录；Release 下载次数不等于真实使用人数。
3. **DEV.to：**在「平台账号」粘贴自己生成的个人 API Key。先发现并确认已发布文章归属，再采集粉丝、浏览、点赞和评论；保留真实来源与采集时间。
4. **网站指标：**只有拥有对应站点的权限时，才配置 Umami、Google Search Console 或 Bing；没有数据就显示不可用，不伪造零值。
5. **ChatGPT / MCP：**本地通过 `-mcp-stdio` 使用；远程 ChatGPT 需要**用户自己的安全 MCP 连接／隧道**。开发者私有插件不是公开一键安装包。参见 [MCP 安装](mcp-setup.zh-CN.md)。

当前只实现了 DEV.to 官方 API 的平台账号统计；**没有**全平台自动采集、后台网页抓取、粉丝真人识别或自动证明下载者就是用户的功能。未核实合法官方统计 API 的平台不宣称支持。

## 更新和回滚

升级前备份 `Data/state.sqlite`、TOML 和私有凭据；更换公共程序和扩展文件后，重新加载已解压扩展。旧程序可能不认识新 SQLite schema，回滚时必须同步恢复相容的旧程序和数据库备份。不要把包含自己密钥或统计数据的 `Data/` 提交到公开 GitHub。

许可证全文汇总在发行包的 `THIRD_PARTY_LICENSES.txt`；另见[安全说明](../SECURITY.md)和[最新发行说明](https://github.com/ThinkerQAQ/reverfolio/releases/latest)。
