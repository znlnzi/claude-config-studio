# 代码架构与技术开源准备评估

> CTO 视角 (Werner Vogels 思维模型) | 2026-03-01

## 总体评价

项目的技术架构**整体健康**，代码组织清晰、职责分离合理。作为一个 MCP Server + 桌面应用的混合项目，模块划分已经比较成熟。以下按优先级列出需要关注的问题。

---

## P0：开源前必须修复

### 1. README 中 Provider 信息过时
- **位置**: `README.md:162`, `README.zh-CN.md:162`
- **问题**: 写着 "Currently supported LLM provider: **Volcengine Doubao**"，但 v0.7.1 已实现多 Provider 支持（OpenAI/DeepSeek/Moonshot/Zhipu/SiliconFlow/Volcengine/Custom）
- **影响**: 国际用户看到只支持火山引擎会直接离开
- **修复**: 更新为多 Provider 列表，突出 OpenAI 兼容性

### 2. npm package.json 版本未同步
- **位置**: `npm/package.json:3` (0.7.1) vs `cmd/mcp-server/main.go:76` (0.7.1)
- **问题**: 新功能（多 Provider、Resources、导入导出）已在代码中但未发布 npm
- **修复**: 更新到 0.7.2 并发布

### 3. CHANGELOG.md 未包含最新功能
- **位置**: `CHANGELOG.md`
- **问题**: 最新版本停留在 0.7.1（国际化），缺少 0.7.2 的多 Provider、MCP Resources、导入导出功能记录
- **修复**: 添加 [0.7.2] 版本条目

### 4. serverInstructions 中的工具列表不完整
- **位置**: `cmd/mcp-server/main.go:22-73`
- **问题**: `serverInstructions` 已包含新工具分类，但部分描述可以更详细
- **建议**: 确保所有新工具都有清晰的使用说明

---

## P1：建议开源前修复

### 5. CONTRIBUTING.md 中 Go 版本不一致
- **位置**: `CONTRIBUTING.md:9` 写 "Go 1.21+"，但 `go.mod:3` 要求 `go 1.23.0`
- **修复**: 统一为 Go 1.23+

### 6. 测试覆盖可以更完善
- **现状**: `internal/luoshu/` 有 14 个测试文件，`internal/exporter/` 有 1 个，**但 `cmd/mcp-server/` 没有单元测试**
- **影响**: MCP tool handler 的逻辑无直接测试覆盖
- **建议**: 至少为核心 handler（recall、config、export）添加表驱动测试

### 7. main.go 文件过长
- **位置**: `cmd/mcp-server/main.go` (420+ 行)
- **问题**: 所有 tool definition 和注册都在 main.go 中，但 handler 已经拆分到独立文件
- **建议**: 将 tool definition（`buildXxxTool()` 函数）也移到对应的 `*_tools.go` 文件中，main.go 只保留注册和启动逻辑

### 8. 缺少 GoDoc 包级注释
- **位置**: `internal/luoshu/`, `internal/exporter/`, `internal/evolution/`
- **问题**: 每个 package 缺少 `doc.go` 或包级文档注释
- **影响**: `go doc` 和 pkg.go.dev 上没有包说明
- **建议**: 在每个包中添加简短的包级文档

---

## P2：开源后逐步改善

### 9. 版本号管理可以更自动化
- **问题**: `serverVersion` 在 main.go 中硬编码，npm 版本在 package.json 中，两处需手动同步
- **建议**: 使用 ldflags 从 git tag 注入版本号，或使用统一的版本文件

### 10. CI 可以增强
- **现状**: 只有 lint + build + test
- **建议**:
  - 添加 test coverage 报告（codecov）
  - 添加跨平台测试（macOS/Linux/Windows matrix）
  - 添加 release workflow 自动化

### 11. 考虑 `services/` 包的处理
- **问题**: `services/` 包是 Wails 桌面应用的服务层，与 MCP Server 无关但在同一仓库
- **建议**: 如果桌面应用不开源，考虑 `.gitignore` 或拆分仓库；如果一起开源，确保文档说明清楚

---

## 安全性评估：通过

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 硬编码密钥 | ✅ | 未发现硬编码的 API Key 或密码 |
| 路径遍历防护 | ✅ | exporter、resources 均有 `..` 和绝对路径检查 |
| 输入验证 | ✅ | MCP 工具参数有 Required() 校验 |
| 敏感信息泄露 | ✅ | API Key 显示时有脱敏处理（前3后4） |
| .gitignore | ✅ | `.luoshu/`、`.env`、`.claude/memory/` 均已排除 |
| 测试中的假路径 | ✅ | 只有 `/Users/test/project` 等明显的测试占位路径 |

---

## 依赖审查：干净

| 依赖 | 版本 | 许可证 | 评估 |
|------|------|--------|------|
| mcp-go | v0.44.0 | MIT | ✅ 核心依赖，活跃维护 |
| wails/v2 | v2.11.0 | MIT | ✅ 桌面应用框架（MCP Server 不依赖） |
| 间接依赖 | - | MIT/BSD/Apache | ✅ 全部 MIT 兼容 |

**结论**: 依赖链干净，没有 GPL 等传染性许可证，与 MIT 完全兼容。
