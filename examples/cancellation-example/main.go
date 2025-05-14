// examples/cancellation-example/main.go
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

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("启动取消示例...")

	// 创建一个长时间运行的任务
	t := task.New(
		task.WithName("长时间任务"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("开始执行任务...")

			// 模拟长时间运行的任务，每秒检查一次是否被取消
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					log.Printf("任务在第 %d 秒被取消: %v", i, ctx.Err())
					return ctx.Err() // 返回取消错误
				case <-time.After(1 * time.Second):
					log.Printf("任务执行中: %d/10 秒", i)
				}
			}

			log.Println("任务完成")
			return nil
		}),
		task.WithStartupDelay(2*time.Second), // 延迟2秒启动
		task.WithLoggerFunc(log.Printf),
		task.WithErrorHandler(func(err error) {
			if err == context.Canceled {
				log.Println("任务被用户取消")
			} else {
				log.Printf("任务错误: %v", err)
			}
		}),
	)

	// 启动任务
	t.Run()
	log.Println("任务已启动，将在2秒后开始执行...")
	log.Println("按 Ctrl+C 取消任务")

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或任务完成
	select {
	case sig := <-sigCh:
		log.Printf("收到信号 %v，正在停止任务...", sig)
		t.Stop()                    // 停止任务
		time.Sleep(1 * time.Second) // 给任务一点时间来清理
	case <-time.After(8 * time.Second):
		log.Println("示例超时，正在停止任务...")
		t.Stop()
	}

	log.Println("示例结束")
}
