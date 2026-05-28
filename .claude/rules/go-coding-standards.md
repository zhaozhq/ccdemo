# Go Coding Standards

> 本项目所有 Go 代码必须遵循以下规范。代码审查时，审查代理（go-reviewer）必须对照本规范逐项检查。

---

## 1. 格式与风格

### 1.1 强制使用自动化格式化工具
- **gofmt**：所有 `.go` 文件必须通过 `gofmt` 格式化
- **goimports**：自动管理 import 分组与排序（标准库 / 第三方 / 项目内部）
- **golines**：行长度超过 120 字符时必须换行

```go
// CORRECT: import 分组
import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"

    "ccdemo/src/pkg/logger"
)
```

### 1.2 代码行规
- 单行不超过 **120** 字符
- 函数体不超过 **50** 行（空行与注释不计）
- 文件不超过 **800** 行
- 缩进使用 **Tab**（gofmt 默认），但在审查中可接受 4 空格项目约定

---

## 2. 命名规范

### 2.1 包名
- 全小写，无下划线，无驼峰
- 包名应简洁且与其目录名一致
- 避免 `common`, `util`, `helper` 等空洞包名，按职责命名

```go
// CORRECT
package fetcher
package aggregator

// WRONG
package commonUtils
package helper_lib
```

### 2.2 变量与常量
- 变量：`camelCase`，在短作用域内可缩写（如 `i`, `id`），长作用域必须描述性
- 常量：`UPPER_SNAKE_CASE` 或 `camelCase`（若导出不使用常量组则首字母大写）
- 布尔值：前缀用 `is`, `has`, `should`, `can`, `enable`

```go
const maxRetryCount = 3
var isDebugMode = false
```

### 2.3 接口命名
- 单方法接口：`方法名 + er`（Go 惯用法）
- 多方法接口：描述性名词

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type DataStore interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, val []byte) error
}
```

### 2.4 结构体与接收者
- 结构体：`PascalCase`（导出）或 `camelCase`（包内）
- 接收者命名：1-2 个字母，反映类型名，全项目统一（要么全用指针，要么全用值，视可变性而定）

```go
type UserService struct { /* ... */ }

func (us *UserService) Fetch(ctx context.Context) error { /* ... */ }
```

### 2.5 测试函数
- 测试：`Test + 被测对象 + _ + 场景/条件`
- 子测试：使用 `t.Run` 描述行为

```go
func TestUserService_Fetch_WithInvalidID_ReturnsError(t *testing.T) { /* ... */ }
```

---

## 3. 错误处理

### 3.1 绝不忽略错误
所有返回 `error` 的调用必须处理。如果确实无需处理，用空白标识符时必须注释原因。

```go
// CORRECT
if err := doSomething(); err != nil {
    return fmt.Errorf("do something: %w", err)
}

// WRONG: 忽略错误
doSomething()
```

### 3.2 Error Wrapping
- 跨层传播时必须包装上下文，使用 `%w` 保留原始错误
-  sentinel errors（预定义错误）用 `errors.New`，比较用 `errors.Is`
-  检查特定错误类型用 `errors.As`

```go
var ErrNotFound = errors.New("not found")

func Find() error {
    if err := db.Query(); err != nil {
        return fmt.Errorf("find user: %w", err)
    }
    return nil
}

// 调用方
if errors.Is(err, ErrNotFound) { /* ... */ }
```

### 3.3 禁止 panic
- 生产代码**严禁**使用 `panic`，除非 `init()` 或 `main()` 中的不可恢复启动失败
- 使用 `log.Fatal` 仅限 CLI 工具，库代码必须返回 `error`

---

## 4. 并发与同步

### 4.1 Goroutine 生命周期
- 每个 `go` 关键字启动的 goroutine 必须有明确的退出机制（`context.Context` 或 `done` channel）
- 禁止在循环中直接 `go func()` 捕获循环变量而不传参

```go
// CORRECT
for _, item := range items {
    item := item // 或作为参数传入
    go func(i Item) {
        process(i)
    }(item)
}

