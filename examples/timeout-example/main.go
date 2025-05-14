// examples/timeout-example/main.go
package main

import (
	"context"
	"log"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("启动超时示例...")

	// 创建一个会超时的任务
	t := task.New(
		task.WithName("超时任务"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("开始执行长时间任务...")
			
			// 模拟耗时操作
			select {
			case <-time.After(5 * time.Second): // 任务需要5秒完成
				log.Println("任务完成")
				return nil
			case <-ctx.Done():
				// 如果上下文被取消（超时或手动取消）
				log.Printf("任务被中断: %v", ctx.Err())
				return ctx.Err()
			}
		}),
		task.WithTimeout(2*time.Second), // 设置2秒超时
		task.WithErrorHandler(func(err error) {
			log.Printf("错误处理: %v", err)
		}),
	)

	// 启动任务
	t.Run()
	log.Println("任务已启动，等待结果...")

	// 等待任务完成
	time.Sleep(10 * time.Second)

	// 停止任务
	t.Stop()
	log.Println("示例结束")
}
