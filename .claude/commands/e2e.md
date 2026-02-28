---
name: e2e
description: 生成和运行端到端测试
---

# /e2e - 端到端测试

使用 Playwright 生成和运行 E2E 测试。

## 工作流

### 1. 分析目标
- 识别要测试的用户流程
- 确定关键交互点
- 列出预期结果

### 2. 生成测试
- 使用 Playwright API
- 使用语义选择器优先（role, label, text）
- 添加 data-testid 作为后备

### 3. 运行测试
- 执行测试套件
- 捕获失败截图
- 记录测试结果

## 测试结构

```typescript
import { test, expect } from '@playwright/test';

test.describe('功能名称', () => {
  test('用户场景', async ({ page }) => {
    // 导航
    await page.goto('/path');

    // 交互
    await page.getByRole('button', { name: '提交' }).click();

    // 断言
    await expect(page.getByText('成功')).toBeVisible();
  });
});
```

## 最佳实践
- 测试真实用户流程，不测试实现细节
- 使用 Page Object 模式组织代码
- 每个测试独立运行，不依赖其他测试
- 使用 fixtures 管理测试数据
