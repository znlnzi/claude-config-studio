import { useEffect, useState, useCallback, useRef } from 'react';
import Editor from '@monaco-editor/react';
import {
  GetAllSkills,
  GetSkillContent,
  ListUserSkills,
  GetUserSkill,
  SaveUserSkill,
  DeleteUserSkill,
  GetMCPServerNames,
  ListSkillFiles,
  ReadSkillFile,
  SaveSkillFile,
  DeleteSkillFile,
  CreateSkillFile,
} from '../../wailsjs/go/services/SkillService';
import {
  SearchOnlineExtensions,
  InstallOnlineExtension,
} from '../../wailsjs/go/services/ExtensionService';
import { SelectDirectory } from '../../wailsjs/go/services/ConfigService';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import HelpTip from '../components/HelpTip';

interface SkillInfo {
  name: string;
  description: string;
  source: string;
  marketplace: string;
  pluginName: string;
  type: string;
  filePath: string;
}

interface UserSkillInfo {
  name: string;
  description: string;
  scope: string;
  dirName: string;
  isFlat: boolean;
}

interface SkillFileInfoItem {
  relativePath: string;
  isDir: boolean;
  size: number;
  isMain: boolean;
}

interface OnlineExtension {
  name: string;
  description: string;
  category: string;
  source: string;
  repoUrl: string;
  downloadUrl: string;
  extType: string;
}

interface ValidationMessage {
  type: 'error' | 'warning';
  text: string;
}

type TabMode = 'custom' | 'installed' | 'market';
type SourceFilter = 'all' | 'builtin' | 'github';

interface SkillTemplate {
  id: string;
  icon: string;
  label: string;
  desc: string;
  name: string;
  description: string;
  content: string;
}

