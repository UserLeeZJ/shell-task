// cmd/shelltask/storage/sqlite.go
package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TaskType 表示任务类型
type TaskType string

const (
	TaskTypeGo   TaskType = "go"   // Go 函数任务
	TaskTypeLua  TaskType = "lua"  // Lua 脚本任务
	TaskTypeShell TaskType = "shell" // Shell 命令任务
)

// TaskStatus 表示任务状态
type TaskStatus string

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

// SQLiteStorage 是基于 SQLite 的任务存储
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage 创建一个新的 SQLite 存储
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	if dbPath == "" {
		// 如果未指定数据库路径，使用默认路径
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dbDir := filepath.Join(homeDir, ".shelltask")
			os.MkdirAll(dbDir, 0755)
			dbPath = filepath.Join(dbDir, "tasks.db")
		} else {
			dbPath = "tasks.db"
		}
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 初始化存储
	storage := &SQLiteStorage{db: db}
	if err := storage.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// Close 关闭存储
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// initialize 初始化数据库表
func (s *SQLiteStorage) initialize() error {
	// 创建任务表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL,
			interval INTEGER NOT NULL,
			max_runs INTEGER NOT NULL,
			retry_times INTEGER NOT NULL,
			timeout INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_run_at TIMESTAMP,
			run_count INTEGER NOT NULL,
			last_error TEXT,
			description TEXT,
			tags TEXT,
			options TEXT
		)
	`)
	if err != nil {
		return err
	}

	// 创建索引
	_, err = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_name ON tasks(name)`)
	if err != nil {
		return err
	}

	return nil
}

// SaveTask 保存任务
func (s *SQLiteStorage) SaveTask(task *TaskInfo) error {
	if task == nil {
		return errors.New("task is nil")
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(task.Tags)
	if err != nil {
		return err
	}

	now := time.Now()
	if task.ID == 0 {
		// 新任务
		task.CreatedAt = now
		task.UpdatedAt = now

		result, err := s.db.Exec(`
			INSERT INTO tasks (
				name, type, content, status, interval, max_runs, retry_times, timeout,
				created_at, updated_at, run_count, last_error, description, tags, options
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			task.Name, task.Type, task.Content, task.Status, task.Interval, task.MaxRuns,
			task.RetryTimes, task.Timeout, task.CreatedAt, task.UpdatedAt, task.RunCount,
			task.LastError, task.Description, string(tagsJSON), task.Options,
		)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		task.ID = id
	} else {
		// 更新任务
		task.UpdatedAt = now

		_, err := s.db.Exec(`
			UPDATE tasks SET
				name = ?, type = ?, content = ?, status = ?, interval = ?, max_runs = ?,
				retry_times = ?, timeout = ?, updated_at = ?, last_run_at = ?, run_count = ?,
				last_error = ?, description = ?, tags = ?, options = ?
			WHERE id = ?
		`,
			task.Name, task.Type, task.Content, task.Status, task.Interval, task.MaxRuns,
			task.RetryTimes, task.Timeout, task.UpdatedAt, task.LastRunAt, task.RunCount,
			task.LastError, task.Description, string(tagsJSON), task.Options, task.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTask 获取任务
func (s *SQLiteStorage) GetTask(id int64) (*TaskInfo, error) {
	row := s.db.QueryRow(`SELECT * FROM tasks WHERE id = ?`, id)
	return s.scanTask(row)
}

// GetTaskByName 根据名称获取任务
func (s *SQLiteStorage) GetTaskByName(name string) (*TaskInfo, error) {
	row := s.db.QueryRow(`SELECT * FROM tasks WHERE name = ?`, name)
	return s.scanTask(row)
}

// ListTasks 列出所有任务
func (s *SQLiteStorage) ListTasks() ([]*TaskInfo, error) {
	rows, err := s.db.Query(`SELECT * FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*TaskInfo
	for rows.Next() {
		task, err := s.scanTaskRows(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (s *SQLiteStorage) DeleteTask(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

// UpdateTaskStatus 更新任务状态
func (s *SQLiteStorage) UpdateTaskStatus(id int64, status TaskStatus) error {
	_, err := s.db.Exec(`
		UPDATE tasks SET
			status = ?,
			updated_at = ?
		WHERE id = ?
	`, status, time.Now(), id)
	return err
}

// UpdateTaskRunInfo 更新任务运行信息
func (s *SQLiteStorage) UpdateTaskRunInfo(id int64, runCount int, lastRunAt time.Time, lastError string) error {
	_, err := s.db.Exec(`
		UPDATE tasks SET
			run_count = ?,
			last_run_at = ?,
			last_error = ?,
			updated_at = ?
		WHERE id = ?
	`, runCount, lastRunAt, lastError, time.Now(), id)
	return err
}

// scanTask 扫描单行任务数据
func (s *SQLiteStorage) scanTask(row *sql.Row) (*TaskInfo, error) {
	var task TaskInfo
	var tagsJSON string
	var lastRunAtNull sql.NullTime

	err := row.Scan(
		&task.ID, &task.Name, &task.Type, &task.Content, &task.Status,
		&task.Interval, &task.MaxRuns, &task.RetryTimes, &task.Timeout,
		&task.CreatedAt, &task.UpdatedAt, &lastRunAtNull, &task.RunCount,
		&task.LastError, &task.Description, &tagsJSON, &task.Options,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	if lastRunAtNull.Valid {
		task.LastRunAt = lastRunAtNull.Time
	}

	// 解析标签
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
			return nil, err
		}
	}

	return &task, nil
}

// scanTaskRows 扫描多行任务数据
func (s *SQLiteStorage) scanTaskRows(rows *sql.Rows) (*TaskInfo, error) {
	var task TaskInfo
	var tagsJSON string
	var lastRunAtNull sql.NullTime

	err := rows.Scan(
		&task.ID, &task.Name, &task.Type, &task.Content, &task.Status,
		&task.Interval, &task.MaxRuns, &task.RetryTimes, &task.Timeout,
		&task.CreatedAt, &task.UpdatedAt, &lastRunAtNull, &task.RunCount,
		&task.LastError, &task.Description, &tagsJSON, &task.Options,
	)
	if err != nil {
		return nil, err
	}

	if lastRunAtNull.Valid {
		task.LastRunAt = lastRunAtNull.Time
	}

	// 解析标签
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &task.Tags); err != nil {
			return nil, err
		}
	}

	return &task, nil
}
