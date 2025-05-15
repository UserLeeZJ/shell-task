// scheduler/worker_pool.go
package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// TaskStatus 表示任务的状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 等待执行
	TaskStatusRunning                     // 正在执行
	TaskStatusCompleted                   // 已完成
	TaskStatusFailed                      // 执行失败
	TaskStatusCancelled                   // 已取消
)

// TaskInfo 存储任务的状态信息
type TaskInfo struct {
	Task      *Task      // 任务引用
	Status    TaskStatus // 任务状态
	WorkerID  int        // 执行该任务的工作协程ID
	StartTime time.Time  // 开始执行时间
	EndTime   time.Time  // 结束执行时间
	Error     error      // 执行错误（如果有）
}

// WorkerPool 管理一组工作协程，限制并发执行的任务数量
type WorkerPool struct {
	size       int                // 工作池大小（最大并发数）
	taskQueue  *PriorityQueue     // 优先级任务队列
	taskChan   chan *Task         // 任务通道，用于工作协程获取任务
	wg         sync.WaitGroup     // 等待所有工作协程完成
	ctx        context.Context    // 上下文，用于取消
	cancelFunc context.CancelFunc // 取消函数
	logger     Logger             // 日志记录器
	mutex      sync.Mutex         // 互斥锁，保护共享数据
	running    bool               // 工作池是否正在运行

	// 任务状态跟踪
	tasksMutex sync.RWMutex         // 保护任务状态映射的互斥锁
	tasks      map[string]*TaskInfo // 任务状态映射，键为任务名称

	// 统计信息
	completedTasks int64 // 已完成任务数量
	failedTasks    int64 // 失败任务数量

	// 生命周期回调
	onTaskStart  func(*Task)        // 任务开始执行时的回调
	onTaskFinish func(*Task, error) // 任务完成执行时的回调
}

// WorkerPoolOption 是配置工作池的函数类型
type WorkerPoolOption func(*WorkerPool)

// WithTaskStartCallback 设置任务开始执行时的回调函数
func WithTaskStartCallback(callback func(*Task)) WorkerPoolOption {
	return func(wp *WorkerPool) {
		wp.onTaskStart = callback
	}
}

// WithTaskFinishCallback 设置任务完成执行时的回调函数
func WithTaskFinishCallback(callback func(*Task, error)) WorkerPoolOption {
	return func(wp *WorkerPool) {
		wp.onTaskFinish = callback
	}
}

// NewWorkerPool 创建一个新的工作池
func NewWorkerPool(size int, logger Logger, opts ...WorkerPoolOption) *WorkerPool {
	if size <= 0 {
		size = 1 // 至少有一个工作协程
	}

	if logger == nil {
		logger = defaultLoggerInstance
	}

	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		size:       size,
		taskQueue:  NewPriorityQueue(),
		taskChan:   make(chan *Task, size*2), // 缓冲区大小为工作池大小的两倍
		ctx:        ctx,
		cancelFunc: cancel,
		logger:     logger,
		running:    false,

		// 初始化任务状态跟踪
		tasks: make(map[string]*TaskInfo),

		// 默认回调函数
		onTaskStart: func(t *Task) {
			// 默认实现为空
		},
		onTaskFinish: func(t *Task, err error) {
			// 默认实现为空
		},
	}

	// 应用所有配置项
	for _, opt := range opts {
		opt(wp)
	}

	return wp
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if wp.running {
		return // 已经在运行
	}

	wp.logger.Info("Starting worker pool with %d workers", wp.size)
	wp.running = true

	// 启动调度协程，将任务从优先级队列移动到任务通道
	go wp.scheduler()

	// 启动工作协程
	wp.wg.Add(wp.size)
	for i := 0; i < wp.size; i++ {
		go wp.worker(i)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if !wp.running {
		return // 已经停止
	}

	wp.logger.Info("Stopping worker pool")
	wp.running = false
	wp.cancelFunc()    // 取消所有工作协程
	close(wp.taskChan) // 关闭任务通道
	wp.wg.Wait()       // 等待所有工作协程完成
}

// Submit 提交任务到工作池
func (wp *WorkerPool) Submit(task *Task) {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	if !wp.running {
		wp.logger.Warn("Worker pool is stopped, cannot submit task: %s", task.name)
		return
	}

	// 记录任务状态
	wp.tasksMutex.Lock()
	wp.tasks[task.name] = &TaskInfo{
		Task:      task,
		Status:    TaskStatusPending,
		StartTime: time.Time{}, // 零值表示未开始
	}
	wp.tasksMutex.Unlock()

	// 将任务添加到优先级队列
	wp.taskQueue.Enqueue(task)
	wp.logger.Debug("Task submitted to worker pool: %s (priority: %d)", task.name, task.priority)
}

