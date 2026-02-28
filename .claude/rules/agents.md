# Agent 编排规则

## 立即使用 Agent

无需用户提示：
1. 复杂功能请求 → 使用 **architect** Agent
2. 代码刚写完/修改 → 使用 **code-reviewer** Agent
3. Bug 修复或新功能 → 使用 **tdd-guide** Agent
4. 架构决策 → 使用 **architect** Agent

## 并行任务执行

**始终**对独立操作使用并行 Task 执行：

```
# 正确: 并行执行
同时启动 3 个 Agent：
1. Agent 1: 认证模块安全分析
2. Agent 2: 缓存系统性能审查
3. Agent 3: 工具函数类型检查

# 错误: 不必要的顺序执行
先 Agent 1, 然后 Agent 2, 然后 Agent 3
```

## 多视角分析

对复杂问题使用分角色子 Agent：
- 事实审查者
- 高级工程师
- 安全专家
- 一致性审查者
- 冗余检查者
