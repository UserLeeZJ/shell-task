// scheduler/options.go
package scheduler

import (
	"context"
	"time"
)

// WithName 设置任务名称
func WithName(name string) TaskOption {
	return func(t *Task) {
		t.name = name
	}
}

// WithJob 设置任务主体函数
func WithJob(job func(context.Context) error) TaskOption {
	return func(t *Task) {
		t.job = job
	}
}

// WithTimeout 设置任务超时时间
func WithTimeout(timeout time.Duration) TaskOption {
	return func(t *Task) {
		t.timeout = timeout
	}
}

// WithRepeat 设置任务以固定间隔重复执行
func WithRepeat(interval time.Duration) TaskOption {
	return func(t *Task) {
		t.interval = interval
	}
}

// WithMaxRuns 设置最大运行次数
func WithMaxRuns(n int) TaskOption {
	return func(t *Task) {
		t.maxRuns = n
	}
}

// WithRetry 出错重试 n 次
func WithRetry(n int) TaskOption {
	return func(t *Task) {
		t.retryTimes = n
	}
}

// 移除 WithParallelism 选项

// WithLogger 自定义日志记录器
func WithLogger(logger Logger) TaskOption {
	return func(t *Task) {
		t.logger = logger
	}
}

// WithLoggerFunc 使用函数作为日志记录器
func WithLoggerFunc(logFunc func(format string, args ...any)) TaskOption {
	return func(t *Task) {
		t.logger = NewFuncLogger(logFunc)
	}
}

// WithRecover 添加 panic 恢复钩子
func WithRecover(hook func(any)) TaskOption {
	return func(t *Task) {
		t.recoverHook = hook
	}
}

// WithStartupDelay 设置延迟启动时间
func WithStartupDelay(delay time.Duration) TaskOption {
	return func(t *Task) {
		t.startupDelay = delay
	}
}

// WithPreHook 添加执行前钩子
func WithPreHook(hook func()) TaskOption {
	return func(t *Task) {
		t.preHook = hook
	}
}

// WithPostHook 添加执行后钩子
func WithPostHook(hook func()) TaskOption {
	return func(t *Task) {
		t.postHook = hook
	}
}

// WithErrorHandler 设置错误处理器
func WithErrorHandler(handler func(error)) TaskOption {
	return func(t *Task) {
		t.errorHandler = handler
	}
}

// WithCancelOnFailure 设置失败时是否取消任务
func WithCancelOnFailure(cancel bool) TaskOption {
	return func(t *Task) {
		t.cancelOnErr = cancel
	}
}

// WithMetricCollector 收集任务指标
func WithMetricCollector(collector func(JobResult)) TaskOption {
	return func(t *Task) {
		t.metricCollector = collector
	}
}

// WithPriority 设置任务优先级
func WithPriority(priority Priority) TaskOption {
	return func(t *Task) {
		t.priority = priority
	}
}

// 移除资源限制相关的选项函数
