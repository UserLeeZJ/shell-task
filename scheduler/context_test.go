// scheduler/context_test.go
package scheduler

import (
	"context"
	"testing"
	"time"
)

// TestTaskContext 测试任务上下文基本功能
func TestTaskContext(t *testing.T) {
	// 创建上下文
	ctx := NewTaskContext()
	
	// 设置值
	ctx.Set("string", "value")
	ctx.Set("int", 42)
	ctx.Set("bool", true)
	
	// 获取值
	if val, ok := ctx.GetString("string"); !ok || val != "value" {
		t.Errorf("Expected string value 'value', got '%v', exists: %v", val, ok)
	}
	
	if val, ok := ctx.GetInt("int"); !ok || val != 42 {
		t.Errorf("Expected int value 42, got %v, exists: %v", val, ok)
	}
	
	if val, ok := ctx.GetBool("bool"); !ok || !val {
		t.Errorf("Expected bool value true, got %v, exists: %v", val, ok)
	}
	
	// 测试不存在的值
	if _, ok := ctx.Get("nonexistent"); ok {
		t.Error("Expected nonexistent key to return not ok")
	}
}

// TestTaskContextParent 测试任务上下文继承
func TestTaskContextParent(t *testing.T) {
	// 创建父上下文
	parent := NewTaskContext()
	parent.Set("parent", "value")
	parent.Set("override", "parent")
	
	// 创建子上下文
	child := NewTaskContext().WithParent(parent)
	child.Set("child", "value")
	child.Set("override", "child") // 覆盖父上下文的值
	
	// 测试子上下文可以访问父上下文的值
	if val, ok := child.GetString("parent"); !ok || val != "value" {
		t.Errorf("Expected parent value 'value', got '%v', exists: %v", val, ok)
	}
	
	// 测试子上下文自己的值
	if val, ok := child.GetString("child"); !ok || val != "value" {
		t.Errorf("Expected child value 'value', got '%v', exists: %v", val, ok)
	}
	
	// 测试子上下文覆盖父上下文的值
	if val, ok := child.GetString("override"); !ok || val != "child" {
		t.Errorf("Expected override value 'child', got '%v', exists: %v", val, ok)
	}
	
	// 测试父上下文不受子上下文影响
	if val, ok := parent.GetString("override"); !ok || val != "parent" {
		t.Errorf("Expected parent override value 'parent', got '%v', exists: %v", val, ok)
	}
}

// TestTaskWithContext 测试任务上下文传递
func TestTaskWithContext(t *testing.T) {
	// 创建上下文
	taskContext := NewTaskContext()
	taskContext.Set("initial", "value")
	
	// 创建任务
	executed := false
	contextUpdated := false
	
	task := NewTask(
		WithName("ContextTest"),
		WithTaskContext(taskContext),
		WithContextPrep(func(ctx *TaskContext) {
			// 在任务执行前准备上下文
			ctx.Set("prep", "done")
		}),
		WithJob(func(ctx context.Context) error {
			// 在任务执行中使用和更新上下文
			executed = true
			
			// 获取上下文值
			if val, ok := taskContext.GetString("initial"); !ok || val != "value" {
				t.Errorf("Expected initial value 'value', got '%v', exists: %v", val, ok)
			}
			
			if val, ok := taskContext.GetString("prep"); !ok || val != "done" {
				t.Errorf("Expected prep value 'done', got '%v', exists: %v", val, ok)
			}
			
			// 更新上下文
			taskContext.Set("job", "executed")
			contextUpdated = true
			
			return nil
		}),
		WithContextClean(func(ctx *TaskContext) {
			// 在任务执行后清理上下文
			ctx.Set("clean", "done")
		}),
	)
	
	// 运行任务
	task.Run()
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行
	
	// 验证任务已执行
	if !executed {
		t.Error("Expected task to be executed, but it wasn't")
	}
	
	// 验证上下文已更新
	if !contextUpdated {
		t.Error("Expected context to be updated, but it wasn't")
	}
	
	// 验证上下文值
	if val, ok := taskContext.GetString("job"); !ok || val != "executed" {
		t.Errorf("Expected job value 'executed', got '%v', exists: %v", val, ok)
	}
	
	if val, ok := taskContext.GetString("clean"); !ok || val != "done" {
		t.Errorf("Expected clean value 'done', got '%v', exists: %v", val, ok)
	}
}

// TestChainTasks 测试任务链上下文传递
func TestChainTasks(t *testing.T) {
	// 创建第一个任务
	task1 := NewTask(
		WithName("Task1"),
		WithJob(func(ctx context.Context) error {
			// 设置上下文值
			task := TaskFromContext(ctx)
			task.SetContextValue("task1", "executed")
			return nil
		}),
	)
	
	// 创建第二个任务
	task2Executed := false
	task2 := NewTask(
		WithName("Task2"),
		WithJob(func(ctx context.Context) error {
			// 获取上下文值
			task := TaskFromContext(ctx)
			val, ok := task.GetContextValue("task1")
			if !ok || val != "executed" {
				t.Errorf("Expected task1 value 'executed', got '%v', exists: %v", val, ok)
			}
			
			// 设置新的上下文值
			task.SetContextValue("task2", "executed")
			task2Executed = true
			return nil
		}),
	)
	
	// 创建任务链
	tasks := ChainTasks(task1, task2)
	
	// 运行任务链
	for _, task := range tasks {
		task.Run()
	}
	
	time.Sleep(100 * time.Millisecond) // 给任务一点时间执行
	
	// 验证任务2已执行
	if !task2Executed {
		t.Error("Expected task2 to be executed, but it wasn't")
	}
	
	// 验证任务2的上下文包含任务1的值
	val, ok := task2.GetContextValue("task1")
	if !ok || val != "executed" {
		t.Errorf("Expected task1 value 'executed' in task2 context, got '%v', exists: %v", val, ok)
	}
}
