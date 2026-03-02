---
name: setup
description: 项目初始化引导。自动检测项目特征，通过 2-3 个问题完成配置。
argument-hint: "[--reset 重新配置] [--upgrade 检查更新]"
---

# /setup -- 项目初始化与升级

根据项目特征自动推荐配置。新项目走完整初始化流程，已配置项目走增量升级流程。

---

## 模式判断

根据项目状态自动选择模式：

| 条件 | 模式 | 行为 |
|------|------|------|
| 无 `.claude/` 目录 | **全新初始化** | 检测项目 → 3 个问题 → 安装模板 → 写 meta |
| 有 `.claude/` + 无 `.setup-meta.json` | **注册现有配置** | 扫描已有配置 → 问用户偏好 → 写 meta（不改已有内容） |
| 有 `.setup-meta.json` 或传 `--upgrade` | **增量升级** | 对比已安装 vs 可用 → 展示新功能 → 用户选择性添加 |
| 传了 `--reset` | **重新配置** | 备份 → 重走完整流程 |

---

## 模式 A: 全新初始化（无 .claude/）

### 1. 检测项目特征（自动，零操作）

调用 `detect_project` 工具，传入当前项目路径。该工具一次调用即返回完整的项目特征：
- **语言/框架**: 自动识别（TypeScript, Go, Python, Rust 等）
- **测试框架**: 从依赖和配置文件中提取（Jest, Vitest, pytest, go test 等）
- **包管理器**: 通过 lock 文件优先匹配（pnpm, yarn, npm, poetry 等）
- **Git 状态**: 是否初始化、最近作者（判断独立/团队）、当前分支
- **Claude 配置**: `.claude/` 目录和 `.setup-meta.json` 状态（可直接用于模式判断）

返回结果中的 `claude_config` 字段可直接用于模式判断：
- `has_claude_dir == false` → 模式 A（全新初始化）
- `has_claude_dir == true && has_setup_meta == false` → 模式 B（注册现有配置）
- `has_setup_meta == true` → 模式 C（增量升级）

简洁呈现检测结果：

```
我快速扫描了一下：
- 项目类型: Next.js 14 + TypeScript
- 测试框架: Jest + React Testing Library
- 包管理器: pnpm
- Git: 已初始化，最近 3 个 commit 都是同一作者
```

### 2. 快速配置（严格控制在 3 个问题以内）

使用 AskUserQuestion 工具，一次性问 3 个问题：

**问题 1: 开发风格偏好**
- (a) 严谨型 -- TDD、代码审查、严格类型检查（推荐给生产项目）
- (b) 快速型 -- 重速度，轻流程（推荐给原型/实验/hackathon）
- (c) 平衡型 -- 关键路径严谨，其余灵活

**问题 2: 团队规模**
- (a) 独立开发（根据 git log 判断是否只有一个作者，如果是则标注 [检测到]）
- (b) 团队协作

**问题 3: 跨会话记忆**
- (a) 启用（推荐）-- Claude 记住上次聊到哪里、你的偏好、项目知识
- (b) 不需要

### 3. 配置映射与安装

根据选择调用 MCP 工具安装模板：

| 选择组合 | 安装的模板 | 核心能力 |
|---------|-----------|---------|
| 严谨 + 独立 + 记忆 | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | TDD、代码审查、记忆、构建诊断 |
| 严谨 + 团队 + 记忆 | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | 同上 + 架构审查、安全审查 |
| 快速 + 独立 + 无记忆 | `hackathon-core` | 基础验证、构建诊断 |
| 快速 + 独立 + 记忆 | `hackathon-core` + `cross-session-memory` | 基础验证 + 记忆 |
| 平衡 + 任意 + 记忆 | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | 规划、验证、代码审查、记忆 |
| 平衡 + 任意 + 无记忆 | `hackathon-core` + `code-review-checkpoint` | 规划、验证、代码审查 |

安装步骤：
1. 调用 `template_install` 安装每个模板（scope 为当前项目路径）
2. 如果启用了记忆，调用 `save_memory` 创建初始的 `session-state.md` 和 `MEMORY.md`

### 4. 写入 setup-meta.json

安装完成后，用 Write 工具写入 `.claude/.setup-meta.json`：

```json
{
  "setup_version": "0.3.0",
  "setup_date": "YYYY-MM-DD",
  "style": "strict|fast|balanced",
  "team": "solo|team",
  "memory": true|false,
  "installed_templates": ["hackathon-core", "cross-session-memory", ...],
  "installed_rules": ["tpl-hackathon-core.md", ...],
  "installed_agents": ["architect.md", ...],
  "installed_commands": ["plan.md", "tdd.md", ...],
  "installed_skills": ["setup", ...]
}
```

