# ccdemo

一个最小化的 Go 模块示例项目。

## 项目简介

本项目展示了 Go 模块的基础结构，包含一个简单的 `Hello, World!` 命令行程序。

## 环境要求

- Go 1.26.1+

## 快速开始

```bash
# 运行程序
go run ./src/cmd/server

# 构建可执行文件
go build -o bin/ccdemo ./src/cmd/server

# 整理依赖
go mod tidy
```

## 项目结构

```
.
├── src/
│   └── cmd/
│   │   └── server/
│   │       └── main.go  # 程序入口
│   └── pkg/         # 库代码
├── go.mod           # 模块定义
├── CLAUDE.md        # 项目说明（Claude Code）
└── .claude/         # Claude Code 配置
```

## 许可证

MIT
