// cmd/shelltask/ui/updates.go
package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/UserLeeZJ/shell-task/cmd/shelltask/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// updateTaskListMode 更新任务列表模式
func (m *ShellTaskModel) updateTaskListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Enter):
		// 查看任务详情
		selectedRow := m.table.SelectedRow()
		if len(selectedRow) > 0 {
			id, err := strconv.ParseInt(selectedRow[0], 10, 64)
			if err != nil {
				m.err = err
				m.statusMsg = errorStyle.Render(fmt.Sprintf("错误: %v", err))
				return m, nil
			}

			for _, task := range m.tasks {
				if task.ID == id {
					m.currentTask = task
					m.mode = taskDetailMode
					return m, nil
				}
			}
		}

	case key.Matches(msg, m.keys.Create):
		// 创建新任务
		m.currentTask = nil
		m.mode = taskCreateMode
		m.initTextInputs()
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		// 刷新任务列表
		return m, m.loadTasks
	}

	// 更新表格
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// updateTaskDetailMode 更新任务详情模式
func (m *ShellTaskModel) updateTaskDetailMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		// 返回任务列表
		m.mode = taskListMode
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		// 编辑任务
		m.mode = taskEditMode
		m.initTextInputs()
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		// 删除任务
		if m.currentTask != nil {
			return m, func() tea.Msg {
				err := m.storage.DeleteTask(m.currentTask.ID)
				if err != nil {
					return errMsg{err}
				}
				m.mode = taskListMode
				m.currentTask = nil
				return statusMsg{msg: successStyle.Render("任务已删除")}
			}
		}

	case key.Matches(msg, m.keys.Run):
		// 运行任务
		if m.currentTask != nil {
			m.mode = taskRunMode
			return m, func() tea.Msg {
				// 更新任务状态
				err := m.storage.UpdateTaskStatus(m.currentTask.ID, storage.TaskStatusRunning)
				if err != nil {
					return errMsg{err}
				}
				m.currentTask.Status = storage.TaskStatusRunning
				return statusMsg{msg: infoStyle.Render("任务已启动")}
			}
		}

	case key.Matches(msg, m.keys.Stop):
		// 停止任务
		if m.currentTask != nil && m.currentTask.Status == storage.TaskStatusRunning {
			return m, func() tea.Msg {
				// 更新任务状态
				err := m.storage.UpdateTaskStatus(m.currentTask.ID, storage.TaskStatusCancelled)
				if err != nil {
					return errMsg{err}
				}
				m.currentTask.Status = storage.TaskStatusCancelled
				return statusMsg{msg: infoStyle.Render("任务已停止")}
			}
		}
	}

	return m, nil
}

// updateTaskEditMode 更新任务编辑模式
func (m *ShellTaskModel) updateTaskEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Cancel):
		// 取消编辑
		if m.mode == taskCreateMode {
			m.mode = taskListMode
		} else {
			m.mode = taskDetailMode
		}
		return m, nil

	case key.Matches(msg, m.keys.Save):
		// 保存任务
		return m, m.saveTask

	case key.Matches(msg, m.keys.NextField):
		// 下一个字段
		m.focusIndex = (m.focusIndex + 1) % len(m.textInputs)
		return m, nil

	case key.Matches(msg, m.keys.PrevField):
		// 上一个字段
		m.focusIndex = (m.focusIndex - 1 + len(m.textInputs)) % len(m.textInputs)
		return m, nil
	}

	// 更新当前文本输入
	if m.focusIndex >= 0 && m.focusIndex < len(m.textInputs) {
		var cmd tea.Cmd
		m.textInputs[m.focusIndex], cmd = m.textInputs[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateTaskRunMode 更新任务运行模式
func (m *ShellTaskModel) updateTaskRunMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		// 返回任务详情
		m.mode = taskDetailMode
		return m, nil

	case key.Matches(msg, m.keys.Stop):
		// 停止任务
		if m.currentTask != nil && m.currentTask.Status == storage.TaskStatusRunning {
			return m, func() tea.Msg {
				// 更新任务状态
				err := m.storage.UpdateTaskStatus(m.currentTask.ID, storage.TaskStatusCancelled)
				if err != nil {
					return errMsg{err}
				}
				m.currentTask.Status = storage.TaskStatusCancelled
				return statusMsg{msg: infoStyle.Render("任务已停止")}
			}
		}
	}

	return m, nil
}

// updateHelpMode 更新帮助模式
func (m *ShellTaskModel) updateHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		// 返回上一个模式
		m.mode = taskListMode
		return m, nil
	}

	return m, nil
}

// saveTask 保存任务
func (m *ShellTaskModel) saveTask() tea.Msg {
	// 解析表单数据
	var task storage.TaskInfo
	if m.currentTask != nil {
		task = *m.currentTask
	} else {
		task.Status = storage.TaskStatusIdle
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
	}

	// 名称
	task.Name = m.textInputs[0].Value()
	if task.Name == "" {
		return errMsg{fmt.Errorf("任务名称不能为空")}
	}

	// 类型
	taskType := m.textInputs[1].Value()
	switch taskType {
	case "lua":
		task.Type = storage.TaskTypeLua
	case "shell":
		task.Type = storage.TaskTypeShell
	default:
		return errMsg{fmt.Errorf("无效的任务类型: %s", taskType)}
	}

	// 内容
	task.Content = m.textInputs[2].Value()
	if task.Content == "" {
		return errMsg{fmt.Errorf("任务内容不能为空")}
	}

	// 间隔
	interval, err := strconv.ParseInt(m.textInputs[3].Value(), 10, 64)
	if err != nil {
		return errMsg{fmt.Errorf("无效的间隔: %v", err)}
	}
	task.Interval = interval

	// 最大运行次数
	maxRuns, err := strconv.Atoi(m.textInputs[4].Value())
	if err != nil {
		return errMsg{fmt.Errorf("无效的最大运行次数: %v", err)}
	}
	task.MaxRuns = maxRuns

	// 重试次数
	retryTimes, err := strconv.Atoi(m.textInputs[5].Value())
	if err != nil {
		return errMsg{fmt.Errorf("无效的重试次数: %v", err)}
	}
	task.RetryTimes = retryTimes

	// 超时
	timeout, err := strconv.ParseInt(m.textInputs[6].Value(), 10, 64)
	if err != nil {
		return errMsg{fmt.Errorf("无效的超时: %v", err)}
	}
	task.Timeout = timeout

	// 描述
	task.Description = m.textInputs[7].Value()

	// 保存任务
	err = m.storage.SaveTask(&task)
	if err != nil {
		return errMsg{err}
	}

	// 更新当前任务
	m.currentTask = &task

	// 返回上一个模式
	if m.mode == taskCreateMode {
		m.mode = taskListMode
		return tasksLoadedMsg{tasks: append(m.tasks, &task)}
	} else {
		m.mode = taskDetailMode
		// 更新任务列表中的任务
		for i, t := range m.tasks {
			if t.ID == task.ID {
				m.tasks[i] = &task
				break
			}
		}
		return statusMsg{msg: successStyle.Render("任务已保存")}
	}
}
