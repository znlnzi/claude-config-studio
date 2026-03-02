# Configuration Management

claude-config-mcp provides MCP tools to manage every aspect of Claude Code's `.claude/` directory.

## Overview

| Category | What It Manages | Key Tools |
|----------|----------------|-----------|
| Memory | `.claude/memory/` files | `save_memory`, `load_memory`, `search_memory` |
| Config | CLAUDE.md, settings.json, .mcp.json | `config_get_global`, `config_save_global`, `config_save_project` |
| Templates | Configuration template packs | `template_list`, `template_install`, `template_uninstall` |
| Extensions | Agents, rules, skills, commands | `extension_list`, `extension_read`, `extension_save`, `extension_delete` |
| Hooks | Event hooks in settings.json | `hooks_list`, `hooks_save` |
| Evolution | Rules health analysis | `evolve_analyze`, `evolve_apply` |

## Memory Management

Memory files live in `.claude/memory/` and persist across Claude Code sessions.

### Save a Memory

```
Tool: save_memory
Parameters:
  project_path: "/path/to/project"  (or "global")
  filename: "MEMORY.md"
  content: "# Project Decisions\n..."
  mode: "overwrite"  (or "append")
```

### Search Memories

`search_memory` performs case-insensitive keyword search and automatically supplements with semantic search when fewer than 3 keyword matches are found.

```
Tool: search_memory
Parameters:
  query: "database decision"
  project_path: "/path/to/project"  (optional)
```

## Global vs Project Configuration

| Scope | Location | Use Case |
|-------|----------|----------|
| Global | `~/.claude/` | Personal preferences, shared across all projects |
| Project | `.claude/` in project root | Project-specific rules, templates, memory |

### Reading Configuration

```
Tool: config_get_global
# Returns: CLAUDE.md, settings.json, .mcp.json contents
```

```
Tool: get_project_config
Parameters:
  project_path: "/path/to/project"
# Returns: overview of .claude/ directory
```

### Saving Configuration

```
Tool: config_save_global
Parameters:
  field: "claude_md"  (or "settings", "mcp")
  content: "Your CLAUDE.md content..."
```

## Templates

Templates are pre-built configuration packs that install rules, agents, skills, and commands.

### Available Templates

Use `template_list` to see all templates grouped by category. See [Templates Reference](/reference/templates) for the full list.

### Installing a Template

```
Tool: template_install
Parameters:
  template_id: "hackathon-core"
  scope: "project"
  project_path: "/path/to/project"
```

### Checking Installed Templates

```
Tool: template_installed
Parameters:
  scope: "project"
  project_path: "/path/to/project"
```

## Extensions

Extensions are individual `.md` files in `.claude/agents/`, `.claude/rules/`, `.claude/skills/`, or `.claude/commands/`.

```
Tool: extension_list
Parameters:
  type: "agents"  (or "rules", "skills", "commands")
  scope: "global"  (or absolute project path)
```

## Hooks

Hooks are shell commands that execute in response to Claude Code events.

### Supported Events

| Event | When It Fires |
|-------|--------------|
| `PreToolUse` | Before a tool executes |
| `PostToolUse` | After a tool executes |
| `SessionStart` | When a session begins |
| `SessionEnd` | When a session ends |
| `Stop` | When Claude stops responding |
| `PreCompact` | Before context compression |
| `UserPromptSubmit` | When user submits a prompt |

### Managing Hooks

```
Tool: hooks_save
Parameters:
  hooks: '{"PreToolUse": [{"matcher": "Write", "hooks": [{"type": "command", "command": "echo Writing..."}]}]}'
  scope: "global"
```

## Evolution Engine

The evolution engine analyzes your rules files for issues:

- **Duplicates** — rules that overlap or contradict
- **Gaps** — missing coverage areas
- **Health issues** — outdated or poorly structured rules

```
Tool: evolve_analyze
Parameters:
  scope: "/path/to/project"
```

Review suggestions with `evolve_status` and apply them with `evolve_apply`.

## Import / Export

Share configurations between machines or team members:

```
Tool: export_config
Parameters:
  scope: "project"
  project_path: "/path/to/project"
# Returns: base64-encoded ZIP of the .claude/ directory
```

```
Tool: import_config
Parameters:
  target_path: "/path/to/new-project"
  data: "base64-encoded-zip-data..."
```