const SKILL_TEMPLATES: SkillTemplate[] = [
  {
    id: 'coding-standards',
    icon: '📏',
    label: '编码规范',
    desc: '定义项目的编码风格和规范',
    name: 'coding-standards',
    description: '当编写或审查代码时，确保遵循项目编码规范。',
    content: `# 编码规范

## 命名规范
- 变量和函数使用 camelCase
- 类和接口使用 PascalCase
- 常量使用 UPPER_SNAKE_CASE
- 文件名使用 kebab-case

## 代码风格
- 使用 2 空格缩进
- 每行不超过 100 个字符
- 函数体不超过 50 行

## 注释规范
- 公共 API 必须有文档注释
- 复杂逻辑需要行内注释说明"为什么"
- 不要注释"是什么"，代码本身应该足够清晰

## 错误处理
- 不要忽略错误，至少记录日志
- 使用自定义错误类型提供上下文
- 在系统边界处验证输入
`,
  },
  {
    id: 'workflow',
    icon: '🔄',
    label: '工作流程',
    desc: '定义固定的工作步骤和流程',
    name: 'my-workflow',
    description: '当需要执行特定工作流程时使用此 Skill。',
    content: `# 工作流程

当用户请求执行此工作流时，按以下步骤进行：

## 步骤 1: 分析需求
- 理解用户的具体需求
- 确认关键约束和边界条件
- 如有不明确之处，先提问确认

## 步骤 2: 制定方案
- 列出可行的方案选项
- 分析每个方案的优缺点
- 推荐最佳方案并说明理由

## 步骤 3: 执行实施
- 按照确认的方案逐步实施
- 每个关键步骤完成后进行验证
- 遇到问题及时调整方案

## 步骤 4: 验证结果
- 检查最终结果是否符合需求
- 运行相关测试确保质量
- 总结完成情况和注意事项
`,
  },
  {
    id: 'tech-stack',
    icon: '🛠️',
    label: '技术栈指南',
    desc: '定义项目使用的技术和最佳实践',
    name: 'tech-stack-guide',
    description: '当进行技术选型或编写代码时，参考项目技术栈指南。',
    content: `# 技术栈指南

## 前端
- **框架**: React 18 + TypeScript
- **状态管理**: Zustand
- **样式**: Tailwind CSS
- **构建工具**: Vite

## 后端
- **语言**: Go / Python / Node.js（按需修改）
- **数据库**: PostgreSQL
- **缓存**: Redis
- **API 风格**: RESTful

## 开发工具
- **包管理器**: pnpm / uv
- **代码格式化**: Prettier + ESLint
- **测试框架**: Vitest / pytest

## 最佳实践
- 优先使用项目已有的库，不要随意引入新依赖
- 遵循现有的目录结构和文件组织方式
- 新功能必须包含单元测试
- API 变更需要更新文档
`,
  },
  {
    id: 'review-checklist',
    icon: '✅',
    label: '代码审查清单',
    desc: '代码审查时的检查要点',
    name: 'code-review-checklist',
    description: '当审查代码或提交 PR 前，使用此清单进行自查。',
    content: `# 代码审查清单

在提交代码或审查 PR 时，请逐项检查以下要点：

## 功能正确性
- [ ] 代码实现了需求描述的功能
- [ ] 边界条件已处理（空值、超长输入、并发等）
- [ ] 错误情况有合理的处理和提示

## 代码质量
- [ ] 没有重复代码，复用了现有工具函数
- [ ] 变量和函数命名清晰、有意义
- [ ] 函数职责单一，不超过 50 行
- [ ] 没有硬编码的魔法数字或字符串

## 安全性
- [ ] 用户输入已验证和清理
- [ ] 没有 SQL 注入、XSS 等安全漏洞
- [ ] 敏感信息没有写入日志或代码

## 测试
- [ ] 新增功能有对应的单元测试
- [ ] 修改的代码没有破坏现有测试
- [ ] 关键路径有集成测试覆盖

## 性能
- [ ] 没有 N+1 查询问题
- [ ] 大列表使用了分页或虚拟滚动
- [ ] 没有不必要的重复计算或渲染
`,
  },
  {
    id: 'workflow-checklist',
    icon: '📋',
    label: '工作流清单',
    desc: '带进度清单和反馈循环',
    name: 'my-workflow-checklist',
    description: '当执行需要按步骤完成的任务时，使用清单跟踪进度并在每步完成后提供反馈。',
    content: `# 工作流清单

按以下步骤执行任务，每步完成后标记状态并报告进度。

## 清单

- [ ] **步骤 1: 需求确认**
  - 理解用户需求
  - 确认约束条件
  - 输出: 需求摘要

- [ ] **步骤 2: 方案设计**
  - 分析可行方案
  - 选择最优方案
  - 输出: 方案文档

- [ ] **步骤 3: 实施**
  - 按方案逐步实施
  - 每个子步骤完成后自检
  - 输出: 实施结果

- [ ] **步骤 4: 验证**
  - 运行测试
  - 检查边界条件
  - 输出: 验证报告

## 反馈循环

每完成一个步骤后：
1. 报告当前进度（已完成 X/4）
2. 列出已完成的关键成果
3. 说明下一步计划
4. 如遇到阻碍，立即说明并提出替代方案
`,
  },
  {
    id: 'progressive-disclosure',
    icon: '📂',
    label: '渐进式披露',
    desc: '多文件结构，主文件引用子文件',
    name: 'my-progressive-skill',
    description: '当处理复杂任务需要分阶段、分模块展开详细指令时使用。',
    content: `# 渐进式披露 Skill

本 Skill 采用多文件组织，根据场景按需加载详细指令。

## 概述
此 Skill 将复杂指令分解为多个子文件，避免一次性加载过多内容。

## 子文件引用

根据当前任务阶段，参考以下子文件：

- **初始化阶段** → 参考 \`setup.md\`
- **核心开发阶段** → 参考 \`development.md\`
- **测试阶段** → 参考 \`testing.md\`
- **发布阶段** → 参考 \`release.md\`

## 使用说明

1. 先阅读本文件了解整体流程
2. 根据当前阶段读取对应子文件
3. 子文件中包含该阶段的详细步骤和检查清单
4. 完成一个阶段后再进入下一个阶段
`,
  },
  {
    id: 'example-pattern',
    icon: '💡',
    label: '示例模式',
    desc: '输入/输出对示例引导',
    name: 'my-example-pattern',
    description: '当需要 Claude 按照特定的输入输出格式处理任务时使用。',
    content: `# 示例模式 Skill

通过输入/输出示例对来指导 Claude 的行为。

## 格式规范

处理用户请求时，按以下示例格式输出。

## 示例 1

**输入：**
\`\`\`
用户描述一个功能需求
\`\`\`

**输出：**
\`\`\`
### 需求分析
- 核心功能点
- 涉及的模块

### 实施方案
1. 步骤一
2. 步骤二

### 预估影响
- 修改文件列表
- 潜在风险
\`\`\`

## 示例 2

**输入：**
\`\`\`
用户报告一个 bug
\`\`\`

**输出：**
\`\`\`
### 问题定位
- 错误现象
- 根因分析

### 修复方案
1. 修改点
2. 验证方法

### 回归测试
- 需要验证的场景
\`\`\`

## 使用说明
根据以上示例模式，对所有同类请求保持一致的输出格式和结构。
`,
  },
  {
    id: 'conditional-workflow',
    icon: '🔀',
    label: '条件工作流',
    desc: '根据条件走不同分支',
    name: 'my-conditional-workflow',
    description: '当任务需要根据不同条件执行不同路径时使用。',
    content: `# 条件工作流

根据输入条件判断执行路径。

## 决策树

### 判断 1: 任务类型
- **如果是新功能** → 执行路径 A
- **如果是 Bug 修复** → 执行路径 B
- **如果是重构** → 执行路径 C

---

## 路径 A: 新功能

1. 检查是否有相关的设计文档
2. 确认测试策略
3. 创建功能分支
4. 实现功能 → 编写测试 → 代码审查

## 路径 B: Bug 修复

1. 复现问题
2. 定位根因
3. 编写回归测试（先写测试，确认失败）
4. 修复代码（确认测试通过）

## 路径 C: 重构

1. 确认重构范围和目标
2. 确保现有测试覆盖足够
3. 小步修改，每步都能编译通过
4. 运行全部测试确保无回归

## 通用收尾

无论哪条路径，最终都需要：
- [ ] 代码编译通过
- [ ] 所有测试通过
- [ ] 变更总结
`,
  },
  {
    id: 'feedback-loop',
    icon: '🔄',
    label: '反馈循环',
    desc: '验证-修复-重复迭代',
    name: 'my-feedback-loop',
    description: '当任务需要迭代验证和修正直到满足质量标准时使用。',
    content: `# 反馈循环 Skill

执行"实施 → 验证 → 修复"的迭代循环，直到满足所有质量标准。

## 质量标准

以下所有条件都必须满足才能结束循环：
- [ ] 代码编译成功，无错误
- [ ] 所有测试通过
- [ ] 无 linter 警告
- [ ] 满足用户需求描述

## 迭代流程

### 第 1 轮: 实施
1. 根据需求编写代码
2. 运行编译检查
3. 运行测试

### 验证检查点
检查上述质量标准：
- 全部通过 → **完成**，输出总结
- 有失败项 → 记录失败原因，进入修复轮

### 修复轮 (最多 3 轮)
1. 分析失败原因
2. 修复问题
3. 重新运行验证检查点

### 超过 3 轮仍未通过
1. 停止迭代
2. 列出所有未解决的问题
3. 分析根因
4. 提出替代方案供用户决定
`,
  },
  {
    id: 'mcp-integration',
    icon: '🔌',
    label: 'MCP 工具集成',
    desc: 'MCP 工具调用模板',
    name: 'my-mcp-integration',
    description: '当任务需要调用 MCP 服务器提供的外部工具时使用。',
    content: `# MCP 工具集成 Skill

指导如何使用项目中配置的 MCP 工具完成任务。

## 可用 MCP 工具

根据项目 .mcp.json 配置，可使用以下工具：
- \`mcp__<server-name>__<tool-name>\` — 调用指定 MCP 服务器的工具

## 工具使用流程

### 1. 确认工具可用性
在使用 MCP 工具前，先确认：
- 工具名称和参数格式正确
- 所需的 MCP 服务器已在配置中启用

### 2. 调用工具
使用标准格式调用工具：
\`\`\`
mcp__server-name__tool-name(参数)
\`\`\`

### 3. 处理结果
- 检查工具返回结果
- 如果失败，检查参数是否正确
- 将结果整合到当前任务中

## 常见场景

### 查询文档
使用文档查询 MCP 获取最新的库文档和 API 参考。

### 浏览器操作
使用浏览器 MCP 进行页面截图、元素交互等操作。

### 文件操作
使用文件系统 MCP 进行批量文件操作。

## 注意事项
- MCP 工具调用可能需要用户授权
- 注意工具的速率限制
- 妥善处理工具调用失败的情况
`,
  },
  {
    id: 'blank',
    icon: '📄',
    label: '空白模板',
    desc: '从零开始编写自定义 Skill',
    name: '',
    description: '',
    content: `# Skill 标题

在这里编写 Skill 的详细指令。Claude 会在匹配到 description 描述的场景时自动使用此 Skill。

## 使用场景
描述何时应该使用此 Skill。

## 具体指令
1. 第一步操作
2. 第二步操作
3. 第三步操作
`,
  },
];

