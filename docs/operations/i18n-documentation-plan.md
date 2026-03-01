# 多语言文档支持方案

> 运营视角 (Paul Graham 思维模型) | 2026-03-01

## 一、当前语言现状分析

### 已有双语内容 ✅
| 文件 | 英文 | 中文 | 备注 |
|------|------|------|------|
| README.md | ✅ | - | 英文版 |
| README.zh-CN.md | - | ✅ | 中文版 |
| 代码注释 | ✅ | - | v0.7.1 已全部翻译为英文 |
| 错误消息 | ✅ | - | v0.7.1 已全部翻译为英文 |
| CONTRIBUTING.md | ✅ | - | 仅英文 |
| CODE_OF_CONDUCT.md | ✅ | - | 仅英文 |
| CHANGELOG.md | ✅ | - | 仅英文 |
| SECURITY.md | ✅ | - | 仅英文 |

### 项目内部配置文件（非公开文档）
| 文件 | 语言 | 需要翻译？ |
|------|------|-----------|
| .claude/rules/*.md | 中文 | ❌ 不需要（用户私有配置，不会开源） |
| .claude/agents/*.md | 中文 | ❌ 不需要（用户私有配置） |
| .claude/skills/*.md | 中文 | ❌ 不需要（用户私有配置） |
| dist/skills/SKILL.md | 英文 | ✅ 已完成英文化 |

### 结论
代码层面已经完成英文化（v0.7.1），主要缺口在**辅助文档的中文版本**和**文档网站**。

---

## 二、多语言目录结构方案对比

### 方案 A: 文件名后缀方式 ⭐ 推荐

```
项目根目录/
├── README.md              # 英文（默认/主要）
├── README.zh-CN.md        # 中文
├── CONTRIBUTING.md         # 英文
├── CONTRIBUTING.zh-CN.md   # 中文
├── CHANGELOG.md            # 仅英文（技术性强，不需翻译）
└── docs/
    ├── guide.md            # 英文详细指南
    └── guide.zh-CN.md      # 中文详细指南
```

| 优点 | 缺点 |
|------|------|
| 零工具依赖 | 文件数翻倍 |
| GitHub 原生支持 | 无自动语言检测 |
| 贡献者门槛最低 | 大量文档时管理困难 |
| 与当前结构一致 | - |

### 方案 B: 目录分隔方式

```
docs/
├── en/
│   ├── guide.md
│   ├── api-reference.md
│   └── faq.md
└── zh-CN/
    ├── guide.md
    ├── api-reference.md
    └── faq.md
```

| 优点 | 缺点 |
|------|------|
| 结构清晰 | GitHub 无法直接渲染 |
| 适合大量文档 | 需要文档框架配合 |
| 翻译覆盖率一目了然 | 贡献者需了解目录结构 |

### 方案 C: 文档框架方式

使用 VitePress / Docusaurus 搭建独立文档站：

```
docs/
├── .vitepress/
│   └── config.ts    # i18n 配置
├── en/
│   ├── guide/
│   └── api/
└── zh-CN/
    ├── guide/
    └── api/
```

| 优点 | 缺点 |
|------|------|
| 最佳用户体验 | 需要维护文档站 |
| 内置 i18n 路由 | 增加项目复杂度 |
| 搜索、导航、主题 | 需要部署（Vercel/Netlify） |

### 推荐：方案 A（当前阶段） → 方案 C（用户量增长后）

**理由**: Paul Graham 说 "Do things that don't scale"。现阶段文档量不大（5-8 个文件），文件名后缀方式最简单、零依赖、贡献者友好。等用户量上来、文档超过 20 页时，再迁移到 VitePress。

---

## 三、文档框架评估（为将来准备）

| 框架 | i18n 支持 | 适合度 | 理由 |
|------|-----------|--------|------|
| **VitePress** | ✅ 内置 | ⭐⭐⭐⭐⭐ | Vue 生态，轻量，i18n 开箱即用，适合中小项目 |
| Docusaurus | ✅ 内置 | ⭐⭐⭐⭐ | React 生态，功能更丰富但更重 |
| MkDocs | 插件支持 | ⭐⭐⭐ | Python 生态，i18n 需要额外插件 |
| Nextra | ✅ 内置 | ⭐⭐⭐⭐ | Next.js 生态，适合已有 Next.js 项目 |

**将来推荐: VitePress** — 轻量、i18n 原生支持、Markdown 为主、部署简单。

---

## 四、翻译工作流

### 推荐：AI 初翻 + 人工审校

| 方式 | 质量 | 速度 | 成本 | 一致性 |
|------|------|------|------|--------|
| 纯人工翻译 | ⭐⭐⭐⭐⭐ | ⭐⭐ | 高 | ⭐⭐⭐ |
| **AI 初翻 + 人工审校** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 低 | ⭐⭐⭐⭐ |
| 纯 AI 翻译 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 极低 | ⭐⭐⭐ |
| 社区翻译 | ⭐⭐⭐⭐ | ⭐ | 免费 | ⭐⭐ |

### 具体工作流

```
1. 英文版作为 Source of Truth
2. 修改英文版后，用 Claude 翻译为中文（保留技术术语英文）
3. 人工审校，确保术语一致性
4. 维护术语表（docs/glossary.md）统一翻译
```

### 翻译同步策略
- 在 PR 模板中添加 checklist: "[ ] 如果修改了英文文档，请同步更新对应的中文版本"
- 可以创建 GitHub Action 检查中英文文件的修改时间差异

---

## 五、优先级排序

### 第一阶段（开源前）
1. ✅ README.md / README.zh-CN.md — 已完成
2. 🔲 CONTRIBUTING.md → CONTRIBUTING.zh-CN.md
3. 🔲 更新 README 中的 Provider 信息

### 第二阶段（开源后 1 个月内）
4. 🔲 添加详细使用指南 docs/guide.md + docs/guide.zh-CN.md
5. 🔲 添加 FAQ docs/faq.md + docs/faq.zh-CN.md
6. 🔲 添加架构图

### 第三阶段（用户量增长后）
7. 🔲 搭建 VitePress 文档站
8. 🔲 完整 API 参考文档
9. 🔲 考虑第三语言（日文？由社区驱动）

---

## 六、README 多语言切换方案

### 当前方案（已实现，推荐保持）

```markdown
[English](README.md) | [中文](README.zh-CN.md)
```

放在 README 最顶部，标题之后第一行。简洁有效。

### 可选增强

添加语言图标：
```markdown
🌍 [English](README.md) | [中文](README.zh-CN.md)
```

---

## 七、国际化社区运营策略

### 短期（开源后 3 个月）
1. **Issue 模板双语化**: Bug Report 和 Feature Request 保持英文，但说明中添加 "中文 issues are welcome"
2. **Discussion 区**: 开启 GitHub Discussions，分中英文分类
3. **响应时间**: 中英文 issue 同等对待，48小时内响应

### 中期（3-6 个月）
4. **贡献者引导**: 标记 "good first issue" + "help wanted" 吸引贡献者
5. **模板贡献**: 鼓励社区贡献新的配置模板（最容易上手的贡献方式）
6. **Provider 贡献**: 鼓励社区添加新的 LLM Provider 预设

### 长期
7. **社区翻译**: 如果有非中英语用户贡献翻译，欢迎接纳
8. **区域性推广**: 中文区走 V2EX/掘金，英文区走 HN/Reddit
