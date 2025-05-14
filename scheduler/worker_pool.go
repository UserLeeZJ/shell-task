// scheduler/worker_pool.go
package scheduler

import (
	"context"
	"sync"
	"time"
)

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
}

// NewWorkerPool 创建一个新的工作池
func NewWorkerPool(size int, logger Logger) *WorkerPool {
	if size <= 0 {
		size = 1 // 至少有一个工作协程
	}

	if logger == nil {
		logger = defaultLoggerInstance
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		size:       size,
		taskQueue:  NewPriorityQueue(),
		taskChan:   make(chan *Task, size*2), // 缓冲区大小为工作池大小的两倍
		ctx:        ctx,
		cancelFunc: cancel,
		logger:     logger,
		running:    false,
	}
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

	// 将任务添加到优先级队列
	wp.taskQueue.Enqueue(task)
	wp.logger.Debug("Task submitted to worker pool: %s (priority: %d)", task.name, task.priority)
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

			// 执行任务
			task.Run()
		}
	}
}
