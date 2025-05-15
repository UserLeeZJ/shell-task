// scheduler/builder.go
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// TaskWithContextMap 创建一个带上下文的任务，使用 map 传递上下文数据
// 这是一个更通用的实现，替代原来的 SimpleTaskWithContext
func TaskWithContextMap(name string, fn func(ctx context.Context, data map[string]interface{}) error) *Task {
	// 创建上下文
	taskContext := NewTaskContext()

	// 创建任务
	return NewTask(
		WithName(name),
		WithTaskContext(taskContext),
		WithJob(func(ctx context.Context) error {
			// 将上下文数据转换为简单的 map 传递给用户函数
			return fn(ctx, taskContext.GetAll())
		}),
	)
}

// ChainTasks 创建任务链，自动传递上下文数据
func ChainTasks(tasks ...*Task) []*Task {
	if len(tasks) <= 1 {
		return tasks
	}

	// 设置任务完成回调，传递上下文数据
	for i := 0; i < len(tasks)-1; i++ {
		currentTask := tasks[i]
		nextTask := tasks[i+1]

		// 设置当前任务的完成回调
		originalPostHook := currentTask.postHook
		currentTask.postHook = func() {
			if originalPostHook != nil {
				originalPostHook()
			}

			// 将当前任务的上下文数据传递给下一个任务
			if currentTask.taskContext != nil {
				if nextTask.taskContext == nil {
					nextTask.taskContext = NewTaskContext()
				}

				// 复制所有上下文值
				for k, v := range currentTask.taskContext.GetAll() {
					nextTask.taskContext.Set(k, v)
				}
			}
		}
	}

	return tasks
}

