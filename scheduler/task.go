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
	syncExec        bool     // 是否同步执行

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

	// 上下文管理
	taskContext  *TaskContext       // 任务上下文
	contextPrep  func(*TaskContext) // 上下文准备钩子
	contextClean func(*TaskContext) // 上下文清理钩子

	// 重试策略
	retryStrategy RetryStrategy // 重试策略

	// 依赖关系管理
	dependencies      []*Task         // 依赖的任务列表
	dependenciesMap   map[string]bool // 依赖任务的完成状态
	dependenciesMutex sync.RWMutex    // 保护依赖相关字段的互斥锁
	onDependenciesMet func()          // 所有依赖满足时的回调
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

		// 初始化依赖关系
		dependencies:    make([]*Task, 0),
		dependenciesMap: make(map[string]bool),

		// 默认依赖满足回调
		onDependenciesMet: func() {
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

// GetContext 获取任务上下文
func (t *Task) GetContext() *TaskContext {
	if t.taskContext == nil {
		t.taskContext = NewTaskContext()
	}
	return t.taskContext
}

// GetName 获取任务名称
func (t *Task) GetName() string {
	return t.name
}

// SetContextValue 设置上下文值
func (t *Task) SetContextValue(key string, value interface{}) {
	t.GetContext().Set(key, value)
}

// GetContextValue 获取上下文值
func (t *Task) GetContextValue(key string) (interface{}, bool) {
	return t.GetContext().Get(key)
}

// DependsOn 设置当前任务依赖的其他任务
func (t *Task) DependsOn(tasks ...*Task) *Task {
	t.dependenciesMutex.Lock()
	defer t.dependenciesMutex.Unlock()

	for _, task := range tasks {
		// 避免重复添加
		exists := false
		for _, dep := range t.dependencies {
			if dep == task {
				exists = true
				break
			}
		}

		if !exists {
			t.dependencies = append(t.dependencies, task)
			t.dependenciesMap[task.name] = false

			// 设置依赖任务的状态变化回调
			originalCallback := task.onStateChange
			task.onStateChange = func(oldState, newState TaskState) {
				if originalCallback != nil {
					originalCallback(oldState, newState)
				}

				// 当依赖任务完成时，更新依赖状态并传递上下文
				if newState == TaskStateCompleted {
					// 传递上下文数据
					t.transferContextFromDependency(task)

					// 更新依赖状态
					t.updateDependencyStatus(task.name, true)
				}
			}
		}
	}

	return t
}

// transferContextFromDependency 从依赖任务传递上下文数据
func (t *Task) transferContextFromDependency(dependency *Task) {
	// 确保两个任务都有上下文
	if dependency.taskContext == nil || t.taskContext == nil {
		return
	}

	// 获取依赖任务的上下文数据
	dependencyContext := dependency.taskContext.GetAll()

	// 将依赖任务的上下文数据复制到当前任务
	for key, value := range dependencyContext {
		// 只复制当前任务上下文中不存在的键，避免覆盖
		if _, exists := t.taskContext.Get(key); !exists {
			t.taskContext.Set(key, value)
		}
	}
}

// GetDependencies 获取当前任务依赖的所有任务
func (t *Task) GetDependencies() []*Task {
	t.dependenciesMutex.RLock()
	defer t.dependenciesMutex.RUnlock()

	// 创建副本
	result := make([]*Task, len(t.dependencies))
	copy(result, t.dependencies)

	return result
}

// AreDependenciesMet 检查所有依赖是否都已满足
func (t *Task) AreDependenciesMet() bool {
	t.dependenciesMutex.RLock()
	defer t.dependenciesMutex.RUnlock()

	return t.areDependenciesMetLocked()
}

// updateDependencyStatus 更新依赖任务的状态
func (t *Task) updateDependencyStatus(taskName string, status bool) {
	t.dependenciesMutex.Lock()

	// 更新依赖状态
	if _, exists := t.dependenciesMap[taskName]; exists {
		t.dependenciesMap[taskName] = status
	}

	// 检查是否所有依赖都已满足
	allMet := true
	for _, met := range t.dependenciesMap {
		if !met {
			allMet = false
			break
		}
	}

	// 保存回调函数的引用，避免在锁内调用
	var callback func()
	if allMet && t.onDependenciesMet != nil {
		callback = t.onDependenciesMet
	}

	t.dependenciesMutex.Unlock()

	// 如果所有依赖都已满足，调用回调函数
	if callback != nil {
		callback()
	}
}

// WithOnDependenciesMet 设置所有依赖满足时的回调函数
func (t *Task) WithOnDependenciesMet(callback func()) *Task {
	t.dependenciesMutex.Lock()

	t.onDependenciesMet = callback

	// 检查依赖是否已满足
	dependenciesMet := t.areDependenciesMetLocked()

	t.dependenciesMutex.Unlock()

	// 如果依赖已经满足，立即调用回调函数
	if dependenciesMet && callback != nil {
		callback()
	}

	return t
}

// areDependenciesMetLocked 在已获取锁的情况下检查依赖是否满足
func (t *Task) areDependenciesMetLocked() bool {
	// 如果没有依赖，则认为依赖已满足
	if len(t.dependencies) == 0 {
		return true
	}

	// 检查所有依赖是否都已完成
	for _, met := range t.dependenciesMap {
		if !met {
			return false
		}
	}

	return true
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

	// 检查依赖是否满足
	if !t.AreDependenciesMet() {
		t.logger.Info("[%s] Task has unmet dependencies, waiting...", t.name)

		// 设置依赖满足时的回调，自动启动任务
		t.WithOnDependenciesMet(func() {
			t.logger.Info("[%s] All dependencies met, starting task", t.name)
			// 递归调用 Run，此时依赖已满足
			t.Run()
		})

		return
	}

	// 更新任务状态为运行中
	t.setState(TaskStateRunning)

	// 根据同步/异步模式决定执行方式
	if t.syncExec {
		// 同步执行
		t.executeTaskSync()
	} else {
		// 异步执行
		go t.executeTaskAsync()
	}
}

// executeTaskSync 同步执行任务
func (t *Task) executeTaskSync() {
	t.executeTaskCore()
}

// executeTaskAsync 异步执行任务
func (t *Task) executeTaskAsync() {
	t.executeTaskCore()
}

// executeTaskCore 执行任务的核心逻辑
func (t *Task) executeTaskCore() {
	defer t.handlePanic()

	// 准备上下文
	t.prepareContext()

	// 处理启动延迟
	if !t.handleStartupDelay() {
		return // 如果在延迟期间被取消，则直接返回
	}

	// 主执行循环
	t.executeMainLoop()
}

// handlePanic 处理任务执行过程中的 panic
func (t *Task) handlePanic() {
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

		// 执行上下文清理
		t.cleanupContext()
	}
}

