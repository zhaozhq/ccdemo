---
name: superpowers-methodology
description: 本项目严格遵守 Superpowers 工程方法论，所有开发任务必须先调用对应 Skill，按流程执行
metadata:
  type: project
---

## Superpowers 工程方法论

本项目强制使用 Superpowers Skill 体系管理开发流程。

### 核心原则
- **任何任务开始前必须调用适用的 Superpowers Skill**，哪怕只有 1% 概率适用。
- **Process Skill > Implementation Skill**：先定流程，再执行。
- Rigid Skill（如 TDD）必须严格遵守，不得跳过。

### 标准开发流程
1. `brainstorming` — 需求分析与方案探索
2. `writing-plans` — 制定详细计划，等待用户确认
3. `test-driven-development` — 严格执行 RED → GREEN → IMPROVE
4. `executing-plans` — 按计划编码实现
5. `requesting-code-review` / `receiving-code-review` — 代码审查
6. `verification-before-completion` — 验证改动生效
7. `finishing-a-development-branch` — 规范提交合并

### 并行与子代理
- 独立任务优先使用 `dispatching-parallel-agents` 并行执行
- 复杂任务使用 `subagent-driven-development` 模式

### 禁止行为
- 不得以"任务简单"为由跳过 Skill
- 不得凭记忆直接执行而不调用 Skill
- Plan 未经确认不得直接编码

**Why:** 确保开发过程系统化、可复现、质量可控，避免凭直觉随意开发导致的质量问题和返工。

**How to apply:** 每次用户提出新需求或任务时，先判断适用哪些 Superpowers Skill，按优先级顺序调用，严格遵循 Skill 输出的流程。