export default function SkillsManager() {
  // Tab
  const [tab, setTab] = useState<TabMode>('custom');

  // === 自定义 Skills 状态 ===
  const [scope, setScope] = useState<string>('global');
  const [projectPath, setProjectPath] = useState('');
  const [userSkills, setUserSkills] = useState<UserSkillInfo[]>([]);
  const [userLoading, setUserLoading] = useState(true);
  const [selectedUserSkill, setSelectedUserSkill] = useState<UserSkillInfo | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editDirName, setEditDirName] = useState('');
  const [originalContent, setOriginalContent] = useState('');
  const [originalName, setOriginalName] = useState('');
  const [originalDesc, setOriginalDesc] = useState('');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [showTemplateChooser, setShowTemplateChooser] = useState(false);
  // 引用菜单
  const [showRefMenu, setShowRefMenu] = useState(false);
  const [mcpServers, setMcpServers] = useState<string[]>([]);
  const [installedSkillNames, setInstalledSkillNames] = useState<string[]>([]);
  const [installedAgentNames, setInstalledAgentNames] = useState<string[]>([]);
  const [nameErrors, setNameErrors] = useState<ValidationMessage[]>([]);
  const [descErrors, setDescErrors] = useState<ValidationMessage[]>([]);
  // 多文件管理
  const [skillFiles, setSkillFiles] = useState<SkillFileInfoItem[]>([]);
  const [activeFile, setActiveFile] = useState<string>('SKILL.md');
  const [fileContents, setFileContents] = useState<Record<string, string>>({});
  const [newFileName, setNewFileName] = useState('');
  const [showNewFileInput, setShowNewFileInput] = useState(false);
  const [isFlat, setIsFlat] = useState(false);
  const editorRef = useRef<any>(null);
  const refMenuRef = useRef<HTMLDivElement>(null);

  // === 已安装 Skills 状态 ===
  const [skills, setSkills] = useState<SkillInfo[]>([]);
  const [installedLoading, setInstalledLoading] = useState(true);
  const [selectedInstalled, setSelectedInstalled] = useState<SkillInfo | null>(null);
  const [installedContent, setInstalledContent] = useState('');
  const [loadingContent, setLoadingContent] = useState(false);
  const [filter, setFilter] = useState('');

  // === 在线市场状态 ===
  const [sourceFilter, setSourceFilter] = useState<SourceFilter>('all');
  const [marketQuery, setMarketQuery] = useState('');
  const [marketResults, setMarketResults] = useState<OnlineExtension[]>([]);
  const [marketLoading, setMarketLoading] = useState(false);
  const [marketError, setMarketError] = useState('');
  const [marketSearched, setMarketSearched] = useState(false);
  const [installingName, setInstallingName] = useState<string | null>(null);
  const [marketVisibleCount, setMarketVisibleCount] = useState(24);
  const [marketInstallScope, setMarketInstallScope] = useState<string>('global');
  const [marketProjectPath, setMarketProjectPath] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ========== 验证逻辑 ==========

  const validateName = (value: string) => {
    const msgs: ValidationMessage[] = [];
    if (value.length === 0) {
      msgs.push({ type: 'error', text: '名称不能为空' });
    } else {
      if (value.length > 64) {
        msgs.push({ type: 'error', text: `名称长度不能超过 64 个字符（当前 ${value.length}）` });
      }
      const namePattern = /^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$/;
      if (!namePattern.test(value)) {
        msgs.push({ type: 'error', text: '只允许小写字母、数字和连字符，不能以连字符开头或结尾' });
      }
      if (/<[^>]+>/.test(value)) {
        msgs.push({ type: 'error', text: '不能包含 XML 标签' });
      }
      const lower = value.toLowerCase();
      if (lower.includes('anthropic') || lower.includes('claude')) {
        msgs.push({ type: 'error', text: '不能包含 "anthropic" 或 "claude"' });
      }
    }
    setNameErrors(msgs);
  };

  const validateDescription = (value: string) => {
    const msgs: ValidationMessage[] = [];
    if (value.length === 0) {
      msgs.push({ type: 'error', text: '描述不能为空' });
    } else {
      if (value.length > 1024) {
        msgs.push({ type: 'error', text: `描述长度不能超过 1024 个字符（当前 ${value.length}）` });
      }
      if (/<[^>]+>/.test(value)) {
        msgs.push({ type: 'error', text: '不能包含 XML 标签' });
      }
    }
    setDescErrors(msgs);
  };

  const hasValidationErrors = nameErrors.some(m => m.type === 'error') || descErrors.some(m => m.type === 'error');

  // 根据文件扩展名推断 Monaco 语言
  const getLanguageByExt = (filename: string): string => {
    const ext = filename.split('.').pop()?.toLowerCase() || '';
    const map: Record<string, string> = {
      md: 'markdown', py: 'python', js: 'javascript', ts: 'typescript',
      json: 'json', yaml: 'yaml', yml: 'yaml', sh: 'shell',
      go: 'go', rs: 'rust', rb: 'ruby', java: 'java',
      css: 'css', html: 'html', xml: 'xml', sql: 'sql',
    };
    return map[ext] || 'plaintext';
  };

  // 加载 Skill 目录下的文件列表
  const loadSkillFiles = async (s: string, dir: string) => {
    try {
      const files: any = await ListSkillFiles(s, dir);
      setSkillFiles((files || []).filter((f: SkillFileInfoItem) => !f.isDir));
    } catch {
      setSkillFiles([]);
    }
  };

  // 切换到某个文件（仅子目录格式使用）
  const switchToFile = async (relativePath: string) => {
    // 缓存当前编辑内容（非主文件需要缓存到 fileContents）
    const currentIsMain = relativePath === 'SKILL.md' ? false : activeFile === 'SKILL.md';
    if (!currentIsMain) {
      setFileContents(prev => ({ ...prev, [activeFile]: editContent }));
    }

    setActiveFile(relativePath);

    if (relativePath === 'SKILL.md') {
      // 恢复 SKILL.md 内容
      const effectiveScope = scope === 'global' ? 'global' : projectPath;
      try {
        const detail: any = await GetUserSkill(effectiveScope, editDirName);
        setEditContent(detail.content || '');
      } catch {}
    } else {
      // 如果已有缓存，使用缓存
      if (fileContents[relativePath] !== undefined) {
        setEditContent(fileContents[relativePath]);
      } else {
        const effectiveScope = scope === 'global' ? 'global' : projectPath;
        try {
          const content = await ReadSkillFile(effectiveScope, editDirName, relativePath);
          setEditContent(content);
          setFileContents(prev => ({ ...prev, [relativePath]: content }));
        } catch {
          setEditContent('');
        }
      }
    }
  };

  // 创建新文件
  const handleCreateFile = async () => {
    if (!newFileName.trim()) return;
    const effectiveScope = scope === 'global' ? 'global' : projectPath;
    try {
      await CreateSkillFile(effectiveScope, editDirName, newFileName.trim());
      setShowNewFileInput(false);
      setNewFileName('');
      await loadSkillFiles(effectiveScope, editDirName);
      switchToFile(newFileName.trim());
    } catch (err) {
      console.error('创建文件失败:', err);
    }
  };

  // 删除文件
  const handleDeleteFile = async (relativePath: string) => {
    if (relativePath === 'SKILL.md') return;
    const effectiveScope = scope === 'global' ? 'global' : projectPath;
    try {
      await DeleteSkillFile(effectiveScope, editDirName, relativePath);
      setFileContents(prev => {
        const next = { ...prev };
        delete next[relativePath];
        return next;
      });
      if (activeFile === relativePath) {
        switchToFile('SKILL.md');
      }
      await loadSkillFiles(effectiveScope, editDirName);
    } catch (err) {
      console.error('删除文件失败:', err);
    }
  };

  // ========== 自定义 Skills 逻辑 ==========

  const loadUserSkills = useCallback(async () => {
    setUserLoading(true);
    try {
      const effectiveScope = scope === 'global' ? 'global' : projectPath;
      if (scope !== 'global' && !projectPath) {
        setUserSkills([]);
        setUserLoading(false);
        return;
      }
      const data: any = await ListUserSkills(effectiveScope);
      setUserSkills(data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setUserLoading(false);
    }
  }, [scope, projectPath]);

  const loadRefData = useCallback(async () => {
    try {
      const [mcpNames, allSkills]: any = await Promise.all([
        GetMCPServerNames(),
        GetAllSkills(),
      ]);
      setMcpServers(mcpNames || []);
      const allList = allSkills || [];
      setInstalledSkillNames(
        [...new Set<string>(allList
          .filter((s: SkillInfo) => s.type === 'skill')
          .map((s: SkillInfo) => s.name))]
      );
      setInstalledAgentNames(
        [...new Set<string>(allList
          .filter((s: SkillInfo) => s.type === 'agent')
          .map((s: SkillInfo) => s.name))]
      );
    } catch (err) {
      console.error(err);
    }
  }, []);

  useEffect(() => {
    loadUserSkills();
  }, [loadUserSkills]);

  useEffect(() => {
    loadRefData();
  }, [loadRefData]);

  const handleSelectUserSkill = async (skill: UserSkillInfo) => {
    setSelectedUserSkill(skill);
    setIsCreating(false);
    setIsFlat(skill.isFlat);
    setActiveFile(skill.isFlat ? skill.dirName + '.md' : 'SKILL.md');
    setFileContents({});
    setShowNewFileInput(false);
    setNewFileName('');
    try {
      const effectiveScope = scope === 'global' ? 'global' : projectPath;
      const detail: any = await GetUserSkill(effectiveScope, skill.dirName);
      setEditName(detail.name || '');
      setEditDesc(detail.description || '');
      setEditContent(detail.content || '');
      setEditDirName(skill.dirName);
      setOriginalName(detail.name || '');
      setOriginalDesc(detail.description || '');
      setOriginalContent(detail.content || '');
      // 仅子目录格式才有多文件
      if (!skill.isFlat) {
        loadSkillFiles(effectiveScope, skill.dirName);
      } else {
        setSkillFiles([]);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreate = () => {
    setIsCreating(true);
    setShowTemplateChooser(true);
    setSelectedUserSkill(null);
    setEditName('');
    setEditDesc('');
    setEditContent('');
    setEditDirName('');
    setOriginalName('');
    setOriginalDesc('');
    setOriginalContent('');
    setActiveFile('SKILL.md');
    setSkillFiles([]);
    setFileContents({});
    setShowNewFileInput(false);
    setNameErrors([]);
    setDescErrors([]);
    setIsFlat(false);
  };

  const handleSelectTemplate = (template: SkillTemplate) => {
    setEditName(template.name);
    setEditDesc(template.description);
    setEditContent(template.content);
    setShowTemplateChooser(false);
  };

  // 判断当前是否是主文件（SKILL.md 或扁平的 name.md）
  const isMainFile = isFlat ? activeFile === editDirName + '.md' : activeFile === 'SKILL.md';

  const handleSave = async () => {
    setSaving(true);
    try {
      const effectiveScope = scope === 'global' ? 'global' : projectPath;

      if (isMainFile) {
        // 保存主文件（含 frontmatter）
        if (!editName.trim()) { setSaving(false); return; }
        const dirName = editDirName || editName.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
        const flatFlag = isFlat ? 'true' : 'false';
        await SaveUserSkill(effectiveScope, dirName, editName.trim(), editDesc.trim(), editContent, flatFlag);
        setOriginalName(editName.trim());
        setOriginalDesc(editDesc.trim());
        setOriginalContent(editContent);
        if (isCreating) {
          setIsCreating(false);
          setEditDirName(dirName);
          if (!isFlat) {
            loadSkillFiles(effectiveScope, dirName);
          }
        }
        loadUserSkills();
      } else {
        // 保存附加文件
        await SaveSkillFile(effectiveScope, editDirName, activeFile, editContent);
        setFileContents(prev => ({ ...prev, [activeFile]: editContent }));
      }

      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error('保存失败:', err);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!editDirName) return;
    try {
      const effectiveScope = scope === 'global' ? 'global' : projectPath;
      await DeleteUserSkill(effectiveScope, editDirName, isFlat ? 'true' : 'false');
      setSelectedUserSkill(null);
      setIsCreating(false);
      setEditName('');
      setEditDesc('');
      setEditContent('');
      setEditDirName('');
      loadUserSkills();
    } catch (err) {
      console.error('删除失败:', err);
    }
  };

  const handleScopeChange = (newScope: string) => {
    if (newScope === 'project' && !projectPath) {
      handleSelectProject();
      return;
    }
    setScope(newScope);
    setSelectedUserSkill(null);
    setIsCreating(false);
  };

  const handleSelectProject = async () => {
    try {
      const dir = await SelectDirectory();
      if (dir) {
        setProjectPath(dir);
        setScope('project');
        setSelectedUserSkill(null);
        setIsCreating(false);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const hasCustomChanges =
    editName !== originalName ||
    editDesc !== originalDesc ||
    editContent !== originalContent;

  // 插入引用
  const insertReference = (type: 'mcp' | 'skill' | 'agent' | 'allowed-tools', name: string) => {
    const editor = editorRef.current;
    let text = '';
    if (type === 'mcp') {
      text = `使用 \`mcp__${name}__<tool_name>\` 工具`;
    } else if (type === 'agent') {
      text = `委派给 ${name} agent 处理`;
    } else if (type === 'allowed-tools') {
      text = `mcp__${name}__`;
    } else {
      text = `使用 /${name} Skill`;
    }
    if (editor) {
      const position = editor.getPosition();
      if (position) {
        editor.executeEdits('insert-ref', [{
          range: {
            startLineNumber: position.lineNumber,
            startColumn: position.column,
            endLineNumber: position.lineNumber,
            endColumn: position.column,
          },
          text,
        }]);
        editor.focus();
      }
    } else {
      setEditContent(prev => prev + text);
    }
    setShowRefMenu(false);
  };

  // 点击外部关闭引用菜单
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (refMenuRef.current && !refMenuRef.current.contains(e.target as Node)) {
        setShowRefMenu(false);
      }
    };
    if (showRefMenu) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [showRefMenu]);

  // 快捷键 Cmd+S
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') {
        e.preventDefault();
        if (tab === 'custom' && (hasCustomChanges || isCreating)) handleSave();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [tab, hasCustomChanges, isCreating, editName, editDesc, editContent]);

  // ========== 已安装 Skills 逻辑 ==========

  const loadInstalledSkills = async () => {
    setInstalledLoading(true);
    try {
      const data: any = await GetAllSkills();
      setSkills(data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setInstalledLoading(false);
    }
  };

  const handleSelectInstalled = async (skill: SkillInfo) => {
    setSelectedInstalled(skill);
    setLoadingContent(true);
    try {
      const text: any = await GetSkillContent(skill.filePath);
      setInstalledContent(text || '');
    } catch (err) {
      console.error(err);
      setInstalledContent('无法读取文件内容');
    } finally {
      setLoadingContent(false);
    }
  };

  // 过滤出仅 Skill 类型（排除 Agent），按来源分组
  const installedSkillsOnly = skills.filter(s => s.type === 'skill');

  const grouped = installedSkillsOnly.reduce<Record<string, SkillInfo[]>>((acc, skill) => {
    const key = skill.source;
    if (!acc[key]) acc[key] = [];
    acc[key].push(skill);
    return acc;
  }, {});

  const filteredGrouped = Object.entries(grouped).reduce<Record<string, SkillInfo[]>>(
    (acc, [source, items]) => {
      const filtered = items.filter(
        s =>
          s.name.toLowerCase().includes(filter.toLowerCase()) ||
          s.description.toLowerCase().includes(filter.toLowerCase())
      );
      if (filtered.length > 0) acc[source] = filtered;
      return acc;
    },
    {}
  );

  const totalInstalled = installedSkillsOnly.length;

  // ========== 在线市场逻辑 ==========

  const doSearch = useCallback((source: string, query: string) => {
    setMarketLoading(true);
    setMarketResults([]);
    setMarketError('');
    setMarketVisibleCount(24);
    SearchOnlineExtensions('skills', source, query)
      .then((data: any) => {
        setMarketResults(data?.extensions || []);
        setMarketSearched(true);
      })
      .catch((err: any) => setMarketError(String(err)))
      .finally(() => setMarketLoading(false));
  }, []);

  const handleMarketSearch = (value: string) => {
    setMarketQuery(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => doSearch(sourceFilter === 'all' ? '' : sourceFilter, value), 300);
  };

  const handleSourceFilterChange = (sf: SourceFilter) => {
    setSourceFilter(sf);
    doSearch(sf === 'all' ? '' : sf, marketQuery);
  };

  const handleTabChange = (t: TabMode) => {
    setTab(t);
    if (t === 'installed' && skills.length === 0) {
      loadInstalledSkills();
    }
    if (t === 'market' && !marketSearched) {
      doSearch('', '');
    }
  };

  useEffect(() => {
    if (tab === 'installed') loadInstalledSkills();
  }, []);

  const handleInstall = async (ext: OnlineExtension, installScope: string = 'global') => {
    setInstallingName(ext.name);
    try {
      await InstallOnlineExtension('skills', ext as any, installScope);
      loadInstalledSkills();
    } catch (err) {
      console.error(err);
    } finally {
      setInstallingName(null);
    }
  };

  const isMarketInstalled = (name: string) => {
    const normalizedName = name.toLowerCase().replace(/[^a-z0-9]/g, '');
    return skills.some(s => {
      const sn = s.name.toLowerCase().replace(/[^a-z0-9]/g, '');
      return sn === normalizedName;
    });
  };

  // ========== 渲染 ==========

  return (
    <div className="page-container page-full">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">Skills 管理</h1>
          <p className="page-subtitle">
            {tab === 'custom'
              ? '创建和管理自定义 Skills'
              : tab === 'installed'
              ? `已安装 ${totalInstalled} 个 Skill（来自插件）`
              : '浏览在线 Skill 市场'}
          </p>
        </div>
        <div className="page-header-right">
          <HelpTip pageId="skills" />
          {tab === 'custom' && saved && <span className="save-indicator">已保存</span>}
          {tab === 'custom' && hasCustomChanges && <span className="unsaved-indicator">未保存</span>}
          {tab === 'custom' && (
            <button className="btn btn-primary" onClick={handleCreate}>
              新建
            </button>
          )}
          {tab === 'installed' && (
            <button className="btn btn-secondary" onClick={loadInstalledSkills}>刷新</button>
          )}
        </div>
      </div>

      <div className="mcp-tabs">
        <button
          className={`mcp-tab ${tab === 'custom' ? 'active' : ''}`}
          onClick={() => handleTabChange('custom')}
        >
          自定义 Skills
          {userSkills.length > 0 && <span className="mcp-tab-count">{userSkills.length}</span>}
        </button>
        <button
          className={`mcp-tab ${tab === 'installed' ? 'active' : ''}`}
          onClick={() => handleTabChange('installed')}
        >
          已安装 Skills
          {totalInstalled > 0 && <span className="mcp-tab-count">{totalInstalled}</span>}
        </button>
        <button
          className={`mcp-tab ${tab === 'market' ? 'active' : ''}`}
          onClick={() => handleTabChange('market')}
        >
          在线市场
        </button>
      </div>

      {/* ========== Tab 1: 自定义 Skills ========== */}
      {tab === 'custom' && (
        <>
          {/* 范围选择器 */}
          <div className="skill-scope-bar">
            <span className="skill-scope-label">范围：</span>
            <div className="marketplace-sources">
              <button
                className={`marketplace-source-btn ${scope === 'global' ? 'active' : ''}`}
                onClick={() => handleScopeChange('global')}
              >
                全局
              </button>
              <button
                className={`marketplace-source-btn ${scope === 'project' ? 'active' : ''}`}
                onClick={() => handleScopeChange('project')}
              >
                项目级
              </button>
            </div>
            {scope === 'project' && (
              <button className="btn btn-ghost skill-scope-path" onClick={handleSelectProject}>
                {projectPath || '选择项目目录'}
              </button>
            )}
          </div>

          {userLoading ? (
            <div className="loading-state">加载中...</div>
          ) : (
            <div className="ext-layout">
              {/* 左侧列表 */}
              <div className="ext-sidebar">
                <div className="ext-file-list">
                  {userSkills.length === 0 && !isCreating && (
                    <div className="ext-empty">
                      <span>🎯</span>
                      <span>暂无自定义 Skill</span>
                      <span style={{ fontSize: 11, color: 'var(--macos-text-tertiary)' }}>点击右上角"新建"创建</span>
                    </div>
                  )}
                  {userSkills.map(skill => (
                    <button
                      key={skill.dirName}
                      className={`ext-file-item ${selectedUserSkill?.dirName === skill.dirName && !isCreating ? 'active' : ''}`}
                      onClick={() => handleSelectUserSkill(skill)}
                    >
                      <div className="ext-file-name">
                        <span style={{ marginRight: 6 }}>🎯</span>
                        {skill.name || skill.dirName}
                      </div>
                    </button>
                  ))}
                </div>
              </div>

              {/* 右侧编辑区 */}
              <div className="ext-editor">
                {isCreating && showTemplateChooser ? (
                  <div className="skill-template-chooser">
                    <div className="skill-template-header">
                      <h3>选择模板开始创建</h3>
                      <p>选择一个模板快速创建 Skill，或从空白开始</p>
                    </div>
                    <div className="skill-template-grid">
                      {SKILL_TEMPLATES.map(t => (
                        <button
                          key={t.id}
                          className="skill-template-card"
                          onClick={() => handleSelectTemplate(t)}
                        >
                          <span className="skill-template-icon">{t.icon}</span>
                          <div className="skill-template-label">{t.label}</div>
                          <div className="skill-template-desc">{t.desc}</div>
                        </button>
                      ))}
                    </div>
                    <div className="skill-template-hint">
                      <strong>Skill 是什么？</strong> Skill 是一组指令，当 Claude 遇到匹配的场景时会自动使用。
                      例如：定义编码规范、固定工作流程、技术栈指南等。文件保存为 <code>SKILL.md</code>。
                    </div>
                  </div>
                ) : (selectedUserSkill || isCreating) ? (
                  <div className="skill-form-container">
                    {/* 表单头部 — 仅在 SKILL.md 激活时显示 */}
                    {isMainFile && (
                      <div className="skill-form-header">
                        <div className="skill-form-fields">
                          <div className="skill-form-group">
                            <div className="skill-form-row">
                              <label className="skill-form-label">名称</label>
                              <input
                                className={`form-input${nameErrors.some(m => m.type === 'error') ? ' has-error' : nameErrors.some(m => m.type === 'warning') ? ' has-warning' : ''}`}
                                value={editName}
                                onChange={e => { setEditName(e.target.value); validateName(e.target.value); }}
                                placeholder="如: my-coding-rules"
                                autoFocus={isCreating}
                              />
                            </div>
                            <div className="skill-field-footer">
                              <div>
                                {nameErrors.map((m, i) => (
                                  <div key={i} className={`skill-validation-msg ${m.type}`}>{m.text}</div>
                                ))}
                                {editName.length === 0 && nameErrors.length === 0 && (
                                  <div className="skill-validation-hint">建议使用动名词形式，如 format-code、review-pr</div>
                                )}
                              </div>
                              <span className={`skill-char-count${editName.length > 64 ? ' over' : ''}`}>
                                {editName.length}/64
                              </span>
                            </div>
                          </div>
                          <div className="skill-form-group">
                            <div className="skill-form-row">
                              <label className="skill-form-label">描述</label>
                              <input
                                className={`form-input${descErrors.some(m => m.type === 'error') ? ' has-error' : descErrors.some(m => m.type === 'warning') ? ' has-warning' : ''}`}
                                value={editDesc}
                                onChange={e => { setEditDesc(e.target.value); validateDescription(e.target.value); }}
                                placeholder="描述何时应该使用此 Skill（第三人称，含触发时机）"
                              />
                            </div>
                            <div className="skill-field-footer">
                              <div>
                                {descErrors.map((m, i) => (
                                  <div key={i} className={`skill-validation-msg ${m.type}`}>{m.text}</div>
                                ))}
                                {editDesc.length > 0 && descErrors.length === 0 && (
                                  <div className="skill-validation-hint">建议使用第三人称描述，包含触发时机</div>
                                )}
                              </div>
                              <span className={`skill-char-count${editDesc.length > 1024 ? ' over' : ''}`}>
                                {editDesc.length}/1024
                              </span>
                            </div>
                          </div>
                        </div>
                        <div className="skill-form-actions">
                          {!isCreating && editDirName && (
                            <button className="btn btn-danger" onClick={handleDelete}>删除</button>
                          )}
                          <button
                            className="btn btn-primary"
                            onClick={handleSave}
                            disabled={(isMainFile && !editName.trim()) || saving || (isMainFile && hasValidationErrors)}
                          >
                            {saving ? '保存中...' : isCreating ? '创建' : '保存'}
                          </button>
                        </div>
                      </div>
                    )}

                    {/* 文件 Tab 栏 — 仅子目录格式、非创建模式且已有 dirName 时显示 */}
                    {!isCreating && editDirName && !isFlat && (
                      <div className="skill-file-tabs">
                        <button
                          className={`skill-file-tab${activeFile === 'SKILL.md' ? ' active' : ''}`}
                          onClick={() => switchToFile('SKILL.md')}
                        >
                          SKILL.md
                        </button>
                        {skillFiles.filter(f => f.relativePath !== 'SKILL.md').map(f => (
                          <button
                            key={f.relativePath}
                            className={`skill-file-tab${activeFile === f.relativePath ? ' active' : ''}`}
                            onClick={() => switchToFile(f.relativePath)}
                          >
                            {f.relativePath}
                            <span
                              className="skill-file-tab-close"
                              onClick={e => { e.stopPropagation(); handleDeleteFile(f.relativePath); }}
                            >
                              ×
                            </span>
                          </button>
                        ))}
                        {showNewFileInput ? (
                          <div className="skill-new-file-bar">
                            <input
                              className="form-input"
                              value={newFileName}
                              onChange={e => setNewFileName(e.target.value)}
                              onKeyDown={e => { if (e.key === 'Enter') handleCreateFile(); if (e.key === 'Escape') { setShowNewFileInput(false); setNewFileName(''); } }}
                              placeholder="文件名如 examples.md"
                              autoFocus
                              style={{ width: 140, padding: '2px 6px', fontSize: 12 }}
                            />
                            <button className="btn btn-ghost btn-sm" onClick={handleCreateFile}>确定</button>
                          </div>
                        ) : (
                          <button
                            className="skill-file-tab-add"
                            onClick={() => setShowNewFileInput(true)}
                            title="新建文件"
                          >
                            +
                          </button>
                        )}
                        {/* 非 SKILL.md 时也需要保存按钮 */}
                        {!isMainFile && (
                          <div style={{ marginLeft: 'auto', display: 'flex', gap: 8, alignItems: 'center' }}>
                            <button
                              className="btn btn-primary btn-sm"
                              onClick={handleSave}
                              disabled={saving}
                            >
                              {saving ? '保存中...' : '保存'}
                            </button>
                          </div>
                        )}
                      </div>
                    )}

                    {/* 内容编辑器 */}
                    <div className="skill-editor-bar">
                      <span className="skill-editor-label">{isMainFile ? '内容' : activeFile}</span>
                      <div className="skill-ref-wrapper" ref={refMenuRef}>
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => setShowRefMenu(!showRefMenu)}
                        >
                          + 插入引用
                        </button>
                        {showRefMenu && (
                          <div className="skill-ref-menu">
                            {mcpServers.length > 0 && (
                              <>
                                <div className="skill-ref-group-title">MCP 工具引用</div>
                                {mcpServers.map(name => (
                                  <button
                                    key={`mcp-${name}`}
                                    className="skill-ref-item"
                                    onClick={() => insertReference('mcp', name)}
                                  >
                                    <span className="skill-ref-icon">🔌</span>
                                    <span className="skill-ref-text">
                                      <span>{name}</span>
                                      <span className="skill-ref-hint">插入 mcp__{name}__&lt;tool&gt; 格式</span>
                                    </span>
                                  </button>
                                ))}
                                <div className="skill-ref-group-title">MCP allowed-tools</div>
                                {mcpServers.map(name => (
                                  <button
                                    key={`at-${name}`}
                                    className="skill-ref-item"
                                    onClick={() => insertReference('allowed-tools', name)}
                                  >
                                    <span className="skill-ref-icon">🔐</span>
                                    <span className="skill-ref-text">
                                      <span>mcp__{name}__</span>
                                      <span className="skill-ref-hint">插入 allowed-tools 前缀</span>
                                    </span>
                                  </button>
                                ))}
                              </>
                            )}
                            {installedAgentNames.length > 0 && (
                              <>
                                <div className="skill-ref-group-title">Agents</div>
                                {installedAgentNames.map(name => (
                                  <button
                                    key={`agent-${name}`}
                                    className="skill-ref-item"
                                    onClick={() => insertReference('agent', name)}
                                  >
                                    <span className="skill-ref-icon">🤖</span>
                                    {name}
                                  </button>
                                ))}
                              </>
                            )}
                            {installedSkillNames.length > 0 && (
                              <>
                                <div className="skill-ref-group-title">Skills</div>
                                {installedSkillNames.map(name => (
                                  <button
                                    key={`skill-${name}`}
                                    className="skill-ref-item"
                                    onClick={() => insertReference('skill', name)}
                                  >
                                    <span className="skill-ref-icon">🎯</span>
                                    /{name}
                                  </button>
                                ))}
                              </>
                            )}
                            {mcpServers.length === 0 && installedSkillNames.length === 0 && installedAgentNames.length === 0 && (
                              <div className="skill-ref-empty">暂无可引用的资源</div>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="ext-monaco">
                      <Editor
                        height="100%"
                        language={getLanguageByExt(activeFile)}
                        value={editContent}
                        onChange={v => setEditContent(v ?? '')}
                        onMount={editor => { editorRef.current = editor; }}
                        theme="vs"
                        options={{
                          fontSize: 13,
                          lineHeight: 20,
                          fontFamily: "'SF Mono', 'Monaco', 'Menlo', monospace",
                          minimap: { enabled: false },
                          wordWrap: 'on',
                          scrollBeyondLastLine: false,
                          padding: { top: 12, bottom: 12 },
                          automaticLayout: true,
                          tabSize: 2,
                          smoothScrolling: true,
                        }}
                      />
                    </div>
                    {/* 编辑器状态栏 */}
                    <div className="skill-editor-statusbar">
                      <span className={`skill-line-count${editContent.split('\n').length > 500 ? ' warning' : ''}`}>
                        {editContent.split('\n').length} 行
                        {editContent.split('\n').length > 500 && ' (建议控制在 500 行以内)'}
                      </span>
                      {!isMainFile && (
                        <span className="skill-file-path">{activeFile}</span>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="ext-empty-editor">
                    <span style={{ fontSize: 36 }}>🎯</span>
                    <p>选择一个 Skill 编辑，或点击"新建"创建</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </>
      )}

      {/* ========== Tab 2: 已安装 Skills ========== */}
      {tab === 'installed' && (
        installedLoading ? (
          <div className="loading-state">加载中...</div>
        ) : totalInstalled === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">🎯</div>
            <h3>暂无已安装 Skills</h3>
            <p>Skills 来自已安装的插件，或从在线市场安装</p>
          </div>
        ) : (
          <div className="skills-layout">
            {/* 左侧列表 */}
            <div className="skills-sidebar">
              <div className="skills-search">
                <input
                  className="form-input"
                  placeholder="搜索 Skill..."
                  value={filter}
                  onChange={e => setFilter(e.target.value)}
                />
              </div>
              <div className="skills-list">
                {Object.entries(filteredGrouped).map(([source, items]) => (
                  <div key={source} className="skills-group">
                    <div className="skills-group-title">{source}</div>
                    {items.map(skill => (
                      <button
                        key={skill.filePath}
                        className={`skills-item ${selectedInstalled?.filePath === skill.filePath ? 'active' : ''}`}
                        onClick={() => handleSelectInstalled(skill)}
                      >
                        <span className="skills-item-icon">
                          {skill.type === 'agent' ? '🤖' : '🎯'}
                        </span>
                        <div className="skills-item-info">
                          <div className="skills-item-name">{skill.name}</div>
                          <div className="skills-item-type">
                            {skill.type === 'agent' ? 'Agent' : 'Skill'}
                          </div>
                        </div>
                      </button>
                    ))}
                  </div>
                ))}
              </div>
            </div>

            {/* 右侧详情 */}
            <div className="skills-detail">
              {selectedInstalled ? (
                <div className="skills-detail-content">
                  <div className="skills-detail-header">
                    <div className="skills-detail-meta">
                      <h2 className="skills-detail-name">
                        <span>{selectedInstalled.type === 'agent' ? '🤖' : '🎯'}</span>
                        {selectedInstalled.name}
                      </h2>
                      <div className="skills-detail-badges">
                        <span className="badge">{selectedInstalled.type === 'agent' ? 'Agent' : 'Skill'}</span>
                        <span className="badge">{selectedInstalled.pluginName}</span>
                        <span className="badge">{selectedInstalled.marketplace}</span>
                      </div>
                      {selectedInstalled.description && (
                        <p className="skills-detail-desc">{selectedInstalled.description}</p>
                      )}
                    </div>
                  </div>
                  <div className="skills-detail-body">
                    {loadingContent ? (
                      <div className="loading-state">加载中...</div>
                    ) : (
                      <pre className="skills-code">{installedContent}</pre>
                    )}
                  </div>
                </div>
              ) : (
                <div className="ext-empty-editor">
                  <span style={{ fontSize: 36 }}>🎯</span>
                  <p>选择一个 Skill 查看详情</p>
                </div>
              )}
            </div>
          </div>
        )
      )}

      {/* ========== Tab 3: 在线市场 ========== */}
      {tab === 'market' && (
        <div className="marketplace-panel">
          <div className="marketplace-sources">
            {(['all', 'builtin', 'github'] as SourceFilter[]).map(sf => (
              <button
                key={sf}
                className={`marketplace-source-btn ${sourceFilter === sf ? 'active' : ''}`}
                onClick={() => handleSourceFilterChange(sf)}
              >
                {sf === 'all' ? '全部' : sf === 'builtin' ? '官方精选' : 'GitHub'}
              </button>
            ))}
            <a href="#" className="marketplace-source-link" onClick={e => {
              e.preventDefault();
              BrowserOpenURL('https://github.com/anthropics/claude-code');
            }}>
              Claude Code ↗
            </a>
          </div>

          <div className="marketplace-search">
            <input
              className="form-input marketplace-search-input"
              value={marketQuery}
              onChange={e => handleMarketSearch(e.target.value)}
              placeholder="搜索 Skill... 如 commit, review, pdf"
            />
          </div>

          <div className="skill-scope-bar">
            <span className="skill-scope-label">安装到：</span>
            <div className="marketplace-sources">
              <button
                className={`marketplace-source-btn ${marketInstallScope === 'global' ? 'active' : ''}`}
                onClick={() => setMarketInstallScope('global')}
              >
                全局
              </button>
              <button
                className={`marketplace-source-btn ${marketInstallScope === 'project' ? 'active' : ''}`}
                onClick={async () => {
                  if (!marketProjectPath) {
                    try {
                      const dir = await SelectDirectory();
                      if (dir) {
                        setMarketProjectPath(dir);
                        setMarketInstallScope('project');
                      }
                    } catch {}
                  } else {
                    setMarketInstallScope('project');
                  }
                }}
              >
                项目级
              </button>
            </div>
            {marketInstallScope === 'project' && (
              <button className="btn btn-ghost skill-scope-path" onClick={async () => {
                try {
                  const dir = await SelectDirectory();
                  if (dir) setMarketProjectPath(dir);
                } catch {}
              }}>
                {marketProjectPath || '选择项目目录'}
              </button>
            )}
          </div>

          <div className="marketplace-scroll">
            {marketError && <div className="json-error">{marketError}</div>}

            {marketLoading ? (
              <div className="loading-state">搜索中...</div>
            ) : marketResults.length === 0 && marketSearched ? (
              <div className="empty-state">
                <div className="empty-icon">🔍</div>
                <h3>未找到结果</h3>
                <p>换个关键词试试</p>
              </div>
            ) : (
              <>
                <div className="marketplace-result-info">
                  共 {marketResults.length} 个 Skill
                  {marketResults.length > marketVisibleCount && `，当前显示 ${marketVisibleCount} 个`}
                </div>
                <div className="mcp-grid">
                  {marketResults.slice(0, marketVisibleCount).map(ext => {
                    const installed = isMarketInstalled(ext.name);
                    return (
                      <div key={`${ext.source}-${ext.name}`} className="mcp-card marketplace-card">
                        <div className="mcp-card-header">
                          <span className="mcp-card-icon">🎯</span>
                          <span className="mcp-card-name">{ext.name}</span>
                          <span className={`badge ${ext.source === 'builtin' ? 'badge-builtin' : 'badge-github'}`}>
                            {ext.source === 'builtin' ? '官方' : 'GitHub'}
                          </span>
                        </div>
                        <div className="marketplace-card-desc-cn">{ext.description}</div>
                        {ext.category && (
                          <div className="marketplace-card-meta">
                            <span className="marketplace-pkg">{ext.category}</span>
                          </div>
                        )}
                        <div className="mcp-card-actions">
                          {ext.repoUrl && (
                            <button className="btn btn-ghost"
                              onClick={() => BrowserOpenURL(ext.repoUrl)}>
                              仓库
                            </button>
                          )}
                          {installed ? (
                            <span className="badge marketplace-installed-badge">已安装</span>
                          ) : (
                            <button className="btn btn-primary marketplace-install-btn"
                              disabled={installingName === ext.name || (marketInstallScope === 'project' && !marketProjectPath)}
                              onClick={() => handleInstall(ext, marketInstallScope === 'project' ? marketProjectPath : 'global')}>
                              {installingName === ext.name ? '安装中...' : '安装'}
                            </button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
                {marketResults.length > marketVisibleCount && (
                  <div className="marketplace-footer">
                    <button
                      className="btn btn-ghost marketplace-load-more"
                      onClick={() => setMarketVisibleCount(prev => prev + 24)}
                    >
                      加载更多（还有 {marketResults.length - marketVisibleCount} 个）
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
