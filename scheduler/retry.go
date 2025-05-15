// scheduler/retry.go
package scheduler

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryStrategy 重试策略接口
type RetryStrategy interface {
	// NextRetryDelay 返回下一次重试的延迟时间
	// attempt: 当前是第几次重试（从0开始）
	// err: 上一次执行的错误
	NextRetryDelay(attempt int, err error) time.Duration
	
	// ShouldRetry 决定是否应该重试
	// err: 上一次执行的错误
	ShouldRetry(err error) bool
	
	// MaxRetries 返回最大重试次数
	MaxRetries() int
}

// FixedDelayRetryStrategy 固定间隔重试策略
type FixedDelayRetryStrategy struct {
	delay        time.Duration
	maxRetries   int
	retryableErrors []error // 可重试的错误类型
	retryPredicate func(error) bool // 自定义重试判断函数
}

// NewFixedDelayRetryStrategy 创建固定间隔重试策略
func NewFixedDelayRetryStrategy(delay time.Duration, maxRetries int) *FixedDelayRetryStrategy {
	return &FixedDelayRetryStrategy{
		delay:      delay,
		maxRetries: maxRetries,
	}
}

// WithRetryableErrors 设置可重试的错误类型
func (s *FixedDelayRetryStrategy) WithRetryableErrors(errors ...error) *FixedDelayRetryStrategy {
	s.retryableErrors = errors
	return s
}

// WithRetryPredicate 设置自定义重试判断函数
func (s *FixedDelayRetryStrategy) WithRetryPredicate(predicate func(error) bool) *FixedDelayRetryStrategy {
	s.retryPredicate = predicate
	return s
}

// NextRetryDelay 实现 RetryStrategy 接口
func (s *FixedDelayRetryStrategy) NextRetryDelay(attempt int, err error) time.Duration {
	if attempt >= s.maxRetries {
		return 0 // 不再重试
	}
	return s.delay
}

// ShouldRetry 实现 RetryStrategy 接口
func (s *FixedDelayRetryStrategy) ShouldRetry(err error) bool {
	// 如果错误为空，不需要重试
	if err == nil {
		return false
	}
	
	// 如果有自定义重试判断函数，使用它
	if s.retryPredicate != nil {
		return s.retryPredicate(err)
	}
	
	// 如果没有指定可重试的错误类型，则所有错误都可重试
	if len(s.retryableErrors) == 0 {
		return true
	}
	
	// 检查错误是否在可重试列表中
	for _, retryableErr := range s.retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}
	
	return false
}

// MaxRetries 实现 RetryStrategy 接口
func (s *FixedDelayRetryStrategy) MaxRetries() int {
	return s.maxRetries
}

// ExponentialBackoffRetryStrategy 指数退避重试策略
type ExponentialBackoffRetryStrategy struct {
	initialDelay   time.Duration
	maxDelay       time.Duration
	factor         float64
	maxRetries     int
	retryableErrors []error
	retryPredicate func(error) bool
	jitter         bool // 是否添加随机抖动
}

// NewExponentialBackoffRetryStrategy 创建指数退避重试策略
func NewExponentialBackoffRetryStrategy(initialDelay, maxDelay time.Duration, factor float64, maxRetries int) *ExponentialBackoffRetryStrategy {
	return &ExponentialBackoffRetryStrategy{
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
		factor:       factor,
		maxRetries:   maxRetries,
		jitter:       true, // 默认启用抖动
	}
}

// WithRetryableErrors 设置可重试的错误类型
func (s *ExponentialBackoffRetryStrategy) WithRetryableErrors(errors ...error) *ExponentialBackoffRetryStrategy {
	s.retryableErrors = errors
	return s
}

// WithRetryPredicate 设置自定义重试判断函数
func (s *ExponentialBackoffRetryStrategy) WithRetryPredicate(predicate func(error) bool) *ExponentialBackoffRetryStrategy {
	s.retryPredicate = predicate
	return s
}

// WithJitter 设置是否添加随机抖动
func (s *ExponentialBackoffRetryStrategy) WithJitter(jitter bool) *ExponentialBackoffRetryStrategy {
	s.jitter = jitter
	return s
}

// NextRetryDelay 实现 RetryStrategy 接口
func (s *ExponentialBackoffRetryStrategy) NextRetryDelay(attempt int, err error) time.Duration {
	if attempt >= s.maxRetries {
		return 0 // 不再重试
	}
	
	// 计算指数退避延迟
	delay := s.initialDelay * time.Duration(math.Pow(s.factor, float64(attempt)))
	
	// 添加随机抖动，避免多个任务同时重试
	if s.jitter {
		jitter := time.Duration(rand.Int63n(int64(delay) / 4))
		delay = delay + jitter
	}
	
	// 确保不超过最大延迟
	if delay > s.maxDelay {
		delay = s.maxDelay
	}
	
	return delay
}

// ShouldRetry 实现 RetryStrategy 接口
func (s *ExponentialBackoffRetryStrategy) ShouldRetry(err error) bool {
	// 如果错误为空，不需要重试
	if err == nil {
		return false
	}
	
	// 如果有自定义重试判断函数，使用它
	if s.retryPredicate != nil {
		return s.retryPredicate(err)
	}
	
	// 如果没有指定可重试的错误类型，则所有错误都可重试
	if len(s.retryableErrors) == 0 {
		return true
	}
	
	// 检查错误是否在可重试列表中
	for _, retryableErr := range s.retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}
	
	return false
}

// MaxRetries 实现 RetryStrategy 接口
func (s *ExponentialBackoffRetryStrategy) MaxRetries() int {
	return s.maxRetries
}
