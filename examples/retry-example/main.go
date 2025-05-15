// examples/retry-example/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

// 自定义错误类型
var (
	ErrTemporary = errors.New("临时错误，可以重试")
	ErrPermanent = errors.New("永久错误，不应重试")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("重试策略示例")

	// 示例1：使用预定义重试策略
	log.Println("\n=== 示例1：预定义重试策略 ===")
	predefinedStrategyExample()

	// 示例2：自定义重试策略
	log.Println("\n=== 示例2：自定义重试策略 ===")
	customStrategyExample()

	// 示例3：使用任务构建器API
	log.Println("\n=== 示例3：任务构建器 ===")
	taskBuilderExample()

	// 示例4：错误类型判断
	log.Println("\n=== 示例4：错误类型判断 ===")
	errorTypeExample()
}

// 示例1：使用预定义重试策略
func predefinedStrategyExample() {
	// 创建一个随机失败的任务，使用简单重试策略
	task1 := task.RetryableTask("简单重试任务", func(ctx context.Context) error {
		log.Println("简单重试任务：尝试执行")

		// 随机失败
		if rand.Float32() < 0.7 {
			log.Println("简单重试任务：执行失败，将重试")
			return fmt.Errorf("随机失败")
		}

		log.Println("简单重试任务：执行成功")
		return nil
	}, task.SimpleRetry)

	// 运行任务
	task1.Run()

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	// 创建一个使用渐进重试策略的任务
	task2 := task.RetryableTask("渐进重试任务", func(ctx context.Context) error {
		log.Println("渐进重试任务：尝试执行")

		// 总是失败
		log.Println("渐进重试任务：执行失败，将使用指数退避重试")
		return fmt.Errorf("总是失败")
	}, task.ProgressiveRetry)

	// 运行任务
	task2.Run()

	// 等待任务完成
	time.Sleep(3 * time.Second)
}

// 示例2：自定义重试策略
func customStrategyExample() {
	// 创建自定义固定间隔重试策略
	customStrategy := task.FixedDelayWithRetryableErrors(
		task.NewFixedDelayRetryStrategy(200*time.Millisecond, 2),
		ErrTemporary)

	// 创建一个任务，根据错误类型决定是否重试
	task1 := task.RetryableTask("自定义重试任务", func(ctx context.Context) error {
		log.Println("自定义重试任务：尝试执行")

		// 随机返回不同类型的错误
		if rand.Float32() < 0.5 {
			log.Println("自定义重试任务：返回临时错误，应该重试")
			return ErrTemporary
		} else {
			log.Println("自定义重试任务：返回永久错误，不应重试")
			return ErrPermanent
		}
	}, customStrategy)

	// 运行任务
	task1.Run()

	// 等待任务完成
	time.Sleep(1 * time.Second)

	// 创建自定义指数退避重试策略
	exponentialStrategy := task.ExponentialBackoffWithJitter(
		task.NewExponentialBackoffRetryStrategy(
			100*time.Millisecond, // 初始延迟
			1*time.Second,        // 最大延迟
			2.0,                  // 指数因子
			3,                    // 最大重试次数
		),
		true, // 启用随机抖动
	)

	// 创建一个使用指数退避策略的任务
	task2 := task.RetryableTask("指数退避任务", func(ctx context.Context) error {
		log.Println("指数退避任务：尝试执行")

		// 总是失败
		log.Println("指数退避任务：执行失败，将使用指数退避重试")
		return fmt.Errorf("总是失败")
	}, exponentialStrategy)

	// 运行任务
	task2.Run()

	// 等待任务完成
	time.Sleep(3 * time.Second)
}

// 示例3：使用任务构建器API
func taskBuilderExample() {
	// 使用任务构建器创建带重试策略的任务
	_ = task.NewTaskBuilder("构建器重试任务").
		WithJob(func(ctx context.Context) error {
			log.Println("构建器重试任务：尝试执行")

			// 随机失败
			if rand.Float32() < 0.7 {
				log.Println("构建器重试任务：执行失败，将重试")
				return fmt.Errorf("随机失败")
			}

			log.Println("构建器重试任务：执行成功")
			return nil
		}).
		WithRetryStrategy(task.SimpleRetry). // 使用简单重试策略
		Run()

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	// 使用任务构建器创建带自定义重试次数的任务
	_ = task.NewTaskBuilder("简单重试任务").
		WithJob(func(ctx context.Context) error {
			log.Println("简单重试任务：尝试执行")

			// 总是失败
			log.Println("简单重试任务：执行失败，将重试")
			return fmt.Errorf("总是失败")
		}).
		WithRetry(2). // 简单设置重试次数
		Run()

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)
}

// 示例4：错误类型判断
func errorTypeExample() {
	// 创建一个网络错误重试策略
	networkRetryStrategy := task.RetryOnNetworkError(task.SimpleRetry)

	// 创建一个模拟网络操作的任务
	task1 := task.RetryableTask("网络操作任务", func(ctx context.Context) error {
		log.Println("网络操作任务：尝试执行")

		// 模拟网络错误
		if rand.Float32() < 0.7 {
			err := fmt.Errorf("connection refused: 连接被拒绝")
			log.Printf("网络操作任务：网络错误 %v，将重试", err)
			return err
		}

		log.Println("网络操作任务：执行成功")
		return nil
	}, networkRetryStrategy)

	// 运行任务
	task1.Run()

	// 等待任务完成
	time.Sleep(1 * time.Second)
}
