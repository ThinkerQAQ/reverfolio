# Reverfolio Edge / Chrome 扩展

扩展现在就是 Reverfolio 工作台，不是跳转器。点击工具栏图标，在浏览器侧边栏直接打开完整影响力管理界面；右上角按钮可以在标签页打开全宽的同一界面。两处复用相同前端代码。

## 安装、自动启动与手动恢复

1. 将 Windows 发行包放在固定目录（例如 `C:\software\Coding\reverfolio`），首次使用执行一次 `Install-Reverfolio.ps1`。安装器保留已有 `config.toml` 与 `Data/`，生成扩展独立访问密钥，并为 Edge、Chrome 注册 `com.reverfolio.bridge` Native Messaging 启动器。
2. 在 `edge://extensions`（Chrome 为 `chrome://extensions`）开启开发者模式，使用「加载解压缩的扩展」选择安装目录下的 `extension/`。固定 Reverfolio 图标。
3. 点击图标后在侧边栏显示完整工作台。**启动时先检测本机服务，未启动则通过 Native Messaging 静默拉起，无需保留黑色控制台窗口。** 如果启动失败，侧边栏显示「启动本地服务」和「重试连接」。
4. 手动启动可直接运行根目录的 `Start-Reverfolio-Hidden.vbs`，没有常驻命令行黑框；重新运行 `Install-Reverfolio.ps1` 可修复 Native Messaging 注册。扩展版本更新通常只需在扩展页点击「重新加载」，保留固定扩展 ID。

浏览器 Native Host 是独立的 `bin/reverfolio-host.exe`（Windows GUI 子系统），只接受服务状态与启动指令，仅检查本机 `127.0.0.1:19452` 并静默拉起原有服务。Native Messaging 不读取或返回 SQLite、API Token、GitHub 令牌与浏览器 Cookie。

安装器还会将旧版「直接执行 `reverfolio.exe`」的 Windows 登录计划任务改为 `wscript.exe` 静默启动，**只更新动作，原触发器、用户身份和数据保持不变**。

## 数据、权限与边界

- 唯一业务服务仍是 Windows 原生 `127.0.0.1:19452`；自动启动不连接 WSL，不复制数据库。
- 扩展仅申请 `sidePanel`、`nativeMessaging` 以及本地 `19452` 的访问权限；Native Host 清单只允许 Reverfolio 的固定扩展身份调用。
- 已运行的服务可直接连接，无需每次唤起 Native Host；未注册启动器的旧安装可以继续连接现有服务。
- 扩展独立访问密钥保存在 `Data/extension-access.key`、`extension/ui/local-auth.js`，与 MCP/API 令牌分离，不随代码发布，也不经 Native Messaging 发送。
- 如果启动失败，会显示连接错误与手动恢复入口，不自动使用其他业务数据源。非 Windows 平台仍按原本方式手动启动服务，Windows 安装脚本不会注册其他平台的浏览器 Native Host。

前端源码同步脚本：scripts/build_extension_ui.py；后端身份验证：internal/echo/http.go。
