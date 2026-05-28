# Git 工作树机制设计文档

**日期**: 2026-05-28
**状态**: 已批准

---

## 1. 背景

当前项目要求严格遵守 Superpowers 工程方法论，开发流程包括 Planning → TDD → Implementation 等阶段。在 Implementation 阶段通过 `executing-plans` Skill 执行计划。

为避免在 main 分支直接开发导致代码污染、冲突或不可回退的修改，需要建立强制的工作树隔离机制：每次执行实施计划时，必须在一个独立的工作树（worktree）中进行。

## 2. 目标

- **强制隔离**: 执行 `executing-plans` 前，必须确认当前处于独立工作树中
- **自动拦截**: 如果不在工作树中，自动阻止 `executing-plans` 调用并给出明确指引
- **降低门槛**: 提供便捷脚本，一键创建带分支的工作树
- **项目本地管理**: 工作树存放于项目本地 `.worktrees/` 目录

## 3. 方案概述

采用 **PreToolUse Hook 拦截 + 便捷脚本辅助** 的组合方案：

1. 新增 `PreToolUse` Hook `pre-skill-use.sh`，在 `executing-plans` 调用前检测工作树状态
2. 若未在工作树中，输出拦截提示并返回非零退出码阻止调用
3. 提供 `.claude/hooks/new-worktree.sh` 便捷脚本封装 `git worktree add` 流程
4. 更新 `.claude/settings.json` 注册新 Hook
5. `.worktrees/` 加入 `.gitignore`，避免误提交

## 4. 详细设计

### 4.1 Hook 设计: `pre-skill-use.sh`

**位置**: `.claude/hooks/pre-skill-use.sh`

**触发时机**: 任何 `PreToolUse` 事件（工具调用前）

**逻辑流程**:

```
1. 读取环境变量/标准输入，判断当前调用的工具类型
2. 如果不是 Skill 工具，直接放行（exit 0）
3. 如果是 Skill 工具，判断 skill 名称是否为 executing-plans
4. 如果不是 executing-plans，直接放行
5. 如果是 executing-plans：
   a. 检测当前是否在工作树中（GIT_DIR != GIT_COMMON，且非 submodule）
   b. 如果已在工作树中，放行
   c. 如果不在工作树中：
      - 输出拦截提示（红色警告）
      - 输出创建工作树的命令指引
      - 返回非零退出码阻止调用
```

**检测工作树状态的命令**:

```bash
GIT_DIR=$(cd "$(git rev-parse --git-dir)" 2>/dev/null && pwd -P)
GIT_COMMON=$(cd "$(git rev-parse --git-common-dir)" 2>/dev/null && pwd -P)
SUPERPROJECT=$(git rev-parse --show-superproject-working-tree 2>/dev/null)

if [ -n "$SUPERPROJECT" ]; then
    # submodule，按非工作树处理
    IS_WORKTREE=false
elif [ "$GIT_DIR" != "$GIT_COMMON" ]; then
    # 已在工作树中
    IS_WORKTREE=true
else
    # 不在工作树中
    IS_WORKTREE=false
fi
```

**拦截提示示例**:

```
[worktree-guard] ============================================================
[worktree-guard]  拦截: executing-plans 必须在独立工作树中执行
[worktree-guard] ============================================================
[worktree-guard]
[worktree-guard] 当前在 main 分支的主工作区，直接执行计划会污染主分支。
[worktree-guard]
[worktree-guard] 请按以下步骤创建工作树：
[worktree-guard]
[worktree-guard]   1. 创建新分支和工作树：
[worktree-guard]      bash .claude/hooks/new-worktree.sh <branch-name>
[worktree-guard]
[worktree-guard]   2. 在新工作树中重新调用 executing-plans
[worktree-guard]
[worktree-guard] ============================================================
```

### 4.2 settings.json 配置更新

在现有 `PostToolUse` 和 `Stop` hooks 基础上，新增 `PreToolUse`：

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [],
        "command": "bash /Users/alvin/CY.localized/aicode/ccdemo/.claude/hooks/pre-skill-use.sh"
      }
    ],
    "PostToolUse": [...],
    "Stop": [...]
  }
}
```

### 4.3 工作树存放规范

**位置**: 项目根目录下的 `.worktrees/` 文件夹

**命名规则**: `.worktrees/<branch-name>/`

**gitignore 处理**: `.worktrees/` 必须加入 `.gitignore`

```bash
# .gitignore
.worktrees/
```

**安全验证**（创建前必须执行）:

```bash
git check-ignore -q .worktrees 2>/dev/null || echo "ERROR: .worktrees 未加入 .gitignore"
```

### 4.4 便捷脚本: `new-worktree.sh`

**位置**: `.claude/hooks/new-worktree.sh`

**功能**: 一键创建新分支 + 工作树

**用法**:

```bash
bash .claude/hooks/new-worktree.sh <branch-name>
```

**行为**:

1. 验证 `.worktrees/` 是否已加入 `.gitignore`
2. 验证分支名是否合法（非空、无特殊字符）
3. 验证分支是否已存在（本地或远程）
4. 执行 `git worktree add .worktrees/<branch-name> -b <branch-name>`
5. 输出成功信息和工作树路径

**示例输出**:

```
[worktree] 创建工作树: feature-x
[worktree] 路径: /Users/alvin/CY.localized/aicode/ccdemo/.worktrees/feature-x
[worktree] 分支: feature-x
[worktree] 完成。请在新工作树中继续开发。
```

## 5. 错误处理

| 场景 | 处理方式 |
|------|---------|
| Hook 检测不到工具信息 | 保守放行，避免误拦截 |
| `.worktrees/` 未加入 `.gitignore` | `new-worktree.sh` 报错并退出，要求先修复 |
| 分支名已存在 | `new-worktree.sh` 报错，提示使用其他名称 |
| `git worktree add` 失败 | 输出原始错误信息，退出码透传 |
| 用户在工作树中执行其他 skill | 不拦截，仅拦截 `executing-plans` |

## 6. 测试验证

验证步骤：

1. **主分支测试**: 在 main 分支执行任意计划，确认 `executing-plans` 被拦截
2. **工作树测试**: 创建新工作树后执行计划，确认正常放行
3. **脚本测试**: 验证 `new-worktree.sh` 能正确创建分支和工作树
4. **gitignore 测试**: 确认 `.worktrees/` 内容不被 git 追踪

## 7. 注意事项

- Hook 脚本返回非零退出码会阻止工具执行，但不应影响其他正常工具调用
- 仅拦截 `executing-plans`，其他 Skill（如 `brainstorming`、`writing-plans`）在主分支执行是允许的
- 如果用户已在工作树中（`GIT_DIR != GIT_COMMON`），无论执行什么都不拦截
- `.worktrees/` 必须可靠地加入 `.gitignore`，否则工作树内容会被误提交

## 8. 相关文件

| 文件 | 作用 |
|------|------|
| `.claude/hooks/pre-skill-use.sh` | PreToolUse Hook，检测并拦截 |
| `.claude/hooks/new-worktree.sh` | 便捷脚本，创建工作树 |
| `.claude/settings.json` | Hook 注册配置 |
| `.gitignore` | 忽略 `.worktrees/` 目录 |
