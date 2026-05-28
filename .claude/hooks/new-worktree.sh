#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/../.."

BRANCH_NAME="$1"

if [ -z "$BRANCH_NAME" ]; then
    printf '%s\n' "[worktree] Error: branch name is required" >&2
    printf '%s\n' "Usage: $0 <branch-name>" >&2
    exit 1
fi

if ! printf '%s\n' "$BRANCH_NAME" | grep -qE '^[a-zA-Z0-9._/-]+$'; then
    printf '%s\n' "[worktree] Error: invalid branch name '$BRANCH_NAME'" >&2
    printf '%s\n' "Only alphanumeric, ., -, _, / are allowed" >&2
    exit 1
fi

if ! git check-ignore -q .worktrees/ 2>/dev/null; then
    printf '%s\n' "[worktree] Error: .worktrees/ is not in .gitignore" >&2
    printf '%s\n' "Please add '.worktrees/' to .gitignore first" >&2
    exit 1
fi

if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME"; then
    printf '%s\n' "[worktree] Error: branch '$BRANCH_NAME' already exists locally" >&2
    exit 1
fi

if git ls-remote --heads origin "$BRANCH_NAME" | grep -q "$BRANCH_NAME"; then
    printf '%s\n' "[worktree] Error: branch '$BRANCH_NAME' already exists on remote" >&2
    exit 1
fi

WORKTREE_DIR=".worktrees/$BRANCH_NAME"

if [ -e "$WORKTREE_DIR" ]; then
    printf '%s\n' "[worktree] Error: directory '$WORKTREE_DIR' already exists" >&2
    exit 1
fi

mkdir -p .worktrees

git worktree add "$WORKTREE_DIR" -b "$BRANCH_NAME"

printf '%s\n' "[worktree] Creating worktree for branch: $BRANCH_NAME"
printf '%s\n' "[worktree] ============================================================"
printf '%s\n' "[worktree]  Worktree created"
printf '%s\n' "[worktree] ============================================================"
printf '%s\n' "[worktree]"
printf '%s\n' "[worktree]  Branch:     $BRANCH_NAME"
printf '%s\n' "[worktree]  Path:       $(pwd)/$WORKTREE_DIR"
printf '%s\n' "[worktree]"
printf '%s\n' "[worktree]  Next steps:"
printf '%s\n' "[worktree]    cd $(pwd)/$WORKTREE_DIR"
printf '%s\n' "[worktree]    # Then invoke executing-plans or continue development"
printf '%s\n' "[worktree]"
printf '%s\n' "[worktree] ============================================================"
