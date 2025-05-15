// storage/sqlite_scan.go
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

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
