#!/usr/bin/env bash
set -euo pipefail

# Claude Config MCP Server — Installer
# Usage: curl -sSL <url>/install.sh | bash
#   or:  git clone ... && ./scripts/install.sh

REPO="https://github.com/anthropics/ClaudeCode-Config-Studio"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="claude-config-mcp"

info()  { echo "  [INFO]  $*"; }
ok()    { echo "  [OK]    $*"; }
err()   { echo "  [ERROR] $*" >&2; }
fatal() { err "$*"; exit 1; }

# ─── Pre-flight checks ──────────────────────────────

check_go() {
    if ! command -v go >/dev/null 2>&1; then
        fatal "Go is required but not installed. See https://go.dev/dl/"
    fi
    local ver
    ver=$(go version | grep -oP 'go(\d+\.\d+)' | head -1)
    info "Found Go ${ver}"
}

check_claude() {
    if command -v claude >/dev/null 2>&1; then
        ok "Claude CLI found"
        return 0
    else
        info "Claude CLI not found — you'll need to register manually"
        return 1
    fi
}

# ─── Build ───────────────────────────────────────────

build_from_source() {
    local project_dir="$1"
    info "Building MCP Server..."
    cd "$project_dir"
    go build -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/mcp-server/
    ok "Built ${BINARY_NAME}"
}

build_with_go_install() {
    info "Installing via go install..."
    go install "${REPO}/cmd/mcp-server@latest"
    # go install puts it in GOBIN or GOPATH/bin
    local gobin
    gobin=$(go env GOBIN)
    if [ -z "$gobin" ]; then
        gobin="$(go env GOPATH)/bin"
    fi
    if [ -f "${gobin}/mcp-server" ]; then
        mkdir -p "$INSTALL_DIR"
        mv "${gobin}/mcp-server" "${INSTALL_DIR}/${BINARY_NAME}"
        ok "Installed to ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

# ─── Install ─────────────────────────────────────────

find_project_root() {
    if [ -f "go.mod" ] && grep -q "ClaudeCode-Config-Studio" go.mod 2>/dev/null; then
        echo "$(pwd)"
    elif [ -f "../go.mod" ] && grep -q "ClaudeCode-Config-Studio" ../go.mod 2>/dev/null; then
        echo "$(cd .. && pwd)"
    else
        echo ""
    fi
}

install_binary() {
    mkdir -p "$INSTALL_DIR"

    local root
    root=$(find_project_root)
    if [ -z "$root" ]; then
        fatal "Please run this script from the project root directory."
    fi
    build_from_source "$root"
}

install_global_skills() {
    local root
    root=$(find_project_root)
    if [ -z "$root" ]; then return; fi

    local dist_dir="$root/dist/skills"
    if [ ! -d "$dist_dir" ]; then
        info "No distributable skills found, skipping."
        return
    fi

    local claude_skills="${HOME}/.claude/skills"
    mkdir -p "$claude_skills"

    # Copy each skill from dist/skills/ to ~/.claude/skills/
    for skill_dir in "$dist_dir"/*/; do
        [ -d "$skill_dir" ] || continue
        local skill_name
        skill_name=$(basename "$skill_dir")
        mkdir -p "${claude_skills}/${skill_name}"
        cp "$skill_dir"* "${claude_skills}/${skill_name}/"
    done
    ok "Global skills installed to ~/.claude/skills/"
}

register_mcp() {
    if check_claude; then
        info "Registering MCP server..."
        claude mcp add claude-config -s user -- "$BINARY_NAME" 2>/dev/null || true
        ok "Registered claude-config MCP server"
    else
        echo ""
        echo "  To register manually, run:"
        echo "    claude mcp add claude-config -s user -- ${BINARY_NAME}"
        echo ""
    fi
}

ensure_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        echo ""
        info "${INSTALL_DIR} is not in your PATH."
        echo "  Add this to your shell profile (~/.zshrc or ~/.bashrc):"
        echo ""
        echo "    export PATH=\"\${HOME}/.local/bin:\${PATH}\""
        echo ""
    fi
}

# ─── Verify ──────────────────────────────────────────

verify() {
    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        local version
        version=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1"}}}' | "${INSTALL_DIR}/${BINARY_NAME}" 2>/dev/null | python3 -c "import sys,json; d=json.loads(sys.stdin.readline()); print(d['result']['serverInfo']['version'])" 2>/dev/null || echo "unknown")
        ok "Installed claude-config-mcp v${version}"
        return 0
    else
        err "Binary not found at ${INSTALL_DIR}/${BINARY_NAME}"
        return 1
    fi
}

# ─── Main ────────────────────────────────────────────

main() {
    echo ""
    echo "  Claude Config MCP Server — Installer"
    echo "  ====================================="
    echo ""

    check_go
    install_binary
    install_global_skills
    ensure_path
    register_mcp

    echo ""
    if verify; then
        echo ""
        echo "  Installation complete!"
        echo "  Restart Claude Code to start using the MCP tools."
        echo "  Type /setup in any project to get started."
        echo ""
    else
        echo ""
        echo "  Installation may have issues. Check the errors above."
        echo ""
        exit 1
    fi
}

main "$@"