// GetTaskInfo 获取任务的状态信息
func (wp *WorkerPool) GetTaskInfo(taskName string) (*TaskInfo, bool) {
	wp.tasksMutex.RLock()
	defer wp.tasksMutex.RUnlock()

	info, exists := wp.tasks[taskName]
	return info, exists
}

// GetAllTasksInfo 获取所有任务的状态信息
func (wp *WorkerPool) GetAllTasksInfo() map[string]*TaskInfo {
	wp.tasksMutex.RLock()
	defer wp.tasksMutex.RUnlock()

	// 创建一个副本以避免并发访问问题
	result := make(map[string]*TaskInfo, len(wp.tasks))
	for k, v := range wp.tasks {
		result[k] = v
	}

	return result
}

// GetStats 获取工作池的统计信息
func (wp *WorkerPool) GetStats() (int, int64, int64) {
	wp.tasksMutex.RLock()
	defer wp.tasksMutex.RUnlock()

	pendingTasks := 0
	for _, info := range wp.tasks {
		if info.Status == TaskStatusPending {
			pendingTasks++
		}
	}

	return pendingTasks, atomic.LoadInt64(&wp.completedTasks), atomic.LoadInt64(&wp.failedTasks)
}

// scheduler 是调度协程的主函数，负责将任务从优先级队列移动到任务通道
func (wp *WorkerPool) scheduler() {
	wp.logger.Debug("Scheduler started")

	for {
		// 检查是否已取消
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Scheduler stopped: context canceled")
			return
		default:
			// 继续执行
		}

		// 从优先级队列中取出任务
		task := wp.taskQueue.Dequeue()
		if task == nil {
			// 队列为空，等待一段时间
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 将任务发送到任务通道
		select {
		case <-wp.ctx.Done():
			return
		case wp.taskChan <- task:
			wp.logger.Debug("Task scheduled: %s (priority: %d)", task.name, task.priority)
		}
	}
}

// worker 是工作协程的主函数
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker %d stopped: context canceled", id)
			return
		case task, ok := <-wp.taskChan:
			if !ok {
				wp.logger.Debug("Worker %d stopped: task channel closed", id)
				return
			}

			wp.logger.Debug("Worker %d executing task: %s", id, task.name)

			// 更新任务状态为运行中
			wp.tasksMutex.Lock()
			if info, exists := wp.tasks[task.name]; exists {
				info.Status = TaskStatusRunning
				info.WorkerID = id
				info.StartTime = time.Now()
			}
			wp.tasksMutex.Unlock()

			// 调用任务开始回调
			wp.onTaskStart(task)

			// 创建一个通道来接收任务完成信号
			done := make(chan struct{})
			var taskErr error

			// 启动一个协程来监控任务执行
			go func() {
				// 设置任务完成回调
				originalPostHook := task.postHook
				task.postHook = func() {
					if originalPostHook != nil {
						originalPostHook()
					}
					close(done)
				}

				// 设置任务错误处理器
				originalErrorHandler := task.errorHandler
				task.errorHandler = func(err error) {
					if originalErrorHandler != nil {
						originalErrorHandler(err)
					}
					taskErr = err
				}

				// 执行任务
				task.Run()
			}()

			// 等待任务完成或工作池停止
			select {
			case <-done:
				// 任务正常完成
				wp.tasksMutex.Lock()
				if info, exists := wp.tasks[task.name]; exists {
					if taskErr != nil {
						info.Status = TaskStatusFailed
						info.Error = taskErr
						atomic.AddInt64(&wp.failedTasks, 1)
					} else {
						info.Status = TaskStatusCompleted
						atomic.AddInt64(&wp.completedTasks, 1)
					}
					info.EndTime = time.Now()
				}
				wp.tasksMutex.Unlock()

				// 调用任务完成回调
				wp.onTaskFinish(task, taskErr)

				wp.logger.Debug("Worker %d completed task: %s, error: %v", id, task.name, taskErr)

			case <-wp.ctx.Done():
				// 工作池停止，取消任务
				task.Stop()

				wp.tasksMutex.Lock()
				if info, exists := wp.tasks[task.name]; exists {
					info.Status = TaskStatusCancelled
					info.EndTime = time.Now()
				}
				wp.tasksMutex.Unlock()

				wp.logger.Debug("Worker %d cancelled task: %s due to pool shutdown", id, task.name)
				return
			}
		}
	}
}
