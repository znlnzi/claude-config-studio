---
name: checkpoint
description: 创建工作检查点，跟踪项目状态变化
---

# /checkpoint - 检查点管理

创建工作快照，跟踪开发进度。

## 子命令

### /checkpoint create [name]
创建一个命名检查点，记录当前状态：
- 文件变更列表
- 测试通过率
- 代码覆盖率
- 构建状态

### /checkpoint verify [name]
将当前状态与指定检查点比较：
- 新增/修改的文件
- 测试结果变化
- 覆盖率变化
- 构建状态变化

### /checkpoint list
显示所有已保存的检查点。

## 使用场景

典型开发周期：
1. 开始工作前创建检查点
2. 每个里程碑创建检查点
3. 提交 PR 前对比检查点
4. 回顾变更确保质量

## 检查点文件格式

检查点保存在 .claude/checkpoints/ 目录下。
