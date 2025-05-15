# Shell-Task

一个简单而灵活的Go语言任务调度系统，专注于可靠性和易用性。Shell-Task 提供了一种优雅的方式来创建、调度和管理定时任务，支持错误重试、超时控制、钩子函数等高级特性。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://github.com/UserLeeZJ/shell-task)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue?style=flat-square)](LICENSE)

## 特性

- **简单易用的API**：函数式选项模式，链式配置
- **定时任务**：支持固定间隔重复执行
- **错误处理**：内置重试机制和自定义错误处理
- **超时控制**：为任务设置最大执行时间
- **优雅停止**：支持平滑关闭正在运行的任务
- **生命周期钩子**：前置钩子、后置钩子和恢复钩子
- **指标收集**：内置任务执行指标收集功能
- **自定义日志**：可配置的日志记录器
- **异常恢复**：自动从 panic 中恢复并继续执行
- **无外部依赖**：仅使用 Go 标准库

## 安装

要求 Go 1.21 或更高版本：

```bash
go get github.com/UserLeeZJ/shell-task
```

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    task "github.com/UserLeeZJ/shell-task"
)

func main() {
    // 创建一个简单的任务
    t := task.New(
        task.WithName("示例任务"),
        task.WithJob(func(ctx context.Context) error {
            fmt.Println("执行任务...")
            return nil
        }),
        task.WithRepeat(5*time.Second),
    )

    // 启动任务
    t.Run()

    // 等待一段时间
    time.Sleep(30 * time.Second)

    // 停止任务
    t.Stop()
}
```

### 错误重试和处理

```go
t := task.New(
    task.WithName("重试任务"),
    task.WithJob(func(ctx context.Context) error {
        // 模拟可能失败的操作
        return fmt.Errorf("操作失败")
    }),
    task.WithRetry(3),  // 失败后重试3次
    task.WithErrorHandler(func(err error) {
        log.Printf("处理错误: %v", err)
    }),
)
```

### 使用钩子函数

```go
t := task.New(
    task.WithName("带钩子的任务"),
    task.WithJob(func(ctx context.Context) error {
        return nil
    }),
    task.WithPreHook(func() {
        log.Println("任务执行前...")
    }),
    task.WithPostHook(func() {
        log.Println("任务执行后...")
    }),
)
```

## 配置选项

Shell-Task 提供了多种配置选项，可以根据需要组合使用：

| 选项 | 描述 |
|------|------|
| `WithName` | 设置任务名称 |
| `WithJob` | 设置任务主体函数 |
| `WithTimeout` | 设置任务超时时间 |
| `WithRepeat` | 设置任务以固定间隔重复执行 |
| `WithMaxRuns` | 设置最大运行次数 |
| `WithRetry` | 设置失败后重试次数 |
| `WithLogger` | 自定义日志记录器 |
| `WithRecover` | 添加 panic 恢复钩子 |
| `WithStartupDelay` | 设置延迟启动时间 |
| `WithPreHook` | 添加执行前钩子 |
| `WithPostHook` | 添加执行后钩子 |
| `WithErrorHandler` | 设置错误处理器 |
| `WithCancelOnFailure` | 设置失败时是否取消任务 |
| `WithMetricCollector` | 设置指标收集器 |

## 高级用法

### 使用任务构建器

```go
// 使用任务构建器创建和配置任务
task1 := task.NewTaskBuilder("构建器示例").
    WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }).
    WithTimeout(5 * time.Second).
    WithRepeat(30 * time.Second).
    WithRetry(3).
    WithPriority(task.PriorityHigh).
    WithErrorHandler(func(err error) {
        log.Printf("处理错误: %v", err)
    }).
    Run() // 构建并立即运行任务

// 或者先构建再运行
task2 := task.NewTaskBuilder("另一个任务").
    WithJob(func(ctx context.Context) error {
        return nil
    }).
    Build() // 仅构建不运行

