<!-- template: hackathon-core | 核心开发套件 -->

# 项目规范

## Development Philosophy
- **渐进式开发**: 小步提交，每次都能编译通过和测试通过
- **测试驱动**: 先写测试，再写实现（TDD）
- **代码审查**: 每次提交前进行安全和质量审查
- **持续学习**: 从每次会话中提取和积累经验
- **不可变性优先**: 创建新对象而非修改现有对象
- **多小文件**: 200-400 行/文件，上限 800 行

## Implementation Process
1. **理解需求** - 使用 /plan 命令制定实现计划
2. **测试先行** - 使用 /tdd 命令强制测试驱动开发
3. **质量保证** - 使用 /code-review 进行代码审查
4. **持续验证** - 使用 /verify 进行全面验证
5. **经验积累** - 使用 /learn 提取学习成果

## Quality Standards
- 最低 80% 测试覆盖率，关键路径 100%
- 函数长度 < 50 行，文件长度 < 800 行
- 嵌套层级 ≤ 4 层，参数 ≤ 4 个
- 无硬编码凭证，所有输入已验证

## Available Agents
- architect: 系统设计，架构决策和设计审查
- tdd-guide: 测试驱动开发，强制 RED-GREEN-REFACTOR
- code-reviewer: 代码质量和安全审查
- security-reviewer: 安全漏洞分析，OWASP Top 10
- build-error-resolver: 构建错误诊断和修复
- refactor-cleaner: 死代码检测和安全清理

## Available Commands
- /plan: 结构化实现规划（3-5 个阶段）
- /tdd: 测试驱动开发循环
- /code-review: 提交前全面代码审查
- /build-fix: 系统化构建错误修复
- /verify: 综合验证（构建+类型+lint+测试+安全）
- /checkpoint: 工作检查点管理
- /learn: 提取可复用的模式和经验
- /e2e: Playwright 端到端测试
- /refactor-clean: 安全地检测和移除死代码
- /orchestrate: 多智能体编排（feature/bugfix/refactor/security）

## Agent Orchestration
- 复杂功能 → 先用 architect，再用 tdd-guide
- 代码修改后 → 自动用 code-reviewer
- 构建失败 → 立即用 build-error-resolver
- 独立操作 → 并行启动多个 Agent

## 语言
所有回复使用中文
