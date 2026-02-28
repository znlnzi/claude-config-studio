# Hooks 系统规则

## Hook 类型

- **PreToolUse**: 工具执行前（验证、参数修改）
- **PostToolUse**: 工具执行后（自动格式化、检查）
- **Stop**: 会话结束时（最终验证）
- **SessionStart**: 会话开始时（加载上下文）
- **SessionEnd**: 会话结束时（持久化状态）
- **PreCompact**: 压缩前（保存状态）

## 自动接受权限

谨慎使用：
- 对信任的、定义明确的计划启用
- 探索性工作时禁用
- 绝不使用 dangerously-skip-permissions 标志
- 使用 allowedTools 配置代替

## TodoWrite 最佳实践

使用 TodoWrite 工具来：
- 跟踪多步骤任务进度
- 验证对指令的理解
- 实现实时引导
- 展示粒度实现步骤

Todo 列表可以暴露：
- 步骤顺序错误
- 缺失的项目
- 多余的项目
- 粒度不当
- 需求误解
