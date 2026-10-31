# OpenEcho 浏览器扩展（Edge / Chrome）

独立的 Manifest V3 扩展，提供 OpenEcho 本地服务状态检测、管理页面快捷入口与侧边栏导航；不连接 BlogCTL，也不持有 GitHub/MCP 密钥。

## Windows 安装

1. 确认服务已运行。双击 C:\software\Coding\opencho\Start-OpenEcho.cmd，或打开 http://127.0.0.1:19452/ 验证。
2. 在 Edge 地址栏输入 edge://extensions（Chrome 对应 chrome://extensions）。
3. 开启 **开发人员模式**。
4. 点击 **加载解压缩的扩展**（Load unpacked）。
5. 选择 C:\software\Coding\opencho\extension。这是包含 manifest.json 的文件夹，不是整个 opencho 或 popup 子文件夹。
6. 在浏览器工具栏的扩展菜单找到 OpenEcho，点击固定图标。可查看连接状态、切换中英文并打开工作台或侧边栏。

本地更新扩展后，前往 edge://extensions，在 OpenEcho 卡片中点击 **重新加载**。

## 权限和边界

- 默认检查 Windows 19452、Ubuntu/WSL 19450 和普通本地端口 18440；也可手动选择实例。
- 只读取本地 GET /healthz 来显示服务是否启动。进入工作台时使用已有的本地安全会话。
- 扩展的语言与选择的服务保存在浏览器扩展专用本地存储；**OpenEcho 主程序的设置依然保存在 config.toml，业务数据在 Data/state.sqlite**。
- 仅申请 storage、sidePanel 和三个本地地址访问权限。不申请 Cookie、浏览历史、网页注入或 Native Messaging 权限，也不上传数据。
- 完整的管理、引用审核和统计仍在 OpenEcho 自身 Web 页面内完成。

扩展不会自行启动 OpenEcho；不需要安装 BlogCTL。复制 extension 文件夹仅完成文件部署，浏览器安装仍需要在扩展管理页手动选择该文件夹。

English instructions: README.md
