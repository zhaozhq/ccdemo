# Git 工作树机制 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立 PreToolUse Hook 自动拦截机制，强制 `executing-plans` 在独立工作树中执行，并提供便捷脚本简化工作树创建。

**Architecture:** 通过 `PreToolUse` Hook 在 `executing-plans` 调用前检测工作树状态（GIT_DIR != GIT_COMMON），若不在工作树中则输出拦截提示并返回非零码；同时提供 `new-worktree.sh` 封装 `git worktree add` 流程。

**Tech Stack:** Bash, git worktree, Claude Code hooks

---

### Task 1: 创建 `pre-skill-use.sh` Hook 脚本

**Files:**
- Create: `.claude/hooks/pre-skill-use.sh`

**Description:** 该 Hook 在每次工具调用前执行。当检测到调用的是 `Skill` 工具且 skill 名称为 `executing-plans` 时，检查当前是否处于独立工作树中。若不在工作树中，输出拦截提示并阻止调用。

**注意：** Hook 脚本通过环境变量获取工具调用信息。Claude Code 的 `PreToolUse` hook 会将工具调用的 JSON 上下文写入标准输入或环境变量。我们的脚本采用保守策略：当无法确定工具类型时直接放行，避免误拦截。

- [ ] **Step 1: 创建 `.claude/hooks/pre-skill-use.sh`**

```bash
#!/bin/bash
set -e

# ============================================================================
# PreToolUse Hook: Worktree Guard
# 拦截 executing-plans，强制在独立工作树中执行
# ============================================================================

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.."

# Helper: check if we are in a linked worktree (not submodule)
_is_in_worktree() {
    local git_dir git_common superproject
    git_dir=$(cd "$(git rev-parse --git-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 1
    git_common=$(cd "$(git rev-parse --git-common-dir 2>/dev/null)" && pwd -P 2>/dev/null) || return 1
    superproject=$(git rev-parse --show-superproject-working-tree 2>/dev/null || true)

    if [ -n "$superproject" ]; then
        # submodule: treat as NOT in worktree
        return 1
    fi

    if [ "$git_dir" != "$git_common" ]; then
        # linked worktree
        return 0
    fi

    return 1
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
IS_SKILL=$(echo "$TOOL_INPUT" | grep -q '"name"[[:space:]]*:[[:space:]]*"Skill"' && echo "yes" || echo "no")
IS_EXECUTING=$(echo "$TOOL_INPUT" | grep -q 'executing-plans' && echo "yes" || echo "no")

if [ "$IS_SKILL" != "yes" ] || [ "$IS_EXECUTING" != "yes" ]; then
    # Not executing-plans, allow pass-through
    exit 0
fi

# We are about to execute executing-plans. Check worktree status.
if _is_in_worktree; then
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
```

- [ ] **Step 2: 赋予执行权限**

```bash
chmod +x .claude/hooks/pre-skill-use.sh
```

- [ ] **Step 3: 本地测试脚本逻辑**

```bash
# Test 1: Verify detection logic when in main repo (should detect NOT in worktree)
bash -c 'source .claude/hooks/pre-skill-use.sh 2>/dev/null; _is_in_worktree && echo "in worktree" || echo "not in worktree"'
# Expected: "not in worktree"

# Test 2: Verify the script exits 0 when stdin is tty (nothing piped)
bash .claude/hooks/pre-skill-use.sh
# Expected: exit 0, no output
```

- [ ] **Step 4: Commit**

```bash
git add .claude/hooks/pre-skill-use.sh
git commit -m "feat: add PreToolUse hook to guard executing-plans in worktree

Blocks executing-plans when not inside a git worktree.
Provides instructions for creating a new worktree.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: 创建 `new-worktree.sh` 便捷脚本

**Files:**
- Create: `.claude/hooks/new-worktree.sh`

**Description:** 封装 `git worktree add` 流程，一键创建新分支和对应工作树。包含 `.gitignore` 验证、分支名校验、重复分支检测。

- [ ] **Step 1: 创建 `.claude/hooks/new-worktree.sh`**

```bash
#!/bin/bash
set -e

# ============================================================================
# Helper: Create a new git worktree with a new branch
# Usage: bash .claude/hooks/new-worktree.sh <branch-name>
# ============================================================================

# Resolve project root (hooks/ is under .claude/)
cd "$(dirname "$0")/../.."

BRANCH_NAME="${1:-}"

# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------

if [ -z "$BRANCH_NAME" ]; then
    echo "[worktree] Error: branch name is required" >&2
    echo "[worktree] Usage: bash .claude/hooks/new-worktree.sh <branch-name>" >&2
    exit 1
fi

