// scheduler/context.go
package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// TaskContext 任务上下文，用于在任务之间传递数据
type TaskContext struct {
	values map[string]interface{}
	mutex  sync.RWMutex
	parent *TaskContext // 父上下文，用于继承
}

// NewTaskContext 创建新的任务上下文
func NewTaskContext() *TaskContext {
	return &TaskContext{
		values: make(map[string]interface{}),
	}
}

// WithParent 设置父上下文
func (tc *TaskContext) WithParent(parent *TaskContext) *TaskContext {
	tc.parent = parent
	return tc
}

// Set 设置上下文值
func (tc *TaskContext) Set(key string, value interface{}) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.values[key] = value
}

// Get 获取上下文值
func (tc *TaskContext) Get(key string) (interface{}, bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 先从当前上下文查找
	if value, exists := tc.values[key]; exists {
		return value, true
	}

	// 如果没有找到且有父上下文，则从父上下文查找
	if tc.parent != nil {
		return tc.parent.Get(key)
	}

	return nil, false
}

// GetString 获取字符串类型的上下文值
func (tc *TaskContext) GetString(key string) (string, bool) {
	value, exists := tc.Get(key)
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetInt 获取整数类型的上下文值
func (tc *TaskContext) GetInt(key string) (int, bool) {
	value, exists := tc.Get(key)
	if !exists {
		return 0, false
	}

	i, ok := value.(int)
	return i, ok
}

// GetBool 获取布尔类型的上下文值
func (tc *TaskContext) GetBool(key string) (bool, bool) {
	value, exists := tc.Get(key)
	if !exists {
		return false, false
	}

	b, ok := value.(bool)
	return b, ok
}

// GetFloat 获取浮点类型的上下文值
func (tc *TaskContext) GetFloat(key string) (float64, bool) {
	value, exists := tc.Get(key)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetAll 获取所有上下文值
func (tc *TaskContext) GetAll() map[string]interface{} {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 创建结果映射
	result := make(map[string]interface{})

	// 如果有父上下文，先获取父上下文的所有值
	if tc.parent != nil {
		parentValues := tc.parent.GetAll()
		for k, v := range parentValues {
			result[k] = v
		}
	}

	// 添加当前上下文的值，覆盖父上下文的同名值
	for k, v := range tc.values {
		result[k] = v
	}

	return result
}

// Filter 根据前缀过滤上下文值
func (tc *TaskContext) Filter(prefix string) map[string]interface{} {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 创建结果映射
	result := make(map[string]interface{})

	// 获取所有值
	allValues := tc.GetAll()

	// 过滤前缀匹配的键
	for k, v := range allValues {
		if strings.HasPrefix(k, prefix) {
			result[k] = v
		}
	}

	return result
}

// Transform 转换上下文值
func (tc *TaskContext) Transform(transformer func(key string, value interface{}) (string, interface{})) *TaskContext {
	// 创建新的上下文
	newContext := NewTaskContext()

	// 获取所有值（这里已经加锁了）
	allValues := tc.GetAll()

	// 应用转换函数
	for k, v := range allValues {
		newKey, newValue := transformer(k, v)
		newContext.Set(newKey, newValue)
	}

	return newContext
}

// CopyTo 将上下文值复制到另一个上下文
func (tc *TaskContext) CopyTo(target *TaskContext, overwrite bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 获取所有值
	allValues := tc.GetAll()

	// 复制值
	for k, v := range allValues {
		if !overwrite {
			// 如果不覆盖且目标上下文已有该键，则跳过
			if _, exists := target.Get(k); exists {
				continue
			}
		}

		target.Set(k, v)
	}
}

// Validator 上下文验证器函数类型
type Validator func(key string, value interface{}) error

// Validate 验证上下文值
func (tc *TaskContext) Validate(validators map[string]Validator) error {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 获取所有值
	allValues := tc.GetAll()

	// 应用验证器
	for k, v := range allValues {
		if validator, exists := validators[k]; exists {
			if err := validator(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// RequiredKeys 验证必需的键是否存在
func (tc *TaskContext) RequiredKeys(keys ...string) error {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// 获取所有值
	allValues := tc.GetAll()

	// 检查必需的键
	for _, key := range keys {
		if _, exists := allValues[key]; !exists {
			return fmt.Errorf("required key not found: %s", key)
		}
	}

	return nil
}

// Clear 清除所有上下文值
func (tc *TaskContext) Clear() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.values = make(map[string]interface{})
}

// taskContextKey 是用于在 context.Context 中存储任务的键
type taskContextKey struct{}

// WithTaskInContext 将任务添加到上下文中
func WithTaskInContext(ctx context.Context, task *Task) context.Context {
	return context.WithValue(ctx, taskContextKey{}, task)
}

// TaskFromContext 从上下文中获取任务
func TaskFromContext(ctx context.Context) *Task {
	task, _ := ctx.Value(taskContextKey{}).(*Task)
	return task
}
