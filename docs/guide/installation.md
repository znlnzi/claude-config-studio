# Installation

## npm (Recommended)

```bash
npm install -g claude-config-mcp
```

This installs the pre-built binary for your platform and registers the `/luoshu.setup` and `/luoshu.config` skills.

Then register the MCP server with Claude Code:

```bash
claude mcp add claude-config -s user -- npx -y claude-config-mcp
```

## From Source

```bash
git clone https://github.com/znlnzi/claude-config-studio.git
cd claude-config-studio
make install
```

This builds the binary, installs it to `~/.local/bin/`, copies skills to `~/.claude/skills/`, and registers with Claude Code.

### Build Prerequisites

- Go 1.23+
- Node.js 18+
- Claude Code CLI (for testing MCP integration)

## Supported Platforms

| Platform | Architecture |
|----------|-------------|
| macOS | ARM64 (Apple Silicon) |
| macOS | x64 (Intel) |
| Linux | x64 |
| Linux | ARM64 |
| Windows | x64 |

## Transport Modes

The server supports two transport modes:

```bash
# stdio (default, for Claude Code integration)
claude-config-mcp

# HTTP (for Docker, shared deployments)
claude-config-mcp --transport http --http-addr localhost:8080
```

HTTP mode exposes a Streamable HTTP endpoint at `/mcp`.

## Verify Installation

After installation, restart Claude Code and run:

```
/luoshu.setup
```

If the setup wizard launches, the installation is working correctly. See [Quick Start](/guide/quickstart) for next steps.
