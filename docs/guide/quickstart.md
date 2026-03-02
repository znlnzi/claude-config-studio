# Quick Start

After [installing](/guide/installation) claude-config-mcp, restart Claude Code and type:

```
/luoshu.setup
```

The setup wizard will:

1. **Detect your project** — identifies the project type and tech stack
2. **Ask 3 quick questions** — your preferences for templates and workflow
3. **Install templates** — matching configuration packs for your project
4. **Guide LLM setup** — optional configuration for intelligent memory

## What Gets Created

After setup, your project will have a `.claude/` directory:

```
.claude/
├── rules/          # Project rules and conventions
├── agents/         # Custom agent definitions
├── skills/         # Project-specific skills
├── commands/       # Custom commands
└── memory/         # Cross-session memory files
    ├── MEMORY.md
    └── session-state.md
```

## Basic Usage

### Save a Memory

Claude automatically saves memories during conversations, but you can also ask explicitly:

> "Remember that we chose PostgreSQL for the database"

The memory is saved to `.claude/memory/` and persists across sessions.

### Search Memories

> "What decisions did we make about the database?"

This triggers semantic search across all stored memories, returning relevant results even if the exact keywords differ.

### Intelligent Recall

> "Summarize everything we've discussed about authentication"

Luoshu combines vector search with LLM synthesis to produce a coherent answer from multiple memory entries.

### Install a Template

> "Install the hackathon-core template"

Templates are pre-built configuration packs. Use `template_list` to see all available templates.

### Manage Configuration

> "Show my global Claude Code configuration"

Access and modify CLAUDE.md, settings.json, and .mcp.json through MCP tools.

## Next Steps

- [Configuration Management](/guide/configuration) — learn about all configuration tools
- [Luoshu Intelligent Memory](/guide/luoshu) — set up LLM for full memory capabilities
- [MCP Tools Reference](/reference/tools) — complete API documentation
