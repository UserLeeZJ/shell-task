# Shell-Task 最佳实践

本文档提供了使用 Shell-Task 库的最佳实践和建议，帮助您更有效地使用该库。

## 目录

- [任务设计](#任务设计)
- [错误处理](#错误处理)
- [资源管理](#资源管理)
- [日志记录](#日志记录)
- [并发控制](#并发控制)
- [性能优化](#性能优化)
- [测试](#测试)

## 任务设计

### 保持任务简单

每个任务应该专注于单一职责，避免在一个任务中做太多事情。这样可以提高代码的可读性和可维护性，也便于测试和调试。

```go
// 好的做法：任务专注于单一职责
task.New(
    task.WithName("处理文件"),
    task.WithJob(func(ctx context.Context) error {
        return processFile("data.txt")
    }),
)

// 不好的做法：任务做太多事情
task.New(
    task.WithName("处理文件并发送邮件"),
    task.WithJob(func(ctx context.Context) error {
        if err := processFile("data.txt"); err != nil {
            return err
        }
        if err := sendEmail("user@example.com"); err != nil {
            return err
        }
        return updateDatabase()
    }),
)
```

### 使用上下文控制任务生命周期

任务函数接收一个上下文参数，应该定期检查上下文是否已取消，以便及时停止任务。

```go
task.New(
    task.WithName("长时间运行的任务"),
    task.WithJob(func(ctx context.Context) error {
        for i := 0; i < 100; i++ {
            // 检查上下文是否已取消
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                // 继续执行
            }
            
            // 执行一部分工作
            if err := doSomeWork(i); err != nil {
                return err
            }
        }
        return nil
    }),
)
```

### 合理设置超时

为任务设置合理的超时时间，避免任务运行时间过长导致资源浪费。

```go
task.New(
    task.WithName("网络请求"),
    task.WithJob(func(ctx context.Context) error {
        // 执行网络请求
        return makeHttpRequest(ctx)
    }),
    task.WithTimeout(5*time.Second), // 设置5秒超时
)
```

## 错误处理

### 使用错误处理器

为任务设置错误处理器，可以集中处理任务执行过程中的错误。

```go
task.New(
    task.WithName("数据处理"),
    task.WithJob(func(ctx context.Context) error {
        // 处理数据
        return processData()
    }),
    task.WithErrorHandler(func(err error) {
        log.Printf("数据处理错误: %v", err)
        // 可以在这里发送告警或执行其他错误处理逻辑
        sendAlert("数据处理失败", err)
    }),
)
```

### 使用重试机制

对于可能因临时原因失败的任务，使用重试机制可以提高任务的成功率。

```go
task.New(
    task.WithName("网络请求"),
    task.WithJob(func(ctx context.Context) error {
        // 执行网络请求
        return makeHttpRequest(ctx)
    }),
    task.WithRetry(3), // 失败后最多重试3次
)
```

### 使用恢复钩子

为任务设置恢复钩子，可以捕获并处理任务执行过程中的 panic。

```go
task.New(
    task.WithName("可能会 panic 的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 可能会 panic 的代码
        return riskyOperation()
    }),
    task.WithRecover(func(r any) {
        log.Printf("捕获到 panic: %v", r)
        // 可以在这里记录详细信息或执行其他恢复逻辑
    }),
)
```

## 资源管理

### 设置资源限制

为任务设置资源限制，避免单个任务占用过多资源。

```go
task.New(
    task.WithName("资源密集型任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行资源密集型操作
        return processLargeData()
    }),
    task.WithResourceLimits(task.ResourceLimits{
        MaxCPU:    50,           // 最多使用50%的CPU
        MaxMemory: 1024,         // 最多使用1GB内存
        MaxTime:   10*time.Minute, // 最多运行10分钟
    }),
)
```

### 使用工作池限制并发

使用工作池可以限制同时运行的任务数量，避免系统过载。

```go
// 创建一个工作池，限制最多5个并发任务
pool := task.NewWorkerPool(5, logger)
pool.Start()

// 创建并提交多个任务
for i := 0; i < 20; i++ {
    t := task.New(
        task.WithName(fmt.Sprintf("任务-%d", i)),
        task.WithJob(func(ctx context.Context) error {
            // 执行任务
            return nil
        }),
    )
    pool.Submit(t)
}

// 等待所有任务完成后停止工作池
time.Sleep(time.Minute)
pool.Stop()
```

## 日志记录

### 使用自定义日志记录器

使用自定义日志记录器可以更好地控制日志输出格式和级别。

```go
// 创建自定义日志记录器
type CustomLogger struct{}

func (l *CustomLogger) Debug(format string, args ...any) {
    // 在开发环境中输出调试日志
    if os.Getenv("ENV") == "dev" {
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
logger := &CustomLogger{}
task.New(
    task.WithName("带日志的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行任务
        return nil
    }),
    task.WithLogger(logger),
)
```

### 记录关键指标

使用指标收集器记录任务执行的关键指标，便于监控和分析。

```go
task.New(
    task.WithName("需要监控的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行任务
        return nil
    }),
    task.WithMetricCollector(func(res task.JobResult) {
        // 记录任务执行时间
        log.Printf("任务 '%s' 耗时 %v, 成功: %t", res.Name, res.Duration, res.Success)
        
        // 可以将指标发送到监控系统
        sendMetric("task_duration", res.Duration.Seconds())
        sendMetric("task_success", boolToInt(res.Success))
    }),
)
```

## 并发控制

### 使用优先级队列

为任务设置优先级，确保重要的任务优先执行。

```go
// 创建高优先级任务
highPriorityTask := task.New(
    task.WithName("高优先级任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行重要任务
        return nil
    }),
    task.WithPriority(task.PriorityHigh),
)

// 创建低优先级任务
lowPriorityTask := task.New(
    task.WithName("低优先级任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行不太重要的任务
        return nil
    }),
    task.WithPriority(task.PriorityLow),
)

// 提交到工作池
pool.Submit(lowPriorityTask)
pool.Submit(highPriorityTask) // 会优先执行
```

## 性能优化

### 避免阻塞主线程

任务执行可能需要较长时间，避免在主线程中等待任务完成。

```go
// 启动任务
t := task.New(
    task.WithName("长时间运行的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 执行长时间运行的操作
        return nil
    }),
)
t.Run()

// 不要这样做：阻塞主线程等待任务完成
// time.Sleep(10 * time.Minute)

// 应该这样做：使用通道或其他机制等待任务完成
done := make(chan struct{})
go func() {
    // 等待任务完成的逻辑
    time.Sleep(time.Minute) // 示例：定期检查任务状态
    if t.GetRunCount() > 0 {
        close(done)
    }
}()

select {
case <-done:
    log.Println("任务完成")
case <-time.After(10 * time.Minute):
    log.Println("任务超时")
    t.Stop()
}
```

## 测试

### 编写单元测试

为任务编写单元测试，确保任务在各种情况下都能正常工作。

```go
func TestMyTask(t *testing.T) {
    // 创建一个测试任务
    myTask := task.New(
        task.WithName("测试任务"),
        task.WithJob(func(ctx context.Context) error {
            // 执行测试逻辑
            return nil
        }),
    )
    
    // 运行任务
    myTask.Run()
    
    // 等待任务完成
    time.Sleep(100 * time.Millisecond)
    
    // 验证任务执行结果
    if myTask.GetRunCount() != 1 {
        t.Errorf("期望运行次数为1，实际为%d", myTask.GetRunCount())
    }
}
```

### 模拟上下文取消

测试任务对上下文取消的响应。

```go
func TestTaskCancellation(t *testing.T) {
    // 创建一个可取消的上下文
    ctx, cancel := context.WithCancel(context.Background())
    
    // 创建一个测试任务
    executed := false
    canceled := false
    
    myTask := task.New(
        task.WithName("可取消的任务"),
        task.WithJob(func(taskCtx context.Context) error {
            executed = true
            
            // 等待取消
            select {
            case <-taskCtx.Done():
                canceled = true
                return taskCtx.Err()
            case <-time.After(5 * time.Second):
                return nil
            }
        }),
    )
    
    // 运行任务
    myTask.Run()
    
    // 取消上下文
    time.Sleep(100 * time.Millisecond) // 确保任务已开始执行
    cancel()
    
    // 等待任务响应取消
    time.Sleep(100 * time.Millisecond)
    
    // 验证任务被取消
    if !executed {
        t.Error("任务未执行")
    }
    if !canceled {
        t.Error("任务未响应取消")
    }
}
```
