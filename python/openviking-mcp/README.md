# openviking-mcp

MCP Server for semantic search over Claude Code memory and rules via OpenViking.

## Install

```bash
pip install openviking-mcp
```

## Register with Claude Code

```bash
claude mcp add openviking -- python -m openviking_mcp
```

## Tools

- `ov_search(query, scope, limit)` - Semantic search over memory and rules
- `ov_index(scope, force)` - Build vector index for a scope
- `ov_status()` - Check server status and index info

## Prerequisites

- Python >= 3.10
- OpenViking (`pip install openviking`)
- `~/.openviking/ov.conf` configured with embedding provider
