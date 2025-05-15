// scheduler/retry_test.go
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// 自定义错误类型，用于测试
var (
	ErrTemporary = errors.New("temporary error")
	ErrPermanent = errors.New("permanent error")
)

// TestFixedDelayRetryStrategy 测试固定间隔重试策略
func TestFixedDelayRetryStrategy(t *testing.T) {
	// 创建固定间隔重试策略
	strategy := NewFixedDelayRetryStrategy(10*time.Millisecond, 3)

	// 测试 NextRetryDelay
	if delay := strategy.NextRetryDelay(0, nil); delay != 10*time.Millisecond {
		t.Errorf("Expected delay 10ms, got %v", delay)
	}

	if delay := strategy.NextRetryDelay(3, nil); delay != 0 {
		t.Errorf("Expected delay 0 for attempt > maxRetries, got %v", delay)
	}

	// 测试 ShouldRetry
	if strategy.ShouldRetry(nil) {
		t.Error("Expected ShouldRetry(nil) to be false")
	}

	if !strategy.ShouldRetry(ErrTemporary) {
		t.Error("Expected ShouldRetry(ErrTemporary) to be true")
	}

	// 测试 MaxRetries
	if strategy.MaxRetries() != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", strategy.MaxRetries())
	}

	// 测试 WithRetryableErrors
	strategyWithRetryableErrors := strategy.WithRetryableErrors(ErrTemporary)
	if !strategyWithRetryableErrors.ShouldRetry(ErrTemporary) {
		t.Error("Expected ShouldRetry(ErrTemporary) to be true with retryable errors")
	}

	if strategyWithRetryableErrors.ShouldRetry(ErrPermanent) {
		t.Error("Expected ShouldRetry(ErrPermanent) to be false with retryable errors")
	}

	// 测试 WithRetryPredicate
	strategyWithPredicate := strategy.WithRetryPredicate(func(err error) bool {
		return err == ErrTemporary
	})

	if !strategyWithPredicate.ShouldRetry(ErrTemporary) {
		t.Error("Expected ShouldRetry(ErrTemporary) to be true with predicate")
	}

	if strategyWithPredicate.ShouldRetry(ErrPermanent) {
		t.Error("Expected ShouldRetry(ErrPermanent) to be false with predicate")
	}
}

// TestExponentialBackoffRetryStrategy 测试指数退避重试策略
func TestExponentialBackoffRetryStrategy(t *testing.T) {
	// 创建指数退避重试策略
	strategy := NewExponentialBackoffRetryStrategy(10*time.Millisecond, 100*time.Millisecond, 2.0, 3)

	// 测试 NextRetryDelay
	delay0 := strategy.NextRetryDelay(0, nil)
	if delay0 < 10*time.Millisecond || delay0 > 15*time.Millisecond {
		t.Errorf("Expected delay around 10ms, got %v", delay0)
	}

	delay1 := strategy.NextRetryDelay(1, nil)
	if delay1 < 20*time.Millisecond || delay1 > 30*time.Millisecond {
		t.Errorf("Expected delay around 20ms, got %v", delay1)
	}

	delay2 := strategy.NextRetryDelay(2, nil)
	if delay2 < 40*time.Millisecond || delay2 > 60*time.Millisecond {
		t.Errorf("Expected delay around 40ms, got %v", delay2)
	}

	// 测试最大延迟
	strategy = NewExponentialBackoffRetryStrategy(10*time.Millisecond, 15*time.Millisecond, 2.0, 3)
	delay3 := strategy.NextRetryDelay(2, nil)
	if delay3 > 15*time.Millisecond {
		t.Errorf("Expected delay to be capped at 15ms, got %v", delay3)
	}

	// 测试 WithJitter
	strategyNoJitter := strategy.WithJitter(false)
	delay := strategyNoJitter.NextRetryDelay(0, nil)
	if delay != 10*time.Millisecond {
		t.Errorf("Expected delay exactly 10ms without jitter, got %v", delay)
	}
}

// TestTaskWithRetryStrategy 测试任务使用重试策略
func TestTaskWithRetryStrategy(t *testing.T) {
	// 创建一个计数器，用于跟踪任务执行次数
	attempts := 0
	maxAttempts := 3

	// 创建一个总是失败的任务
	task := NewTask(
		WithName("RetryTest"),
		WithJob(func(ctx context.Context) error {
			attempts++
			return fmt.Errorf("attempt %d failed", attempts)
		}),
		WithRetryStrategy(NewFixedDelayRetryStrategy(10*time.Millisecond, maxAttempts)),
	)

	// 运行任务
	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	// 验证任务重试了正确的次数
	expectedAttempts := maxAttempts + 1 // 初始尝试 + 重试次数
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

// TestRetryableTask 测试简化的重试任务API
func TestRetryableTask(t *testing.T) {
	// 创建一个计数器，用于跟踪任务执行次数
	attempts := 0

	// 创建一个自定义的重试策略，使用非常短的延迟
	strategy := NewFixedDelayRetryStrategy(10*time.Millisecond, 3)

	// 创建一个使用自定义重试策略的任务
	task := RetryableTask("SimpleRetryTest", func(ctx context.Context) error {
		attempts++
		return ErrTemporary
	}, strategy)

	// 运行任务
	task.Run()
	time.Sleep(200 * time.Millisecond) // 给任务一点时间执行

	// 验证任务重试了正确的次数
	expectedAttempts := 4 // 初始尝试 + 3次重试
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

// TestRetryOnNetworkError 测试网络错误重试
func TestRetryOnNetworkError(t *testing.T) {
	// 创建一个网络错误
	networkErr := fmt.Errorf("connection refused")

	// 创建一个包装了网络错误判断的重试策略
	strategy := RetryOnNetworkError(SimpleRetry)

	// 测试网络错误应该重试
	if !strategy.ShouldRetry(networkErr) {
		t.Error("Expected network error to be retryable")
	}

	// 测试其他错误不应该重试
	if strategy.ShouldRetry(ErrPermanent) {
		t.Error("Expected non-network error to be non-retryable")
	}
}
