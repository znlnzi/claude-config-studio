---
name: orchestrate
description: 多智能体编排，为复杂任务启用顺序工作流
---

# /orchestrate - 智能体编排

为复杂开发任务启用顺序多智能体工作流。

## 内置工作流

### Feature - 完整功能开发
```
规划 → TDD → 代码审查 → 安全审查
```

### Bugfix - Bug 修复流程
```
问题分析 → 复现测试 → 修复 → 回归测试
```

### Refactor - 安全重构
```
架构审查 → 测试加固 → 重构 → 验证
```

### Security - 安全审查
```
漏洞扫描 → 风险评估 → 修复 → 验证
```

## 使用方式

- /orchestrate feature "添加用户认证"
- /orchestrate bugfix "修复登录失败"
- /orchestrate refactor "重构数据访问层"
- /orchestrate security "全面安全审计"

## 自定义工作流

/orchestrate custom "architect,tdd-guide,code-reviewer" "任务描述"

## 交接格式

每个阶段生成结构化交接文档：
- 上下文
- 发现和修改
- 开放问题
- 建议

## 最终输出

编排报告总结：
- 所有智能体的贡献
- 修改的文件列表
- 测试结果
- 安全发现
- 最终建议: SHIP / NEEDS WORK / BLOCKED
