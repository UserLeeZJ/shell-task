package scheduler

import (
	"context"
	"testing"
)

// TestNewPriorityQueue 测试创建新优先级队列
func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	if pq.Len() != 0 {
		t.Errorf("Expected new priority queue to be empty, got length %d", pq.Len())
	}

	if !pq.IsEmpty() {
		t.Error("Expected new priority queue to be empty, but it wasn't")
	}
}

// TestPriorityQueueEnqueueDequeue 测试入队和出队
func TestPriorityQueueEnqueueDequeue(t *testing.T) {
	pq := NewPriorityQueue()

	// 创建一个任务
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)

	// 入队
	pq.Enqueue(task)

	if pq.Len() != 1 {
		t.Errorf("Expected priority queue length to be 1 after enqueue, got %d", pq.Len())
	}

	if pq.IsEmpty() {
		t.Error("Expected priority queue to not be empty after enqueue, but it was")
	}

	// 出队
	dequeuedTask := pq.Dequeue()

	if dequeuedTask != task {
		t.Error("Expected dequeued task to be the same as enqueued task, but it wasn't")
	}

	if pq.Len() != 0 {
		t.Errorf("Expected priority queue length to be 0 after dequeue, got %d", pq.Len())
	}

	if !pq.IsEmpty() {
		t.Error("Expected priority queue to be empty after dequeue, but it wasn't")
	}

	// 从空队列出队
	dequeuedTask = pq.Dequeue()
	if dequeuedTask != nil {
		t.Errorf("Expected dequeued task from empty queue to be nil, got %v", dequeuedTask)
	}
}

// TestPriorityQueuePriority 测试优先级排序
func TestPriorityQueuePriority(t *testing.T) {
	pq := NewPriorityQueue()

	// 创建低优先级任务
	lowTask := NewTask(
		WithName("LowTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPriority(PriorityLow),
	)

	// 创建普通优先级任务
	normalTask := NewTask(
		WithName("NormalTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPriority(PriorityNormal),
	)

	// 创建高优先级任务
	highTask := NewTask(
		WithName("HighTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPriority(PriorityHigh),
	)

	// 按优先级从低到高入队
	pq.Enqueue(lowTask)
	pq.Enqueue(normalTask)
	pq.Enqueue(highTask)

	// 出队，应该按优先级从高到低
	task1 := pq.Dequeue()
	task2 := pq.Dequeue()
	task3 := pq.Dequeue()

	if task1 != highTask {
		t.Errorf("Expected first dequeued task to be high priority task, got %v", task1.name)
	}

	if task2 != normalTask {
		t.Errorf("Expected second dequeued task to be normal priority task, got %v", task2.name)
	}

	if task3 != lowTask {
		t.Errorf("Expected third dequeued task to be low priority task, got %v", task3.name)
	}
}

// TestPriorityQueueConcurrency 测试并发安全性
func TestPriorityQueueConcurrency(t *testing.T) {
	pq := NewPriorityQueue()

	// 创建一个任务
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)

	// 并发入队和出队
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			pq.Enqueue(task)
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 50; i++ {
			pq.Dequeue()
		}
		done <- struct{}{}
	}()

	// 等待两个协程完成
	<-done
	<-done

	// 检查队列长度 - 注意：由于并发操作，实际长度可能不确定
	// 我们只检查队列不为空
	if pq.IsEmpty() {
		t.Error("Expected priority queue to not be empty after concurrent operations, but it was")
	}

	// 清空队列
	for pq.Len() > 0 {
		pq.Dequeue()
	}

	if !pq.IsEmpty() {
		t.Error("Expected priority queue to be empty after clearing, but it wasn't")
	}
}
