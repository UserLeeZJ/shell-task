// scheduler/task.go
package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Job 定义任务函数
type Job func(ctx context.Context) error

// JobResult 用于记录任务执行结果
type JobResult struct {
	Name     string
	Duration time.Duration
	Success  bool
	Err      error
}

// TaskOption 是配置任务的函数类型
type TaskOption func(*Task)

// Task 表示一个可配置的任务
type Task struct {
	name            string
	job             Job
	timeout         time.Duration
	interval        time.Duration
	maxRuns         int
	retryTimes      int
	parallelism     int
	startupDelay    time.Duration
	preHook         func()
	postHook        func()
	errorHandler    func(error)
	cancelOnErr     bool
	logger          func(string, ...any)
	recoverHook     func(interface{})
	metricCollector func(JobResult)

	ctx        context.Context
	cancelFunc context.CancelFunc
	runCount   int
}

// NewTask 创建新任务，并应用所有配置项
func NewTask(opts ...TaskOption) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ctx:        ctx,
		cancelFunc: cancel,

		// 默认值
		logger: func(format string, args ...any) {
			log.Printf(format, args...)
		},
	}

	// 应用所有配置项
	for _, opt := range opts {
		opt(task)
	}

	return task
}

// Run 启动任务
func (t *Task) Run() {
	if t.job == nil {
		panic("job is not set")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.logger("[%s] Recovered from panic: %v", t.name, r)
				if t.recoverHook != nil {
					t.recoverHook(r)
				}
			}
		}()

		// 延迟启动
		if t.startupDelay > 0 {
			t.logger("[%s] Startup delay: %v", t.name, t.startupDelay)
			select {
			case <-t.ctx.Done():
				return
			case <-time.After(t.startupDelay):
			}
		}

		for {
			select {
			case <-t.ctx.Done():
				t.logger("[%s] Task stopped: %v", t.name, t.ctx.Err())
				return
			default:
				if t.preHook != nil {
					t.preHook()
				}

				var err error
				for attempt := 0; attempt <= t.retryTimes; attempt++ {
					start := time.Now()

					// 创建任务执行上下文，如果设置了超时，则使用带超时的上下文
					jobCtx := t.ctx
					var cancel context.CancelFunc
					if t.timeout > 0 {
						jobCtx, cancel = context.WithTimeout(t.ctx, t.timeout)
						defer cancel()
					}

					// 使用可能带有超时的上下文执行任务
					err = t.job(jobCtx)
					duration := time.Since(start)

					// 检查是否因为超时而取消
					if jobCtx.Err() == context.DeadlineExceeded {
						t.logger("[%s] Task timed out after %v", t.name, t.timeout)
						err = fmt.Errorf("task timed out after %v: %w", t.timeout, jobCtx.Err())
					}

					result := JobResult{
						Name:     t.name,
						Duration: duration,
						Success:  err == nil,
						Err:      err,
					}

					if t.metricCollector != nil {
						t.metricCollector(result)
					}

					if err == nil {
						break
					}

					if attempt < t.retryTimes {
						t.logger("[%s] Attempt %d failed: %v, retrying...", t.name, attempt+1, err)
					} else {
						t.logger("[%s] Failed after %d attempts: %v", t.name, t.retryTimes, err)
						if t.errorHandler != nil {
							t.errorHandler(err)
						}
						if t.cancelOnErr {
							t.cancelFunc()
						}
					}
				}

				if t.postHook != nil {
					t.postHook()
				}

				// 判断最大运行次数
				t.runCount++
				if t.maxRuns > 0 && t.runCount >= t.maxRuns {
					t.logger("[%s] Reached max runs (%d), stopping.", t.name, t.maxRuns)
					t.cancelFunc()
					return
				}

				// 如果不是周期性任务，执行一次就退出
				if t.interval <= 0 {
					return
				}

				// 等待下一次执行
				select {
				case <-t.ctx.Done():
					return
				case <-time.After(t.interval):
				}
			}
		}
	}()
}

// Stop 停止任务
func (t *Task) Stop() {
	t.cancelFunc()
}

// GetRunCount 返回当前运行次数
func (t *Task) GetRunCount() int {
	return t.runCount
}