// TaskBuilder 提供流式API创建和配置任务
type TaskBuilder struct {
	task *Task
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder(name string) *TaskBuilder {
	return &TaskBuilder{
		task: NewTask(WithName(name)),
	}
}

// WithJob 设置任务函数
func (tb *TaskBuilder) WithJob(fn func(context.Context) error) *TaskBuilder {
	tb.task.job = fn
	return tb
}

// WithContextJob 设置带上下文的任务函数
func (tb *TaskBuilder) WithContextJob(fn func(ctx context.Context, taskCtx *TaskContext) error) *TaskBuilder {
	tb.task.job = func(ctx context.Context) error {
		return fn(ctx, tb.task.GetContext())
	}
	return tb
}

// WithMapContextJob 设置带映射上下文的任务函数
// 替代原来的 WithSimpleContextJob
func (tb *TaskBuilder) WithMapContextJob(fn func(ctx context.Context, data map[string]interface{}) error) *TaskBuilder {
	tb.task.job = func(ctx context.Context) error {
		return fn(ctx, tb.task.GetContext().GetAll())
	}
	return tb
}

// WithTimeout 设置超时
func (tb *TaskBuilder) WithTimeout(timeout time.Duration) *TaskBuilder {
	tb.task.timeout = timeout
	return tb
}

// WithRepeat 设置重复执行
func (tb *TaskBuilder) WithRepeat(interval time.Duration) *TaskBuilder {
	tb.task.interval = interval
	return tb
}

// WithMaxRuns 设置最大运行次数
func (tb *TaskBuilder) WithMaxRuns(n int) *TaskBuilder {
	tb.task.maxRuns = n
	return tb
}

// WithContextValue 设置上下文数据
func (tb *TaskBuilder) WithContextValue(key string, value interface{}) *TaskBuilder {
	tb.task.SetContextValue(key, value)
	return tb
}

// WithContextPrep 设置上下文准备钩子
func (tb *TaskBuilder) WithContextPrep(prep func(*TaskContext)) *TaskBuilder {
	tb.task.contextPrep = prep
	return tb
}

// WithContextClean 设置上下文清理钩子
func (tb *TaskBuilder) WithContextClean(clean func(*TaskContext)) *TaskBuilder {
	tb.task.contextClean = clean
	return tb
}

// WithTaskContext 设置完整的任务上下文
func (tb *TaskBuilder) WithTaskContext(ctx *TaskContext) *TaskBuilder {
	tb.task.taskContext = ctx
	return tb
}

// WithLogger 设置日志记录器
func (tb *TaskBuilder) WithLogger(logger Logger) *TaskBuilder {
	tb.task.logger = logger
	return tb
}

// WithLoggerFunc 设置日志函数
func (tb *TaskBuilder) WithLoggerFunc(logFunc func(format string, args ...any)) *TaskBuilder {
	tb.task.logger = NewFuncLogger(logFunc)
	return tb
}

// WithPriority 设置任务优先级
func (tb *TaskBuilder) WithPriority(priority Priority) *TaskBuilder {
	tb.task.priority = priority
	return tb
}

// WithPreHook 设置前置钩子
func (tb *TaskBuilder) WithPreHook(hook func()) *TaskBuilder {
	tb.task.preHook = hook
	return tb
}

// WithPostHook 设置后置钩子
func (tb *TaskBuilder) WithPostHook(hook func()) *TaskBuilder {
	tb.task.postHook = hook
	return tb
}

// WithErrorHandler 设置错误处理器
func (tb *TaskBuilder) WithErrorHandler(handler func(error)) *TaskBuilder {
	tb.task.errorHandler = handler
	return tb
}

// WithCancelOnFailure 设置失败时是否取消任务
func (tb *TaskBuilder) WithCancelOnFailure(cancel bool) *TaskBuilder {
	tb.task.cancelOnErr = cancel
	return tb
}

// WithRecover 设置恢复钩子
func (tb *TaskBuilder) WithRecover(hook func(any)) *TaskBuilder {
	tb.task.recoverHook = hook
	return tb
}

// WithMetricCollector 设置指标收集器
func (tb *TaskBuilder) WithMetricCollector(collector func(JobResult)) *TaskBuilder {
	tb.task.metricCollector = collector
	return tb
}

// WithStartupDelay 设置启动延迟
func (tb *TaskBuilder) WithStartupDelay(delay time.Duration) *TaskBuilder {
	tb.task.startupDelay = delay
	return tb
}

// WithContextTransformer 设置上下文转换器
func (tb *TaskBuilder) WithContextTransformer(transformer func(key string, value interface{}) (string, interface{})) *TaskBuilder {
	if tb.task.taskContext == nil {
		tb.task.taskContext = NewTaskContext()
	}

	// 创建一个新的上下文，应用转换器
	newContext := tb.task.taskContext.Transform(transformer)

	// 将新上下文设置为任务上下文
	tb.task.taskContext = newContext
	return tb
}

// WithContextFilter 设置上下文过滤器
func (tb *TaskBuilder) WithContextFilter(prefix string) *TaskBuilder {
	if tb.task.taskContext == nil {
		tb.task.taskContext = NewTaskContext()
		return tb
	}

	// 创建一个新的上下文
	newContext := NewTaskContext()

	// 获取过滤后的值
	filteredValues := tb.task.taskContext.Filter(prefix)

	// 将过滤后的值设置到新上下文
	for k, v := range filteredValues {
		newContext.Set(k, v)
	}

	// 将新上下文设置为任务上下文
	tb.task.taskContext = newContext
	return tb
}

// WithContextValidator 设置上下文验证器
func (tb *TaskBuilder) WithContextValidator(validators map[string]Validator) *TaskBuilder {
	if tb.task.taskContext == nil {
		tb.task.taskContext = NewTaskContext()
	}

	// 设置上下文准备钩子，在任务执行前验证上下文
	originalPrep := tb.task.contextPrep
	tb.task.contextPrep = func(ctx *TaskContext) {
		if originalPrep != nil {
			originalPrep(ctx)
		}

		// 验证上下文
		if err := ctx.Validate(validators); err != nil {
			panic(fmt.Sprintf("Context validation failed: %v", err))
		}
	}
	return tb
}

// WithRequiredContextKeys 设置必需的上下文键
func (tb *TaskBuilder) WithRequiredContextKeys(keys ...string) *TaskBuilder {
	if tb.task.taskContext == nil {
		tb.task.taskContext = NewTaskContext()
	}

	// 设置上下文准备钩子，在任务执行前验证必需的键
	originalPrep := tb.task.contextPrep
	tb.task.contextPrep = func(ctx *TaskContext) {
		if originalPrep != nil {
			originalPrep(ctx)
		}

		// 验证必需的键
		if err := ctx.RequiredKeys(keys...); err != nil {
			panic(fmt.Sprintf("Required context keys check failed: %v", err))
		}
	}
	return tb
}

// Build 构建任务
func (tb *TaskBuilder) Build() *Task {
	return tb.task
}

// WithRetry 设置简单重试
func (tb *TaskBuilder) WithRetry(times int) *TaskBuilder {
	tb.task.retryTimes = times
	return tb
}

// WithRetryStrategy 设置重试策略
func (tb *TaskBuilder) WithRetryStrategy(strategy RetryStrategy) *TaskBuilder {
	tb.task.retryStrategy = strategy
	if strategy != nil {
		tb.task.retryTimes = strategy.MaxRetries()
	}
	return tb
}

// DependsOn 设置任务依赖
func (tb *TaskBuilder) DependsOn(tasks ...*Task) *TaskBuilder {
	tb.task.DependsOn(tasks...)
	return tb
}

// WithDependenciesCallback 设置依赖满足时的回调
func (tb *TaskBuilder) WithDependenciesCallback(callback func()) *TaskBuilder {
	tb.task.WithOnDependenciesMet(callback)
	return tb
}

// WithStateChangeCallback 设置状态变化回调
func (tb *TaskBuilder) WithStateChangeCallback(callback func(oldState, newState TaskState)) *TaskBuilder {
	tb.task.onStateChange = callback
	return tb
}

// WithSync 设置是否同步执行
func (tb *TaskBuilder) WithSync(sync bool) *TaskBuilder {
	tb.task.syncExec = sync
	return tb
}

// Run 构建并运行任务
func (tb *TaskBuilder) Run() *Task {
	task := tb.Build()
	task.Run()
	return task
}

// 预定义常用的重试策略
var (
	// NoRetry 不重试
	NoRetry = NewFixedDelayRetryStrategy(0, 0)

	// SimpleRetry 简单重试（固定间隔3次）
	SimpleRetry = NewFixedDelayRetryStrategy(time.Second, 3)

	// ProgressiveRetry 渐进重试（指数退避5次）
	ProgressiveRetry = NewExponentialBackoffRetryStrategy(
		time.Second, // 初始延迟
		time.Minute, // 最大延迟
		2.0,         // 指数因子
		5,           // 最大重试次数
	)
)

// RetryableTask 创建一个带重试策略的简化任务
func RetryableTask(name string, fn func(ctx context.Context) error, strategy RetryStrategy) *Task {
	return NewTask(
		WithName(name),
		WithJob(fn),
		WithRetryStrategy(strategy),
	)
}

// RetryOnNetworkError 包装重试策略，添加网络错误判断
func RetryOnNetworkError(strategy RetryStrategy) RetryStrategy {
	// 创建一个新的固定间隔重试策略，复制原策略的参数
	if fixed, ok := strategy.(*FixedDelayRetryStrategy); ok {
		return fixed.WithRetryPredicate(func(err error) bool {
			// 检查是否为网络错误
			var netErr net.Error
			return err != nil && (errors.As(err, &netErr) || strings.Contains(err.Error(), "connection"))
		})
	}

	// 创建一个新的指数退避重试策略，复制原策略的参数
	if exp, ok := strategy.(*ExponentialBackoffRetryStrategy); ok {
		return exp.WithRetryPredicate(func(err error) bool {
			// 检查是否为网络错误
			var netErr net.Error
			return err != nil && (errors.As(err, &netErr) || strings.Contains(err.Error(), "connection"))
		})
	}

	// 如果不是已知类型，返回原策略
	return strategy
}

// FixedDelayWithRetryableErrors 设置固定间隔重试策略的可重试错误类型
func FixedDelayWithRetryableErrors(strategy *FixedDelayRetryStrategy, errs ...error) *FixedDelayRetryStrategy {
	return strategy.WithRetryableErrors(errs...)
}

// FixedDelayWithJitter 设置固定间隔重试策略的随机抖动
func FixedDelayWithJitter(strategy *FixedDelayRetryStrategy, jitter bool) *FixedDelayRetryStrategy {
	// 固定间隔策略没有 WithJitter 方法，所以这里返回原策略
	return strategy
}

// ExponentialBackoffWithRetryableErrors 设置指数退避重试策略的可重试错误类型
func ExponentialBackoffWithRetryableErrors(strategy *ExponentialBackoffRetryStrategy, errs ...error) *ExponentialBackoffRetryStrategy {
	return strategy.WithRetryableErrors(errs...)
}

// ExponentialBackoffWithJitter 设置指数退避重试策略的随机抖动
func ExponentialBackoffWithJitter(strategy *ExponentialBackoffRetryStrategy, jitter bool) *ExponentialBackoffRetryStrategy {
	return strategy.WithJitter(jitter)
}

// RunAfter 设置任务依赖关系，并返回依赖任务
func RunAfter(task *Task, dependencies ...*Task) *Task {
	// 设置依赖关系
	task.DependsOn(dependencies...)
	return task
}

// Sequence 创建一个任务序列，每个任务依赖于前一个任务
func Sequence(tasks ...*Task) []*Task {
	if len(tasks) <= 1 {
		return tasks
	}

	// 设置依赖链
	for i := 1; i < len(tasks); i++ {
		tasks[i].DependsOn(tasks[i-1])
	}

	return tasks
}

// Parallel 创建一个并行任务组，返回一个汇聚任务
func Parallel(name string, tasks ...*Task) *Task {
	if len(tasks) == 0 {
		return nil
	}

	// 创建一个汇聚任务，依赖所有并行任务
	joinTask := NewTask(
		WithName(name+"-join"),
		WithJob(func(ctx context.Context) error {
			// 这个任务不做实际工作，只是等待所有依赖完成
			return nil
		}),
		WithDependencies(tasks...),
	)

	return joinTask
}

// NewDefaultTaskGroup 创建一个使用默认日志记录器的任务组
// 这是一个更通用的实现，替代原来的 SimpleTaskGroup
func NewDefaultTaskGroup(name string) *TaskGroup {
	return NewTaskGroup(name, nil)
}

// ContextTransformerOption 设置上下文转换器
func ContextTransformerOption(transformer func(key string, value interface{}) (string, interface{})) TaskOption {
	return func(t *Task) {
		if t.taskContext == nil {
			t.taskContext = NewTaskContext()
		}

		// 创建一个新的上下文，应用转换器
		newContext := t.taskContext.Transform(transformer)

		// 将新上下文设置为任务上下文
		t.taskContext = newContext
	}
}

// ContextFilterOption 设置上下文过滤器
func ContextFilterOption(prefix string) TaskOption {
	return func(t *Task) {
		if t.taskContext == nil {
			t.taskContext = NewTaskContext()
			return
		}

		// 创建一个新的上下文
		newContext := NewTaskContext()

		// 获取过滤后的值
		filteredValues := t.taskContext.Filter(prefix)

		// 将过滤后的值设置到新上下文
		for k, v := range filteredValues {
			newContext.Set(k, v)
		}

		// 将新上下文设置为任务上下文
		t.taskContext = newContext
	}
}

// ContextValidatorOption 设置上下文验证器
func ContextValidatorOption(validators map[string]Validator) TaskOption {
	return func(t *Task) {
		if t.taskContext == nil {
			t.taskContext = NewTaskContext()
		}

		// 设置上下文准备钩子，在任务执行前验证上下文
		originalPrep := t.contextPrep
		t.contextPrep = func(ctx *TaskContext) {
			if originalPrep != nil {
				originalPrep(ctx)
			}

			// 验证上下文
			if err := ctx.Validate(validators); err != nil {
				panic(fmt.Sprintf("Context validation failed: %v", err))
			}
		}
	}
}

// RequiredContextKeysOption 设置必需的上下文键
func RequiredContextKeysOption(keys ...string) TaskOption {
	return func(t *Task) {
		if t.taskContext == nil {
			t.taskContext = NewTaskContext()
		}

		// 设置上下文准备钩子，在任务执行前验证必需的键
		originalPrep := t.contextPrep
		t.contextPrep = func(ctx *TaskContext) {
			if originalPrep != nil {
				originalPrep(ctx)
			}

			// 验证必需的键
			if err := ctx.RequiredKeys(keys...); err != nil {
				panic(fmt.Sprintf("Required context keys check failed: %v", err))
			}
		}
	}
}
