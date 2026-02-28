#!/bin/bash
# SessionStart Hook: 检测项目状态，为 /setup 提供引导信息。
#
# 三种场景:
#   1. 无 .claude/ → 新项目，检测语言/框架，建议 /setup
#   2. 有 .claude/ 但无 .setup-meta.json → 已配置但未注册，建议 /setup 同步
#   3. 有 .setup-meta.json → 检查是否过期（>30天），过期则提示检查更新
#
# 退出码:
#   0 = 正常（永远不阻塞会话启动）
#
# 输出通过 stdout（SessionStart hook 的 stdout 会注入为 Claude context）。
# 同一提示不重复（per-project flag 文件去重）。

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
CLAUDE_DIR="$PROJECT_DIR/.claude"
META_FILE="$CLAUDE_DIR/.setup-meta.json"
# 每个项目一个 flag 文件，避免重复提示
HINT_FLAG="/tmp/.claude-setup-hint-$(echo "$PROJECT_DIR" | md5sum 2>/dev/null | cut -c1-8 || echo "$PROJECT_DIR" | shasum 2>/dev/null | cut -c1-8 || echo "default")"

# ─── 场景 1: 新项目（无 .claude/）──────────────────────
if [ ! -d "$CLAUDE_DIR" ]; then
    # 检测项目类型
    LANG=""
    FRAMEWORK=""
    PKG_MANAGER=""
    TEST_FRAMEWORK=""
    HAS_GIT="false"

    [ -d "$PROJECT_DIR/.git" ] && HAS_GIT="true"

    # Node.js
    if [ -f "$PROJECT_DIR/package.json" ]; then
        LANG="node"
        if [ -f "$PROJECT_DIR/pnpm-lock.yaml" ]; then PKG_MANAGER="pnpm"
        elif [ -f "$PROJECT_DIR/yarn.lock" ]; then PKG_MANAGER="yarn"
        elif [ -f "$PROJECT_DIR/bun.lockb" ]; then PKG_MANAGER="bun"
        else PKG_MANAGER="npm"; fi

        if grep -q '"next"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="nextjs"
        elif grep -q '"nuxt"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="nuxt"
        elif grep -q '"react"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="react"
        elif grep -q '"vue"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="vue"
        elif grep -q '"svelte"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="svelte"
        elif grep -q '"express"' "$PROJECT_DIR/package.json" 2>/dev/null; then FRAMEWORK="express"; fi

        if grep -q '"vitest"' "$PROJECT_DIR/package.json" 2>/dev/null; then TEST_FRAMEWORK="vitest"
        elif grep -q '"jest"' "$PROJECT_DIR/package.json" 2>/dev/null; then TEST_FRAMEWORK="jest"
        elif grep -q '"playwright"' "$PROJECT_DIR/package.json" 2>/dev/null; then TEST_FRAMEWORK="playwright"; fi
    fi

    # Python
    if [ -f "$PROJECT_DIR/pyproject.toml" ] || [ -f "$PROJECT_DIR/setup.py" ] || [ -f "$PROJECT_DIR/requirements.txt" ]; then
        LANG="${LANG:+$LANG+}python"
        if [ -z "$PKG_MANAGER" ]; then
            if [ -f "$PROJECT_DIR/pyproject.toml" ] && grep -q 'uv' "$PROJECT_DIR/pyproject.toml" 2>/dev/null; then PKG_MANAGER="uv"
            elif [ -f "$PROJECT_DIR/poetry.lock" ]; then PKG_MANAGER="poetry"
            else PKG_MANAGER="pip"; fi
        fi
        if [ -f "$PROJECT_DIR/pyproject.toml" ]; then
            if grep -q 'fastapi' "$PROJECT_DIR/pyproject.toml" 2>/dev/null; then FRAMEWORK="${FRAMEWORK:+$FRAMEWORK+}fastapi"
            elif grep -q 'django' "$PROJECT_DIR/pyproject.toml" 2>/dev/null; then FRAMEWORK="${FRAMEWORK:+$FRAMEWORK+}django"
            elif grep -q 'flask' "$PROJECT_DIR/pyproject.toml" 2>/dev/null; then FRAMEWORK="${FRAMEWORK:+$FRAMEWORK+}flask"; fi
        fi
        if [ -f "$PROJECT_DIR/pyproject.toml" ] && grep -q 'pytest' "$PROJECT_DIR/pyproject.toml" 2>/dev/null; then
            TEST_FRAMEWORK="${TEST_FRAMEWORK:+$TEST_FRAMEWORK+}pytest"
        fi
    fi

    # Go
    if [ -f "$PROJECT_DIR/go.mod" ]; then
        LANG="${LANG:+$LANG+}go"
        PKG_MANAGER="${PKG_MANAGER:-go}"
        TEST_FRAMEWORK="${TEST_FRAMEWORK:+$TEST_FRAMEWORK+}go-test"
        if grep -q 'github.com/gin-gonic/gin' "$PROJECT_DIR/go.mod" 2>/dev/null; then FRAMEWORK="${FRAMEWORK:+$FRAMEWORK+}gin"
        elif grep -q 'github.com/labstack/echo' "$PROJECT_DIR/go.mod" 2>/dev/null; then FRAMEWORK="${FRAMEWORK:+$FRAMEWORK+}echo"; fi
    fi

    # Rust
    if [ -f "$PROJECT_DIR/Cargo.toml" ]; then
        LANG="${LANG:+$LANG+}rust"
        PKG_MANAGER="${PKG_MANAGER:-cargo}"
        TEST_FRAMEWORK="${TEST_FRAMEWORK:+$TEST_FRAMEWORK+}cargo-test"
    fi

    # 没检测到任何语言 → 静默退出
    [ -z "$LANG" ] && exit 0

    echo "[detect-project] 新项目（无 .claude/ 配置）: lang=$LANG framework=$FRAMEWORK pkg=$PKG_MANAGER test=$TEST_FRAMEWORK git=$HAS_GIT — 建议运行 /setup 初始化"
    exit 0
