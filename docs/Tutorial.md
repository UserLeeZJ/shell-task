# Shell-Task 使用教程

本教程将指导您从基础到高级使用 Shell-Task 库，通过实际示例帮助您快速上手。

## 目录

- [安装](#安装)
- [基础用法](#基础用法)
  - [创建和运行简单任务](#创建和运行简单任务)
  - [定时重复执行](#定时重复执行)
  - [设置最大运行次数](#设置最大运行次数)
- [错误处理](#错误处理)
  - [使用错误处理器](#使用错误处理器)
  - [重试机制](#重试机制)
  - [恢复钩子](#恢复钩子)
- [高级用法](#高级用法)
  - [使用钩子函数](#使用钩子函数)
  - [自定义日志记录](#自定义日志记录)
  - [指标收集](#指标收集)
  - [超时控制](#超时控制)
  - [优雅取消](#优雅取消)
- [并发控制](#并发控制)
  - [使用工作池](#使用工作池)
  - [任务优先级](#任务优先级)
  - [资源限制](#资源限制)
- [实际应用](#实际应用)
  - [文件处理](#文件处理)
  - [网络请求](#网络请求)
  - [数据库操作](#数据库操作)

## 安装

使用 Go 模块安装 Shell-Task 库：

```bash
go get github.com/UserLeeZJ/shell-task
```

## 基础用法

### 创建和运行简单任务

以下是创建和运行简单任务的基本示例：

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
        task.WithName("简单任务"),
        task.WithJob(func(ctx context.Context) error {
            fmt.Println("执行任务...")
            return nil
        }),
    )

    // 启动任务
    t.Run()

    // 等待任务完成
    time.Sleep(1 * time.Second)

    fmt.Println("程序结束")
}
```

### 定时重复执行

使用 `WithRepeat` 选项可以让任务以固定间隔重复执行：

```go
t := task.New(
    task.WithName("定时任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行定时任务...")
        return nil
    }),
    task.WithRepeat(5*time.Second), // 每5秒执行一次
)

// 启动任务
t.Run()

// 等待一段时间
time.Sleep(30 * time.Second)

// 停止任务
t.Stop()
```

### 设置最大运行次数

使用 `WithMaxRuns` 选项可以限制任务的最大运行次数：

```go
t := task.New(
    task.WithName("有限次数任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }),
    task.WithRepeat(2*time.Second), // 每2秒执行一次
    task.WithMaxRuns(5),            // 最多执行5次
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(15 * time.Second)
```

## 错误处理

### 使用错误处理器

使用 `WithErrorHandler` 选项可以设置错误处理器，处理任务执行过程中的错误：

```go
t := task.New(
    task.WithName("可能出错的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 模拟错误
        return fmt.Errorf("任务执行失败")
    }),
    task.WithErrorHandler(func(err error) {
        log.Printf("处理错误: %v", err)
        // 可以在这里发送告警或执行其他错误处理逻辑
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(1 * time.Second)
```

### 重试机制

使用 `WithRetry` 选项可以设置任务失败后的重试次数：

```go
t := task.New(
    task.WithName("需要重试的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 模拟随机错误
        if time.Now().UnixNano()%2 == 0 {
            return fmt.Errorf("随机错误")
        }
        fmt.Println("任务成功")
        return nil
    }),
    task.WithRetry(3), // 失败后最多重试3次
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(5 * time.Second)
```

### 恢复钩子

使用 `WithRecover` 选项可以设置恢复钩子，捕获并处理任务执行过程中的 panic：

```go
t := task.New(
    task.WithName("可能会 panic 的任务"),
    task.WithJob(func(ctx context.Context) error {
        // 模拟 panic
        panic("意外错误")
        return nil
    }),
    task.WithRecover(func(r any) {
        log.Printf("从 panic 恢复: %v", r)
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(1 * time.Second)
```

## 高级用法

### 使用钩子函数

使用 `WithPreHook` 和 `WithPostHook` 选项可以设置任务执行前后的钩子函数：

```go
t := task.New(
    task.WithName("带钩子的任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }),
    task.WithPreHook(func() {
        fmt.Println("任务执行前...")
    }),
    task.WithPostHook(func() {
        fmt.Println("任务执行后...")
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(1 * time.Second)
```

### 自定义日志记录

使用 `WithLogger` 选项可以设置自定义日志记录器：

```go
// 创建自定义日志记录器
type CustomLogger struct{}

func (l *CustomLogger) Debug(format string, args ...any) {
    log.Printf("[DEBUG] "+format, args...)
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
t := task.New(
    task.WithName("带日志的任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }),
    task.WithLogger(logger),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(1 * time.Second)
```

也可以使用 `WithLoggerFunc` 选项，更简单地设置日志函数：

```go
t := task.New(
    task.WithName("带日志的任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }),
    task.WithLoggerFunc(func(format string, args ...any) {
        log.Printf("[TASK] "+format, args...)
    }),
)
```

### 指标收集

使用 `WithMetricCollector` 选项可以收集任务执行的指标：

```go
t := task.New(
    task.WithName("指标收集任务"),
    task.WithJob(func(ctx context.Context) error {
        fmt.Println("执行任务...")
        return nil
    }),
    task.WithMetricCollector(func(res task.JobResult) {
        log.Printf("任务 '%s' 耗时 %v, 成功: %t", 
            res.Name, res.Duration, res.Success)
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(1 * time.Second)
```

### 超时控制

使用 `WithTimeout` 选项可以设置任务的最大执行时间：

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

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(10 * time.Second)
```

### 优雅取消

任务可以通过检查上下文是否已取消来实现优雅取消：

```go
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

// 等待一段时间后取消任务
time.Sleep(5 * time.Second)
t.Stop()

// 等待任务完全停止
time.Sleep(1 * time.Second)
```

## 并发控制

### 使用工作池

使用工作池可以限制同时运行的任务数量：

```go
// 创建一个自定义日志记录器
logger := task.NewFuncLogger(func(format string, args ...any) {
    log.Printf("[WorkerPool] "+format, args...)
})

// 创建工作池，限制最多3个并发任务
pool := task.NewWorkerPool(3, logger)
pool.Start()

// 创建多个任务
for i := 1; i <= 10; i++ {
    id := i // 捕获变量
    t := task.New(
        task.WithName(fmt.Sprintf("任务-%d", id)),
        task.WithJob(func(ctx context.Context) error {
            log.Printf("执行任务-%d", id)
            // 模拟工作负载
            time.Sleep(2 * time.Second)
            log.Printf("完成任务-%d", id)
            return nil
        }),
    )
    
    // 提交任务到工作池
    pool.Submit(t)
}

// 等待所有任务完成
time.Sleep(10 * time.Second)

// 停止工作池
pool.Stop()
```

### 任务优先级

使用 `WithPriority` 选项可以设置任务的优先级：

```go
// 创建高优先级任务
highPriorityTask := task.New(
    task.WithName("高优先级任务"),
    task.WithJob(func(ctx context.Context) error {
        log.Println("执行高优先级任务")
        time.Sleep(2 * time.Second)
        return nil
    }),
    task.WithPriority(task.PriorityHigh),
)

// 创建低优先级任务
lowPriorityTask := task.New(
    task.WithName("低优先级任务"),
    task.WithJob(func(ctx context.Context) error {
        log.Println("执行低优先级任务")
        time.Sleep(2 * time.Second)
        return nil
    }),
    task.WithPriority(task.PriorityLow),
)

// 提交到工作池
pool.Submit(lowPriorityTask)
pool.Submit(highPriorityTask) // 会优先执行
```

### 资源限制

使用 `WithResourceLimits` 选项可以设置任务的资源限制：

```go
t := task.New(
    task.WithName("资源受限任务"),
    task.WithJob(func(ctx context.Context) error {
        log.Println("执行资源密集型任务")
        // 模拟资源密集型操作
        time.Sleep(3 * time.Second)
        return nil
    }),
    task.WithResourceLimits(task.ResourceLimits{
        MaxCPU:    50,           // 最多使用50%的CPU
        MaxMemory: 1024,         // 最多使用1GB内存
        MaxTime:   5*time.Second, // 最多运行5秒
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(10 * time.Second)
```

## 实际应用

### 文件处理

使用 Shell-Task 处理文件：

```go
t := task.New(
    task.WithName("文件处理"),
    task.WithJob(func(ctx context.Context) error {
        // 打开文件
        file, err := os.Open("data.txt")
        if err != nil {
            return err
        }
        defer file.Close()
        
        // 处理文件内容
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            // 检查是否已取消
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                // 处理一行
                line := scanner.Text()
                log.Println("处理行:", line)
            }
        }
        
        return scanner.Err()
    }),
    task.WithErrorHandler(func(err error) {
        log.Printf("文件处理错误: %v", err)
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(10 * time.Second)
```

### 网络请求

使用 Shell-Task 执行网络请求：

```go
t := task.New(
    task.WithName("网络请求"),
    task.WithJob(func(ctx context.Context) error {
        // 创建请求
        req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
        if err != nil {
            return err
        }
        
        // 发送请求
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        // 处理响应
        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("请求失败: %s", resp.Status)
        }
        
        // 读取响应内容
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return err
        }
        
        log.Printf("响应内容: %s", body)
        return nil
    }),
    task.WithTimeout(5*time.Second), // 设置5秒超时
    task.WithRetry(3),               // 失败后最多重试3次
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(10 * time.Second)
```

### 数据库操作

使用 Shell-Task 执行数据库操作：

```go
t := task.New(
    task.WithName("数据库操作"),
    task.WithJob(func(ctx context.Context) error {
        // 连接数据库
        db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
        if err != nil {
            return err
        }
        defer db.Close()
        
        // 设置上下文
        db.SetConnMaxLifetime(time.Minute * 3)
        db.SetMaxOpenConns(10)
        db.SetMaxIdleConns(10)
        
        // 执行查询
        rows, err := db.QueryContext(ctx, "SELECT id, name FROM users LIMIT 10")
        if err != nil {
            return err
        }
        defer rows.Close()
        
        // 处理结果
        for rows.Next() {
            // 检查是否已取消
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                // 处理一行
                var id int
                var name string
                if err := rows.Scan(&id, &name); err != nil {
                    return err
                }
                log.Printf("用户: %d, %s", id, name)
            }
        }
        
        return rows.Err()
    }),
    task.WithErrorHandler(func(err error) {
        log.Printf("数据库操作错误: %v", err)
    }),
)

// 启动任务
t.Run()

// 等待任务完成
time.Sleep(10 * time.Second)
```
