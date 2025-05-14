// scheduler/logger.go
package scheduler

// Logger 定义了日志接口，支持不同级别的日志记录
type Logger interface {
	// Debug 记录调试级别的日志
	Debug(format string, args ...any)

	// Info 记录信息级别的日志
	Info(format string, args ...any)

	// Warn 记录警告级别的日志
	Warn(format string, args ...any)

	// Error 记录错误级别的日志
	Error(format string, args ...any)
}

// defaultLogger 是默认的日志实现，使用标准库的 log 包
type defaultLogger struct{}

func (l *defaultLogger) Debug(format string, args ...any) {
	// 默认实现中，Debug 级别的日志不输出
}

func (l *defaultLogger) Info(format string, args ...any) {
	// 使用标准库的 log 包记录信息
	stdLog("[INFO] "+format, args...)
}

func (l *defaultLogger) Warn(format string, args ...any) {
	stdLog("[WARN] "+format, args...)
}

func (l *defaultLogger) Error(format string, args ...any) {
	stdLog("[ERROR] "+format, args...)
}

// 全局默认日志实例
var defaultLoggerInstance = &defaultLogger{}

// FuncLogger 是一个适配器，将单一日志函数转换为 Logger 接口
// 用于兼容旧的日志函数
type FuncLogger struct {
	logFunc func(format string, args ...any)
}

func (l *FuncLogger) Debug(format string, args ...any) {
	// 默认不输出 Debug 级别日志
}

func (l *FuncLogger) Info(format string, args ...any) {
	l.logFunc(format, args...)
}

func (l *FuncLogger) Warn(format string, args ...any) {
	l.logFunc(format, args...)
}

func (l *FuncLogger) Error(format string, args ...any) {
	l.logFunc(format, args...)
}

// NewFuncLogger 创建一个新的 FuncLogger
func NewFuncLogger(logFunc func(format string, args ...any)) Logger {
	return &FuncLogger{logFunc: logFunc}
}