task2.Run() // 稍后运行
```

### 指标收集

```go
t := task.New(
    task.WithName("指标收集任务"),
    task.WithJob(func(ctx context.Context) error {
        // 任务逻辑
        return nil
    }),
    task.WithMetricCollector(func(res task.JobResult) {
        log.Printf("任务 '%s' 耗时 %v, 成功: %t",
            res.Name, res.Duration, res.Success)
    }),
)
```

### Panic 恢复

```go
t := task.New(
    task.WithName("恢复任务"),
    task.WithJob(func(ctx context.Context) error {
        // 可能会 panic 的代码
        panic("意外错误")
        return nil
    }),
    task.WithRecover(func(r interface{}) {
        log.Printf("从 panic 恢复: %v", r)
    }),
)
```

### 任务超时控制

```go
t := task.New(
    task.WithName("超时任务"),
    task.WithJob(func(ctx context.Context) error {
        // 长时间运行的任务
        select {
        case <-time.After(5 * time.Second): // 任务需要5秒完成
            return nil
        case <-ctx.Done():
            // 如果上下文被取消（超时或手动取消）
            return ctx.Err()
        }
    }),
    task.WithTimeout(2*time.Second), // 设置2秒超时
    task.WithErrorHandler(func(err error) {
        log.Printf("处理超时错误: %v", err)
    }),
)
```

### 优雅取消任务

```go
// 创建任务
t := task.New(
    task.WithName("可取消任务"),
    task.WithJob(func(ctx context.Context) error {
        // 在任务中定期检查取消信号
        for i := 1; i <= 10; i++ {
            select {
            case <-ctx.Done():
                log.Printf("任务被取消: %v", ctx.Err())
                return ctx.Err()
            case <-time.After(1 * time.Second):
                log.Printf("任务执行中: %d/10", i)
            }
        }
        return nil
    }),
    task.WithErrorHandler(func(err error) {
        if err == context.Canceled {
            log.Println("任务被用户取消")
        }
    }),
)

// 启动任务
t.Run()

// 在需要时停止任务
t.Stop()
```

### 上下文传递

```go
// 使用上下文传递数据
task1 := task.TaskWithContextMap("数据准备", func(ctx context.Context, data map[string]interface{}) error {
    // 将数据存储到上下文中
    data["result"] = "准备好的数据"
    data["timestamp"] = time.Now().Format(time.RFC3339)
    return nil
})

// 或者使用构建器 API
task2 := task.NewTaskBuilder("数据处理").
    WithMapContextJob(func(ctx context.Context, data map[string]interface{}) error {
        // 获取上一个任务传递的数据
        result := data["result"]
        log.Printf("处理数据: %v", result)
        return nil
    }).
    Build()

// 创建任务链，自动传递上下文数据
tasks := task.ChainTasks(task1, task2)

// 运行所有任务
for _, t := range tasks {
    t.Run()
}
```

### 自定义日志记录器

Shell-Task 支持自定义日志记录器，可以实现 `Logger` 接口来控制不同级别的日志输出：

```go
// 自定义日志实现
type CustomLogger struct {
    debugEnabled bool
}

func (l *CustomLogger) Debug(format string, args ...any) {
    if l.debugEnabled {
        log.Printf("[DEBUG] "+format, args...)
    }
}

func (l *CustomLogger) Info(format string, args ...any) {
    log.Printf("[INFO] "+format, args...)
}

func (l *CustomLogger) Warn(format string, args ...any) {
    log.Printf("[WARN] "+format, args...)
}

func (l *CustomLogger) Error(format string, args ...any) {
    log.Printf("[ERROR] "+format, args...)
}

// 使用自定义日志记录器
logger := &CustomLogger{debugEnabled: true}
t := task.New(
    task.WithName("日志示例"),
    task.WithJob(func(ctx context.Context) error {
        return nil
    }),
    task.WithLogger(logger),
)
```

也可以使用兼容旧版本的日志函数：

```go
// 使用函数式日志适配器
oldStyleLogger := task.NewFuncLogger(func(format string, args ...any) {
    log.Printf("[LOG] "+format, args...)
})

t := task.New(
    task.WithName("旧风格日志"),
    task.WithJob(func(ctx context.Context) error {
        return nil
    }),
    task.WithLogger(oldStyleLogger),
)
```

更多示例请查看 [examples](./examples) 目录。

## 构建和测试

```bash
# 构建
make build

# 运行测试
make test

# 运行示例
make run
```

## 可能的改进方向

- 添加 cron 表达式支持，实现更灵活的调度
- 添加持久化支持，允许任务状态保存和恢复
- 提供 HTTP API 接口，便于远程管理和监控
- 实现分布式任务调度，支持多节点协作
- 添加更多单元测试和基准测试

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

[Apache License 2.0](LICENSE)