// prepareContext 准备任务上下文
func (t *Task) prepareContext() {
	// 确保任务上下文存在
	if t.taskContext == nil {
		t.taskContext = NewTaskContext()
	}

	// 执行上下文准备
	if t.contextPrep != nil {
		t.contextPrep(t.taskContext)
	}
}

// handleStartupDelay 处理启动延迟，返回是否应该继续执行
func (t *Task) handleStartupDelay() bool {
	if t.startupDelay <= 0 {
		return true
	}

	t.logger.Info("[%s] Startup delay: %v", t.name, t.startupDelay)
	select {
	case <-t.ctx.Done():
		t.logger.Warn("[%s] Startup delay interrupted: %v", t.name, t.ctx.Err())
		t.setState(TaskStateCancelled)
		t.cleanupContext()
		return false
	case <-time.After(t.startupDelay):
		return true
	}
}

// executeMainLoop 执行主循环
func (t *Task) executeMainLoop() {
	for {
		select {
		case <-t.ctx.Done():
			t.handleCancellation()
			return
		default:
			if !t.executeOneIteration() {
				return // 如果不需要继续执行，则返回
			}
		}
	}
}

// handleCancellation 处理任务取消
func (t *Task) handleCancellation() {
	t.logger.Info("[%s] Task stopped: %v", t.name, t.ctx.Err())
	t.setState(TaskStateCancelled)
	t.cleanupContext()
}

// executeOneIteration 执行一次任务迭代，返回是否应该继续执行
func (t *Task) executeOneIteration() bool {
	// 执行前置钩子
	if t.preHook != nil {
		t.preHook()
	}

	// 记录开始时间
	start := time.Now()
	t.stateMutex.Lock()
	t.lastRunTime = start
	t.stateMutex.Unlock()

	// 执行任务并处理重试
	err := t.executeJobWithRetry(start)

	// 处理执行结果
	if !t.handleJobResult(err) {
		return false // 如果不需要继续执行，则返回 false
	}

	// 执行后置钩子
	if t.postHook != nil {
		t.postHook()
	}

	// 更新运行次数并检查是否达到最大运行次数
	if !t.checkMaxRuns() {
		return false // 如果达到最大运行次数，则返回 false
	}

	// 如果不是周期性任务，执行一次就退出
	if t.interval <= 0 {
		t.setState(TaskStateCompleted)
		t.cleanupContext()
		return false
	}

	// 等待下一次执行
	return t.waitForNextRun()
}

