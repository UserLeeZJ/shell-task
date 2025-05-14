package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestNewTask 测试创建新任务
func TestNewTask(t *testing.T) {
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
	)

	if task.name != "TestTask" {
		t.Errorf("Expected task name to be 'TestTask', got '%s'", task.name)
	}

	if task.job == nil {
		t.Error("Expected task job to be set, got nil")
	}

	if task.priority != PriorityNormal {
		t.Errorf("Expected task priority to be %d, got %d", PriorityNormal, task.priority)
	}
}

// TestTaskRun 测试任务运行
func TestTaskRun(t *testing.T) {
	executed := false
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			executed = true
			return nil
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if !executed {
		t.Error("Expected task to be executed, but it wasn't")
	}

	if task.GetRunCount() != 1 {
		t.Errorf("Expected run count to be 1, got %d", task.GetRunCount())
	}
}

// TestTaskStop 测试任务停止
func TestTaskStop(t *testing.T) {
	stopCalled := false
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			// 等待取消
			<-ctx.Done()
			stopCalled = true
			return ctx.Err()
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间开始执行
	task.Stop()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间响应停止

	if !stopCalled {
		t.Error("Expected task to be stopped, but it wasn't")
	}
}

// TestTaskRepeat 测试任务重复执行
func TestTaskRepeat(t *testing.T) {
	count := 0
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			count++
			return nil
		}),
		WithRepeat(50*time.Millisecond), // 每50毫秒执行一次
		WithMaxRuns(3),                  // 最多执行3次
	)

	task.Run()
	time.Sleep(200 * time.Millisecond) // 给任务一点时间执行

	if count != 3 {
		t.Errorf("Expected task to be executed 3 times, got %d", count)
	}

	if task.GetRunCount() != 3 {
		t.Errorf("Expected run count to be 3, got %d", task.GetRunCount())
	}
}

// TestTaskRetry 测试任务重试
func TestTaskRetry(t *testing.T) {
	attempts := 0
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			attempts++
			return errors.New("test error")
		}),
		WithRetry(2), // 失败后重试2次
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if attempts != 3 { // 1次初始尝试 + 2次重试
		t.Errorf("Expected task to be attempted 3 times, got %d", attempts)
	}
}

// TestTaskTimeout 测试任务超时
func TestTaskTimeout(t *testing.T) {
	timedOut := false
	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			// 尝试运行5秒
			select {
			case <-time.After(5 * time.Second):
				return nil
			case <-ctx.Done():
				timedOut = true
				return ctx.Err()
			}
		}),
		WithTimeout(100*time.Millisecond), // 设置100毫秒超时
	)

	task.Run()
	time.Sleep(200 * time.Millisecond) // 给任务一点时间执行和超时

	if !timedOut {
		t.Error("Expected task to time out, but it didn't")
	}
}

// TestTaskErrorHandler 测试错误处理器
func TestTaskErrorHandler(t *testing.T) {
	handlerCalled := false
	expectedErr := errors.New("test error")
	var actualErr error

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return expectedErr
		}),
		WithErrorHandler(func(err error) {
			handlerCalled = true
			actualErr = err
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if !handlerCalled {
		t.Error("Expected error handler to be called, but it wasn't")
	}

	if actualErr != expectedErr {
		t.Errorf("Expected error to be '%v', got '%v'", expectedErr, actualErr)
	}
}

// TestTaskRecover 测试恢复钩子
func TestTaskRecover(t *testing.T) {
	recoverCalled := false
	var recovered any

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			panic("test panic")
		}),
		WithRecover(func(r any) {
			recoverCalled = true
			recovered = r
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if !recoverCalled {
		t.Error("Expected recover hook to be called, but it wasn't")
	}

	if recovered != "test panic" {
		t.Errorf("Expected recovered value to be 'test panic', got '%v'", recovered)
	}
}

// TestTaskHooks 测试钩子函数
func TestTaskHooks(t *testing.T) {
	preHookCalled := false
	postHookCalled := false

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithPreHook(func() {
			preHookCalled = true
		}),
		WithPostHook(func() {
			postHookCalled = true
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if !preHookCalled {
		t.Error("Expected pre-hook to be called, but it wasn't")
	}

	if !postHookCalled {
		t.Error("Expected post-hook to be called, but it wasn't")
	}
}

// TestTaskMetricCollector 测试指标收集器
func TestTaskMetricCollector(t *testing.T) {
	collectorCalled := false
	var result JobResult

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithMetricCollector(func(res JobResult) {
			collectorCalled = true
			result = res
		}),
	)

	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行

	if !collectorCalled {
		t.Error("Expected metric collector to be called, but it wasn't")
	}

	if result.Name != "TestTask" {
		t.Errorf("Expected result name to be 'TestTask', got '%s'", result.Name)
	}

	if !result.Success {
		t.Error("Expected result to be successful, but it wasn't")
	}

	if result.Err != nil {
		t.Errorf("Expected result error to be nil, got '%v'", result.Err)
	}
}
