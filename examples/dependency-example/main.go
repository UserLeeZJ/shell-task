// examples/dependency-example/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("任务依赖关系示例")

	// 示例1：基本依赖关系
	log.Println("\n=== 示例1：基本依赖关系 ===")
	basicDependencyExample()

	// 示例2：任务序列
	log.Println("\n=== 示例2：任务序列 ===")
	sequenceExample()

	// 示例3：并行任务
	log.Println("\n=== 示例3：并行任务 ===")
	parallelExample()

	// 示例4：复杂依赖图
	log.Println("\n=== 示例4：复杂依赖图 ===")
	complexDependencyExample()
}

// 示例1：基本依赖关系
func basicDependencyExample() {
	// 创建工作池
	pool := task.NewWorkerPool(2, nil)
	pool.Start()
	defer pool.Stop()

	// 创建第一个任务
	task1 := task.New(
		task.WithName("数据准备"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("任务1：准备数据")
			time.Sleep(500 * time.Millisecond) // 模拟工作
			log.Println("任务1：数据准备完成")
			return nil
		}),
	)

	// 创建第二个任务，依赖于第一个任务
	task2 := task.New(
		task.WithName("数据处理"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("任务2：处理数据")
			time.Sleep(300 * time.Millisecond) // 模拟工作
			log.Println("任务2：数据处理完成")
			return nil
		}),
	)

	// 设置依赖关系
	task2.DependsOn(task1)

	// 提交任务（顺序不重要，因为有依赖关系）
	pool.Submit(task2) // 先提交依赖任务
	pool.Submit(task1) // 后提交被依赖任务

	// 等待任务完成
	time.Sleep(1 * time.Second)
}

// 示例2：任务序列
func sequenceExample() {
	// 创建工作池
	pool := task.NewWorkerPool(2, nil)
	pool.Start()
	defer pool.Stop()

	// 创建一系列任务
	tasks := make([]*task.Task, 5)
	for i := 0; i < 5; i++ {
		// 使用局部变量避免闭包捕获问题
		index := i
		taskName := fmt.Sprintf("序列任务%d", index+1)
		tasks[index] = task.New(
			task.WithName(taskName),
			task.WithJob(func(ctx context.Context) error {
				// 从上下文中获取任务，然后获取任务名称
				currentTask := task.TaskFromContext(ctx)
				currentTaskName := currentTask.GetName()
				log.Printf("执行任务: %s", currentTaskName)
				time.Sleep(200 * time.Millisecond) // 模拟工作
				log.Printf("完成任务: %s", currentTaskName)
				return nil
			}),
		)
	}

	// 使用简化API创建任务序列
	sequenceTasks := task.Sequence(tasks...)

	// 提交所有任务
	for _, t := range sequenceTasks {
		pool.Submit(t)
	}

	// 等待任务完成
	time.Sleep(2 * time.Second)
}

// 示例3：并行任务
func parallelExample() {
	// 创建工作池
	pool := task.NewWorkerPool(4, nil)
	pool.Start()
	defer pool.Stop()

	// 创建一组并行任务
	parallelTasks := make([]*task.Task, 3)
	for i := 0; i < 3; i++ {
		// 使用局部变量避免闭包捕获问题
		index := i
		taskName := fmt.Sprintf("并行任务%d", index+1)
		parallelTasks[index] = task.New(
			task.WithName(taskName),
			task.WithJob(func(ctx context.Context) error {
				// 从上下文中获取任务，然后获取任务名称
				currentTask := task.TaskFromContext(ctx)
				currentTaskName := currentTask.GetName()
				log.Printf("执行任务: %s", currentTaskName)
				time.Sleep(300 * time.Millisecond) // 模拟工作
				log.Printf("完成任务: %s", currentTaskName)
				return nil
			}),
		)
	}

	// 创建一个汇聚任务，等待所有并行任务完成
	joinTask := task.Parallel("并行组", parallelTasks...)

	// 创建一个最终任务，依赖于汇聚任务
	finalTask := task.New(
		task.WithName("最终任务"),
		task.WithJob(func(ctx context.Context) error {
			log.Println("执行最终任务")
			time.Sleep(200 * time.Millisecond) // 模拟工作
			log.Println("所有并行任务和最终任务都已完成")
			return nil
		}),
	)

	// 设置依赖关系
	finalTask.DependsOn(joinTask)

	// 提交所有任务
	for _, t := range parallelTasks {
		pool.Submit(t)
	}
	pool.Submit(joinTask)
	pool.Submit(finalTask)

	// 等待任务完成
	time.Sleep(1 * time.Second)
}

// 示例4：复杂依赖图
func complexDependencyExample() {
	// 创建工作池
	pool := task.NewWorkerPool(4, nil)
	pool.Start()
	defer pool.Stop()

	// 创建任务构建器
	taskA := task.NewTaskBuilder("任务A").
		WithJob(func(ctx context.Context) error {
			log.Println("执行任务A")
			time.Sleep(200 * time.Millisecond)
			return nil
		}).
		Build()

	taskB := task.NewTaskBuilder("任务B").
		WithJob(func(ctx context.Context) error {
			log.Println("执行任务B")
			time.Sleep(300 * time.Millisecond)
			return nil
		}).
		DependsOn(taskA). // B 依赖 A
		Build()

	taskC := task.NewTaskBuilder("任务C").
		WithJob(func(ctx context.Context) error {
			log.Println("执行任务C")
			time.Sleep(250 * time.Millisecond)
			return nil
		}).
		DependsOn(taskA). // C 依赖 A
		Build()

	taskD := task.NewTaskBuilder("任务D").
		WithJob(func(ctx context.Context) error {
			log.Println("执行任务D")
			time.Sleep(200 * time.Millisecond)
			return nil
		}).
		DependsOn(taskB, taskC). // D 依赖 B 和 C
		WithDependenciesCallback(func() {
			log.Println("任务D的所有依赖已满足，即将开始执行")
		}).
		Build()

	// 提交所有任务
	pool.Submit(taskD)
	pool.Submit(taskC)
	pool.Submit(taskB)
	pool.Submit(taskA)

	// 等待任务完成
	time.Sleep(2 * time.Second)
}
