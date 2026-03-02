# MCP Tools

All tools are exposed via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) over stdio or HTTP transport.

## Memory Management

### save_memory

Save a memory entry to `.claude/memory/`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_path` | string | yes | Absolute project path, or `"global"` for `~/.claude/memory/` |
| `filename` | string | yes | Filename (e.g., `MEMORY.md`, `session-state.md`) |
| `content` | string | yes | File content to save |
| `mode` | string | no | `"overwrite"` (default) or `"append"` |

**Returns:** `{ success, path }`

### load_memory

Load memory files from `.claude/memory/`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_path` | string | yes | Absolute project path, or `"global"` |
| `filename` | string | no | Specific file to load; omit to list all `.md` files |

**Returns (single file):** `{ filename, content, path, modified_at }`

**Returns (list):** `{ files: [{ filename, content, size, modified_at }] }`

### search_memory

Keyword search over memory files (case-insensitive). Automatically supplements with semantic search when fewer than 3 keyword matches are found.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | yes | Search keyword |
| `project_path` | string | no | Limit to specific project; omit to search all |
| `max_results` | number | no | Maximum results (default: 20) |

**Returns:** `{ results, total_matches, query, search_method, semantic_results? }`

## Configuration

### config_get_global

Get global configuration overview including CLAUDE.md, settings.json, and .mcp.json.

*No parameters.*

**Returns:** `{ claude_home, claude_md?, settings?, mcp_config?, has_claude_md, has_settings, has_mcp, has_agents, has_rules, has_skills, has_commands, has_memory }`

### config_save_global

Save a global configuration field. Validates JSON for `settings` and `mcp` fields.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `field` | string | yes | `"claude_md"`, `"settings"`, or `"mcp"` |
| `content` | string | yes | Content to write |

**Returns:** `{ success, path, field }`

### config_save_project

Save a project-level configuration field.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_path` | string | yes | Absolute project path |
| `field` | string | yes | `"claude_md"`, `"settings"`, or `"mcp"` |
| `content` | string | yes | Content to write |

**Returns:** `{ success, path, field, project_path }`

### get_project_config

Get an overview of a project's `.claude/` directory.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_path` | string | yes | Absolute project path |

**Returns:** `{ path, name, has_claude_md, claude_md_preview, has_settings, has_mcp, has_hooks, has_agents, has_commands, has_skills, has_rules, has_memory, config_count }`

### list_projects

List all Claude Code managed projects found in `~/.claude/projects/`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | number | no | Maximum projects to return (default: 50) |

**Returns:** `{ projects: [{ name, path, config_count, has_memory, last_modified }], total }`

## Templates

### template_list

List all available configuration templates.

*No parameters.*

**Returns:** `{ categories: [{ id, name, icon, templates: [{ id, name, category, description, tags }] }], total_templates }`

### template_install

Install a template to project or global scope.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `template_id` | string | yes | Template ID (e.g., `"hackathon-core"`) |
| `scope` | string | no | `"project"` (default) or `"global"` |
| `project_path` | string | no | Required for project scope |
| `overwrite` | string | no | `"true"` to overwrite existing files |

**Returns:** `{ success, template_id, template_name, scope, installed_files, total_files }`

### template_uninstall

Remove a template's rules file (`tpl-{id}.md`).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `template_id` | string | yes | Template ID to uninstall |
| `scope` | string | no | `"project"` (default) or `"global"` |
| `project_path` | string | no | Required for project scope |

**Returns:** `{ success, template_id, scope, removed }`

### template_installed

List installed templates by scanning for `tpl-*.md` files.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `scope` | string | no | `"project"` (default) or `"global"` |
| `project_path` | string | no | Required for project scope |

**Returns:** `{ installed: [{ template_id, name, file_path }], scope, total }`

## Extensions

### extension_list

List extension files (agents, rules, skills, or commands).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | yes | `"agents"`, `"rules"`, `"skills"`, or `"commands"` |
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ files: [{ name, filename, is_dir, size, modified_at }], type, scope, total }`

### extension_read

Read the content of a specific extension file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | yes | Extension type |
| `name` | string | yes | Extension name (without `.md`) |
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ name, type, scope, content, path }`

### extension_save

Create or update an extension file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | yes | Extension type |
| `name` | string | yes | Extension name (without `.md`) |
| `content` | string | yes | File content |
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ success, name, type, scope, path }`

### extension_delete

Delete an extension file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | yes | Extension type |
| `name` | string | yes | Extension name (without `.md`) |
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ success, name, type, scope }`

## Hooks

### hooks_list

