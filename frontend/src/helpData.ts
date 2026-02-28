export interface PageHelp {
  title: string;
  summary: string;
  tips: string[];
  shortcuts?: { key: string; desc: string }[];
}

export const helpData: Record<string, PageHelp> = {
  dashboard: {
    title: '仪表盘',
    summary: '快速查看 Claude Code 的全局配置状态和项目统计信息。',
    tips: [
      '点击状态卡片可直接跳转到对应的配置页面',
      '「未配置」状态的项建议优先完成配置',
      '快速操作区提供常用功能的入口',
    ],
  },
  'global-claudemd': {
    title: '全局指令文件',
    summary: '编辑 ~/.claude/CLAUDE.md，这是 Claude Code 在所有项目中都会读取的全局指令文件。',
    tips: [
      '全局 CLAUDE.md 会被所有项目继承，适合放通用规范',
      '项目级 CLAUDE.md 会与全局合并，可覆盖全局设置',
      '建议写明你的编码风格、语言偏好和常用技术栈',
      '支持 Markdown 格式，可使用标题分组管理',
    ],
    shortcuts: [
      { key: '⌘S', desc: '保存文件' },
    ],
  },
  'global-settings': {
    title: '全局设置',
    summary: '管理 ~/.claude/settings.json 中的全局配置项。',
    tips: [
      '「表单视图」提供可视化的配置编辑',
      '「JSON 视图」可直接编辑原始 JSON 内容',
      '环境变量会注入到 Claude Code 的运行环境中',
      'Thinking 模式开启后 Claude 会先推理再回答',
      '修改语言设置后所有新对话将使用指定语言',
    ],
    shortcuts: [
      { key: '⌘S', desc: '保存设置' },
    ],
  },
  projects: {
    title: '项目列表',
    summary: '查看和管理所有已配置 Claude Code 的项目。',
    tips: [
      '列表自动扫描 ~/.claude/projects/ 下的已知项目',
      '点击「添加项目」选择本地目录注册新项目',
      '点击项目卡片进入项目详情编辑页面',
      '项目状态显示已配置的文件数量',
    ],
  },
  'project-wizard': {
    title: '新建配置向导',
    summary: '通过三步向导为项目快速创建 Claude Code 配置。',
    tips: [
      '第一步：选择目标项目目录',
      '第二步：从内置模板中选择模板（推荐「黑客马拉松冠军」系列）',
      '第三步：确认配置并一键生成',
      '模板可自动创建 CLAUDE.md、Agents、Commands、Skills、Rules 等全套配置',
    ],
  },
  mcp: {
    title: 'MCP 服务管理',
    summary: '管理 Model Context Protocol 服务器配置（~/.claude/.mcp.json）。',
    tips: [
      'MCP 服务器有两种类型：HTTP 远程服务和 CLI 命令行工具',
      'HTTP 类型需填写 URL，可选配置请求头和超时时间',
      'CLI 类型需填写命令和参数（如 npx 调用的本地工具）',
      '添加新服务器后 Claude Code 会在下次启动时自动连接',
      '常用 MCP：Context7（文档查询）、Playwright（浏览器自动化）',
    ],
  },
  hooks: {
    title: 'Hooks 管理',
    summary: '配置 Claude Code 生命周期钩子（settings.json 中的 hooks 字段）。',
    tips: [
      '支持的事件：PreToolUse / PostToolUse / SessionStart / Stop / UserPromptSubmit',
      '每个事件可配置多个 Hook 条目',
      'Matcher 用于限定触发条件（如只针对某个工具触发）',
      '每个条目可包含多个 shell 命令，按顺序执行',
      'Timeout 默认为空（使用系统默认值），可自定义毫秒数',
    ],
  },
  commands: {
    title: 'Commands 管理',
    summary: '管理 ~/.claude/commands/ 下的自定义斜杠命令。',
    tips: [
      '每个 .md 文件对应一个 /command 斜杠命令',
      '文件名就是命令名（如 review.md → /review）',
      'frontmatter 中用 description 描述命令用途',
      'allowed-tools 限定命令可使用的工具列表',
      '命令体中使用 $ARGUMENTS 引用用户传入的参数',
    ],
    shortcuts: [
      { key: '⌘S', desc: '保存文件' },
    ],
  },
  agents: {
    title: 'Agents 管理',
    summary: '管理 ~/.claude/agents/ 下的 Agent 角色模板。',
    tips: [
      'Agent 是预定义的 AI 助手角色，在 Task 工具中使用',
      'frontmatter 中定义 name、description 和可选的 model',
      '正文部分定义 Agent 的能力、工作流程和行为规范',
      '文件名就是 Agent 的 subagent_type 标识',
    ],
    shortcuts: [
      { key: '⌘S', desc: '保存文件' },
    ],
  },
  skills: {
    title: 'Skills 管理',
    summary: '创建和管理自定义 Skills，浏览已安装 Skills，从在线市场安装新 Skills。',
    tips: [
      '自定义 Skills 支持多文件结构：SKILL.md 为主文件，可添加辅助文件',
      '名称规则：仅小写字母、数字和连字符，不以连字符开头/结尾，不超过 64 字符',
      '名称不能包含 "anthropic" 或 "claude"，建议使用动名词形式（如 format-code）',
      '描述建议使用第三人称，说明何时触发此 Skill，不超过 1024 字符',
      'SKILL.md 建议控制在 500 行以内，复杂指令可拆分到子文件',
      '使用 11 个内置模板快速创建：编码规范、工作流、技术栈、审查清单等',
      'Cmd+S 快捷键可快速保存当前编辑内容',
      '从在线市场可一键安装社区共享的 Skills',
    ],
    shortcuts: [
      { key: '⌘S', desc: '保存文件' },
    ],
  },
  plugins: {
    title: '插件管理',
    summary: '管理已安装的 Claude Code 插件的启用/禁用状态。',
    tips: [
      '开关控制插件是否在 Claude Code 中激活',
      '禁用插件不会卸载，可随时重新启用',
      '插件来源显示在名称下方',
      '如需安装新插件，请在 Claude Code CLI 中操作',
    ],
  },
  templates: {
    title: '模板库',
    summary: '浏览内置的最佳实践模板，一键应用到项目。',
    tips: [
      '模板按类别分组：通用、前端、后端、最佳实践、黑客马拉松冠军',
      '「黑客马拉松冠军」模板来自 Anthropic 黑客马拉松冠军的实战配置',
      '高级模板可同时安装 Agents、Commands、Skills、Rules 等扩展文件',
      '点击模板卡片查看详细内容，展开可预览每个扩展文件的完整内容',
      '「应用到项目」会将模板的所有配置写入选定的项目 .claude/ 目录',
      '已有配置的项目应用模板会覆盖同名文件，建议先备份',
    ],
  },
  'import-export': {
    title: '导入 / 导出',
    summary: '通过 ZIP 压缩包备份和迁移 Claude Code 配置。',
    tips: [
      '全局导出：打包 ~/.claude/ 下的所有配置文件',
      '全局导入：从 ZIP 恢复配置到 ~/.claude/ 目录',
      '项目导出：打包指定项目的 .claude/ 目录和 CLAUDE.md',
      '项目导入：从 ZIP 恢复到指定项目目录',
      '适合在不同机器间迁移配置或定期备份',
    ],
  },
};

