// Package shell_task 提供了一个简单而灵活的任务调度系统
package shell_task

import "github.com/UserLeeZJ/shell-task/scheduler"

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

// New 创建新的任务实例
func New(opts ...TaskOption) *Task {
	return scheduler.NewTask(opts...)
}

// NewFuncLogger 创建一个新的函数式日志适配器
// 用于将单一日志函数转换为 Logger 接口，兼容旧的日志函数
func NewFuncLogger(logFunc func(format string, args ...any)) Logger {
	return scheduler.NewFuncLogger(logFunc)
}

// 导出所有任务配置选项
var (
	WithName            = scheduler.WithName
	WithJob             = scheduler.WithJob
	WithTimeout         = scheduler.WithTimeout
	WithRepeat          = scheduler.WithRepeat
	WithMaxRuns         = scheduler.WithMaxRuns
	WithRetry           = scheduler.WithRetry
	WithParallelism     = scheduler.WithParallelism
	WithLogger          = scheduler.WithLogger
	WithRecover         = scheduler.WithRecover
	WithStartupDelay    = scheduler.WithStartupDelay
	WithPreHook         = scheduler.WithPreHook
	WithPostHook        = scheduler.WithPostHook
	WithErrorHandler    = scheduler.WithErrorHandler
	WithCancelOnFailure = scheduler.WithCancelOnFailure
	WithMetricCollector = scheduler.WithMetricCollector
)
