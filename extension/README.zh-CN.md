# OpenEcho Edge / Chrome 扩展

扩展现在就是 OpenEcho 工作台，不是跳转器。点击工具栏图标，在浏览器侧边栏直接打开完整影响力管理界面；右上角按钮可以在标签页打开全宽的同一界面。两处复用相同前端代码。

## 安装与更新

1. 在 Windows 启动 OpenEcho：C:\software\Coding\opencho\Start-OpenEcho.cmd，确认 http://127.0.0.1:19452/ 可以打开。
2. 首次安装执行 OpenEcho 安装目录中的 Initialize-OpenEchoExtension.ps1，或直接通过 Start-OpenEcho.cmd 启动（它会自动初始化）。这会在 Data 生成独立随机访问密钥，并同步到扩展 ui/local-auth.js，无需手动填 Token。
3. 打开 Edge 的 edge://extensions（Chrome 使用 chrome://extensions），开启开发人员模式。
4. 点击加载解压缩的扩展，选中 C:\software\Coding\opencho\extension。
5. 固定 OpenEcho 图标，点击后直接进入完整侧边栏，包含总览、项目、渠道、引用审核、发现任务、历史统计、递归设置。
6. 不再显示自动检测、Windows 优先、其他服务之类的选项。

从 v0.1 升级：v0.2 为扩展固定了新的公开身份密钥。请在 edge://extensions 删除旧的解压缩扩展，再重新加载本目录，确保浏览器注册正确的扩展 ID。后续更新可直接重新加载。

## 数据与权限

- 唯一服务地址 http://127.0.0.1:19452，不自动连接 WSL。WSL 的 19450 仍供原有服务和 MCP 使用。
- 扩展与本地 Web 使用同一套交互。业务数据在 Windows 的 Data/state.sqlite，配置保存在 config.toml。
- 扩展只申请 sidePanel 权限和本机 19452 访问权限，不读取浏览器 Cookie，不复制 GitHub/MCP 密钥；专用访问密钥由本地初始化脚本生成，保存在 Data/extension-access.key 和扩展私有文件 ui/local-auth.js，不得上传到 GitHub，不连接其他网站。
- 本地服务验证 OpenEcho 扩展固定身份；其他网页或扩展不能直接访问它的管理 API。MCP 鉴权保持独立。
- 本地服务未运行时展示明确错误，不会切换到其他数据源。

前端源码同步脚本：scripts/build_extension_ui.py；后端身份验证：internal/echo/http.go。