// executeJobWithRetry 执行任务并处理重试逻辑，返回最终错误
func (t *Task) executeJobWithRetry(start time.Time) error {
	var err error
	maxRetries := t.getMaxRetries()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 创建任务执行上下文
		jobCtx, cancel := t.createJobContext()
		if cancel != nil {
			defer cancel()
		}

		// 执行任务
		err = t.job(jobCtx)
		duration := time.Since(start)

		// 检查是否因为超时而取消
		if jobCtx.Err() == context.DeadlineExceeded {
			t.logger.Error("[%s] Task timed out after %v", t.name, t.timeout)
			err = fmt.Errorf("task timed out after %v: %w", t.timeout, jobCtx.Err())
		}

		// 收集指标
		t.collectMetrics(JobResult{
			Name:     t.name,
			Duration: duration,
			Success:  err == nil,
			Err:      err,
		})

		// 如果成功，则跳出重试循环
		if err == nil {
			break
		}

		// 如果需要重试，则等待后重试
		if !t.shouldRetry(err, attempt, maxRetries) {
			break
		}
	}

	return err
}

// getMaxRetries 获取最大重试次数
func (t *Task) getMaxRetries() int {
	maxRetries := t.retryTimes
	if t.retryStrategy != nil {
		maxRetries = t.retryStrategy.MaxRetries()
	}
	return maxRetries
}

// createJobContext 创建任务执行上下文
func (t *Task) createJobContext() (context.Context, context.CancelFunc) {
	jobCtx := t.ctx
	var cancel context.CancelFunc

	if t.timeout > 0 {
		jobCtx, cancel = context.WithTimeout(t.ctx, t.timeout)
	}

	// 将任务实例添加到上下文中，便于在任务函数中访问
	return WithTaskInContext(jobCtx, t), cancel
}

// collectMetrics 收集任务执行指标
func (t *Task) collectMetrics(result JobResult) {
	if t.metricCollector != nil {
		t.metricCollector(result)
	}
}

// shouldRetry 判断是否应该重试
func (t *Task) shouldRetry(err error, attempt, maxRetries int) bool {
	// 如果是最后一次尝试，不需要重试
	if attempt >= maxRetries {
		return false
	}

	if t.retryStrategy != nil {
		// 检查是否应该重试
		if !t.retryStrategy.ShouldRetry(err) {
			t.logger.Warn("[%s] Error not retryable: %v", t.name, err)
			return false
		}

		// 获取下一次重试的延迟时间
		delay := t.retryStrategy.NextRetryDelay(attempt, err)
		if delay == 0 {
			t.logger.Warn("[%s] Retry strategy decided not to retry", t.name)
			return false // 策略决定不再重试
		}

		t.logger.Warn("[%s] Attempt %d failed: %v, retrying after %v...",
			t.name, attempt+1, err, delay)

		// 等待重试
		select {
		case <-t.ctx.Done():
			t.logger.Warn("[%s] Retry interrupted: %v", t.name, t.ctx.Err())
			return false
		case <-time.After(delay):
			return true // 继续下一次重试
		}
	} else {
		// 使用原有的重试逻辑
		t.logger.Warn("[%s] Attempt %d failed: %v, retrying...", t.name, attempt+1, err)
		return true
	}

	return false
}

// handleJobResult 处理任务执行结果，返回是否应该继续执行
func (t *Task) handleJobResult(err error) bool {
	if err == nil {
		return true
	}

	t.logger.Error("[%s] Failed after retries: %v", t.name, err)

	// 更新任务状态和错误信息
	t.stateMutex.Lock()
	t.lastError = err
	t.stateMutex.Unlock()

	if t.errorHandler != nil {
		t.errorHandler(err)
	}

	if t.cancelOnErr {
		t.setState(TaskStateFailed)
		t.cleanupContext()
		t.cancelFunc()
		return false
	}

	return true
}

// checkMaxRuns 检查是否达到最大运行次数，返回是否应该继续执行
func (t *Task) checkMaxRuns() bool {
	newCount := atomic.AddInt64(&t.runCount, 1)
	if t.maxRuns > 0 && int(newCount) >= t.maxRuns {
		t.logger.Info("[%s] Reached max runs (%d), stopping.", t.name, t.maxRuns)
		t.setState(TaskStateCompleted)
		t.cleanupContext()
		t.cancelFunc()
		return false
	}
	return true
}

// waitForNextRun 等待下一次执行，返回是否应该继续执行
func (t *Task) waitForNextRun() bool {
	select {
	case <-t.ctx.Done():
		t.logger.Info("[%s] Next execution canceled: %v", t.name, t.ctx.Err())
		t.setState(TaskStateCancelled)
		t.cleanupContext()
		return false
	case <-time.After(t.interval):
		return true
	}
}

// cleanupContext 清理上下文
func (t *Task) cleanupContext() {
	if t.contextClean != nil && t.taskContext != nil {
		t.contextClean(t.taskContext)
	}
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
