# Contributing to claude-config-mcp

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- Claude Code CLI (for testing MCP integration)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/znlnzi/claude-config-studio.git
cd claude-config-studio

# Build MCP server
make mcp

# Run tests
go test ./internal/luoshu/ -v

# Install locally
make install
```

### Project Structure

```
cmd/mcp-server/       # MCP server entry point
internal/luoshu/      # Core logic (memory, search, indexing)
internal/templatedata/ # Built-in templates
dist/skills/          # Global skills (luoshu.setup, luoshu.config)
npm/                  # npm package for distribution
```

## How to Contribute

### Reporting Issues

- Check existing issues first to avoid duplicates
- Include reproduction steps, expected vs actual behavior
- Attach relevant logs or screenshots

### Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes following the code style below
4. Add tests for new functionality
5. Ensure all tests pass: `go test ./internal/...`
6. Ensure the build succeeds: `make mcp`
7. Commit with [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation
   - `refactor:` code restructuring
   - `test:` adding tests
   - `chore:` build/tooling
8. Open a Pull Request

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions under 50 lines
- Keep files under 800 lines
- Write meaningful error messages
- Add tests for new functionality (target 80%+ coverage)

### Adding Templates

Templates are defined in `internal/templatedata/`. To add a new template:

1. Create the template definition in the appropriate file
2. Register it in `catalog.go`
3. Build and verify: `go build ./cmd/mcp-server/`

### Adding MCP Tools

MCP tools are defined in `cmd/mcp-server/`. Follow the existing patterns for tool registration and handler implementation.

## Code of Conduct

Be respectful and constructive. We are all here to build something useful together.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
