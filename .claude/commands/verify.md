---
name: verify
description: 综合代码库验证协议
---

# /verify - 验证协议

执行六个顺序检查的综合验证。

## 验证步骤

### 1. 构建验证
运行构建命令，必须通过才能继续。

### 2. 类型检查
报告所有类型错误及其位置。

### 3. Linting
运行 lint 工具，报告样式和质量问题。

### 4. 测试执行
运行测试套件，报告通过率和覆盖率。

### 5. 调试代码审计
检查源代码中的 console.log、debugger 等调试语句。

### 6. 版本控制状态
显示未提交的更改摘要。

## 执行模式

- **quick**: 仅构建 + 类型检查
- **full**: 全部六个步骤
- **pre-commit**: 构建 + 类型 + lint + 调试审计
- **pre-pr**: 全部步骤 + 安全扫描

## 使用方式
- /verify quick - 快速验证
- /verify full - 完整验证
- /verify pre-commit - 提交前验证
- /verify pre-pr - PR 前验证
