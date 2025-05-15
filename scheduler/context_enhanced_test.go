// scheduler/context_enhanced_test.go
package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestContextFilter 测试上下文过滤功能
func TestContextFilter(t *testing.T) {
	// 创建上下文
	ctx := NewTaskContext()

	// 设置值
	ctx.Set("user.name", "Alice")
	ctx.Set("user.age", 30)
	ctx.Set("system.version", "1.0")
	ctx.Set("app.name", "TestApp")

	// 过滤 user 前缀的键
	userValues := ctx.Filter("user.")

	// 验证过滤结果
	if len(userValues) != 2 {
		t.Errorf("Expected 2 user values, got %d", len(userValues))
	}

	if name, ok := userValues["user.name"].(string); !ok || name != "Alice" {
		t.Errorf("Expected user.name to be 'Alice', got %v", userValues["user.name"])
	}

	if age, ok := userValues["user.age"].(int); !ok || age != 30 {
		t.Errorf("Expected user.age to be 30, got %v", userValues["user.age"])
	}
}

// TestContextTransform 测试上下文转换功能
func TestContextTransform(t *testing.T) {
	// 创建上下文
	ctx := NewTaskContext()

	// 设置值
	ctx.Set("name", "Alice")
	ctx.Set("age", 30)

	// 创建转换函数
	transformer := func(key string, value interface{}) (string, interface{}) {
		// 将所有键转换为大写
		newKey := "transformed." + key

		// 如果值是字符串，转换为大写
		if strValue, ok := value.(string); ok {
			return newKey, "TRANSFORMED: " + strValue
		}

		return newKey, value
	}

	// 应用转换
	newCtx := ctx.Transform(transformer)

	// 验证转换结果
	transformedName, nameOk := newCtx.Get("transformed.name")
	if !nameOk {
		t.Error("Expected transformed.name to exist, but it doesn't")
	} else if transformedName != "TRANSFORMED: Alice" {
		t.Errorf("Expected transformed.name to be 'TRANSFORMED: Alice', got %v", transformedName)
	}

	transformedAge, ageOk := newCtx.Get("transformed.age")
	if !ageOk {
		t.Error("Expected transformed.age to exist, but it doesn't")
	} else if transformedAge != 30 {
		t.Errorf("Expected transformed.age to be 30, got %v", transformedAge)
	}
}

// TestContextValidation 测试上下文验证功能
func TestContextValidation(t *testing.T) {
	// 创建上下文
	ctx := NewTaskContext()

	// 设置值
	ctx.Set("name", "Alice")
	ctx.Set("age", -5) // 无效年龄

	// 创建验证器
	validators := map[string]Validator{
		"name": func(key string, value interface{}) error {
			strValue, ok := value.(string)
			if !ok {
				return errors.New("name must be a string")
			}
			if len(strValue) == 0 {
				return errors.New("name cannot be empty")
			}
			return nil
		},
		"age": func(key string, value interface{}) error {
			intValue, ok := value.(int)
			if !ok {
				return errors.New("age must be an integer")
			}
			if intValue < 0 {
				return errors.New("age cannot be negative")
			}
			return nil
		},
	}

	// 验证上下文
	err := ctx.Validate(validators)

	// 验证结果
	if err == nil {
		t.Error("Expected validation to fail, but it succeeded")
	}

	// 修复无效值
	ctx.Set("age", 30)

	// 再次验证
	err = ctx.Validate(validators)

	// 验证结果
	if err != nil {
		t.Errorf("Expected validation to succeed, but it failed: %v", err)
	}
}

// TestRequiredKeys 测试必需键验证功能
func TestRequiredKeys(t *testing.T) {
	// 创建上下文
	ctx := NewTaskContext()

	// 设置部分值
	ctx.Set("name", "Alice")

	// 验证必需的键
	err := ctx.RequiredKeys("name", "age")

	// 验证结果
	if err == nil {
		t.Error("Expected required keys check to fail, but it succeeded")
	}

	// 添加缺失的键
	ctx.Set("age", 30)

	// 再次验证
	err = ctx.RequiredKeys("name", "age")

	// 验证结果
	if err != nil {
		t.Errorf("Expected required keys check to succeed, but it failed: %v", err)
	}
}

// TestTaskGroup 测试任务组上下文共享功能
func TestTaskGroup(t *testing.T) {
	// 创建任务组
	group := NewTaskGroup("TestGroup", nil)

	// 设置组上下文值
	group.SetContextValue("shared", "value")

	// 创建任务
	task1 := NewTask(
		WithName("Task1"),
		WithJob(func(ctx context.Context) error {
			// 获取任务实例
			task := TaskFromContext(ctx)

			// 验证可以访问组上下文值
			value, exists := task.GetContextValue("shared")
			if !exists || value != "value" {
				t.Errorf("Expected shared value 'value', got %v, exists: %v", value, exists)
			}

			// 设置任务特定的上下文值
			task.SetContextValue("task1", "done")

			return nil
		}),
	)

	// 将任务添加到组
	group.AddTask(task1)

	// 运行任务
	task1.Run()

	// 等待任务完成
	time.Sleep(50 * time.Millisecond)

	// 验证任务上下文值
	value, exists := task1.GetContextValue("task1")
	if !exists || value != "done" {
		t.Errorf("Expected task1 value 'done', got %v, exists: %v", value, exists)
	}
}
