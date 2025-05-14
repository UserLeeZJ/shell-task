package scheduler

import (
	"context"
	"testing"
	"time"
)

// BenchmarkTaskCreation 基准测试任务创建
func BenchmarkTaskCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				return nil
			}),
		)
	}
}

// BenchmarkTaskRun 基准测试任务运行
func BenchmarkTaskRun(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				return nil
			}),
		)
		task.Run()
		task.Stop()
	}
}

// BenchmarkTaskWithHooks 基准测试带钩子的任务
func BenchmarkTaskWithHooks(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				return nil
			}),
			WithPreHook(func() {}),
			WithPostHook(func() {}),
		)
		task.Run()
		task.Stop()
	}
}

// BenchmarkTaskWithLogger 基准测试带日志的任务
func BenchmarkTaskWithLogger(b *testing.B) {
	logger := NewFuncLogger(func(format string, args ...any) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				return nil
			}),
			WithLogger(logger),
		)
		task.Run()
		task.Stop()
	}
}

// BenchmarkTaskWithMetrics 基准测试带指标的任务
func BenchmarkTaskWithMetrics(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				return nil
			}),
			WithMetricCollector(func(res JobResult) {}),
		)
		task.Run()
		task.Stop()
	}
}

// BenchmarkPriorityQueueEnqueue 基准测试优先级队列入队
func BenchmarkPriorityQueueEnqueue(b *testing.B) {
	pq := NewPriorityQueue()
	task := NewTask(
		WithName("BenchmarkTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Enqueue(task)
	}
}

// BenchmarkPriorityQueueDequeue 基准测试优先级队列出队
func BenchmarkPriorityQueueDequeue(b *testing.B) {
	pq := NewPriorityQueue()
	task := NewTask(
		WithName("BenchmarkTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)
	// 预先入队足够多的任务
	for i := 0; i < b.N+1000; i++ {
		pq.Enqueue(task)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Dequeue()
	}
}

// BenchmarkWorkerPoolSubmit 基准测试工作池提交任务
func BenchmarkWorkerPoolSubmit(b *testing.B) {
	pool := NewWorkerPool(10, nil)
	pool.Start()
	defer pool.Stop()

	task := NewTask(
		WithName("BenchmarkTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(task)
	}
}

// BenchmarkWorkerPoolWithPriority 基准测试带优先级的工作池
func BenchmarkWorkerPoolWithPriority(b *testing.B) {
	pool := NewWorkerPool(10, nil)
	pool.Start()
	defer pool.Stop()

	// 创建不同优先级的任务
	lowTask := NewTask(
		WithName("LowTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPriority(PriorityLow),
	)

	highTask := NewTask(
		WithName("HighTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPriority(PriorityHigh),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			pool.Submit(lowTask)
		} else {
			pool.Submit(highTask)
		}
	}
}

// BenchmarkConcurrentTasks 基准测试并发任务
func BenchmarkConcurrentTasks(b *testing.B) {
	// 创建一个通道，用于等待所有任务完成
	done := make(chan struct{}, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := NewTask(
			WithName("BenchmarkTask"),
			WithJob(func(ctx context.Context) error {
				// 模拟一些工作
				time.Sleep(1 * time.Microsecond)
				done <- struct{}{}
				return nil
			}),
		)
		task.Run()
	}

	// 等待所有任务完成
	for i := 0; i < b.N; i++ {
		<-done
	}
}
