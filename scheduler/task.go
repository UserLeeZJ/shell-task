// scheduler/task.go
package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// 使用标准库的 log 包，便于默认 logger 实现
var stdLog = log.Printf

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

// Priority 定义任务优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 5
	PriorityHigh   Priority = 10
)

// 移除 ResourceLimits 结构体

// TaskState 表示任务的状态
type TaskState int

const (
	TaskStateIdle      TaskState = iota // 空闲状态，尚未运行
	TaskStateRunning                    // 正在运行
	TaskStatePaused                     // 已暂停
	TaskStateCompleted                  // 已完成
	TaskStateFailed                     // 执行失败
	TaskStateCancelled                  // 已取消
)

// Task 表示一个可配置的任务
type Task struct {
	name            string
	job             Job
	timeout         time.Duration
	interval        time.Duration
	maxRuns         int
	retryTimes      int
	startupDelay    time.Duration
	preHook         func()
	postHook        func()
	errorHandler    func(error)
	cancelOnErr     bool
	logger          Logger
	recoverHook     func(any)
	metricCollector func(JobResult)
	priority        Priority // 任务优先级

	ctx        context.Context
	cancelFunc context.CancelFunc
	runCount   int64

	// 任务状态管理
	state       TaskState    // 当前状态
	stateMutex  sync.RWMutex // 保护状态的互斥锁
	lastRunTime time.Time    // 上次运行时间
	lastError   error        // 上次错误

	// 生命周期事件
	onStateChange func(oldState, newState TaskState) // 状态变化回调
}

// NewTask 创建新任务，并应用所有配置项
func NewTask(opts ...TaskOption) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ctx:        ctx,
		cancelFunc: cancel,

		// 默认值
		logger:   defaultLoggerInstance,
		priority: PriorityNormal,
		state:    TaskStateIdle,

		// 默认状态变化回调
		onStateChange: func(oldState, newState TaskState) {
			// 默认实现为空
		},
	}

	// 应用所有配置项
	for _, opt := range opts {
		opt(task)
	}

	return task
}

// GetState 获取任务当前状态
func (t *Task) GetState() TaskState {
	t.stateMutex.RLock()
	defer t.stateMutex.RUnlock()
	return t.state
}

// setState 设置任务状态（内部方法）
func (t *Task) setState(newState TaskState) {
	t.stateMutex.Lock()
	oldState := t.state
	t.state = newState
	t.stateMutex.Unlock()

	// 调用状态变化回调
	if t.onStateChange != nil {
		t.onStateChange(oldState, newState)
	}
}

// GetLastRunTime 获取上次运行时间
func (t *Task) GetLastRunTime() time.Time {
	t.stateMutex.RLock()
	defer t.stateMutex.RUnlock()
	return t.lastRunTime
}

// GetLastError 获取上次错误
func (t *Task) GetLastError() error {
	t.stateMutex.RLock()
	defer t.stateMutex.RUnlock()
	return t.lastError
}

