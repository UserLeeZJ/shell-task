// scheduler/dependency_test.go
package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestDependsOn 测试基本的依赖关系设置
func TestDependsOn(t *testing.T) {
	// 创建两个任务
	task1 := NewTask(WithName("Task1"))
	task2 := NewTask(WithName("Task2"))

	// 设置依赖关系
	task2.DependsOn(task1)

	// 验证依赖关系
	deps := task2.GetDependencies()
	if len(deps) != 1 || deps[0] != task1 {
		t.Errorf("Expected task2 to depend on task1, got %v", deps)
	}

	// 验证依赖状态
	if task2.AreDependenciesMet() {
		t.Error("Expected dependencies not met, but they are")
	}

	// 完成依赖任务
	task1.setState(TaskStateCompleted)

	// 等待状态变化传播
	time.Sleep(10 * time.Millisecond)

	// 验证依赖状态
	if !task2.AreDependenciesMet() {
		t.Error("Expected dependencies met, but they are not")
	}
}

// TestDependenciesCallback 测试依赖满足时的回调
func TestDependenciesCallback(t *testing.T) {
	// 创建两个任务
	task1 := NewTask(WithName("Task1"))
	task2 := NewTask(WithName("Task2"))

	// 设置回调（在设置依赖关系之前）
	callbackCalled := false
	task2.WithOnDependenciesMet(func() {
		callbackCalled = true
	})

	// 设置依赖关系
	task2.DependsOn(task1)

	// 完成依赖任务
	task1.setState(TaskStateCompleted)

	// 等待回调执行
	time.Sleep(50 * time.Millisecond)

	// 验证回调是否被调用
	if !callbackCalled {
		t.Error("Expected callback to be called, but it wasn't")
	}
}

// TestTaskExecution 测试任务执行顺序
func TestTaskExecution(t *testing.T) {
	// 创建一个互斥锁和执行顺序记录
	var mu sync.Mutex
	executionOrder := make([]string, 0)

	// 创建三个任务
	task1 := NewTask(
		WithName("Task1"),
		WithJob(func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "Task1")
			mu.Unlock()
			return nil
		}),
	)

	task2 := NewTask(
		WithName("Task2"),
		WithJob(func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "Task2")
			mu.Unlock()
			return nil
		}),
	)

	task3 := NewTask(
		WithName("Task3"),
		WithJob(func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "Task3")
			mu.Unlock()
			return nil
		}),
	)

	// 设置依赖关系: task2 依赖 task1, task3 依赖 task2
	task2.DependsOn(task1)
	task3.DependsOn(task2)

	// 创建工作池
	pool := NewWorkerPool(1, nil)
	pool.Start()

	// 提交任务（顺序不重要，因为有依赖关系）
	pool.Submit(task3) // 提交最后一个任务
	pool.Submit(task2) // 提交中间任务
	pool.Submit(task1) // 提交第一个任务

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)

	// 停止工作池
	pool.Stop()

	// 验证执行顺序
	expectedOrder := []string{"Task1", "Task2", "Task3"}
	mu.Lock()
	defer mu.Unlock()

	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 tasks to be executed, got %d", len(executionOrder))
	}

	for i, taskName := range executionOrder {
		if i < len(expectedOrder) && taskName != expectedOrder[i] {
			t.Errorf("Expected task %s at position %d, got %s", expectedOrder[i], i, taskName)
		}
	}
}

// TestSimplifiedAPI 测试简化的依赖关系API
func TestSimplifiedAPI(t *testing.T) {
	// 创建一个互斥锁和执行顺序记录
	var mu sync.Mutex
	executionOrder := make([]string, 0)

	// 创建任务
	task1 := NewTask(
		WithName("Task1"),
		WithJob(func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "Task1")
			mu.Unlock()
			return nil
		}),
	)

	task2 := NewTask(
		WithName("Task2"),
		WithJob(func(ctx context.Context) error {
			mu.Lock()
			executionOrder = append(executionOrder, "Task2")
			mu.Unlock()
			return nil
		}),
	)

	// 使用 RunAfter
	RunAfter(task2, task1)

	// 创建工作池
	pool := NewWorkerPool(1, nil)
	pool.Start()

	// 提交任务
	pool.Submit(task2)
	pool.Submit(task1)

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)

	// 停止工作池
	pool.Stop()

	// 验证执行顺序
	expectedOrder := []string{"Task1", "Task2"}
	mu.Lock()
	defer mu.Unlock()

	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 tasks to be executed, got %d", len(executionOrder))
	}

	for i, taskName := range executionOrder {
		if i < len(expectedOrder) && taskName != expectedOrder[i] {
			t.Errorf("Expected task %s at position %d, got %s", expectedOrder[i], i, taskName)
		}
	}
}
