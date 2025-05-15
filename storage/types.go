// storage/types.go
package storage

import (
	"time"
)

// TaskType 表示任务类型
type TaskType string

// 任务类型常量
const (
	TaskTypeGo    TaskType = "go"    // Go 函数任务
	TaskTypeLua   TaskType = "lua"   // Lua 脚本任务
	TaskTypeShell TaskType = "shell" // Shell 命令任务
)

// TaskStatus 表示任务状态
type TaskStatus string

// 任务状态常量
const (
	TaskStatusIdle       TaskStatus = "idle"       // 空闲
	TaskStatusRunning    TaskStatus = "running"    // 运行中
	TaskStatusPaused     TaskStatus = "paused"     // 暂停
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
	TaskStatusCancelled  TaskStatus = "cancelled"  // 已取消
)

// TaskInfo 表示任务信息
type TaskInfo struct {
	ID          int64      `json:"id"`           // 任务ID
	Name        string     `json:"name"`         // 任务名称
	Type        TaskType   `json:"type"`         // 任务类型
	Content     string     `json:"content"`      // 任务内容（脚本内容或命令）
	Status      TaskStatus `json:"status"`       // 任务状态
	Interval    int64      `json:"interval"`     // 重复间隔（秒）
	MaxRuns     int        `json:"max_runs"`     // 最大运行次数
	RetryTimes  int        `json:"retry_times"`  // 重试次数
	Timeout     int64      `json:"timeout"`      // 超时时间（秒）
	CreatedAt   time.Time  `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`   // 更新时间
	LastRunAt   time.Time  `json:"last_run_at"`  // 上次运行时间
	RunCount    int        `json:"run_count"`    // 运行次数
	LastError   string     `json:"last_error"`   // 上次错误
	Description string     `json:"description"`  // 任务描述
	Tags        []string   `json:"tags"`         // 标签
	Options     string     `json:"options"`      // 其他选项（JSON格式）
}