// WRONG: 闭包捕获循环变量
for _, item := range items {
    go func() {
        process(item) // 可能只处理到最后一个
    }()
}
```

### 4.2 Channel 使用
- 发送端关闭 channel，接收端不关闭
- 明确 channel 方向：函数参数中 `chan<- T`（只写）或 `<-chan T`（只读）
- 优先使用 `select` 处理多路复用，始终包含 `ctx.Done()` 分支

### 4.3 同步原语
- 优先使用 `sync.Mutex` / `RWMutex` 保护共享状态，而非 channel
- `sync.Once` 用于一次性初始化，`sync.Pool` 用于高频临时对象复用
- 使用 `go test -race` 检测数据竞争

### 4.4 Context 传播
- 所有阻塞/IO 操作必须接受 `context.Context` 并向下传播
- 禁止存储 `context.Context` 到结构体长期持有，仅作为函数第一个参数传递

```go
// CORRECT
func (s *Service) Fetch(ctx context.Context, id string) error { /* ... */ }

// WRONG
type Service struct {
    ctx context.Context // 不要这样做
}
```

---

## 5. 接口与组合

### 5.1 小接口原则
- 接口应该很小（1-3 个方法最佳），在**消费端**定义，而非实现端
- 不要提前设计接口，等真的有多个实现时再抽象

```go
// CORRECT: 在消费处定义
type Fetcher interface {
    Fetch(ctx context.Context) ([]Item, error)
}

func Process(f Fetcher) { /* ... */ }
```

### 5.2 嵌入与组合
- 优先使用组合（`has-a`）而非嵌入（`is-a`），除非确实需要接口转发
- 嵌入会暴露被嵌入类型的全部方法，增加 API 表面积

---

## 6. 资源管理

### 6.1 延迟关闭
- `defer` 必须在资源获取后立即注册，按 LIFO 顺序释放
- `io.Closer` 资源（文件、HTTP Body、DB Rows）必须关闭

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

### 6.2 HTTP 客户端
- 禁止在循环内重复创建 `http.Client`，应复用实例
- 总是读取并关闭 `resp.Body`，即使不处理响应内容

---

## 7. 性能与内存

### 7.1 减少内存分配
- 预分配 slice 容量：`make([]T, 0, estimatedLen)`
- 复用大型对象或 buffer，使用 `sync.Pool`
- 注意 `string` 与 `[]byte` 的转换会产生拷贝

### 7.2 字符串拼接
- 少量拼接：`+` 或 `fmt.Sprintf`
- 循环/大量拼接：使用 `strings.Builder`

```go
var b strings.Builder
b.Grow(estimatedLen)
for _, s := range items {
    b.WriteString(s)
}
result := b.String()
```

### 7.3 Map 与 Slice 零值检查
- 判断 slice/map 为空用 `len(v) == 0`，不直接与 `nil` 比较（除非语义确实需要区分 nil 与 empty）

---

## 8. 测试规范

### 8.1 表格驱动测试
- 所有单元测试优先使用表格驱动，覆盖正常、边界、异常路径

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    int
        want    int
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"divide by zero", 10, 0, 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Fatalf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### 8.2 Mock 与依赖注入
- 通过接口注入依赖，测试时使用 mock/fake 实现
- 不要为了达到测试目的而导出未导出的函数或变量

### 8.3 并行测试
- 无共享状态的测试使用 `t.Parallel()`
- 涉及共享资源（数据库、全局状态）的测试禁止并行

### 8.4 覆盖率
- 新代码行覆盖率不得低于 **80%**
- 使用 `go test -cover` 或 `go test -coverprofile=coverage.out`

---

## 9. 项目结构

遵循标准 Go 项目布局：

```
ccdemo/
├── cmd/                    # 可执行入口（main 包）
│   └── ccdemo/
│       └── main.go
├── internal/               # 私有应用代码，不允许外部导入
│   └── service/
├── pkg/                    # 可被外部项目导入的库代码
│   ├── logger/
│   ├── fetcher/
│   └── aggregator/
├── api/                    # API 定义（protobuf, openapi）
├── configs/                # 配置文件模板
├── scripts/                # 构建、部署脚本
├── docs/                   # 文档
├── go.mod
└── go.sum
```

- `internal/` 放置业务逻辑，防止被外部依赖
- `pkg/` 放置通用、可复用的库代码
- 不要在根目录放置过多 `.go` 文件

---

## 10. 日志与监控

### 10.1 结构化日志
- 使用结构化日志库（如 `zap`, `zerolog`, `slog`）
- 关键字段必须提取为独立 field，不拼接进消息字符串

```go
// CORRECT
logger.Info("fetch completed",
    zap.String("source", "weibo"),
    zap.Int("count", len(items)),
    zap.Duration("elapsed", elapsed),
)