# Validate branch name (no spaces, no special chars except - _ /)
if ! echo "$BRANCH_NAME" | grep -qE '^[a-zA-Z0-9._/-]+$'; then
    echo "[worktree] Error: invalid branch name '$BRANCH_NAME'" >&2
    echo "[worktree] Only alphanumeric, ., -, _, / are allowed" >&2
    exit 1
fi

# Check .gitignore contains .worktrees/
if ! git check-ignore -q .worktrees 2>/dev/null; then
    echo "[worktree] Error: .worktrees/ is not in .gitignore" >&2
    echo "[worktree] Please add '.worktrees/' to .gitignore first" >&2
    exit 1
fi

# Check if branch already exists locally
if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME"; then
    echo "[worktree] Error: branch '$BRANCH_NAME' already exists locally" >&2
    exit 1
fi

# Check if branch already exists remotely
if git ls-remote --heads origin "$BRANCH_NAME" | grep -q "$BRANCH_NAME"; then
    echo "[worktree] Error: branch '$BRANCH_NAME' already exists on remote" >&2
    exit 1
fi

# Check if worktree directory already exists
WORKTREE_PATH=".worktrees/$BRANCH_NAME"
if [ -e "$WORKTREE_PATH" ]; then
    echo "[worktree] Error: directory '$WORKTREE_PATH' already exists" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Create worktree
# ---------------------------------------------------------------------------

echo "[worktree] Creating worktree for branch: $BRANCH_NAME"
mkdir -p .worktrees
git worktree add "$WORKTREE_PATH" -b "$BRANCH_NAME"

echo "[worktree] ============================================================"
echo "[worktree]  Worktree created"
echo "[worktree] ============================================================"
echo "[worktree]"
echo "[worktree]  Branch:     $BRANCH_NAME"
echo "[worktree]  Path:       $(pwd)/$WORKTREE_PATH"
echo "[worktree]"
echo "[worktree]  Next steps:"
echo "[worktree]    cd $(pwd)/$WORKTREE_PATH"
echo "[worktree]    # Then invoke executing-plans or continue development"
echo "[worktree]"
echo "[worktree] ============================================================"
```

- [ ] **Step 2: 赋予执行权限**

```bash
chmod +x .claude/hooks/new-worktree.sh
```

- [ ] **Step 3: 测试脚本帮助和错误场景**

```bash
# Test 1: No arguments
bash .claude/hooks/new-worktree.sh
# Expected: Error "branch name is required", exit 1

# Test 2: Invalid branch name
bash .claude/hooks/new-worktree.sh "feature@123"
# Expected: Error "invalid branch name", exit 1

# Test 3: Branch already exists (use existing main branch as test)
bash .claude/hooks/new-worktree.sh main
# Expected: Error "branch 'main' already exists locally", exit 1
```

- [ ] **Step 4: Commit**

```bash
git add .claude/hooks/new-worktree.sh
git commit -m "feat: add new-worktree helper script

One-command creation of git worktree + branch.
Validates branch name, .gitignore, and duplicate checks.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: 更新 `.claude/settings.json` 注册 PreToolUse Hook

**Files:**
- Modify: `.claude/settings.json`

**Description:** 在现有 `PostToolUse` 和 `Stop` hooks 配置中，新增 `PreToolUse` 段，注册 `pre-skill-use.sh`。

- [ ] **Step 1: 读取当前 settings.json 确认结构**

```bash
cat .claude/settings.json
```

- [ ] **Step 2: 修改 `.claude/settings.json`**

在 `hooks` 对象下新增 `PreToolUse` 段。修改后的完整 `hooks` 部分应如下所示（保留原有的 `PostToolUse` 和 `Stop`）：

```json
{
  "permissions": {
    "allow": [
      "Bash(go mod *)",
      "Bash(go run *)",
      "Bash(go get *)",
      "Bash(go test *)",
      "Bash(go build *)",
      "Bash(go vet *)",
      "Bash(gofmt *)",
      "Bash(goimports *)",
      "Bash(git add *)",
      "Bash(git commit *)",
      "Bash(git -C *)",
      "Bash(git push *)",
      "Bash(git diff *)",
      "Bash(git status *)",
      "Bash(.claude/hooks/*.sh)"
    ],
    "additionalDirectories": [
      "/Users/alvin/CY.localized/aicode/ccdemo/.claude"
    ]
  },
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [],
        "command": "bash /Users/alvin/CY.localized/aicode/ccdemo/.claude/hooks/pre-skill-use.sh"
      }
    ],
    "PostToolUse": [
      {
        "hooks": [],
        "command": "bash /Users/alvin/CY.localized/aicode/ccdemo/.claude/hooks/post-edit.sh"
      }
    ],
    "Stop": [
      {
        "hooks": [],
        "command": "bash /Users/alvin/CY.localized/aicode/ccdemo/.claude/hooks/stop-check.sh"
      }
    ]
  }
}
```

