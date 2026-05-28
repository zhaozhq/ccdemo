# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A minimal Go module (`hello`) using Go 1.26.1. Entry point is [src/cmd/server/main.go](src/cmd/server/main.go).

## Common Commands

- **Run**: `go run ./src/cmd/server`
- **Build**: `go build -o bin/ccdemo ./src/cmd/server`
- **Tidy dependencies**: `go mod tidy`

## Project Permissions

The local `.claude/settings.json` auto-allows `go mod *` and `go run *` commands.

## Superpowers 工程方法论（强制执行）

本项目严格遵守 Superpowers 工程方法论。所有在本项目上的开发工作必须遵循以下原则：

### 1. Skill 优先原则
- **任何任务开始前，必须优先检查并调用适用的 Superpowers Skill**。
- 只要任务有 1% 的概率适用某个 Skill，就必须调用它。
- **Process Skill 优先于 Implementation Skill**：
  - 新增功能 / 重构 → 先调用 `brainstorming`，再调用 `writing-plans`，最后才是执行相关的 Skill。
  - 修复 Bug → 先调用 `systematic-debugging`，再执行修复。
- Rigid Skill（如 `test-driven-development`）必须严格遵守，不得跳过或简化流程。

### 2. 开发流程顺序
所有功能开发必须按以下顺序执行：
1. **Research & Brainstorming** - 调用 `brainstorming` Skill 进行需求分析和方案探索。
2. **Planning** - 调用 `writing-plans` Skill 制定详细实施计划，等待用户确认后再编码。
3. **Test-Driven Development** - 调用 `test-driven-development` Skill，严格执行：写测试（RED）→ 写实现（GREEN）→ 重构（IMPROVE）。
4. **Implementation** - 调用 `executing-plans` Skill 执行计划。
5. **Code Review** - 代码完成后，调用 `requesting-code-review` 或 `receiving-code-review` Skill 进行审查。
6. **Verification** - 调用 `verification-before-completion` Skill 验证改动确实生效。
7. **Finish Branch** - 调用 `finishing-a-development-branch` Skill 规范地提交和合并代码。

### 3. 并行与子代理
- **优先使用并行 Agent 执行独立任务**，调用 `dispatching-parallel-agents` Skill。
- 复杂任务应使用 `subagent-driven-development` 模式，通过子代理分担工作。

### 4. 指令优先级
当发生冲突时，优先级如下：
1. 用户明确指令（本文件、直接请求）
2. Superpowers Skill 的规范要求
3. 默认系统提示行为

### 5. 禁止行为
- 不允许以"这个任务很简单"为由跳过 Skill 调用。
- 不允许以"我记得这个 Skill"为由不调用而直接执行。
- 不允许在 Plan 未经用户确认前直接开始编码。
