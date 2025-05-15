// scheduler/group.go
package scheduler

import (
	"sync"
	"time"
)

// TaskGroup 管理一组相关任务
type TaskGroup struct {
	name   string
	tasks  []*Task
	mutex  sync.RWMutex
	logger Logger

	// 共享上下文
	context *TaskContext

	// 组级别的回调函数
	onAllCompleted func()
	onAnyFailed    func([]*Task)
}

// NewTaskGroup 创建新的任务组
func NewTaskGroup(name string, logger Logger) *TaskGroup {
	if logger == nil {
		logger = defaultLoggerInstance
	}

	return &TaskGroup{
		name:    name,
		tasks:   make([]*Task, 0),
		logger:  logger,
		context: NewTaskContext(),
	}
}

// AddTask 添加任务到组
func (tg *TaskGroup) AddTask(task *Task) *TaskGroup {
	tg.mutex.Lock()
	defer tg.mutex.Unlock()

	// 将任务添加到组
	tg.tasks = append(tg.tasks, task)

	// 设置任务使用组的共享上下文
	if task.taskContext == nil {
		task.taskContext = NewTaskContext()
	}
	task.taskContext.WithParent(tg.context)

	// 设置任务状态变化回调，用于跟踪组内任务状态
	originalCallback := task.onStateChange
	task.onStateChange = func(oldState, newState TaskState) {
		if originalCallback != nil {
			originalCallback(oldState, newState)
		}

		// 检查组内所有任务是否完成
		tg.checkGroupCompletion()
	}

	return tg
}

// AddTasks 添加多个任务到组
func (tg *TaskGroup) AddTasks(tasks ...*Task) *TaskGroup {
	for _, task := range tasks {
		tg.AddTask(task)
	}
	return tg
}

// GetContext 获取组的共享上下文
func (tg *TaskGroup) GetContext() *TaskContext {
	return tg.context
}

// SetContextValue 设置组上下文值
func (tg *TaskGroup) SetContextValue(key string, value interface{}) {
	tg.context.Set(key, value)
}

// GetContextValue 获取组上下文值
func (tg *TaskGroup) GetContextValue(key string) (interface{}, bool) {
	return tg.context.Get(key)
}

// RunAll 启动组内所有任务
func (tg *TaskGroup) RunAll() {
	tg.mutex.RLock()
	defer tg.mutex.RUnlock()

	tg.logger.Info("Starting all tasks in group: %s", tg.name)

	for _, task := range tg.tasks {
		task.Run()
	}
}

// StopAll 停止组内所有任务
func (tg *TaskGroup) StopAll() {
	tg.mutex.RLock()
	defer tg.mutex.RUnlock()

	tg.logger.Info("Stopping all tasks in group: %s", tg.name)

	for _, task := range tg.tasks {
		task.Stop()
	}
}

// GetGroupStats 获取组的统计信息
func (tg *TaskGroup) GetGroupStats() (total, running, completed, failed int) {
	tg.mutex.RLock()
	defer tg.mutex.RUnlock()

	total = len(tg.tasks)

	for _, task := range tg.tasks {
		state := task.GetState()
		switch state {
		case TaskStateRunning:
			running++
		case TaskStateCompleted:
			completed++
		case TaskStateFailed:
			failed++
		}
	}

	return
}

// OnAllCompleted 设置所有任务完成时的回调
func (tg *TaskGroup) OnAllCompleted(callback func()) *TaskGroup {
	tg.mutex.Lock()
	defer tg.mutex.Unlock()

	tg.onAllCompleted = callback

	// 检查是否已经全部完成
	if tg.areAllTasksCompleted() && callback != nil {
		callback()
	}

	return tg
}

// OnAnyFailed 设置任何任务失败时的回调
func (tg *TaskGroup) OnAnyFailed(callback func([]*Task)) *TaskGroup {
	tg.mutex.Lock()
	defer tg.mutex.Unlock()

	tg.onAnyFailed = callback

	// 检查是否已经有失败的任务
	failedTasks := tg.getFailedTasks()
	if len(failedTasks) > 0 && callback != nil {
		callback(failedTasks)
	}

	return tg
}

// RunAndWait 运行所有任务并等待完成
func (tg *TaskGroup) RunAndWait(timeout time.Duration) error {
	// 创建完成通知通道
	done := make(chan struct{})
	var groupErr error

	// 设置完成回调
	tg.OnAllCompleted(func() {
		close(done)
	}).OnAnyFailed(func(failedTasks []*Task) {
		if len(failedTasks) > 0 {
			groupErr = failedTasks[0].GetLastError()
		}
	})

	// 启动所有任务
	tg.RunAll()

	// 等待完成或超时
	select {
	case <-done:
		return groupErr
	case <-time.After(timeout):
		tg.StopAll()
		return ErrTimeout
	}
}

// checkGroupCompletion 检查组内所有任务是否完成
func (tg *TaskGroup) checkGroupCompletion() {
	tg.mutex.Lock()
	defer tg.mutex.Unlock()

	// 检查是否有失败的任务
	failedTasks := tg.getFailedTasksLocked()
	if len(failedTasks) > 0 && tg.onAnyFailed != nil {
		tg.onAnyFailed(failedTasks)
	}

	// 检查是否所有任务都完成了
	if tg.areAllTasksCompletedLocked() && tg.onAllCompleted != nil {
		tg.onAllCompleted()
	}
}

// areAllTasksCompleted 检查是否所有任务都已完成
func (tg *TaskGroup) areAllTasksCompleted() bool {
	tg.mutex.RLock()
	defer tg.mutex.RUnlock()

	return tg.areAllTasksCompletedLocked()
}

// areAllTasksCompletedLocked 在已获取锁的情况下检查是否所有任务都已完成
func (tg *TaskGroup) areAllTasksCompletedLocked() bool {
	for _, task := range tg.tasks {
		state := task.GetState()
		if state != TaskStateCompleted && state != TaskStateFailed && state != TaskStateCancelled {
			return false
		}
	}

	return true
}

// getFailedTasks 获取所有失败的任务
func (tg *TaskGroup) getFailedTasks() []*Task {
	tg.mutex.RLock()
	defer tg.mutex.RUnlock()

	return tg.getFailedTasksLocked()
}

// getFailedTasksLocked 在已获取锁的情况下获取所有失败的任务
func (tg *TaskGroup) getFailedTasksLocked() []*Task {
	failedTasks := make([]*Task, 0)

	for _, task := range tg.tasks {
		if task.GetState() == TaskStateFailed {
			failedTasks = append(failedTasks, task)
		}
	}

	return failedTasks
}