- [ ] **Step 3: 验证 JSON 格式**

```bash
python3 -m json.tool .claude/settings.json > /dev/null && echo "JSON valid" || echo "JSON invalid"
# Expected: "JSON valid"
```

- [ ] **Step 4: Commit**

```bash
git add .claude/settings.json
git commit -m "chore: register PreToolUse hook for worktree guard

Adds pre-skill-use.sh to PreToolUse hooks in settings.json.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: 更新 `.gitignore` 忽略 `.worktrees/`

**Files:**
- Modify: `.gitignore`

**Description:** 将 `.worktrees/` 加入 `.gitignore`，防止工作树内容被误提交。

- [ ] **Step 1: 检查 `.gitignore` 是否已存在 `.worktrees/`**

```bash
grep -q "^\.worktrees/$" .gitignore 2>/dev/null && echo "already ignored" || echo "not ignored"
# If "already ignored", skip to Step 4
```

- [ ] **Step 2: 追加 `.worktrees/` 到 `.gitignore`**

```bash
echo "" >> .gitignore
echo "# Git worktrees" >> .gitignore
echo ".worktrees/" >> .gitignore
```

- [ ] **Step 3: 验证 `.gitignore` 生效**

```bash
mkdir -p .worktrees/test-ignore
echo "test" > .worktrees/test-ignore/file.txt
git check-ignore -q .worktrees/test-ignore/file.txt && echo "ignored correctly" || echo "NOT ignored"
rm -rf .worktrees/test-ignore
# Expected: "ignored correctly"
```

- [ ] **Step 4: Commit**

```bash
git add .gitignore
git commit -m "chore: ignore .worktrees/ directory in git

Prevents worktree contents from being accidentally committed.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: 端到端验证

**Files:**
- (no new files)

**Description:** 在 main 分支上测试 Hook 拦截逻辑，然后创建一个真实工作树验证流程贯通。

- [ ] **Step 1: 确认当前在 main 分支且不在工作树中**

```bash
git branch --show-current
# Expected: main

cd "$(git rev-parse --git-dir)" && pwd -P
cd "$(git rev-parse --git-common-dir)" && pwd -P
# Expected: two identical paths (not in worktree)
```

- [ ] **Step 2: 测试 `new-worktree.sh` 创建真实工作树**

```bash
bash .claude/hooks/new-worktree.sh test-worktree-guard
# Expected: success, creates .worktrees/test-worktree-guard/
```

- [ ] **Step 3: 验证工作树中的 git 状态**

```bash
cd .worktrees/test-worktree-guard
git branch --show-current
# Expected: test-worktree-guard

cd "$(git rev-parse --git-dir)" && pwd -P
cd "$(git rev-parse --git-common-dir)" && pwd -P
# Expected: two different paths (in worktree)
```

- [ ] **Step 4: 清理测试工作树**

```bash
cd /Users/alvin/CY.localized/aicode/ccdemo
git worktree remove .worktrees/test-worktree-guard
rm -rf .worktrees/test-worktree-guard
git branch -D test-worktree-guard
# Verify cleanup:
ls .worktrees/ 2>/dev/null || echo "worktrees dir empty or removed"
```

- [ ] **Step 5: 更新项目记忆（可选但推荐）**

```bash
# 在工作树机制生效后，更新 MEMORY.md 添加工作树相关条目
```

更新 `.claude/memory/MEMORY.md`，追加：

```markdown
- [Git 工作树机制](git-worktree-mechanism.md) — 执行 executing-plans 前必须处于独立工作树，通过 PreToolUse Hook 强制拦截
```

创建 `.claude/memory/git-worktree-mechanism.md`：

```markdown
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
```

- [ ] **Step 6: Commit 记忆更新**

```bash
git add .claude/memory/
git commit -m "docs: add git worktree mechanism to project memory

Records the worktree isolation rule for future sessions.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Self-Review Checklist

**1. Spec coverage:**
- [x] PreToolUse Hook 检测拦截 — Task 1
- [x] 工作树状态检测逻辑（GIT_DIR != GIT_COMMON，排除 submodule）— Task 1 Step 1
- [x] 拦截提示输出 — Task 1 Step 1
- [x] `new-worktree.sh` 便捷脚本 — Task 2
- [x] `.gitignore` 处理 — Task 4
- [x] `settings.json` 注册 — Task 3
- [x] 验证测试 — Task 5

**2. Placeholder scan:**
- [x] 无 TBD / TODO / "implement later"
- [x] 所有步骤包含完整代码或命令
- [x] 无 "Similar to Task N" 引用

**3. Type consistency:**
- [x] 文件路径一致：`.claude/hooks/pre-skill-use.sh`、`.claude/hooks/new-worktree.sh`
- [x] 环境变量和函数名在脚本内部一致
