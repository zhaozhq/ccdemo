---
name: git-worktree-mechanism
description: 强制在独立工作树中执行 executing-plans，通过 PreToolUse Hook 拦截
metadata:
  type: project
---

## Git 工作树强制隔离机制

- 执行 `executing-plans` 前，必须处于独立的 git worktree 中
- 通过 `.claude/hooks/pre-skill-use.sh` (PreToolUse Hook) 自动检测和拦截
- 如果不在工作树中，Hook 会阻止调用并提示创建工作树
- 使用 `.claude/hooks/new-worktree.sh <branch-name>` 一键创建
- 工作树存放于 `.worktrees/<branch-name>/`

**Why:** 防止在 main 分支直接开发导致代码污染和不可回退的修改。

**How to apply:** 每次要开始 Implementation 阶段时，先执行 `bash .claude/hooks/new-worktree.sh <branch-name>`，然后在新工作树中继续。