setup_version 通过 `config_get_global` 的 MCP Server 版本获取，或直接使用当天日期。

### 5. 确认与引导

分两层展示结果：

**已启用**（自动生效）：
- 用用户能理解的语言，例如"代码质量守护"而不是"code-reviewer agent"
- 列出 2-3 个

**可用命令**（需要时用）：
- 只列最重要的 3-4 个，每个一句话说明
- 例如：`/plan -- 开始新功能前做规划`

**结尾**回到真实目标：
```
现在可以开始写代码了。有什么要做的？
随时可以说"还有什么功能"或 /setup --upgrade 查看更多。
```

---

## 模式 B: 注册现有配置（有 .claude/ 但无 meta）

这个项目已经有配置了，但不是通过 /setup 安装的（手动配置或旧版本）。

### 1. 扫描现有配置

调用 MCP 工具获取当前状态：
- `template_installed` — 看装了哪些模板
- `extension_list(type="rules")` — 看有哪些规则
- `extension_list(type="agents")` — 看有哪些 agent
- `extension_list(type="commands")` — 看有哪些命令
- `extension_list(type="skills")` — 看有哪些 skill
- `hooks_list` — 看有哪些 hooks

### 2. 呈现当前状态

```
我扫描了你的现有配置：

已安装的模板: hackathon-core, cross-session-memory
已有规则: 15 个
已有 Agent: 6 个
已有命令: 11 个
已有 Hooks: SessionStart, PreCompact, Stop

看起来配置挺完善的。我帮你记录一下，
以后有新功能时可以对比看哪些是新的。
```

### 3. 快速补充

只问一个问题（如果能从已有配置推断就不问）：

**问题: 你的开发风格更偏向哪种？**
- (a) 严谨型 [检测到已安装 TDD + code-review]
- (b) 快速型
- (c) 平衡型

### 4. 写入 meta 文件

根据扫描结果写入 `.claude/.setup-meta.json`，不修改任何现有配置。

---

## 模式 C: 增量升级（有 meta 或 --upgrade）

这是核心升级流程。用户已经配置过，现在想看看有什么新东西。

### 1. 读取 meta + 对比

读取 `.claude/.setup-meta.json` 获取上次安装记录。
调用 `template_list` 获取当前所有可用模板。
对比得出：
- **已安装**: 上次装的（meta 中记录的）
- **新增可用**: 现在有但上次没装的
- **未选择**: 上次就有但用户选择不装的

### 2. 展示更新内容

```
上次配置: 2026-01-15（33 天前）
你的风格: 严谨型 + 独立开发 + 跨会话记忆

自上次以来的新功能:
  [NEW] 上下文功能提示 — Claude 在工作中自然推荐相关命令
  [NEW] 配置健康检查 — 检测规则冗余/冲突/缺失
  [NEW] Hooks 管理 — 直接管理 SessionStart/Stop 等钩子

你之前跳过的:
  [ ] codemap 模板 — 项目代码地图，帮助 Claude 快速理解项目结构
  [ ] continuous-learning — 自动积累经验

要添加哪些？（可多选，或直接回车跳过）
```

### 3. 增量安装

用户选择的新功能，逐个安装：
- 调用 `template_install` 安装新模板
- 调用 `extension_save` 安装新规则/agent

**不修改、不覆盖**已有配置。只做增量添加。

### 4. 更新 meta 文件

将新安装的内容合并到 `.claude/.setup-meta.json`，更新 `setup_date` 和 `setup_version`。

### 5. 确认

```
已添加 2 个新功能:
- 上下文功能提示（自动生效）
- 配置健康检查（运行 evolve_analyze 使用）

你的配置已是最新版本。
```

---

## 模式 D: 重新配置（--reset）

读取现有 `.claude/.setup-meta.json` 了解用户偏好（如果有），然后重走模式 A 的完整流程。
安装时使用 overwrite=true 覆盖已有文件。

---

## 关键原则

- **检测代替询问**: 能自动检测的不要问用户
- **增量不覆盖**: 升级模式只添加新内容，不修改已有配置
- **默认值代替选择**: 每个选项都有推荐默认值
- **结果导向描述**: 用"代码质量守护"而不是"code-reviewer agent"
- **3 个问题是上限**: 全新初始化 3 个，注册 1 个，升级 0 个（多选即可）
- **meta 文件是关键**: 每次操作后必须更新 `.claude/.setup-meta.json`
