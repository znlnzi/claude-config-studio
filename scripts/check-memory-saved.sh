#!/bin/bash
# Stop hook gate: block Claude from stopping if session-state.md wasn't updated this session.
#
# Logic:
#   - Find session-state.md in the project's .claude/memory/ directory
#   - Check if it was modified within the last 10 minutes
#   - If not modified recently → exit 2 (block stop, force Claude to save)
#   - If modified → exit 0 (allow stop)
#
# Safety: counter file prevents infinite loops (max 2 retries)

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
STATE_FILE="$PROJECT_DIR/.claude/memory/session-state.md"
COUNTER_FILE="/tmp/.claude-memory-check-$$"

# Safety: if we've already asked twice, let it go
if [ -f "$COUNTER_FILE" ]; then
    count=$(cat "$COUNTER_FILE")
    if [ "$count" -ge 2 ]; then
        rm -f "$COUNTER_FILE"
        exit 0
    fi
    echo $((count + 1)) > "$COUNTER_FILE"
else
    echo 1 > "$COUNTER_FILE"
fi

# If no .claude/memory directory exists, skip check
if [ ! -d "$PROJECT_DIR/.claude/memory" ]; then
    rm -f "$COUNTER_FILE"
    exit 0
fi

# Check if session-state.md was modified in the last 10 minutes
if [ -f "$STATE_FILE" ]; then
    if find "$STATE_FILE" -mmin -10 -print -quit | grep -q .; then
        # Recently updated, allow stop
        rm -f "$COUNTER_FILE"
        exit 0
    fi
fi

# Not updated recently — block stop and tell Claude to save
echo "session-state.md 未在本次会话中更新。请先保存会话状态到 .claude/memory/session-state.md，记录当前进度和下一步行动，然后再结束。" >&2
exit 2