fi

# ─── 场景 2: 已配置但无 meta 文件 ──────────────────────
if [ ! -f "$META_FILE" ]; then
    # 只提示一次（用 flag 文件去重）
    if [ ! -f "$HINT_FLAG" ]; then
        echo "1" > "$HINT_FLAG"
        echo "[detect-project] 项目已配置但未记录版本信息。运行 /setup 可同步最新功能并注册当前配置。"
    fi
    exit 0
fi

# ─── 场景 3: 有 meta 文件，检查是否过期 ──────────────────
# 读取 setup_date，比较是否超过 30 天
SETUP_DATE=$(python3 -c "import json,sys; print(json.load(open('$META_FILE')).get('setup_date',''))" 2>/dev/null || echo "")

if [ -z "$SETUP_DATE" ]; then
    # meta 文件损坏或无日期，当作需要更新
    if [ ! -f "$HINT_FLAG" ]; then
        echo "1" > "$HINT_FLAG"
        echo "[detect-project] 配置版本信息不完整。运行 /setup 可检查并同步最新功能。"
    fi
    exit 0
fi

# 计算天数差
DAYS_OLD=$(python3 -c "
from datetime import datetime, date
try:
    d = datetime.strptime('$SETUP_DATE', '%Y-%m-%d').date()
    print((date.today() - d).days)
except:
    print(999)
" 2>/dev/null || echo "999")

if [ "$DAYS_OLD" -gt 30 ]; then
    # 只提示一次
    if [ ! -f "$HINT_FLAG" ]; then
        echo "1" > "$HINT_FLAG"
        echo "[detect-project] 配置已 ${DAYS_OLD} 天未更新。运行 /setup 查看是否有新功能可用。"
    fi
fi

exit 0
