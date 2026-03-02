# Roadmap

## Current: v0.8.0 (Released)

- Multi-provider support (OpenAI, DeepSeek, Moonshot, Zhipu, SiliconFlow, Volcengine, Custom)
- Unified recall: JSONL memory + file semantic search merged in `luoshu_recall`
- MCP Resources for read-only access to CLAUDE.md and memory files
- Configuration import/export (base64-zip)
- 153 handler unit tests, Mermaid architecture diagrams
- CI with codecov coverage and cross-platform build matrix

## v0.9.0 — Developer Experience

- [ ] Interactive setup wizard improvements (smarter project detection)
- [ ] Template marketplace: community-contributed templates via GitHub
- [ ] Better error messages with actionable fix suggestions
- [ ] `--version` flag for the binary

## v1.0.0 — Production Ready

- [ ] Stable API: all MCP tool schemas frozen, breaking changes require major version
- [ ] 80%+ test coverage across all packages
- [ ] VitePress documentation site with guides and API reference
- [ ] Performance benchmarks for semantic search and recall
- [ ] Windows CI testing in build matrix

## Future Ideas

- **Multi-language memory**: auto-detect and index memories in any language
- **Memory expiration**: automatic cleanup of stale memories based on retention policy
- **Shared team memory**: export/import memory snapshots for team onboarding
- **MCP Prompts**: expose curated prompt templates via MCP Prompts capability
- **Plugin system**: third-party extensions for custom memory backends
- **Wails desktop app**: visual configuration editor (paused, MCP-first approach)

## Contributing

Have an idea? [Open an issue](https://github.com/znlnzi/claude-config-studio/issues) or check [CONTRIBUTING.md](CONTRIBUTING.md) to get started.
