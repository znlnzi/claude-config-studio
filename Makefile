.PHONY: mcp build install uninstall clean test npm-build npm-prepare npm-publish

# MCP Server binary name
MCP_BIN = build/bin/claude-config-mcp
MCP_PKG = ./cmd/mcp-server/
INSTALL_DIR = $(HOME)/.local/bin
CLAUDE_DIR = $(HOME)/.claude

# Version from npm/package.json (single source of truth)
VERSION := $(shell node -p "require('./npm/package.json').version" 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.serverVersion=$(VERSION)"

# npm multi-platform build
PLATFORMS = darwin-arm64 darwin-amd64 linux-amd64 linux-arm64 windows-amd64

# Build MCP Server only (no Wails dependency)
mcp:
	go build $(LDFLAGS) -o $(MCP_BIN) $(MCP_PKG)
	@echo "Built $(MCP_BIN) (v$(VERSION))"

# Build everything (includes Wails desktop app)
build: mcp
	~/go/bin/wails build

# Install MCP Server + global skills to user scope
install: mcp
	@echo ""
	@echo "=== Installing claude-config ==="
	@echo ""
	@# 1. Install binary
	@mkdir -p $(INSTALL_DIR)
	@cp $(MCP_BIN) $(INSTALL_DIR)/
	@echo "[1/3] Binary → $(INSTALL_DIR)/claude-config-mcp"
	@# 2. Install global skills
	@mkdir -p $(CLAUDE_DIR)/skills/luoshu.setup
	@cp dist/skills/luoshu.setup/SKILL.md $(CLAUDE_DIR)/skills/luoshu.setup/SKILL.md
	@mkdir -p $(CLAUDE_DIR)/skills/luoshu.config
	@cp dist/skills/luoshu.config/SKILL.md $(CLAUDE_DIR)/skills/luoshu.config/SKILL.md
	@echo "[2/3] Skills → ~/.claude/skills/luoshu.*/"
	@# 3. Register MCP server
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp add claude-config -s user -- claude-config-mcp 2>/dev/null; \
		echo "[3/3] MCP server registered with Claude Code."; \
	else \
		echo "[3/3] Claude CLI not found. Register manually:"; \
		echo "      claude mcp add claude-config -s user -- claude-config-mcp"; \
	fi
	@echo ""
	@echo "Done! Restart Claude Code to activate."
	@echo "Then type /luoshu.setup in any project to get started."
	@echo ""

# Uninstall MCP Server + global skills
uninstall:
	@if command -v claude >/dev/null 2>&1; then \
		claude mcp remove claude-config -s user 2>/dev/null || true; \
		echo "Unregistered MCP server from Claude Code."; \
	fi
	@rm -f $(INSTALL_DIR)/claude-config-mcp
	@rm -rf $(CLAUDE_DIR)/skills/luoshu.setup
	@rm -rf $(CLAUDE_DIR)/skills/luoshu.config
	@rm -rf $(CLAUDE_DIR)/skills/setup
	@echo "Uninstalled claude-config-mcp and global skills."

# Run lint, build, and tests
test:
	go vet $(MCP_PKG) ./internal/... ./services/...
	go build $(LDFLAGS) $(MCP_PKG)
	go test ./cmd/mcp-server/ ./internal/... -count=1
	@echo "All checks passed."

# Cross-compile MCP Server for all npm platforms
npm-build:
	@for p in $(PLATFORMS); do \
		os=$$(echo $$p | cut -d- -f1); \
		arch=$$(echo $$p | cut -d- -f2); \
		npm_platform=$$(echo $$p | sed 's/amd64/x64/' | sed 's/windows/win32/'); \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		mkdir -p npm/platforms/$$npm_platform/bin; \
		echo "Building $$os/$$arch → npm/platforms/$$npm_platform/bin/claude-config-mcp$$ext"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o npm/platforms/$$npm_platform/bin/claude-config-mcp$$ext $(MCP_PKG); \
	done
	@echo "All platforms built."

# Generate platform package.json files
npm-prepare: npm-build
	@node scripts/generate-platform-packages.js

# Publish all npm packages (platform packages first, then main)
npm-publish: npm-prepare
	@node scripts/publish-all.js

# Clean build artifacts
clean:
	rm -f $(MCP_BIN)
	rm -rf npm/platforms/
