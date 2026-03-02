---
name: luoshu.setup
description: 项目初始化引导。自动检测项目特征，通过 2-3 个问题完成配置，并引导配置智能记忆。
argument-hint: "[--reset 重新配置] [--upgrade 检查更新]"
---

# /luoshu.setup -- 项目初始化与升级

根据项目特征自动推荐配置。新项目走完整初始化流程，已配置项目走增量升级流程。

---

## 模式判断

根据项目状态自动选择模式：

| 条件 | 模式 | 行为 |
|------|------|------|
| 无 `.claude/` 目录 | **全新初始化** | 检测项目 → 3 个问题 → 安装模板 → LLM 配置 → 写 meta |
| 有 `.claude/` + 无 `.setup-meta.json` | **注册现有配置** | 扫描已有配置 → 问用户偏好 → LLM 配置 → 写 meta |
| 有 `.setup-meta.json` 或传 `--upgrade` | **增量升级** | 对比已安装 vs 可用 → 展示新功能 → 检查 LLM 配置 |
| 传了 `--reset` | **重新配置** | 备份 → 重走完整流程 |

---

## 模式 A: 全新初始化（无 .claude/）

### Step 1. 检测项目特征（自动，零操作）

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

### Step 2. 快速配置（严格控制在 3 个问题以内）

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

### Step 3. 配置映射与安装

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

### Step 4. 配置智能记忆 AI 引擎（强烈推荐）

模板安装完成后，引导配置 luoshu 智能记忆的 LLM 服务。

#### 4.1 检查当前配置状态

调用 `luoshu_config_get` 检查 LLM 是否已配置：
- **已配置且有效** → 调用 `luoshu_config_validate` 验证连接，成功则静默跳过
- **已配置但失效** → 引导更新 Key
- **未配置** → 进入配置引导

#### 4.2 配置引导话术

```
项目配置已安装完成。

最后一步 -- 配置智能记忆的 AI 引擎。

这个产品的核心能力是跨会话记忆：
- 自动记住你的偏好和上下文
- 下次对话时无缝恢复
- 用自然语言搜索任何之前的内容

这需要连接一个 AI 服务来驱动。
目前支持火山引擎 Doubao（国内延迟低、费用低）。

你有火山引擎的 API Key 吗？
```

#### 4.3 用户有 Key

用户输入 Key 后：
1. 调用 `luoshu_config_set`（key=`llm.api_key`, value=用户输入）
   - 工具会自动做 PreValidateKey 格式预检
   - 如果识别为其他平台的 Key（OpenAI/Anthropic 等），展示具体诊断
2. 调用 `luoshu_config_validate` 验证连接
3. 成功 → 展示配置信息（Key 脱敏）+ 安全声明

```
连接成功。配置信息：
- 服务商：火山引擎
- API Key：sk-****xxxx
- 模型：doubao-1.5-pro-256k（默认）

Key 保存在本地 ~/.luoshu/config.json
不会发送到 Anthropic 或其他外部服务。
```

#### 4.4 用户没有 Key

分步引导获取：

**第 1 步：注册**
```
没问题，我带你一步步获取。大概需要 3 分钟。

火山引擎是字节跳动的云服务平台。

请在浏览器中访问：
https://console.volcengine.com/auth/signup

完成后告诉我，我带你进行下一步。
```

**第 2 步：获取 Key**
```
访问 API Key 管理页面：
https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey

点击"创建 API Key"，复制创建的 Key（以 sk- 开头）。

注意：页面关掉后就看不到完整 Key 了，请确保已经复制。
```

用户粘贴 Key 后，同 4.3 流程验证。

#### 4.5 用户选择跳过

```
跳过 LLM 配置后，以下核心功能将不可用：
- 跨会话智能记忆
- 语义搜索
- 自动提取会话要点

基础配置管理功能不受影响。

[1] 跳过，先用基础功能
[2] 我去获取 Key（约 3 分钟，我教你）
[3] 我已经有 Key 了
```

