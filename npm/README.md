# claude-config-mcp

MCP Server for Claude Code — configuration management + cross-session intelligent memory.

## Quick Start

```bash
# 1. Install
npm install -g claude-config-mcp

# 2. Register with Claude Code
claude mcp add claude-config -s user -- npx -y claude-config-mcp

# 3. Restart Claude Code, then type in any project:
/luoshu.setup
```

That's it. Claude will guide you through project setup.

## What It Does

### Configuration Management

Manage your `.claude/` directory programmatically — rules, agents, skills, commands, hooks, and templates.

- **Templates**: Install pre-built configuration sets (`/luoshu.setup` handles this)
- **Extensions**: CRUD for rules, agents, skills, and commands files
- **Hooks**: Configure pre/post tool hooks
- **Evolution**: Analyze rules for duplicates and health issues

### Cross-Session Memory (Luoshu)

Claude remembers context across sessions — decisions, architecture, patterns, and progress.

- **Auto-extract**: LLM extracts key points from conversations and saves them as structured memory entries
- **Semantic search**: Find related memories using natural language, even with different wording
- **Intelligent recall**: Ask "what decisions were made about auth?" and get a synthesized answer
- **Keyword search**: Fast exact-match fallback, always available

### File Semantic Search

Semantic search over your `.claude/memory/*.md`, `.claude/rules/*.md`, and project root `MEMORY.md` (official auto-memory) files.

- **ov_search**: Find rules and memory by meaning, not just keywords
- **Auto-indexing**: Files are indexed on first search, incrementally synced after
- **Auto-memory compatible**: Automatically indexes Anthropic's official `MEMORY.md` at project root
- **No setup required**: Works automatically once embedding is configured

### Templates & Skills Highlights

- **context-fields**: Auto-inject project context (git branch, recent commits, project structure) at session start via hooks
- **/council**: Multi-perspective code review — 3 parallel Agents (architect, security, performance) review your changes independently, then merge findings into a unified report
- **PreCompact hook**: The cross-session-memory template now triggers a reminder before context compression, helping preserve critical session state

## Configuration

### Basic (No API Key)

Works out of the box:
- Configuration management (templates, extensions, hooks)
- Keyword search over memory files
- Session state tracking

### Full Features (With API Key)

To enable semantic search, auto-extraction, and intelligent recall:

```
/luoshu.config
```

This configures the LLM and Embedding providers. Keys are stored locally in `~/.luoshu/config.json`.

Currently supports Volcengine (Doubao) and any OpenAI-compatible API.

## Tools Reference

### Memory (3 tools)

| Tool | Description |
|------|-------------|
| `save_memory` | Save a memory file to `.claude/memory/` |
| `load_memory` | Load memory files from `.claude/memory/` |
| `search_memory` | Keyword search over memory files |

### Config (5 tools)

| Tool | Description |
|------|-------------|
| `config_get_global` | Read global Claude Code config |
| `config_save_global` | Write global config fields |
| `config_save_project` | Write project-level config fields |
| `get_project_config` | Overview of project's `.claude/` directory |
| `list_projects` | List all managed projects |

### Templates (4 tools)

| Tool | Description |
|------|-------------|
| `template_list` | List available templates |
| `template_install` | Install a template |
| `template_uninstall` | Remove a template |
| `template_installed` | List installed templates |

### Extensions (4 tools)

| Tool | Description |
|------|-------------|
| `extension_list` | List extensions (agents/rules/skills/commands) |
| `extension_read` | Read an extension file |
| `extension_save` | Create or update an extension |
| `extension_delete` | Delete an extension |

### Hooks (2 tools)

| Tool | Description |
|------|-------------|
| `hooks_list` | List configured hooks |
| `hooks_save` | Save hooks configuration |

### Evolution (3 tools)

| Tool | Description |
|------|-------------|
| `evolve_status` | Check evolution system status |
| `evolve_analyze` | Analyze rules for issues |
| `evolve_apply` | Approve or reject suggestions |

### Luoshu Memory (5 tools)

| Tool | Description |
|------|-------------|
| `memory_extract` | Auto-extract key points from session summaries (auto-triggered on PreCompact) |
| `memory_semantic_search` | Semantic search over structured memories |
| `luoshu_recall` | Intelligent recall with LLM synthesis |
| `luoshu_status` | System status and statistics |
| `luoshu_reindex` | Rebuild memory vector index |

### Luoshu Config (3 tools)

| Tool | Description |
|------|-------------|
| `luoshu_config_get` | Read luoshu configuration |
| `luoshu_config_set` | Set a configuration value |
| `luoshu_config_validate` | Validate config and test API connection |

### File Semantic Search (3 tools)

| Tool | Description |
|------|-------------|
| `ov_search` | Semantic search over `.claude/` files and root `MEMORY.md` |
| `ov_index` | Build or rebuild file index |
| `ov_status` | Check index status |

## Supported Platforms

| Platform | Architecture |
|----------|-------------|
| macOS | Apple Silicon (arm64), Intel (x64) |
| Linux | x64, arm64 |
| Windows | x64 |

## How It Works

This package distributes a pre-compiled Go binary via npm's `optionalDependencies` multi-platform pattern. On `npm install`, only the binary for your platform is downloaded. A thin Node.js wrapper (`bin/cli.js`) locates and executes the native binary.

The postinstall script also copies `/luoshu.setup` and `/luoshu.config` skills to `~/.claude/skills/` for quick access.

## License

MIT
