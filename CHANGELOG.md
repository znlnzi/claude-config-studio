# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.2] - 2026-03-01

### Added
- Multi-provider support: 7 built-in presets (OpenAI, DeepSeek, Moonshot, Zhipu, SiliconFlow, Volcengine, Custom)
- `luoshu_provider_list` tool to list all available LLM provider presets
- Auto-fill endpoint/model defaults when setting `llm.provider`
- MCP Resources: read-only access to CLAUDE.md and memory files via `claude://` URIs
- Config import/export via `export_config` and `import_config` tools (base64-zip format)
- Unified recall: `luoshu_recall` now merges JSONL memory search with file semantic search

### Changed
- Renamed `VolcengineProvider` to `OpenAICompatProvider` for generic OpenAI-compatible API support
- Enhanced `Recaller` with optional `ClaudeIndex` integration for cross-source search
- Added `sk-` generic prefix detection in API key validation

## [0.7.1] - 2026-03-01

### Changed
- Full English internationalization of all Go source code (comments, error messages, string literals)
- Translated template content: builtin.go, hackathon.go, solopreneur.go (agents, skills, rules, CLAUDE.md variants)
- Translated luoshu core: config, validation, search, recall, extractor, vector index, volcengine provider
- Translated evolution engine: analyzer, store, types
- Translated all service files: extension, MCP, skill, template, project, config, export, hooks, plugin
- Translated skills: `/luoshu.setup` and `/luoshu.config` SKILL.md files
- Language sections changed to "All responses in the user's preferred language"

### Added
- LICENSE (MIT)
- CONTRIBUTING.md with development setup and contribution guidelines
- CODE_OF_CONDUCT.md (Contributor Covenant v2.1)
- CI workflow (.github/workflows/ci.yml)
- Issue templates (bug report, feature request) and PR template
- README badges (CI, npm version, license)

## [0.7.0] - 2026-02-28

### Added
- Official auto-memory cooperation: `ov_search` now indexes project root `MEMORY.md` (Anthropic's auto-memory file)
- `context-fields` template: SessionStart hook for dynamic context injection (git branch, recent commits, project structure)
- `/council` skill: Multi-perspective code review with 3 parallel agents (architecture, security, performance)
- `PreCompact` hook in cross-session-memory template for auto-saving state before context compression
- Community resource references in server instructions (aimtpl.com, awesome-claude-code-subagents)

### Changed
- Updated `memory_extract` tool description for PreCompact trigger awareness
- Enhanced `ov_search` description to mention root MEMORY.md support

## [0.6.0] - 2026-02-27

### Added
- npm multi-platform distribution (darwin-arm64, darwin-x64, linux-x64, linux-arm64, win32-x64)
- `/luoshu.setup` and `/luoshu.config` global skills distributed via npm
- Cross-compile Makefile targets (`npm-build`, `npm-prepare`, `npm-publish`)
- Platform-specific npm packages (`@claude-config/darwin-arm64`, etc.)

### Changed
- Simplified installation: `npm install -g claude-config-mcp`

## [0.5.0] - 2026-02-25

### Added
- Luoshu (洛书) intelligent memory system
  - `memory_extract`: Auto-extract key decisions from conversations
  - `memory_semantic_search`: Vector similarity search over memories
  - `luoshu_recall`: LLM-powered intelligent recall with synthesized answers
- Luoshu configuration tools (`luoshu_config_get`, `luoshu_config_set`, `luoshu_config_validate`)
- OpenViking semantic search integration (`ov_search`, `ov_index`, `ov_status`)
- Graceful degradation: keyword search without LLM, full semantic with LLM

## [0.4.0] - 2026-02-22

### Added
- Template system with built-in packs (hackathon-core, cross-session-memory, solo-full, etc.)
- Extension CRUD operations (agents, rules, skills, commands)
- Evolution engine: analyze rules for duplicates, gaps, and health issues
- Hooks management (PreToolUse, PostToolUse, SessionStart, Stop, etc.)

## [0.3.0] - 2026-02-18

### Added
- MCP server with stdio transport
- HTTP transport mode for Docker/shared deployments
- Core configuration tools (config_get_global, config_save_global, config_save_project)
- Memory management tools (save_memory, load_memory, search_memory)
- Project listing and overview tools

## [0.1.0] - 2026-02-10

### Added
- Initial release as Wails desktop application
- Visual configuration editor for `.claude/` directory
- Monaco Editor integration for JSON/Markdown editing
- Project discovery and management UI
