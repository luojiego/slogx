# Slogx

Slogx 是一个基于 Go 1.21+ 内置的 `slog` 包封装的结构化日志库。它提供了简单易用的接口，同时具备日志轮转、环境感知和灵活配置等特性。

[English](README.md)

## 特性

- 基于 Go 1.21+ 的 `slog` 包，支持结构化日志
- 自动日志轮转（基于 lumberjack）
- 支持 JSON 和文本两种输出格式
- 自动使用程序名作为日志文件名
- 环境感知（测试/生产环境自动配置）
- 支持通过环境变量配置
- 支持同时输出到文件和控制台
- 支持动态调整日志级别（通过系统信号）
- 支持添加额外字段（With 方法）

## 安装

```bash
go get github.com/luojiedev/slogx
```

## 快速开始

```go
package main

import "github.com/luojiedev/slogx"

func main() {
    // 直接使用包级别的函数
    slogx.Info("应用启动")
    slogx.Debug("调试信息")
    slogx.Error("发生错误", "error", err)

    // 使用 With 添加额外字段
    logger := slogx.With("module", "user-service")
    logger.Info("用户登录", "userId", 123)
}
```

## 配置说明

### 默认配置

- 日志文件位置：`./logs/<程序名>.log`
- 日志级别：Debug
- 输出格式：文本格式
- 单个日志文件大小：50MB
- 保留日志文件数：100
- 日志保留天数：30天

### 环境变量配置

可以通过以下环境变量调整配置：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| LOG_MAX_SIZE | 单个日志文件大小上限(MB) | 50 |
| LOG_MAX_BACKUPS | 保留的日志文件数量 | 100 |
| LOG_MAX_AGE | 日志文件保留天数 | 30 |
| GO_ENV | 运行环境(production/prod表示生产环境) | - |

### 环境相关行为

测试环境（默认）：
- 同时输出到控制台和文件
- 不压缩旧日志文件

生产环境（GO_ENV=production/prod）：
- 只输出到文件
- 自动压缩旧日志文件

## 自定义配置

如果需要自定义配置，可以使用 `NewLogger` 函数：

```go
logger := slogx.NewLogger(slogx.Config{
    Level:      "debug",
    Format:     "json",
    Filename:   "custom.log",
    MaxSize:    100,    // MB
    MaxBackups: 10,     // 文件个数
    MaxAge:     7,      // 天数
    Compress:   true,   // 是否压缩
    Stdout:     true,   // 是否输出到控制台
})

// 设置为默认logger（可选）
slogx.SetDefaultLogger(logger)
```

## 动态调整日志级别

支持通过系统信号动态调整日志级别：

- `SIGHUP`: 设置为 Debug 级别
- `SIGUSR1`: 设置为 Info 级别
- `SIGUSR2`: 设置为 Warn 级别

示例（Unix/Linux）：
```bash
# 调整为 Debug 级别
kill -HUP <pid>

# 调整为 Info 级别
kill -USR1 <pid>

# 调整为 Warn 级别
kill -USR2 <pid>
```

## 依赖

- Go 1.21+
- gopkg.in/natefinch/lumberjack.v2

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

[luojiedev](https://github.com/luojiedev) 
