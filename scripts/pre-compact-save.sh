#!/bin/bash
# PreCompact hook: force Claude to save session state before compaction.
#
# On auto-compact: check if session-state.md was updated recently.
#   - If not → exit 2 (block compact, force Claude to save first)
#   - If yes → exit 0 (allow compact)
#
# On manual /compact: always allow (user explicitly asked for it)
#
# Safety: max 1 retry to prevent infinite loops at high context usage.

# Read hook input from stdin
INPUT=$(cat)
TRIGGER=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('trigger','auto'))" 2>/dev/null || echo "auto")

# Manual compact: always allow
if [ "$TRIGGER" = "manual" ]; then
    exit 0
fi

# Auto compact: check if session-state.md was updated recently
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
STATE_FILE="$PROJECT_DIR/.claude/memory/session-state.md"
COUNTER_FILE="/tmp/.claude-precompact-gate"

# Safety: only block once
if [ -f "$COUNTER_FILE" ]; then
    rm -f "$COUNTER_FILE"
    exit 0
fi

# If no memory directory, skip
if [ ! -d "$PROJECT_DIR/.claude/memory" ]; then
    exit 0
fi

# Check if session-state.md was updated in the last 15 minutes
if [ -f "$STATE_FILE" ]; then
    if find "$STATE_FILE" -mmin -15 -print -quit | grep -q .; then
        exit 0
    fi
fi

# Block compact and force save
echo 1 > "$COUNTER_FILE"
echo "[PreCompact] Context 即将被压缩。请立即保存当前工作状态到 .claude/memory/session-state.md，包括：已完成的任务、未完成的任务、关键决策、遇到的问题、下一步行动。保存完成后会话将继续。" >&2
exit 2
