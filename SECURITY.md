# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email: **zenglingzi@gmail.com**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix release**: Depends on severity, typically within 2 weeks for critical issues

## Scope

This policy applies to:
- The `claude-config-mcp` npm package
- The Go MCP server binary
- The luoshu memory system (local config at `~/.luoshu/`)

## Security Design

- API keys are stored locally at `~/.luoshu/config.json` and never transmitted to third parties
- Keys are always displayed in masked form (first 3 + last 4 characters)
- Memory files are stored locally in `.claude/memory/` directories
- The MCP server runs locally and does not expose network ports by default (stdio mode)

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.7.x   | Yes       |
| < 0.7   | No        |