// Run 启动任务
func (t *Task) Run() {
	if t.job == nil {
		panic("job is not set")
	}

	// 检查任务状态，如果已经在运行则不重复启动
	currentState := t.GetState()
	if currentState == TaskStateRunning {
		t.logger.Warn("[%s] Task is already running", t.name)
		return
	}

	// 更新任务状态为运行中
	t.setState(TaskStateRunning)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.logger.Error("[%s] Recovered from panic: %v", t.name, r)
				if t.recoverHook != nil {
					t.recoverHook(r)
				}

				// 更新任务状态为失败
				t.setState(TaskStateFailed)

				// 记录错误信息
				t.stateMutex.Lock()
				t.lastError = fmt.Errorf("panic: %v", r)
				t.stateMutex.Unlock()
			}
		}()

		// 延迟启动
		if t.startupDelay > 0 {
			t.logger.Info("[%s] Startup delay: %v", t.name, t.startupDelay)
			select {
			case <-t.ctx.Done():
				t.logger.Warn("[%s] Startup delay interrupted: %v", t.name, t.ctx.Err())
				t.setState(TaskStateCancelled)
				return
			case <-time.After(t.startupDelay):
			}
		}

		for {
			select {
			case <-t.ctx.Done():
				t.logger.Info("[%s] Task stopped: %v", t.name, t.ctx.Err())
				t.setState(TaskStateCancelled)
				return
			default:
				if t.preHook != nil {
					t.preHook()
				}

				var err error
				start := time.Now()

				// 更新上次运行时间
				t.stateMutex.Lock()
				t.lastRunTime = start
				t.stateMutex.Unlock()

				for attempt := 0; attempt <= t.retryTimes; attempt++ {
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
						t.logger.Error("[%s] Task timed out after %v", t.name, t.timeout)
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
						t.logger.Warn("[%s] Attempt %d failed: %v, retrying...", t.name, attempt+1, err)
					} else {
						t.logger.Error("[%s] Failed after %d attempts: %v", t.name, t.retryTimes, err)

						// 更新任务状态和错误信息
						t.stateMutex.Lock()
						t.lastError = err
						t.stateMutex.Unlock()

						if t.errorHandler != nil {
							t.errorHandler(err)
						}

						if t.cancelOnErr {
							t.setState(TaskStateFailed)
							t.cancelFunc()
							return
						}
					}
				}

				if t.postHook != nil {
					t.postHook()
				}

				// 判断最大运行次数
				newCount := atomic.AddInt64(&t.runCount, 1)
				if t.maxRuns > 0 && int(newCount) >= t.maxRuns {
					t.logger.Info("[%s] Reached max runs (%d), stopping.", t.name, t.maxRuns)
					t.setState(TaskStateCompleted)
					t.cancelFunc()
					return
				}

				// 如果不是周期性任务，执行一次就退出
				if t.interval <= 0 {
					t.setState(TaskStateCompleted)
					return
				}

				// 等待下一次执行
				select {
				case <-t.ctx.Done():
					t.logger.Info("[%s] Next execution canceled: %v", t.name, t.ctx.Err())
					t.setState(TaskStateCancelled)
					return
				case <-time.After(t.interval):
				}
			}
		}
	}()
}

// Pause 暂停任务（仅对周期性任务有效）
func (t *Task) Pause() bool {
	currentState := t.GetState()
	if currentState != TaskStateRunning {
		return false
	}

	t.setState(TaskStatePaused)
	return true
}

// Resume 恢复暂停的任务
func (t *Task) Resume() bool {
	currentState := t.GetState()
	if currentState != TaskStatePaused {
		return false
	}

	t.setState(TaskStateRunning)
	return true
}

// Stop 停止任务
func (t *Task) Stop() {
	currentState := t.GetState()
	if currentState == TaskStateCancelled || currentState == TaskStateCompleted {
		return // 任务已经停止
	}

	if t.ctx.Err() == nil { // 只有在任务未停止时才记录日志和取消
		t.logger.Info("[%s] Stopping task...", t.name)
		t.setState(TaskStateCancelled)
		t.cancelFunc()
	}
}

// Reset 重置任务状态，允许重新运行
func (t *Task) Reset() {
	currentState := t.GetState()
	if currentState == TaskStateRunning {
		t.Stop() // 如果任务正在运行，先停止它
	}

	// 创建新的上下文
	ctx, cancel := context.WithCancel(context.Background())

	t.stateMutex.Lock()
	// 重置状态
	t.state = TaskStateIdle
	t.lastError = nil
	t.lastRunTime = time.Time{}

	// 重置上下文
	t.ctx = ctx
	t.cancelFunc = cancel

	// 重置运行计数
	atomic.StoreInt64(&t.runCount, 0)
	t.stateMutex.Unlock()

	t.logger.Info("[%s] Task has been reset", t.name)
}

// WithStateChangeCallback 设置状态变化回调
func WithStateChangeCallback(callback func(oldState, newState TaskState)) TaskOption {
	return func(t *Task) {
		t.onStateChange = callback
	}
}

// GetRunCount 返回当前运行次数
func (t *Task) GetRunCount() int {
	return int(atomic.LoadInt64(&t.runCount))
}
