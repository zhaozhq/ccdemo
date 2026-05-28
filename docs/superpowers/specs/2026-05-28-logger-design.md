# 日志系统设计方案

## 背景

项目为一个极简 Go 模块，需要在不引入过重依赖的前提下，实现生产级日志功能。

## 需求

1. 能按天切割生成日志文件。
2. 错误日志文件独立。
3. 能按级别控制日志是否输出到文件。
4. 屏幕打印要输出所有级别日志。

## 方案选型

采用 **zap + 自定义 DailyRotateWriter** 方案：

- `uber-go/zap` 提供高性能结构化日志与级别过滤。
- 自定义 `DailyRotateWriter` 实现按天文件切割，无需引入额外切割库。
- 通过 `zapcore.NewTee` 实现多路输出（屏幕 + 普通文件 + 错误文件）。

## 目录结构

```
pkg/
  logger/
    config.go       // 配置结构体定义与级别解析
    rotate.go       // DailyRotateWriter 实现
    logger.go       // zap 初始化、全局 Logger 封装
    logger_test.go  // 单元测试与集成测试
```

## 配置设计

```go
type Config struct {
    Dir           string // 日志目录，默认 "./logs"
    AppFilename   string // 普通日志文件名前缀，默认 "app"
    ErrorFilename string // 错误日志文件名前缀，默认 "error"
    FileMinLevel  string // 写入普通文件的最低级别，默认 "info"
}
```

- `FileMinLevel` 支持：`debug`, `info`, `warn`, `error`, `fatal`, `panic`。
- 配置非法时降级为 `info`，不 panic。

## DailyRotateWriter

实现 `zapcore.WriteSyncer` 接口。

行为：
- 文件名格式：`{prefix}.{YYYY-MM-DD}.log`，例如 `app.2026-05-28.log`。
- 每次 `Write` 检查当前日期，跨天则关闭旧文件、打开新文件。
- 使用 `sync.Mutex` 保证并发安全。
- `Close()` 安全关闭当前文件句柄。

## 多路输出设计

通过 `zapcore.NewTee` 合并三路 Core：

| Core | 目的地 | 编码 | 级别过滤 |
|------|--------|------|----------|
| 屏幕 | `os.Stdout` | `ConsoleEncoder` | `DebugLevel`（始终全量） |
| 普通文件 | `DailyRotateWriter(app)` | `JSONEncoder` | `>= Config.FileMinLevel` |
| 错误文件 | `DailyRotateWriter(error)` | `JSONEncoder` | `>= ErrorLevel` |

关键细节：
- 普通文件包含 Error 及以上日志；错误文件是独立副本，方便单独排查。
- 屏幕使用 `ConsoleEncoder` 带颜色，便于开发调试。

## 输出效果示例

屏幕输出（全量、带颜色）：

```
2026-05-28T14:32:01.123+0800	DEBUG	user trace	{uid: 123}
2026-05-28T14:32:01.124+0800	INFO 	user login	{uid: 123}
2026-05-28T14:32:01.125+0800	ERROR	db connection failed	{error: connection refused}
```

普通文件 `app.2026-05-28.log`（Info 及以上）：

```json
{"t":"2026-05-28T14:32:01.124+0800","l":"info","m":"user login","uid":"123"}
{"t":"2026-05-28T14:32:01.125+0800","l":"error","m":"db connection failed","error":"connection refused"}
```

错误文件 `error.2026-05-28.log`（Error 及以上）：

```json
{"t":"2026-05-28T14:32:01.125+0800","l":"error","m":"db connection failed","error":"connection refused"}
```

## 存储优化

为降低日志存储占用，文件 JSON 采用精简策略：

- 键名缩写：`t`（time）、`l`（level）、`m`（msg），减少重复字段长度。
- 禁用 `caller`、`stacktrace`、`function` 等冗余字段。
- 时间格式使用紧凑的 RFC3339 毫秒精度，不额外填充。
- 屏幕输出保留完整键名和换行可读性，文件输出优先紧凑。

## 对外 API

```go
func Init(cfg Config) error
func Sync() error
func Debug(msg string, fields ...zap.Field)
func Info(msg string, fields ...zap.Field)
func Warn(msg string, fields ...zap.Field)
func Error(msg string, fields ...zap.Field)
func Fatal(msg string, fields ...zap.Field)
```

使用示例：

```go
func main() {
    logger.Init(logger.Config{
        Dir:          "./logs",
        FileMinLevel: "info",
    })
    defer logger.Sync()

    logger.Debug("this prints to screen only")
    logger.Info("user login", zap.String("uid", "123"))
    logger.Error("db connection failed", zap.Error(err))
}
```

## 错误处理与降级策略

- 目录自动创建：若 `./logs` 不存在，自动 `os.MkdirAll`。
- 文件打开失败降级：若磁盘故障导致无法创建日志文件，Logger 退化为仅屏幕输出，不中断业务。
- `Sync()` 忽略非致命错误，避免程序退出时 panic。

## 测试策略

| 测试 | 内容 |
|------|------|
| `TestDailyRotateWriter` | 验证同一天写入同一文件，跨天自动切到新文件 |
| `TestLoggerInit` | 验证不同级别日志是否正确路由到屏幕/普通文件/错误文件 |
| `TestConfigValidation` | 验证非法 `FileMinLevel` 降级为 `info` |
