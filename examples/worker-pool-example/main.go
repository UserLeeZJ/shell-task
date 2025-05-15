// examples/worker-pool-example/main.go
package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

// 创建一个模拟任务，根据优先级执行不同的工作
func createTask(name string, priority task.Priority, duration time.Duration) *task.Task {
	return task.New(
		task.WithName(name),
		task.WithPriority(priority),
		task.WithJob(func(ctx context.Context) error {
			log.Printf("开始执行任务: %s (优先级: %d, 预计耗时: %v)", name, priority, duration)

			// 模拟工作负载
			select {
			case <-ctx.Done():
				log.Printf("任务被取消: %s", name)
				return ctx.Err()
			case <-time.After(duration):
				log.Printf("任务完成: %s", name)
				return nil
			}
		}),
	)
}

func main() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("启动工作池示例...")

	// 创建一个自定义日志记录器
	logger := task.NewFuncLogger(func(format string, args ...any) {
		log.Printf("[WorkerPool] "+format, args...)
	})

	// 创建工作池，限制最多3个并发任务
	pool := task.NewWorkerPool(3, logger)
	pool.Start()

	// 创建一些任务，具有不同的优先级
	tasks := []*task.Task{
		createTask("低优先级任务-1", task.PriorityLow, 3*time.Second),
		createTask("普通优先级任务-1", task.PriorityNormal, 2*time.Second),
		createTask("高优先级任务-1", task.PriorityHigh, 1*time.Second),
		createTask("低优先级任务-2", task.PriorityLow, 2*time.Second),
		createTask("普通优先级任务-2", task.PriorityNormal, 3*time.Second),
		createTask("高优先级任务-2", task.PriorityHigh, 2*time.Second),
		createTask("低优先级任务-3", task.PriorityLow, 1*time.Second),
		createTask("普通优先级任务-3", task.PriorityNormal, 1*time.Second),
		createTask("高优先级任务-3", task.PriorityHigh, 3*time.Second),
	}

	// 提交任务到工作池
	for _, t := range tasks {
		pool.Submit(t)
		// 稍微延迟一下，便于观察
		time.Sleep(100 * time.Millisecond)
	}

	// 等待所有任务完成
	log.Println("等待所有任务完成...")
	time.Sleep(15 * time.Second)

	// 停止工作池
	pool.Stop()
	log.Println("工作池已停止，示例结束")
}
