# Shell-Task API 文档

本文档详细介绍了 Shell-Task 库的 API，包括核心类型、函数和选项。

## 目录

- [核心类型](#核心类型)
  - [Task](#task)
  - [Job](#job)
  - [JobResult](#jobresult)
  - [Logger](#logger)
  - [Priority](#priority)
  - [WorkerPool](#workerpool)
- [函数](#函数)
  - [New](#new)
  - [NewFuncLogger](#newfunclogger)
  - [NewWorkerPool](#newworkerpool)
- [选项](#选项)
  - [基本选项](#基本选项)
  - [日志选项](#日志选项)
  - [钩子选项](#钩子选项)
  - [错误处理选项](#错误处理选项)
  - [优先级选项](#优先级选项)

## 核心类型

### Task

`Task` 表示一个可配置的任务，是 Shell-Task 库的核心类型。

**方法：**

- `Run()`: 启动任务
- `Stop()`: 停止任务
- `GetRunCount() int`: 获取当前运行次数

### Job

`Job` 是任务函数的类型定义：

```go
type Job func(ctx context.Context) error
```

任务函数接收一个上下文参数，可以用于取消任务，并返回一个错误。

### JobResult

`JobResult` 表示任务执行的结果：

```go
type JobResult struct {
    Name     string
    Duration time.Duration
    Success  bool
    Err      error
}
```

### Logger

`Logger` 是日志接口，支持不同级别的日志记录：

```go
type Logger interface {
    Debug(format string, args ...any)
    Info(format string, args ...any)
    Warn(format string, args ...any)
    Error(format string, args ...any)
}
```

### Priority

`Priority` 定义任务优先级：

```go
type Priority int

const (
    PriorityLow    Priority = 1
    PriorityNormal Priority = 5
    PriorityHigh   Priority = 10
)
```

### WorkerPool

`WorkerPool` 管理一组工作协程，限制并发执行的任务数量。

**方法：**

- `Start()`: 启动工作池
- `Stop()`: 停止工作池
- `Submit(task *Task)`: 提交任务到工作池

## 函数

### New

```go
func New(opts ...TaskOption) *Task
```

创建一个新的任务实例，可以通过选项配置任务。

### NewFuncLogger

```go
func NewFuncLogger(logFunc func(format string, args ...any)) Logger
```

创建一个新的函数式日志适配器，用于将单一日志函数转换为 Logger 接口。

### NewWorkerPool

```go
func NewWorkerPool(size int, logger Logger) *WorkerPool
```

创建一个新的工作池，限制并发执行的任务数量。

## 选项

### 基本选项

- `WithName(name string)`: 设置任务名称
- `WithJob(job Job)`: 设置任务主体函数
- `WithTimeout(timeout time.Duration)`: 设置任务超时时间
- `WithRepeat(interval time.Duration)`: 设置任务以固定间隔重复执行
- `WithMaxRuns(n int)`: 设置最大运行次数
- `WithRetry(n int)`: 设置失败后重试次数
- `WithStartupDelay(delay time.Duration)`: 设置延迟启动时间

### 日志选项

- `WithLogger(logger Logger)`: 设置自定义日志记录器
- `WithLoggerFunc(logFunc func(format string, args ...any))`: 使用函数作为日志记录器

### 钩子选项

- `WithPreHook(hook func())`: 添加执行前钩子
- `WithPostHook(hook func())`: 添加执行后钩子
- `WithRecover(hook func(any))`: 添加 panic 恢复钩子
- `WithMetricCollector(collector func(JobResult))`: 设置指标收集器

### 错误处理选项

- `WithErrorHandler(handler func(error))`: 设置错误处理器
- `WithCancelOnFailure(cancel bool)`: 设置失败时是否取消任务

### 优先级选项

- `WithPriority(priority Priority)`: 设置任务优先级
