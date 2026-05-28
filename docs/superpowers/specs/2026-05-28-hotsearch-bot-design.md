# 热搜推送 Bot 设计文档

## 1. 概述

构建一个 Go 应用，提供：
- 多平台热搜聚合（微博、百度、知乎、抖音）
- MCP Server 暴露查询能力
- 内置定时任务每日推送 Top20 到 Telegram Bot
- 用户可通过 Telegram Bot 主动查询
- SQLite 持久化，仅保留当日数据

## 2. 架构

单体应用，单一进程内运行三个子系统：

```
┌─────────────────────────────────────────────────────┐
│                    Go Application                     │
├─────────────┬──────────────┬────────────────────────┤
│  MCP Server │ Cron Scheduler│   Telegram Bot Client  │
│  (stdio/sse)│              │   (Long Polling)       │
└──────┬──────┴──────┬───────┴───────────┬────────────┘
       │             │                   │
       ▼             ▼                   ▼
┌─────────────────────────────────────────────────────┐
│              HotSearch Aggregator                     │
│   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│   │  微博    │ │  百度    │ │  知乎    │ │  抖音    │  │
│   │ Fetcher │ │ Fetcher │ │ Fetcher │ │ Fetcher │  │
│   └─────────┘ └─────────┘ └─────────┘ └─────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │   SQLite (当日)  │
              │  hot_searches   │
              └─────────────────┘
```

## 3. 核心模块

| 模块 | 职责 | 路径 |
|------|------|------|
| MCP Server | 实现 MCP 协议，暴露查询 Tools | `pkg/mcp/` |
| Telegram Bot | 接收命令、推送消息 | `pkg/bot/` |
| Fetcher | 各平台热搜抓取 | `pkg/fetcher/` |
| Aggregator | 聚合、排序、去重 | `pkg/aggregator/` |
| Storage | SQLite 数据访问 | `pkg/storage/` |
| Scheduler | Cron 定时任务 | `pkg/scheduler/` |

## 4. 数据流

### 4.1 定时推送
```
Cron 触发 → 并行抓取 → 聚合排序 Top20 → 存入 SQLite → 推送到 Telegram
```

### 4.2 MCP 查询
```
MCP Client → MCP Server → 查询 SQLite → 返回结构化数据
```

### 4.3 Bot 交互
```
用户 /hot → Bot → 查询 SQLite → 回复消息
```

## 5. MCP Tools

- `get_hot_searches` — 获取聚合 Top20
- `get_hot_searches_by_platform` — 按平台查询
- `get_platforms` — 获取支持平台列表

## 6. 数据模型

```go
type HotSearch struct {
    ID        int64
    Title     string
    URL       string
    Platform  string // weibo, baidu, zhihu, douyin
    Rank      int
    Heat      int64
    Category  string
    CreatedAt time.Time
}
```

## 7. 配置（环境变量）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TELEGRAM_BOT_TOKEN` | Telegram Bot Token | 必填 |
| `TELEGRAM_CHAT_ID` | 推送目标 Chat ID | 必填 |
| `PUSH_CRON` | 推送 Cron 表达式 | `0 9 * * *` |
| `FETCH_INTERVAL` | 数据刷新间隔 | `30m` |
| `DB_PATH` | SQLite 路径 | `./data/hotsearch.db` |

## 8. 非功能性设计

- 抓取失败时保留上一次成功数据
- Telegram 推送失败记录日志并重试
- 每日凌晨清理过期数据（保留当日）
- 各平台抓取并行执行