// 帮助页面使用的完整指南内容
export interface GuideSection {
  id: string;
  icon: string;
  title: string;
  content: string;
}

export const guideSections: GuideSection[] = [
  {
    id: 'overview',
    icon: '📖',
    title: '总览',
    content: `ClaudeCode Config Studio 是一款 macOS 桌面工具，用于可视化管理 Claude Code 的所有配置。

Claude Code 的配置存储在以下位置：
• ~/.claude/CLAUDE.md — 全局指令文件
• ~/.claude/settings.json — 全局设置（语言、环境变量、Hooks、插件等）
• ~/.claude/.mcp.json — 全局 MCP 服务器配置
• ~/.claude/commands/ — 自定义斜杠命令
• ~/.claude/agents/ — Agent 角色模板
• ~/.claude/plugins/ — 已安装的插件

项目级配置存储在项目目录下：
• <project>/.claude/CLAUDE.md — 项目指令文件
• <project>/.claude/settings.json — 项目设置
• <project>/.claude/.mcp.json — 项目 MCP 配置`,
  },
  {
    id: 'claudemd',
    icon: '📝',
    title: 'CLAUDE.md 指令文件',
    content: `CLAUDE.md 是 Claude Code 最重要的配置文件，用于定义 AI 的行为准则。

层级关系：
1. ~/.claude/CLAUDE.md — 全局指令，所有项目共享
2. <project>/CLAUDE.md — 项目根目录的指令
3. <project>/.claude/CLAUDE.md — 项目 .claude 目录的指令

合并规则：Claude Code 会自动合并所有层级的内容，项目级优先。

推荐内容结构：
• 编码规范（命名风格、代码格式）
• 语言偏好（如「用中文回复」）
• 技术栈约束（框架版本、依赖限制）
• 工作流程要求（测试、提交规范）
• 禁止事项（不要修改的文件/目录）`,
  },
  {
    id: 'settings',
    icon: '⚙️',
    title: '全局设置',
    content: `settings.json 控制 Claude Code 的行为参数。

主要配置项：
• language — 设定 Claude 回复的语言
• alwaysThinkingEnabled — 启用思考模式，让 Claude 先推理再回答
• env — 环境变量，注入到 Claude Code 运行时
• hooks — 生命周期钩子（详见 Hooks 章节）
• enabledPlugins — 已启用的插件列表
• statusLine — 状态栏配置

环境变量常见用途：
• HTTP_PROXY/HTTPS_PROXY — 网络代理
• CLAUDE_CODE_MAX_OUTPUT_TOKENS — 最大输出 token 数`,
  },
  {
    id: 'mcp',
    icon: '🔌',
    title: 'MCP 服务',
    content: `MCP（Model Context Protocol）让 Claude Code 能连接外部工具和数据源。

两种类型：
1. HTTP 远程服务 — 提供 URL 地址，Claude Code 通过网络调用
2. CLI 命令行工具 — 通过 npx/node 等命令启动本地进程

配置文件：~/.claude/.mcp.json（全局）或 <project>/.claude/.mcp.json（项目级）

常用 MCP 服务：
• Context7 — 查询最新的库和框架文档
• Playwright MCP — 浏览器自动化和测试
• GitHub MCP — GitHub API 集成
• Filesystem MCP — 增强文件系统操作`,
  },
  {
    id: 'hooks',
    icon: '🪝',
    title: 'Hooks 钩子',
    content: `Hooks 可在 Claude Code 的关键时刻自动执行 shell 命令。

支持的事件类型：
• PreToolUse — 工具调用前触发
• PostToolUse — 工具调用后触发
• SessionStart — 会话启动时触发
• Stop — 会话结束时触发
• UserPromptSubmit — 用户提交提示时触发

配置结构：
每个事件 → 多个条目 → 每个条目包含 matcher（可选）和 hooks 数组

Matcher 说明：
• 留空表示匹配所有
• PreToolUse/PostToolUse 支持工具名匹配（如 "Bash"、"Write"）

使用场景举例：
• SessionStart: 自动检查 Git 状态
• PostToolUse(Bash): 执行完命令后自动格式化
• Stop: 会话结束时发送通知`,
  },
  {
    id: 'commands',
    icon: '⌨️',
    title: 'Commands 自定义命令',
    content: `Commands 是用户自定义的斜杠命令，存储在 ~/.claude/commands/ 目录。

文件格式（Markdown + frontmatter）：
---
description: 命令的简短描述
allowed-tools: [Read, Glob, Grep, Bash]
---

# 命令说明
$ARGUMENTS 可获取用户传入的参数。

使用方式：文件名为 review.md，则在 Claude Code 中输入 /review 即可调用。

子目录命名：commands/project/init.md → /project:init`,
  },
  {
    id: 'agents',
    icon: '🤖',
    title: 'Agents 角色模板',
    content: `Agents 定义可复用的 AI 助手角色，存储在 ~/.claude/agents/ 目录。

文件格式（Markdown + frontmatter）：
---
name: agent-name
description: Agent 角色描述
model: sonnet（可选，默认继承父级）
---

角色定义正文...

使用方式：在 Task 工具中通过 subagent_type 参数指定 Agent 名称。

最佳实践：
• 为不同场景创建专用 Agent（如代码审查、测试编写、文档生成）
• 在角色定义中明确能力边界和工作流程
• 使用 model 字段为简单任务指定轻量模型（如 haiku）`,
  },
  {
    id: 'skills-plugins',
    icon: '🎯',
    title: 'Skills 与插件',
    content: `Skills 是自定义指令集，当 Claude 遇到匹配场景时会自动使用。

自定义 Skills 目录结构：
• ~/.claude/skills/<skill-name>/SKILL.md — 主文件（必需，包含 frontmatter）
• ~/.claude/skills/<skill-name>/其他文件 — 辅助文件（可选，子文件引用）
• 项目级：<project>/.claude/skills/<skill-name>/

SKILL.md 格式：
---
name: skill-name
description: 第三人称描述，说明何时触发
---
（Markdown 正文内容）

验证规则：
• name: 仅 [a-z0-9-]，不以连字符开头或结尾，不超过 64 字符，不含 "anthropic"/"claude"
• description: 不超过 1024 字符，不含 XML 标签

最佳实践：
• 命名使用动名词形式（如 format-code、review-pr）
• 描述使用第三人称，明确说明触发时机
• 主文件建议控制在 500 行以内
• 复杂指令采用渐进式披露：主文件引用子文件
• 使用反馈循环模式确保质量：实施 → 验证 → 修复

内置模板（11 个）：
• 编码规范、工作流程、技术栈指南、代码审查清单
• 工作流清单、渐进式披露、示例模式
• 条件工作流、反馈循环、MCP 工具集成
• 空白模板

插件 Skills：
• 在「已安装 Skills」页面浏览插件提供的 Skill 内容
• 在「在线市场」搜索和安装社区共享的 Skills`,
  },
  {
    id: 'templates',
    icon: '📋',
    title: '模板库',
    content: `内置模板提供开箱即用的最佳实践配置。

模板分类：
• 通用 — 极简配置、标准配置
• 前端 — React、Vue 项目配置
• 后端 — Go、Python FastAPI、Node Express
• 最佳实践 — 跨会话记忆、持续学习、代码审查检查点、CodeMap
• 🏆 黑客马拉松冠军 — 来自 Anthropic 黑客马拉松冠军的完整配置集

黑客马拉松冠军模板：
基于 everything-claude-code 项目，经过 10+ 个月实战检验的生产级配置。
• 核心开发套件 — 6 Agents + 10 Commands + 6 Skills + 4 Rules，推荐大多数项目使用
• 安全与质量 — 专注安全审查和代码质量保证
• 完整配置 — 包含所有组件 + Hooks，适合高级用户

模板可包含的配置类型：
• CLAUDE.md — 项目指令文件
• Settings — settings.json 配置（含 Hooks）
• MCP — .mcp.json 服务器配置
• Agents — .claude/agents/ 下的 Agent 角色模板
• Commands — .claude/commands/ 下的自定义命令
• Skills — .claude/skills/ 下的技能定义
• Rules — .claude/rules/ 下的规则文件

应用方式：
1. 在模板库中选择模板，点击查看预览
2. 展开各分类预览 Agents/Commands/Skills/Rules 的完整内容
3. 点击「应用到项目」选择目标目录
4. 模板会自动创建所有配置文件到 .claude/ 目录

注意：应用模板会覆盖项目中已有的同名配置文件，建议先导出备份。`,
  },
  {
    id: 'import-export',
    icon: '💾',
    title: '导入 / 导出',
    content: `通过 ZIP 压缩包备份和迁移配置。

全局配置导出包含：
• CLAUDE.md、settings.json、cclsp.json、.mcp.json
• commands/ 目录所有文件
• agents/ 目录所有文件

项目配置导出包含：
• 项目根目录的 CLAUDE.md
• .claude/ 目录全部内容

使用场景：
• 在新电脑上快速恢复所有配置
• 团队间共享统一的配置模板
• 定期备份防止配置丢失`,
  },
  {
    id: 'shortcuts',
    icon: '⌨️',
    title: '快捷键',
    content: `全局快捷键：
• ⌘S — 保存当前编辑内容（在编辑器页面生效）

编辑器功能：
• 支持 Monaco Editor 的全部快捷键
• ⌘Z / ⌘⇧Z — 撤销 / 重做
• ⌘F — 搜索
• ⌘H — 替换
• ⌘D — 选择下一个匹配
• ⌥↑ / ⌥↓ — 上下移动行`,
  },
];
