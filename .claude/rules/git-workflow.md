# Git 工作流规则

## 提交格式

使用 Conventional Commits:
- feat: 新功能
- fix: Bug 修复
- refactor: 代码重构
- test: 添加或修改测试
- docs: 文档更新
- chore: 构建/工具/依赖更新
- perf: 性能优化
- style: 代码格式调整

## 提交规范

- 每个提交只做一件事
- 提交消息简洁明了（50 字符以内）
- 必要时添加详细描述（Body）
- 引用相关 Issue 编号

## 分支策略

- main: 生产就绪代码
- develop: 集成分支
- feature/*: 功能分支
- fix/*: Bug 修复分支
- release/*: 发布准备

## PR 流程

- 禁止直接提交到 main
- 所有 PR 需要代码审查
- 合并前测试必须通过
- PR 描述包含变更说明和测试计划
- 链接相关 Issue

## 分支管理

- 功能分支从 develop 创建
- 完成后合并回 develop
- 定期清理已合并的分支
- 保持分支名称有意义
