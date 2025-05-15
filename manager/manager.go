// manager/manager.go
package manager

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/UserLeeZJ/shell-task/lua"
	"github.com/UserLeeZJ/shell-task/scheduler"
	"github.com/UserLeeZJ/shell-task/storage"
)

// TaskManager 任务管理器
type TaskManager struct {
	storage    *storage.SQLiteStorage
	executor   *lua.Executor
	workerPool *scheduler.WorkerPool
	tasks      map[int64]*scheduler.Task
	mutex      sync.RWMutex
}

// NewTaskManager 创建一个新的任务管理器
func NewTaskManager(storage *storage.SQLiteStorage, executor *lua.Executor) *TaskManager {
	return &TaskManager{
		storage:    storage,
		executor:   executor,
		workerPool: scheduler.NewWorkerPool(5, nil), // 创建一个有5个工作协程的工作池
		tasks:      make(map[int64]*scheduler.Task),
	}
}

// Start 启动任务管理器
func (m *TaskManager) Start() error {
	// 启动工作池
	m.workerPool.Start()

	// 加载所有任务
	return m.LoadAllTasks()
}

// Stop 停止任务管理器
func (m *TaskManager) Stop() {
	// 停止工作池
	m.workerPool.Stop()

	// 停止所有任务
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, task := range m.tasks {
		task.Stop()
	}
}

// LoadAllTasks 加载所有任务
func (m *TaskManager) LoadAllTasks() error {
	// 获取所有任务
	tasks, err := m.storage.ListTasks()
	if err != nil {
		return err
	}

	// 加载每个任务
	for _, taskInfo := range tasks {
		if taskInfo.Status == storage.TaskStatusRunning {
			// 如果任务状态为运行中，则启动任务
			if err := m.StartTask(taskInfo.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// StartTask 启动任务
func (m *TaskManager) StartTask(id int64) error {
	// 获取任务信息
	taskInfo, err := m.storage.GetTask(id)
	if err != nil {
		return err
	}

	// 检查任务是否已经在运行
	m.mutex.RLock()
	_, exists := m.tasks[id]
	m.mutex.RUnlock()
	if exists {
		return fmt.Errorf("task %d is already running", id)
	}

	// 创建任务
	task, err := m.createTask(taskInfo)
	if err != nil {
		return err
	}

	// 添加到任务映射
	m.mutex.Lock()
	m.tasks[id] = task
	m.mutex.Unlock()

	// 更新任务状态
	taskInfo.Status = storage.TaskStatusRunning
	if err := m.storage.SaveTask(taskInfo); err != nil {
		return err
	}

	// 提交任务到工作池
	m.workerPool.Submit(task)

	return nil
}

// StopTask 停止任务
func (m *TaskManager) StopTask(id int64) error {
	// 获取任务
	m.mutex.RLock()
	task, exists := m.tasks[id]
	m.mutex.RUnlock()
	if !exists {
		return fmt.Errorf("task %d is not running", id)
	}

	// 停止任务
	task.Stop()

	// 从任务映射中移除
	m.mutex.Lock()
	delete(m.tasks, id)
	m.mutex.Unlock()

	// 更新任务状态
	taskInfo, err := m.storage.GetTask(id)
	if err != nil {
		return err
	}
	taskInfo.Status = storage.TaskStatusCancelled
	return m.storage.SaveTask(taskInfo)
}

// createTask 创建任务
func (m *TaskManager) createTask(taskInfo *storage.TaskInfo) (*scheduler.Task, error) {
	// 创建任务选项
	options := []scheduler.TaskOption{
		scheduler.WithName(taskInfo.Name),
		scheduler.WithTimeout(time.Duration(taskInfo.Timeout) * time.Second),
		scheduler.WithRetry(taskInfo.RetryTimes),
	}

	// 设置重复间隔
	if taskInfo.Interval > 0 {
		options = append(options, scheduler.WithRepeat(time.Duration(taskInfo.Interval)*time.Second))
	}

	// 设置最大运行次数
	if taskInfo.MaxRuns > 0 {
		options = append(options, scheduler.WithMaxRuns(taskInfo.MaxRuns))
	}

	// 创建任务函数
	var job scheduler.Job
	switch taskInfo.Type {
	case storage.TaskTypeLua:
		// Lua 脚本任务
		job = m.executor.CreateLuaJob(taskInfo.Content)
	case storage.TaskTypeShell:
		// Shell 命令任务
		job = func(ctx context.Context) error {
			cmd := exec.CommandContext(ctx, "cmd", "/C", taskInfo.Content)
			return cmd.Run()
		}
	default:
		return nil, fmt.Errorf("unsupported task type: %s", taskInfo.Type)
	}

	// 添加任务函数
	options = append(options, scheduler.WithJob(job))

	// 添加错误处理
	options = append(options, scheduler.WithErrorHandler(func(err error) {
		// 更新任务错误信息
		taskInfo.LastError = err.Error()
		m.storage.UpdateTaskRunInfo(taskInfo.ID, taskInfo.RunCount, taskInfo.LastRunAt, taskInfo.LastError)
	}))

	// 添加完成回调
	options = append(options, scheduler.WithPostHook(func() {
		// 更新任务运行信息
		taskInfo.RunCount++
		taskInfo.LastRunAt = time.Now()
		m.storage.UpdateTaskRunInfo(taskInfo.ID, taskInfo.RunCount, taskInfo.LastRunAt, taskInfo.LastError)

		// 如果达到最大运行次数，更新状态为已完成
		if taskInfo.MaxRuns > 0 && taskInfo.RunCount >= taskInfo.MaxRuns {
			taskInfo.Status = storage.TaskStatusCompleted
			m.storage.SaveTask(taskInfo)

			// 从任务映射中移除
			m.mutex.Lock()
			delete(m.tasks, taskInfo.ID)
			m.mutex.Unlock()
		}
	}))

	// 创建任务
	return scheduler.NewTask(options...), nil
}

// GetTaskStatus 获取任务状态
func (m *TaskManager) GetTaskStatus(id int64) (storage.TaskStatus, error) {
	taskInfo, err := m.storage.GetTask(id)
	if err != nil {
		return "", err
	}
	return taskInfo.Status, nil
}

// IsTaskRunning 检查任务是否正在运行
func (m *TaskManager) IsTaskRunning(id int64) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.tasks[id]
	return exists
}

// GetRunningTasks 获取所有正在运行的任务
func (m *TaskManager) GetRunningTasks() []int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var ids []int64
	for id := range m.tasks {
		ids = append(ids, id)
	}
	return ids
}