使用 AskUserQuestion 展示选项。用户选 [1] 则标记跳过，继续到 Step 5。

### Step 5. 写入 setup-meta.json

安装完成后，用 Write 工具写入 `.claude/.setup-meta.json`：

```json
{
  "setup_version": "0.6.0",
  "setup_date": "YYYY-MM-DD",
  "style": "strict|fast|balanced",
  "team": "solo|team",
  "memory": true|false,
  "luoshu_configured": true|false,
  "installed_templates": ["hackathon-core", "cross-session-memory"],
  "installed_rules": ["tpl-hackathon-core.md"],
  "installed_agents": ["architect.md"],
  "installed_commands": ["plan.md", "tdd.md"],
  "installed_skills": ["luoshu.setup", "luoshu.config"]
}
```

### Step 6. 确认与引导

分两层展示结果：

**已启用**（自动生效）：
- 用用户能理解的语言描述，例如"代码质量守护"而不是"code-reviewer agent"
- 根据 luoshu 配置状态区分展示

**luoshu 已配置时：**
```
Setup 完成！项目已配置以下能力：

[已启用] 配置管理、模板系统、智能记忆

试试说"你记得我什么"或"帮我继续上次的工作"。
```

**luoshu 未配置时：**
```
Setup 完成。当前状态：

[已启用] 配置管理、模板系统
[未启用] 智能记忆（需要 LLM 配置）

随时运行 /luoshu.config 来配置。
现在可以开始写代码了。有什么要做的？
```

---

## 模式 B: 注册现有配置（有 .claude/ 但无 meta）

### 1. 扫描现有配置

调用 MCP 工具获取当前状态：
- `template_installed` — 看装了哪些模板
- `extension_list(type="rules")` — 看有哪些规则
- `extension_list(type="agents")` — 看有哪些 agent
- `extension_list(type="commands")` — 看有哪些命令
- `hooks_list` — 看有哪些 hooks

### 2. 呈现当前状态 + 快速补充

只问一个问题（如果能从已有配置推断就不问）。

### 3. 检查 LLM 配置

调用 `luoshu_config_get` 检查状态。如未配置，走 Step 4 的配置引导（同模式 A）。

### 4. 写入 meta 文件

根据扫描结果写入 `.claude/.setup-meta.json`，不修改任何现有配置。

---

## 模式 C: 增量升级（有 meta 或 --upgrade）

### 1. 读取 meta + 对比

读取 `.claude/.setup-meta.json`，对比已安装 vs 可用模板，展示新功能。

### 2. 增量安装

用户选择的新功能，逐个安装。**不修改、不覆盖**已有配置。

### 3. 检查 LLM 配置

如果 meta 中 `luoshu_configured` 为 false 或字段不存在：

```
注意：你还未配置 LLM 服务。
智能记忆是本产品的核心功能，需要 LLM 驱动。

现在配置吗？（约 3 分钟）

[1] 现在配置（推荐）
[2] 稍后再说
```

已配置则调用 `luoshu_config_validate` 静默验证，失效时引导更新。

### 4. 更新 meta 文件

将新安装的内容合并到 meta，更新 `setup_date` 和 `setup_version`。

---

## 模式 D: 重新配置（--reset）

读取现有 meta 了解用户偏好，然后重走模式 A 的完整流程。安装时使用 overwrite=true。

---

## 关键原则

- **检测代替询问**: 能自动检测的不要问用户
- **增量不覆盖**: 升级模式只添加新内容，不修改已有配置
- **默认值代替选择**: 每个选项都有推荐默认值
- **结果导向描述**: 用"代码质量守护"而不是"code-reviewer agent"
- **3 个问题是上限**: 全新初始化 3 个，注册 1 个，升级 0 个
- **LLM 配置强烈推荐但可跳过**: 确保用户在任何情况下都能完成 setup
- **meta 文件是关键**: 每次操作后必须更新 `.claude/.setup-meta.json`
