# Code Review Standards (Project-Level)

> 本文件覆盖并扩展全局 `~/.claude/rules/code-review.md`，专门针对 **Go 项目** 的审查流程。

## Go 审查强制规范

### 审查前置条件
- 所有变更必须通过 `go build ./...`
- 所有变更必须通过 `go test ./...`
- 所有变更必须通过 `go vet ./...`
- 新代码覆盖率 >= 80%

### 审查时必须引用的规范
**go-reviewer 代理在审查 Go 代码时，必须对照以下规范逐项检查：**
- [go-coding-standards.md](./go-coding-standards.md) — Go 语言专项规范（命名、错误处理、并发、资源管理等）

### 审查输出格式
审查结果必须按以下结构输出：

```
## 审查概要
- 文件数：X
- 问题总数：Y（CRITICAL: a, HIGH: b, MEDIUM: c, LOW: d）

## 按规范检查项

### 格式与结构
- [ ] PASS / FAIL — 具体说明

### 错误处理
- [ ] PASS / FAIL — 具体说明

### 并发安全
- [ ] PASS / FAIL — 具体说明

### 资源管理
- [ ] PASS / FAIL — 具体说明

### 测试覆盖
- [ ] PASS / FAIL — 具体说明

### 安全
- [ ] PASS / FAIL — 具体说明

## 详细问题列表

### [Level] 文件:行号 — 问题描述
- 规范依据：go-coding-standards.md 第 X 节
- 修复建议：...
```

### 严重级别定义

| Level | 含义 | 处理要求 |
|-------|------|----------|
| **CRITICAL** | 安全漏洞、数据竞态、panic 风险、资源泄漏 | **阻塞合入** — 必须修复 |
| **HIGH** | 逻辑错误、未处理错误、并发 Bug、测试缺失 | **警告** — 应当修复 |
| **MEDIUM** | 可维护性问题（大函数、嵌套深、命名不清）| **建议** — 考虑修复 |
| **LOW** | 风格问题、注释缺失、微小优化 | **备注** — 可选修复 |

## 审查触发条件（Go 项目补充）

除全局规则中的通用触发条件外，Go 项目还必须在以下场景触发审查：

- 新增或修改 `go` 关键字（goroutine 生命周期审查）
- 新增或修改 `chan` / `select` / `sync` 相关代码（并发安全审查）
- 新增或修改数据库/SQL/文件操作（注入与资源泄漏审查）
- 新增或修改错误返回值的处理路径（错误处理审查）
- 引入新的第三方依赖（安全性与许可证审查）

## 审查代理配置

```
Agent: go-reviewer
必做：
  1. 运行 git diff 获取变更范围
  2. 运行 go test ./... 验证测试
  3. 运行 go vet ./... 静态分析
  4. 逐条检查 go-coding-standards.md 中的 Checklist
  5. 对 goroutine/channel/mutex 使用进行专项审查
  6. 输出按严重级别分类的问题列表

禁止：
  - 仅检查格式问题而忽略并发/错误处理
  - 仅检查实现代码而忽略测试代码
```

## 自动化工具链

建议在 CI 中集成以下工具，作为人工审查的前置关卡：

| 工具 | 用途 | 命令 |
|------|------|------|
| gofmt | 格式化检查 | `gofmt -l .` |
| goimports | Import 管理与分组 | `goimports -l .` |
| go vet | 静态分析 | `go vet ./...` |
| golangci-lint | 综合 Lint | `golangci-lint run` |
| go test -race | 数据竞争检测 | `go test -race ./...` |
| go test -cover | 覆盖率检查 | `go test -coverprofile=coverage.out ./...` |

## 与全局规则的集成

本文件与全局规则的关系：
- [全局 code-review.md](../../../../../.claude/rules/code-review.md) — 通用审查框架（触发条件、严重级别、Approval Criteria）
- [全局 coding-style.md](../../../../../.claude/rules/coding-style.md) — 通用编码原则（KISS、DRY、Immutability）
- **本文件** — Go 项目审查流程与代理配置
- [go-coding-standards.md](./go-coding-standards.md) — Go 语言专项规范

审查时代理必须同时参考以上所有文件。
