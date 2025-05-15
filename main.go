// Package shell_task 提供了一个简单而灵活的任务调度系统
package shell_task

import (
	"context"
	"time"

	"github.com/UserLeeZJ/shell-task/scheduler"
)

// Task 代表一个可配置的任务
type Task = scheduler.Task

// Job 定义任务函数类型
type Job = scheduler.Job

// JobResult 表示任务执行结果
type JobResult = scheduler.JobResult

// TaskOption 配置任务的函数类型
type TaskOption = scheduler.TaskOption

// Logger 定义了日志接口，支持不同级别的日志记录
type Logger = scheduler.Logger

// Priority 定义任务优先级
type Priority = scheduler.Priority

// TaskState 表示任务状态
type TaskState = scheduler.TaskState

// TaskContext 任务上下文，用于在任务之间传递数据
type TaskContext = scheduler.TaskContext

// TaskFromContext 从上下文中获取任务
func TaskFromContext(ctx context.Context) *Task {
	return scheduler.TaskFromContext(ctx)
}

// RetryStrategy 重试策略接口
type RetryStrategy = scheduler.RetryStrategy

// FixedDelayRetryStrategy 固定间隔重试策略
type FixedDelayRetryStrategy = scheduler.FixedDelayRetryStrategy

// ExponentialBackoffRetryStrategy 指数退避重试策略
type ExponentialBackoffRetryStrategy = scheduler.ExponentialBackoffRetryStrategy

// TaskBuilder 提供流式API创建和配置任务
type TaskBuilder = scheduler.TaskBuilder

// TaskGroup 管理一组相关任务
type TaskGroup = scheduler.TaskGroup

// Validator 上下文验证器函数类型
type Validator = scheduler.Validator

// 预定义任务状态常量
const (
	TaskStateIdle      = scheduler.TaskStateIdle
	TaskStateRunning   = scheduler.TaskStateRunning
	TaskStatePaused    = scheduler.TaskStatePaused
	TaskStateCompleted = scheduler.TaskStateCompleted
	TaskStateCancelled = scheduler.TaskStateCancelled
	TaskStateFailed    = scheduler.TaskStateFailed
)

// 预定义优先级常量
const (
	PriorityLow    = scheduler.PriorityLow
	PriorityNormal = scheduler.PriorityNormal
	PriorityHigh   = scheduler.PriorityHigh
)

