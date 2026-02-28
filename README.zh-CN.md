# claude-config-mcp

[English](README.md) | [中文](README.zh-CN.md)

[![CI](https://github.com/znlnzi/claude-config-studio/actions/workflows/ci.yml/badge.svg)](https://github.com/znlnzi/claude-config-studio/actions/workflows/ci.yml)
[![npm version](https://img.shields.io/npm/v/claude-config-mcp.svg)](https://www.npmjs.com/package/claude-config-mcp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Claude Code 在会话之间会遗忘一切。这个 MCP Server 解决了这个问题。**

claude-config-mcp 通过 MCP 工具管理 Claude Code 的 `.claude/` 配置，并提供 **洛书 (luoshu)** — 一个跨会话智能记忆系统，让 Claude 在不同对话之间记住你的决策、偏好和项目上下文。

## 功能特性

### 配置管理

通过 MCP 工具管理 Claude Code 的 `.claude/` 目录：

- **记忆** — 跨项目读写搜索 `.claude/memory/` 文件
- **配置** — 获取/保存全局和项目级配置（CLAUDE.md、settings.json、.mcp.json）
- **模板** — 安装/卸载配置模板包（规则、Agent、技能、命令）
- **扩展** — Agent、规则、技能、命令的增删改查
- **Hooks** — 管理事件钩子（PreToolUse、PostToolUse、SessionStart、Stop 等）
- **演化** — 分析规则的重复、缺失和健康问题

### 洛书 (Luoshu) 智能记忆

基于 LLM 和向量搜索的跨会话记忆系统：

- **自动提取** — 自动从对话中提取关键决策、模式和上下文
- **语义搜索** — 使用向量相似度查找相关记忆，而非仅关键词匹配
- **智能回忆** — 用自然语言提问，如"之前关于认证的决策是什么？"，获得综合回答
- **优雅降级** — 无 LLM 配置时使用关键词搜索，配置后解锁全部能力

## 安装

### npm（推荐）

```bash
npm install -g claude-config-mcp
```

这会为你的平台安装二进制文件，并注册 `/luoshu.setup` 和 `/luoshu.config` 技能。

然后注册 MCP 服务器：

```bash
claude mcp add claude-config -s user -- npx -y claude-config-mcp
```

### 从源码构建

```bash
git clone https://github.com/znlnzi/claude-config-studio.git
cd claude-config-studio
make install
```

这会构建二进制文件、安装到 `~/.local/bin/`、复制技能到 `~/.claude/skills/`，并注册到 Claude Code。

## 快速开始

安装后，重启 Claude Code 并输入：

```
/luoshu.setup
```

它会：
1. 检测你的项目类型和技术栈
2. 问 3 个快速问题了解你的偏好
3. 安装匹配的配置模板
4. 引导你设置 LLM 以启用智能记忆（可选）

## MCP 工具

### 记忆管理

| 工具 | 说明 |
|------|------|
| `save_memory` | 保存记忆条目到 `.claude/memory/` |
| `load_memory` | 从项目加载记忆文件 |
| `search_memory` | 关键词搜索，自动补充语义搜索结果 |

### 配置

| 工具 | 说明 |
|------|------|
| `config_get_global` | 获取全局 Claude Code 配置 |
| `config_save_global` | 保存全局配置字段 |
| `config_save_project` | 保存项目级配置字段 |
| `get_project_config` | 获取项目 `.claude/` 目录概览 |
| `list_projects` | 列出所有受管项目 |

### 模板

| 工具 | 说明 |
|------|------|
| `template_list` | 列出可用配置模板 |
| `template_install` | 安装模板到项目或全局范围 |
| `template_uninstall` | 卸载模板 |
| `template_installed` | 列出已安装模板 |

### 扩展

| 工具 | 说明 |
|------|------|
| `extension_list` | 列出扩展（agents/rules/skills/commands） |
| `extension_read` | 读取扩展文件 |
| `extension_save` | 创建或更新扩展 |
| `extension_delete` | 删除扩展 |

### Hooks

| 工具 | 说明 |
|------|------|
| `hooks_list` | 列出已配置的 hooks |
| `hooks_save` | 保存 hooks 配置 |

### 演化引擎

| 工具 | 说明 |
|------|------|
| `evolve_status` | 获取演化系统状态 |
| `evolve_analyze` | 分析规则问题 |
| `evolve_apply` | 批准或拒绝建议 |

### 洛书配置

| 工具 | 说明 |
|------|------|
| `luoshu_config_get` | 获取当前洛书配置（API Key 已脱敏） |
| `luoshu_config_set` | 设置配置字段（含 Key 格式预校验） |
| `luoshu_config_validate` | 测试 LLM/Embedding 连接 |

### 洛书记忆

| 工具 | 说明 |
|------|------|
| `memory_extract` | 从对话文本中提取关键要点 |
| `memory_semantic_search` | 向量相似度搜索记忆 |
| `luoshu_recall` | LLM 综合的智能回忆 |

### 洛书状态

| 工具 | 说明 |
|------|------|
| `luoshu_status` | 获取系统统计（记忆数、索引大小、缓存） |
| `luoshu_reindex` | 重建向量索引 |

## 洛书配置

洛书使用本地配置文件 `~/.luoshu/config.json`。通过 `/luoshu.config` 或环境变量配置：

| 环境变量 | 说明 |
|---------|------|
| `LUOSHU_LLM_API_KEY` | LLM 服务 API Key |
| `LUOSHU_LLM_MODEL` | LLM 模型名称 |
| `LUOSHU_EMBEDDING_API_KEY` | Embedding 服务 API Key |
| `LUOSHU_EMBEDDING_MODEL` | Embedding 模型名称 |

当前支持的 LLM 提供商：**火山引擎豆包** (Volcengine Doubao)。

## 传输模式

```bash
# stdio（默认，用于 Claude Code 集成）
claude-config-mcp

# HTTP（用于 Docker、共享部署）
claude-config-mcp --transport http --http-addr localhost:8080
```

HTTP 模式在 `/mcp` 端点暴露 Streamable HTTP 服务。

## 支持平台

| 平台 | 架构 |
|------|------|
| macOS | ARM64 (Apple Silicon) |
| macOS | x64 (Intel) |
| Linux | x64 |
| Linux | ARM64 |
| Windows | x64 |

## 项目结构

```
claude-config-studio/
├── cmd/mcp-server/       # MCP Server 入口 (Go)
├── internal/
│   ├── luoshu/           # 洛书记忆引擎（向量索引、回忆、配置）
│   ├── templatedata/     # 内置模板定义
│   └── evolution/        # 规则演化引擎
├── npm/                  # npm 包分发
├── dist/skills/          # 全局技能（/luoshu.setup、/luoshu.config）
├── python/openviking-mcp/  # OpenViking 语义搜索 MCP Server (Python)
├── scripts/              # 构建和发布脚本
└── Makefile              # 构建、测试、安装目标
```

核心组件是 **Go MCP Server**（`cmd/mcp-server/`），通过 npm 分发预编译二进制。

`python/openviking-mcp/` 是配套的 Python MCP Server，通过 [OpenViking](https://pypi.org/project/openviking/) 提供对 Claude Code 记忆和规则文件的语义搜索。

## 开发

```bash
# 仅构建 MCP Server
make mcp

# 运行 lint + 构建 + 测试
make test

# 详细输出运行测试
go test ./internal/... -v

# 交叉编译所有平台
make npm-build

# 清理构建产物
make clean
```

## 贡献

请参阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发环境设置、代码风格和 Pull Request 指南。

## 许可证

MIT
