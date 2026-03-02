# MCP Resources

claude-config-mcp exposes read-only access to configuration files via MCP resource URIs using the `claude://` scheme.

## Resource URIs

| URI | Description |
|-----|-------------|
| `claude://global/claude-md` | Global `~/.claude/CLAUDE.md` |
| `claude://global/memory/{filename}` | Global memory file |
| `claude://project/{project_path}/memory/{filename}` | Project memory file |

## Usage

Resources are read-only and accessed through the MCP Resources capability. They allow MCP clients to read configuration and memory files without using tools.

### Global CLAUDE.md

```
URI: claude://global/claude-md
```

Returns the content of `~/.claude/CLAUDE.md`, which contains the user's global instructions for Claude Code.

### Global Memory Files

```
URI: claude://global/memory/MEMORY.md
URI: claude://global/memory/session-state.md
```

Returns memory files from `~/.claude/memory/`.

### Project Memory Files

```
URI: claude://project//Users/me/my-project/memory/decisions.md
```

Returns memory files from a specific project's `.claude/memory/` directory.

## Constraints

- Filenames must end in `.md`
- Filenames cannot contain `..` or path separators
- Access is read-only — use `save_memory` or `config_save_*` tools to write
