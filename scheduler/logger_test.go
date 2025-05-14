package scheduler

import (
	"context"
	"testing"
)

// TestDefaultLogger 测试默认日志记录器
func TestDefaultLogger(t *testing.T) {
	logger := defaultLoggerInstance

	// 这些调用不应该导致 panic
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
}

// TestFuncLogger 测试函数式日志适配器
func TestFuncLogger(t *testing.T) {
	var lastFormat string
	var lastArgs []any

	logFunc := func(format string, args ...any) {
		lastFormat = format
		lastArgs = args
	}

	logger := NewFuncLogger(logFunc)

	// 测试 Debug 方法（默认不输出）
	logger.Debug("Debug message", 1, 2, 3)
	if lastFormat == "Debug message" {
		t.Error("Expected Debug method to not call logFunc, but it did")
	}

	// 测试 Info 方法
	logger.Info("Info message", 1, 2, 3)
	if lastFormat != "Info message" {
		t.Errorf("Expected format to be 'Info message', got '%s'", lastFormat)
	}
	if len(lastArgs) != 3 || lastArgs[0] != 1 || lastArgs[1] != 2 || lastArgs[2] != 3 {
		t.Errorf("Expected args to be [1, 2, 3], got %v", lastArgs)
	}

	// 测试 Warn 方法
	logger.Warn("Warn message", 4, 5, 6)
	if lastFormat != "Warn message" {
		t.Errorf("Expected format to be 'Warn message', got '%s'", lastFormat)
	}
	if len(lastArgs) != 3 || lastArgs[0] != 4 || lastArgs[1] != 5 || lastArgs[2] != 6 {
		t.Errorf("Expected args to be [4, 5, 6], got %v", lastArgs)
	}

	// 测试 Error 方法
	logger.Error("Error message", 7, 8, 9)
	if lastFormat != "Error message" {
		t.Errorf("Expected format to be 'Error message', got '%s'", lastFormat)
	}
	if len(lastArgs) != 3 || lastArgs[0] != 7 || lastArgs[1] != 8 || lastArgs[2] != 9 {
		t.Errorf("Expected args to be [7, 8, 9], got %v", lastArgs)
	}
}

// 自定义测试日志记录器
type testLogger struct {
	debugCalled bool
	infoCalled  bool
	warnCalled  bool
	errorCalled bool
	lastFormat  string
	lastArgs    []any
}

func (l *testLogger) Debug(format string, args ...any) {
	l.debugCalled = true
	l.lastFormat = format
	l.lastArgs = args
}

func (l *testLogger) Info(format string, args ...any) {
	l.infoCalled = true
	l.lastFormat = format
	l.lastArgs = args
}

func (l *testLogger) Warn(format string, args ...any) {
	l.warnCalled = true
	l.lastFormat = format
	l.lastArgs = args
}

func (l *testLogger) Error(format string, args ...any) {
	l.errorCalled = true
	l.lastFormat = format
	l.lastArgs = args
}

// TestTaskWithLogger 测试任务使用自定义日志记录器
func TestTaskWithLogger(t *testing.T) {
	// 这个测试只是验证自定义日志记录器可以被设置
	// 不测试实际的日志调用，因为那依赖于内部实现
	logger := &testLogger{}

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithLogger(logger),
	)

	// 验证日志记录器被正确设置
	if task.logger != logger {
		t.Error("Expected task.logger to be set to the custom logger, but it wasn't")
	}
}

// TestTaskWithLoggerFunc 测试任务使用函数式日志记录器
func TestTaskWithLoggerFunc(t *testing.T) {
	// 这个测试只是验证函数式日志记录器可以被设置
	// 不测试实际的日志调用，因为那依赖于内部实现
	logFunc := func(format string, args ...any) {}

	task := NewTask(
		WithName("TestTask"),
		WithJob(func(ctx context.Context) error {
			return nil
		}),
		WithLoggerFunc(logFunc),
	)

	// 验证日志记录器被正确设置
	if task.logger == nil {
		t.Error("Expected task.logger to be set, but it was nil")
	}

	// 验证日志记录器是 FuncLogger 类型
	_, ok := task.logger.(*FuncLogger)
	if !ok {
		t.Error("Expected task.logger to be a FuncLogger, but it wasn't")
	}
}