// New 创建新的任务实例
func New(opts ...TaskOption) *Task {
	return scheduler.NewTask(opts...)
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder(name string) *TaskBuilder {
	return scheduler.NewTaskBuilder(name)
}

// NewTaskContext 创建新的任务上下文
func NewTaskContext() *TaskContext {
	return scheduler.NewTaskContext()
}

// NewTaskGroup 创建新的任务组
func NewTaskGroup(name string, logger Logger) *TaskGroup {
	return scheduler.NewTaskGroup(name, logger)
}

// NewDefaultTaskGroup 创建一个使用默认日志记录器的任务组
func NewDefaultTaskGroup(name string) *TaskGroup {
	return scheduler.NewDefaultTaskGroup(name)
}

// NewFuncLogger 创建一个新的函数式日志适配器
// 用于将单一日志函数转换为 Logger 接口，兼容旧的日志函数
func NewFuncLogger(logFunc func(format string, args ...any)) Logger {
	return scheduler.NewFuncLogger(logFunc)
}

// WorkerPool 表示一个工作池，用于限制并发执行的任务数量
type WorkerPool = scheduler.WorkerPool

// NewWorkerPool 创建一个新的工作池
func NewWorkerPool(size int, logger Logger) *WorkerPool {
	return scheduler.NewWorkerPool(size, logger)
}

// TaskWithContextMap 创建一个带上下文的任务，使用 map 传递上下文数据
func TaskWithContextMap(name string, fn func(ctx context.Context, data map[string]interface{}) error) *Task {
	return scheduler.TaskWithContextMap(name, fn)
}

// RetryableTask 创建一个带重试策略的简化任务
func RetryableTask(name string, fn func(ctx context.Context) error, strategy RetryStrategy) *Task {
	return scheduler.RetryableTask(name, fn, strategy)
}

// ChainTasks 创建任务链，自动传递上下文数据
func ChainTasks(tasks ...*Task) []*Task {
	return scheduler.ChainTasks(tasks...)
}

// Sequence 创建一个任务序列，每个任务依赖于前一个任务
func Sequence(tasks ...*Task) []*Task {
	return scheduler.Sequence(tasks...)
}

// Parallel 创建一个并行任务组，返回一个汇聚任务
func Parallel(name string, tasks ...*Task) *Task {
	return scheduler.Parallel(name, tasks...)
}

// RunAfter 设置任务依赖关系，并返回依赖任务
func RunAfter(task *Task, dependencies ...*Task) *Task {
	return scheduler.RunAfter(task, dependencies...)
}

// 重试策略相关函数
// NewFixedDelayRetryStrategy 创建固定间隔重试策略
func NewFixedDelayRetryStrategy(delay time.Duration, maxRetries int) *scheduler.FixedDelayRetryStrategy {
	return scheduler.NewFixedDelayRetryStrategy(delay, maxRetries)
}

// NewExponentialBackoffRetryStrategy 创建指数退避重试策略
func NewExponentialBackoffRetryStrategy(initialDelay, maxDelay time.Duration, factor float64, maxRetries int) *scheduler.ExponentialBackoffRetryStrategy {
	return scheduler.NewExponentialBackoffRetryStrategy(initialDelay, maxDelay, factor, maxRetries)
}

// RetryOnNetworkError 包装重试策略，添加网络错误判断
func RetryOnNetworkError(strategy RetryStrategy) RetryStrategy {
	return scheduler.RetryOnNetworkError(strategy)
}

// FixedDelayWithRetryableErrors 设置固定间隔重试策略的可重试错误类型
func FixedDelayWithRetryableErrors(strategy *FixedDelayRetryStrategy, errs ...error) *FixedDelayRetryStrategy {
	return scheduler.FixedDelayWithRetryableErrors(strategy, errs...)
}

// FixedDelayWithJitter 设置固定间隔重试策略的随机抖动
func FixedDelayWithJitter(strategy *FixedDelayRetryStrategy, jitter bool) *FixedDelayRetryStrategy {
	return scheduler.FixedDelayWithJitter(strategy, jitter)
}

// ExponentialBackoffWithRetryableErrors 设置指数退避重试策略的可重试错误类型
func ExponentialBackoffWithRetryableErrors(strategy *ExponentialBackoffRetryStrategy, errs ...error) *ExponentialBackoffRetryStrategy {
	return scheduler.ExponentialBackoffWithRetryableErrors(strategy, errs...)
}

// ExponentialBackoffWithJitter 设置指数退避重试策略的随机抖动
func ExponentialBackoffWithJitter(strategy *ExponentialBackoffRetryStrategy, jitter bool) *ExponentialBackoffRetryStrategy {
	return scheduler.ExponentialBackoffWithJitter(strategy, jitter)
}

// 预定义重试策略
var (
	// NoRetry 不重试
	NoRetry = scheduler.NoRetry

	// SimpleRetry 简单重试（固定间隔3次）
	SimpleRetry = scheduler.SimpleRetry

	// ProgressiveRetry 渐进重试（指数退避5次）
	ProgressiveRetry = scheduler.ProgressiveRetry
)

// 导出所有任务配置选项
var (
	// 基本选项
	WithName            = scheduler.WithName
	WithJob             = scheduler.WithJob
	WithTimeout         = scheduler.WithTimeout
	WithRepeat          = scheduler.WithRepeat
	WithMaxRuns         = scheduler.WithMaxRuns
	WithRetry           = scheduler.WithRetry
	WithLogger          = scheduler.WithLogger
	WithLoggerFunc      = scheduler.WithLoggerFunc
	WithRecover         = scheduler.WithRecover
	WithStartupDelay    = scheduler.WithStartupDelay
	WithPreHook         = scheduler.WithPreHook
	WithPostHook        = scheduler.WithPostHook
	WithErrorHandler    = scheduler.WithErrorHandler
	WithCancelOnFailure = scheduler.WithCancelOnFailure
	WithMetricCollector = scheduler.WithMetricCollector

	// 优先级选项
	WithPriority = scheduler.WithPriority

	// 上下文相关选项
	WithTaskContext     = scheduler.WithTaskContext
	WithContextValue    = scheduler.WithContextValue
	WithContextPrep     = scheduler.WithContextPrep
	WithContextClean    = scheduler.WithContextClean
	ContextTransformer  = scheduler.ContextTransformerOption
	ContextFilter       = scheduler.ContextFilterOption
	ContextValidator    = scheduler.ContextValidatorOption
	RequiredContextKeys = scheduler.RequiredContextKeysOption

	// 依赖相关选项
	WithDependencies = scheduler.WithDependencies

	// 重试相关选项
	WithRetryStrategy = scheduler.WithRetryStrategy
)
