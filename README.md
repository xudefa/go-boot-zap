# go-boot-zap

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-zap)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-zap)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-zap/test.yml?branch=master)](https://github.com/xudefa/go-boot-zap/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-zap.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-zap) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-zap)](https://goreportcard.com/report/github.com/xudefa/go-boot-zap)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 Zap 日志集成模块。将 go.uber.org/zap 无缝集成到 go-boot 的 IoC 容器和自动配置体系中，提供高性能、结构化的日志记录能力。

> 设计理念：遵循 go-boot 的开发规范，通过函数式选项模式和自动配置实现零代码启动 Zap 日志服务。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │     go-boot-zap Starter       │
                    │  ┌─────────────────────────┐  │
                    │  │ ZapAdapter Bean         │  │
                    │  │ Logger Implementation   │  │
                    │  │ Output Writer           │  │
                    │  │ Encoder Config          │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [日志记录](#日志记录)
- [高级配置](#高级配置)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 Zap 集成模块
go get github.com/xudefa/go-boot-zap
```

### 最小示例

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/log"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-log-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProperty("zap.enabled", "true"),
        boot.WithProperty("zap.format", "console"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 启动应用（自动配置 Zap 日志）
    app.Start()

    // 获取日志实例并记录日志
    logger := app.Container().Get("zapLogger").(log.Logger)
    
    logger.Info("Application started", 
        "version", "1.0.0",
        "mode", "production")
    logger.Warn("Deprecated API usage detected")
    logger.Error("Failed to connect to database", 
        "error", "connection refused")

    // 等待终止信号
    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| Zap 集成 | 将 Zap SugaredLogger 适配到 go-boot `log.Logger` 接口 |
| 自动配置 | 通过 `zap.enabled=true` 自动注册日志 Bean |
| 多格式输出 | 支持 JSON 格式（生产）和 Console 格式（开发） |
| 级别控制 | 支持 Debug、Info、Warn、Error 等日志级别 |
| 调用者追踪 | 支持记录文件名和行号 |
| 文件输出 | 支持日志文件输出和追加模式 |
| 高性能 | Zap 提供零分配的结构化日志记录 |

## 日志记录

### 基本日志记录

```go
logger := app.Container().Get("zapLogger").(log.Logger)

// 不同级别的日志
logger.Debug("Debug message with details", "key", "value")
logger.Info("Information message", "user", "alice")
logger.Warn("Warning message", "retry", 3)
logger.Error("Error occurred", "error", err)
```

### 结构化日志

```go
logger.Info("User login",
    "user_id", 1001,
    "username", "alice",
    "ip", "192.168.1.100",
    "duration_ms", 150)
```

### 创建独立日志实例

```go
import "github.com/xudefa/go-boot-zap/zap"

// 创建自定义日志适配器
logger := zap.NewZapAdapter(
    zap.WithZapLevel(log.DebugLevel),
    zap.WithZapFormat("json"),
    zap.WithZapOutputPath("/var/log/app.log"),
    zap.WithZapAddCaller(true),
)

logger.Info("Custom logger initialized")
```

## 高级配置

### 输出格式

```go
// JSON 格式（适合生产环境日志收集）
logger := zap.NewZapAdapter(
    zap.WithZapFormat("json"),
)

// Console 格式（适合本地开发调试，带彩色输出）
logger := zap.NewZapAdapter(
    zap.WithZapFormat("console"),
)

// Text 格式（类似 Console）
logger := zap.NewZapAdapter(
    zap.WithZapFormat("text"),
)
```

### 文件输出

```go
// 写入日志文件
logger := zap.NewZapAdapter(
    zap.WithZapOutputPath("/var/log/myapp.log"),
    zap.WithZapLevel(log.InfoLevel),
)

// 或使用底层 WriteSyncer
import "go.uber.org/zap/zapcore"

file, _ := os.OpenFile("/var/log/myapp.log", 
    os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
logger := zap.NewZapAdapter(
    zap.WithZapOutput(zapcore.AddSync(file)),
)
```

### 调用者追踪

```go
logger := zap.NewZapAdapter(
    zap.WithZapAddCaller(true),
    zap.WithZapCallerSkip(1), // 多层包装时调整跳过帧数
)
```

### 自定义时间格式

```go
logger := zap.NewZapAdapter(
    zap.WithZapTimeFormat("2006-01-02"), // 仅日期
    zap.WithZapTimeFormat("15:04:05"),   // 仅时间
)
```

### 自定义编码器

```go
import "go.uber.org/zap/zapcore"

encoderConfig := zapcore.EncoderConfig{
    TimeKey:        "ts",
    LevelKey:       "level",
    NameKey:        "logger",
    CallerKey:      "caller",
    MessageKey:     "msg",
    StacktraceKey:  "stacktrace",
    LineEnding:     zapcore.DefaultLineEnding,
    EncodeLevel:    zapcore.LowercaseLevelEncoder,
    EncodeTime:     zapcore.ISO8601TimeEncoder,
    EncodeDuration: zapcore.SecondsDurationEncoder,
    EncodeCaller:   zapcore.ShortCallerEncoder,
}

logger := zap.NewZapAdapter(
    zap.WithZapEncoder(zapcore.NewJSONEncoder(encoderConfig)),
)
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `zap.enabled` | `false` | 是否启用 Zap 日志 |
| `zap.level` | `info` | 日志级别（debug/info/warn/error） |
| `zap.format` | `json` | 输出格式（json/console/text） |
| `zap.time-format` | `2006-01-02 15:04:05` | 时间格式 |
| `zap.add-caller` | `false` | 是否记录调用位置 |
| `zap.caller-skip` | `1` | 调用者跳过栈帧数 |
| `zap.output` | `` | 日志文件路径（空则输出到 stdout） |

### 示例配置

```yaml
# application.yml
zap:
  enabled: true
  level: info
  format: json
  time-format: "2006-01-02 15:04:05"
  add-caller: true
  caller-skip: 1
  output: /var/log/myapp.log
```

## 项目结构

```
go-boot-zap/
├── zap.go                  # Zap 日志适配器实现
├── autoconfig.go           # 自动配置注册
├── zap_test.go             # 单元测试
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。