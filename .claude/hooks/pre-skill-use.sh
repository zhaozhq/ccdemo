#!/usr/bin/env bash
set -euo pipefail

# PreToolUse hook: guard executing-plans by enforcing worktree isolation.
# If the tool call is Skill with executing-plans and we are NOT inside a
# git worktree (excluding submodules), block the call.

# cd to project root first
cd "$(dirname "$0")/../.."

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_is_in_worktree() {
    local GIT_DIR GIT_COMMON SUPERPROJECT

    GIT_DIR=$(cd "$(git rev-parse --git-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 1
    GIT_COMMON=$(cd "$(git rev-parse --git-common-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 1
    SUPERPROJECT=$(git rev-parse --show-superproject-working-tree 2>/dev/null || true)

    if [ -n "$SUPERPROJECT" ]; then
        # submodule, treat as NOT in worktree
        return 1
    elif [ "$GIT_DIR" != "$GIT_COMMON" ]; then
        # linked worktree
        return 0
    fi
    return 1
}

_print_banner() {
    cat >&2 <<'BANNER'
[worktree-guard] ============================================================
[worktree-guard]  拦截: executing-plans 必须在独立工作树中执行
[worktree-guard] ============================================================
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
[worktree-guard] ============================================================
BANNER
}

# ---------------------------------------------------------------------------
# Conservative fallback: if stdin is a tty, empty, or unreadable, allow.
# ---------------------------------------------------------------------------

if [ -t 0 ]; then
    exit 0
fi

INPUT=$(cat) || true
if [ -z "${INPUT:-}" ]; then
    exit 0
fi

# ---------------------------------------------------------------------------
# Detect if this is a Skill call for executing-plans.
# Best-effort grep on JSON with possible whitespace variations.
# ---------------------------------------------------------------------------

# If we cannot determine tool type, allow pass-through.
if ! printf '%s\n' "$INPUT" | grep -qiE '"name"\s*:\s*"Skill"'; then
    exit 0
fi

if ! printf '%s\n' "$INPUT" | grep -qiF 'executing-plans'; then
    exit 0
fi

# ---------------------------------------------------------------------------
# We are executing-plans: enforce worktree isolation.
# ---------------------------------------------------------------------------

if _is_in_worktree; then
    exit 0
fi

_print_banner
exit 1
