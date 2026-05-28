#!/usr/bin/env bash
set -e

# ============================================================================
# PreToolUse Hook: Worktree Guard
# 拦截 executing-plans，强制在独立工作树中执行
# ============================================================================

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.."

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
IS_SKILL=$(echo "$TOOL_INPUT" | grep -q '"name"[[:space:]]*:[[:space:]]*"Skill"' && echo "yes" || echo "no")
IS_EXECUTING=$(echo "$TOOL_INPUT" | grep -q 'executing-plans' && echo "yes" || echo "no")

if [ "$IS_SKILL" != "yes" ] || [ "$IS_EXECUTING" != "yes" ]; then
    # Not executing-plans, allow pass-through
    exit 0
fi

# We are about to execute executing-plans. Check worktree status.
GIT_DIR=$(cd "$(git rev-parse --git-dir 2>/dev/null)" && pwd -P 2>/dev/null) || true
GIT_COMMON=$(cd "$(git rev-parse --git-common-dir 2>/dev/null)" && pwd -P 2>/dev/null) || true
SUPERPROJECT=$(git rev-parse --show-superproject-working-tree 2>/dev/null || true)

IS_WORKTREE=false
if [ -n "$SUPERPROJECT" ]; then
    # submodule: treat as NOT in worktree
    IS_WORKTREE=false
elif [ "$GIT_DIR" != "$GIT_COMMON" ]; then
    # linked worktree
    IS_WORKTREE=true
fi

if [ "$IS_WORKTREE" = "true" ]; then
    # Already in a worktree, allow
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