// WRONG
logger.Infof("fetch completed from %s, got %d items in %v", source, len, elapsed)
```

### 10.2 日志级别
- `Debug`：开发调试信息
- `Info`：正常业务流程节点
- `Warn`：可恢复异常、降级、重试
- `Error`：业务错误、需人工关注
- `Fatal/Panic`：**禁止**在库代码中使用

---

## 11. 安全

### 11.1 输入校验
- 所有外部输入（HTTP 参数、文件、环境变量、配置）必须校验长度、类型、范围
- 使用 `html/template` 输出 HTML，防止 XSS

### 11.2 注入防护
- SQL 必须使用参数化查询，禁止字符串拼接
- 命令执行禁止直接拼接用户输入到 `exec.Command`
- 文件操作校验路径，防止目录遍历

```go
// CORRECT
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", userID)

// WRONG
rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))
```

### 11.3 敏感信息
- 密钥、Token、密码禁止硬编码，必须通过环境变量或配置中心注入
- 日志中禁止输出密码、Token、完整 SQL 中的敏感字段

---

## 12. 代码审查检查表（Checklist）

审查代理必须对照以下清单逐项确认：

### 格式与结构
- [ ] 通过 `gofmt` / `goimports` 检查
- [ ] 函数长度 < 50 行
- [ ] 文件长度 < 800 行
- [ ] 无深层嵌套（> 4 层）

### 命名
- [ ] 包名全小写、有意义
- [ ] 导出符号有文档注释
- [ ] 布尔变量有 `is/has/should/can` 前缀
- [ ] 测试函数名描述行为

### 错误处理
- [ ] 无忽略的错误
- [ ] 跨层错误使用 `fmt.Errorf("...: %w", err)` 包装
- [ ] 无 `panic`（除非 `init`/`main` 启动失败）
- [ ] 库代码返回 `error` 而非 `log.Fatal`

### 并发
- [ ] Goroutine 有退出机制（context/channel）
- [ ] 无循环变量闭包捕获问题
- [ ] 共享状态有同步保护（mutex 或 channel）
- [ ] 函数签名中 channel 有方向限定

### 资源与性能
- [ ] `defer Close()` 紧跟资源获取后
- [ ] HTTP Body / DB Rows 已关闭
- [ ] Slice 预分配容量（在已知大小时）
- [ ] 大量字符串拼接使用 `strings.Builder`

### 测试
- [ ] 新增功能有单元测试
- [ ] 使用表格驱动测试覆盖边界情况
- [ ] 测试覆盖率 >= 80%
- [ ] 并发测试使用 `go test -race`

### 安全
- [ ] 无硬编码密钥/密码
- [ ] SQL/命令/文件操作为参数化或已校验
- [ ] 日志无敏感信息泄漏

### 架构
- [ ] 接口定义在消费端（小接口）
- [ ] 优先组合而非嵌入
- [ ] `context.Context` 作为函数首参数向下传播