List all configured hooks grouped by event type.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ scope, events: [{ event, entries }], total_events }`

### hooks_save

Save hooks configuration. Only modifies the `hooks` field in `settings.json`, preserving all other settings.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `hooks` | string | yes | Complete hooks JSON object |
| `scope` | string | no | `"global"` (default) or absolute project path |

Valid event names: `PreToolUse`, `PostToolUse`, `SessionStart`, `SessionEnd`, `Stop`, `PreCompact`, `UserPromptSubmit`

**Returns:** `{ success, scope, events_saved, path }`

## Evolution

### evolve_status

Get evolution system status: pending suggestions, history statistics.

*No parameters.*

**Returns:** `{ status, pending_suggestions: [{ id, type, title, confidence }] }`

### evolve_analyze

Analyze rules for duplicates, missing coverage, and health issues.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `scope` | string | no | `"global"` (default) or absolute project path |

**Returns:** `{ scope, rules_scanned, suggestions_found, duration_ms, suggestions }`

### evolve_apply

Approve or reject a suggestion.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `suggestion_id` | string | yes | Suggestion ID |
| `action` | string | yes | `"approve"` or `"reject"` |

**Returns:** `{ success, suggestion_id, action, new_status, title }`

## Luoshu Config

### luoshu_config_get

Read luoshu configuration. API Keys are automatically masked.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `section` | string | no | `"llm"`, `"embedding"`, `"memory"`, or `"reminder"`; omit for all |

**Returns:** Full or partial config object with masked API keys.

### luoshu_config_set

Set a configuration value. Connection test runs automatically after setting an API key.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `key` | string | yes | Config path (e.g., `"llm.api_key"`, `"llm.provider"`) |
| `value` | string | yes | Value to set |

**Available keys:** `llm.provider`, `llm.api_key`, `llm.endpoint`, `llm.model`, `llm.max_tokens`, `llm.temperature`, `embedding.provider`, `embedding.api_key`, `embedding.endpoint`, `embedding.model`, `embedding.dimensions`, `memory.auto_extract`, `memory.retention_days`, `memory.max_entries`, `memory.vector_search_top_k`, `reminder.dismissed`, `reminder.permanently_dismissed`

**Returns:** `{ success, key, connection_test? }`

### luoshu_config_validate

Validate configuration completeness and test LLM API connection.

*No parameters.*

**Returns:** `{ valid, issues, connection_test: { connected, status, error? } }`

### luoshu_provider_list

List all available LLM/Embedding provider presets.

*No parameters.*

**Returns:** `{ providers: [{ name, display_name, llm_endpoint, llm_model, embed_endpoint, embed_model, embed_dimension }], usage }`

## Luoshu Memory

### memory_extract

Extract key points from a session summary and save as memory entries. Requires LLM configuration.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_summary` | string | yes | Session summary text |
| `project_path` | string | no | Associated project path |
| `tags` | string | no | Additional tags, comma-separated |

**Returns:** `{ extracted, entries }`

### memory_semantic_search

Semantic search over memories using natural language.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | yes | Natural language query |
| `project_path` | string | no | Limit to specific project |
| `max_results` | number | no | Maximum results (default: 10) |
| `mode` | string | no | `"auto"` (default), `"semantic"`, or `"keyword"` |

**Returns:** `{ results, total, search_method, query }`

### luoshu_recall

Intelligent recall: semantic search + LLM synthesis into a coherent answer. Merges JSONL memory with file semantic search results.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | yes | Natural language query |
| `project_path` | string | no | Limit to specific project |
| `max_results` | number | no | Maximum search sources (default: 5) |

**Returns:** `{ answer, sources, file_sources?, search_method, query }`

## Luoshu Status

### luoshu_status

Return overall system status.

*No parameters.*

**Returns:** `{ version, llm_configured, embedding_configured, memory_entries, vector_index_entries, embedding_cache_size, config_path, features }`

### luoshu_reindex

Rebuild the vector index for all memory entries. Requires Embedding configuration.

*No parameters.*

**Returns:** `{ reindexed, total, duration_ms }`

## File Semantic Search (OpenViking)

### ov_search

Semantic search over `.claude/memory/*.md`, `.claude/rules/*.md`, and project root `MEMORY.md`.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | yes | Natural language query |
| `scope` | string | no | `"global"` or absolute project path |
| `limit` | number | no | Maximum results (default: 10) |

**Returns:** `{ results, total, query, scope }`

### ov_index

Build or update the vector index. Usually not needed — `ov_search` auto-syncs.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `scope` | string | no | `"global"` or absolute project path |
| `force` | boolean | no | Full rebuild (default: false) |

**Returns:** `{ changes, total_indexed, vector_entries, duration_ms, force, scope }`

### ov_status

Check index status and configuration.

*No parameters.*

**Returns:** `{ initialized, embedding_configured, embedding_model, global_index: { indexed_files, vector_entries, last_reconciled, index_dir } }`

## Import / Export

### export_config

Export configuration as base64-encoded ZIP.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `scope` | string | yes | `"global"` or `"project"` |
| `project_path` | string | no | Required when scope is `"project"` |

**Returns:** `{ success, format: "base64-zip", size, files, data }`

### import_config

Import configuration from base64-encoded ZIP.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `target_path` | string | yes | Absolute path to import into |
| `data` | string | yes | Base64-encoded ZIP data |

**Returns:** `{ success, files_imported, file_names }`
