// cmd/shelltask/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

// Version 由构建时的 -ldflags 设置
var Version = "dev"

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Shell Task 版本: %s", Version)

	// 创建一个简单的示例任务
	t := task.New(
		task.WithName("示例任务"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("执行任务...")
			// 这里可以放置实际的任务逻辑
			return nil
		}),
		task.WithRepeat(5*time.Second),
		task.WithLoggerFunc(log.Printf),
		task.WithMaxRuns(5), // 最多运行5次
	)

	// 启动任务
	t.Run()
	log.Println("任务已启动，将自动运行5次后退出")

	// 创建一个完成通道
	done := make(chan struct{})

	// 等待中断信号或任务完成
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("收到中断信号")
	case <-done:
		log.Println("任务已完成所有运行")
	case <-time.After(30 * time.Second):
		log.Println("超时，强制退出")
	}

	// 停止任务
	t.Stop()
	log.Println("任务已停止")
}
