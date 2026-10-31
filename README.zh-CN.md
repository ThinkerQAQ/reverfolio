# OpenEcho

[English](README.md) · [架构设计](docs/architecture.zh-CN.md) · [实施计划](docs/implementation-plan.md)

本地运行的开源与内容影响力观测工具。关联开源项目、文章中英文版本、多平台分发页面、指标历史，以及第三方引用和使用证据。

**当前状态：规划与仓库初始化。尚无可运行应用或 Release。**

计划采用 Go 单服务＋SQLite＋Web UI＋MCP。平台适配器保存指标快照；AI 客户端调用搜索工具发现线索，通过 MCP 提交证据；统一业务层负责归属、去重、审核和报告。

第一版覆盖作品登记、GitHub 历史指标、历史导入、传播证据审核、Web UI 与 MCP。详细步骤与验收要求见实施计划。

通过功能分支和 PR 开发，仅使用 Squash Merge，合并后删除分支。规则执行情况见[仓库治理](docs/governance.md)。真实账号配置、凭据、个人统计和历史调查资料保存在运行数据目录中，不提交到代码库。

采用 [MIT](LICENSE) 许可证。引入第三方代码时保留适用的版权和许可声明。
