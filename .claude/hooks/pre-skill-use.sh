#!/usr/bin/env bash
set -e

# ============================================================================
# PreToolUse Hook: Worktree Guard
# 拦截 executing-plans，强制在独立工作树中执行
# ============================================================================

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.." || {
    echo "[worktree-guard] ERROR: cannot resolve project root" >&2
    exit 0  # conservative allow on unexpected failure
}

# Helper: print banner
_banner() {
    echo "[worktree-guard] ============================================================" >&2
    echo "[worktree-guard]  $1" >&2
    echo "[worktree-guard] ============================================================" >&2
}

# ---------------------------------------------------------------------------
# Main logic
# ---------------------------------------------------------------------------

# Try to detect the tool being invoked from environment/context.
# Claude Code PreToolUse hooks receive tool call info via stdin as JSON.
# We read stdin and check for Skill tool with executing-plans.
TOOL_INPUT=""
if [ -t 0 ]; then
    # stdin is a tty, nothing piped — cannot detect, allow pass-through
    exit 0
else
    TOOL_INPUT=$(cat)
fi

# If we have no input, allow pass-through
if [ -z "$TOOL_INPUT" ]; then
    exit 0
fi

# Check if this is a Skill tool call for executing-plans
# The JSON structure from Claude Code hooks contains tool name and arguments.
# We do a best-effort grep check; if ambiguous, we allow pass-through.
# Tightened patterns to reduce false positives.
# Skill tool: look for '"name" : "Skill"' followed by comma or closing brace.
IS_SKILL=$(printf '%s\n' "$TOOL_INPUT" | grep -qE '"name"[[:space:]]*:[[:space:]]*"Skill"[[:space:]]*[,}]' && echo "yes" || echo "no")
# executing-plans: look for it as a standalone word to avoid matching inside strings.
IS_EXECUTING=$(printf '%s\n' "$TOOL_INPUT" | grep -qE '\bexecuting-plans\b' && echo "yes" || echo "no")

if [ "$IS_SKILL" != "yes" ] || [ "$IS_EXECUTING" != "yes" ]; then
    # Not executing-plans, allow pass-through
    exit 0
fi

# We are about to execute executing-plans. Check worktree status.
_is_in_worktree() {
    local git_dir git_common superproject
    git_dir=$(cd "$(git rev-parse --git-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 2
    git_common=$(cd "$(git rev-parse --git-common-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 2
    superproject=$(git rev-parse --show-superproject-working-tree 2>/dev/null || true)

    if [ -n "$superproject" ]; then
        # submodule: treat as NOT in worktree
        return 1
    elif [ "$git_dir" != "$git_common" ]; then
        # linked worktree
        return 0
    fi
    return 1
}

status=0
_is_in_worktree || status=$?
if [ "$status" -eq 2 ]; then
    # Cannot determine worktree status — conservative allow
    exit 0
fi
if [ "$status" -eq 0 ]; then
    # In worktree, allow
    exit 0
fi

# Not in a worktree — block the call
_banner "拦截: executing-plans 必须在独立工作树中执行"

cat >&2 <<'MSG'
[worktree-guard]
[worktree-guard] 当前在主工作区，直接执行计划会污染主分支。
[worktree-guard]
[worktree-guard] 请按以下步骤创建工作树：
[worktree-guard]
[worktree-guard]   1. 创建新分支和工作树：
[worktree-guard]      bash .claude/hooks/new-worktree.sh <branch-name>
[worktree-guard]
[worktree-guard]   2. 在新工作树中重新调用 executing-plans
[worktree-guard]
MSG

exit 1
