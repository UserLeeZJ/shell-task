// examples/logger-example/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

// 自定义日志实现
type CustomLogger struct {
	// 是否启用调试日志
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

// 创建新的自定义日志记录器
func NewCustomLogger(debugEnabled bool) *CustomLogger {
	return &CustomLogger{debugEnabled: debugEnabled}
}

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("启动日志示例...")

	// 创建自定义日志记录器
	logger := NewCustomLogger(true)

	// 创建一个任务，使用自定义日志记录器
	t := task.New(
		task.WithName("日志示例任务"),
		task.WithJob(func(ctx context.Context) error {
			fmt.Println("执行任务...")
			// 模拟一个错误
			return fmt.Errorf("模拟错误")
		}),
		task.WithLogger(logger),
		task.WithMaxRuns(1),
		task.WithErrorHandler(func(err error) {
			fmt.Printf("处理错误: %v\n", err)
		}),
	)

	// 启动任务
	t.Run()
	log.Println("任务已启动...")

	// 等待任务完成
	time.Sleep(2 * time.Second)

	// 使用兼容旧版本的日志函数
	oldStyleLogger := task.NewFuncLogger(func(format string, args ...any) {
		log.Printf("[旧风格] "+format, args...)
	})

	// 创建另一个任务，使用旧风格的日志记录器
	t2 := task.New(
		task.WithName("旧风格日志任务"),
		task.WithJob(func(ctx context.Context) error {
			fmt.Println("执行任务...")
			return nil
		}),
		task.WithLogger(oldStyleLogger),
		task.WithMaxRuns(1),
	)

	// 启动任务
	t2.Run()
	log.Println("第二个任务已启动...")

	// 等待任务完成
	time.Sleep(2 * time.Second)

	log.Println("示例结束")
}
