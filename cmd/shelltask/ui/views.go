// cmd/shelltask/ui/views.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/UserLeeZJ/shell-task/cmd/shelltask/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// taskListView 任务列表视图
func (m *ShellTaskModel) taskListView() string {
	var sb strings.Builder

	// 标题
	title := titleStyle.Render("Shell Task 管理器")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// 表格
	sb.WriteString(m.table.View())
	sb.WriteString("\n\n")

	// 状态栏
	status := m.statusMsg
	if status == "" {
		status = fmt.Sprintf("共 %d 个任务", len(m.tasks))
	}
	sb.WriteString(statusBarStyle.Render(status))
	sb.WriteString("\n\n")

	// 帮助
	sb.WriteString(m.help.View(m.keys))

	return sb.String()
}

// taskDetailView 任务详情视图
func (m *ShellTaskModel) taskDetailView() string {
	var sb strings.Builder

	if m.currentTask == nil {
		return "未选择任务"
	}

	// 标题
	title := titleStyle.Render(fmt.Sprintf("任务详情: %s", m.currentTask.Name))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// 详情
	detailStyle := lipgloss.NewStyle().Width(m.width - 4).Padding(0, 2)
	
	sb.WriteString(detailStyle.Render(fmt.Sprintf("ID: %d", m.currentTask.ID)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("名称: %s", m.currentTask.Name)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("类型: %s", m.currentTask.Type)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("状态: %s", m.currentTask.Status)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("间隔: %d 秒", m.currentTask.Interval)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("最大运行次数: %d", m.currentTask.MaxRuns)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("重试次数: %d", m.currentTask.RetryTimes)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("超时: %d 秒", m.currentTask.Timeout)))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("创建时间: %s", m.currentTask.CreatedAt.Format(time.RFC3339))))
	sb.WriteString("\n")
	sb.WriteString(detailStyle.Render(fmt.Sprintf("更新时间: %s", m.currentTask.UpdatedAt.Format(time.RFC3339))))
	sb.WriteString("\n")
	
	if !m.currentTask.LastRunAt.IsZero() {
		sb.WriteString(detailStyle.Render(fmt.Sprintf("上次运行: %s", m.currentTask.LastRunAt.Format(time.RFC3339))))
	} else {
		sb.WriteString(detailStyle.Render("上次运行: 从未运行"))
	}
	sb.WriteString("\n")
	
	sb.WriteString(detailStyle.Render(fmt.Sprintf("运行次数: %d", m.currentTask.RunCount)))
	sb.WriteString("\n")
	
	if m.currentTask.LastError != "" {
		sb.WriteString(detailStyle.Render(fmt.Sprintf("上次错误: %s", m.currentTask.LastError)))
	} else {
		sb.WriteString(detailStyle.Render("上次错误: 无"))
	}
	sb.WriteString("\n")
	
	if m.currentTask.Description != "" {
		sb.WriteString(detailStyle.Render(fmt.Sprintf("描述: %s", m.currentTask.Description)))
	} else {
		sb.WriteString(detailStyle.Render("描述: 无"))
	}
	sb.WriteString("\n")
	
	if len(m.currentTask.Tags) > 0 {
		sb.WriteString(detailStyle.Render(fmt.Sprintf("标签: %s", strings.Join(m.currentTask.Tags, ", "))))
	} else {
		sb.WriteString(detailStyle.Render("标签: 无"))
	}
	sb.WriteString("\n\n")
	
	// 内容
	sb.WriteString(detailStyle.Render("内容:"))
	sb.WriteString("\n")
	contentStyle := lipgloss.NewStyle().Width(m.width - 8).Padding(0, 4).BorderStyle(lipgloss.RoundedBorder())
	sb.WriteString(contentStyle.Render(m.currentTask.Content))
	sb.WriteString("\n\n")

	// 状态栏
	sb.WriteString(statusBarStyle.Render(m.statusMsg))
	sb.WriteString("\n\n")

	// 帮助
	sb.WriteString(m.help.View(m.keys))

	return sb.String()
}

// taskEditView 任务编辑视图
func (m *ShellTaskModel) taskEditView() string {
	var sb strings.Builder

	// 标题
	var title string
	if m.mode == taskCreateMode {
		title = titleStyle.Render("创建新任务")
	} else {
		title = titleStyle.Render(fmt.Sprintf("编辑任务: %s", m.currentTask.Name))
	}
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// 表单
	for i, input := range m.textInputs {
		if i == m.focusIndex {
			sb.WriteString(selectedItemStyle.Render("> " + input.View()))
		} else {
			sb.WriteString("  " + input.View())
		}
		sb.WriteString("\n\n")
	}

	// 状态栏
	sb.WriteString(statusBarStyle.Render(m.statusMsg))
	sb.WriteString("\n\n")

	// 帮助
	sb.WriteString(m.help.View(m.keys))

	return sb.String()
}

// taskRunView 任务运行视图
func (m *ShellTaskModel) taskRunView() string {
	var sb strings.Builder

	if m.currentTask == nil {
		return "未选择任务"
	}

	// 标题
	title := titleStyle.Render(fmt.Sprintf("运行任务: %s", m.currentTask.Name))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// 状态
	statusStyle := lipgloss.NewStyle().Width(m.width - 4).Padding(0, 2)
	sb.WriteString(statusStyle.Render(fmt.Sprintf("状态: %s", m.currentTask.Status)))
	sb.WriteString("\n")
	sb.WriteString(statusStyle.Render(fmt.Sprintf("运行次数: %d/%d", m.currentTask.RunCount, m.currentTask.MaxRuns)))
	sb.WriteString("\n")
	
	if !m.currentTask.LastRunAt.IsZero() {
		sb.WriteString(statusStyle.Render(fmt.Sprintf("上次运行: %s", m.currentTask.LastRunAt.Format(time.RFC3339))))
	} else {
		sb.WriteString(statusStyle.Render("上次运行: 从未运行"))
	}
	sb.WriteString("\n")
	
	if m.currentTask.LastError != "" {
		sb.WriteString(statusStyle.Render(fmt.Sprintf("上次错误: %s", m.currentTask.LastError)))
	} else {
		sb.WriteString(statusStyle.Render("上次错误: 无"))
	}
	sb.WriteString("\n\n")

	// 日志输出
	sb.WriteString(statusStyle.Render("日志输出:"))
	sb.WriteString("\n")
	logStyle := lipgloss.NewStyle().Width(m.width - 8).Padding(0, 4).BorderStyle(lipgloss.RoundedBorder())
	sb.WriteString(logStyle.Render("任务运行中..."))
	sb.WriteString("\n\n")

	// 状态栏
	sb.WriteString(statusBarStyle.Render(m.statusMsg))
	sb.WriteString("\n\n")

	// 帮助
	sb.WriteString(m.help.View(m.keys))

	return sb.String()
}

// helpView 帮助视图
func (m *ShellTaskModel) helpView() string {
	var sb strings.Builder

	// 标题
	title := titleStyle.Render("帮助")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// 帮助内容
	helpStyle := lipgloss.NewStyle().Width(m.width - 4).Padding(0, 2)
	sb.WriteString(helpStyle.Render("Shell Task 是一个任务调度器，可以执行各种类型的任务。"))
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("支持的任务类型:"))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("- Lua 脚本任务: 执行 Lua 脚本"))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("- Shell 命令任务: 执行 Shell 命令"))
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("键盘快捷键:"))
	sb.WriteString("\n")
	sb.WriteString(m.help.View(m.keys))

	// 状态栏
	sb.WriteString("\n\n")
	sb.WriteString(statusBarStyle.Render("按 ESC 返回"))

	return sb.String()
}

// 初始化文本输入
func (m *ShellTaskModel) initTextInputs() {
	m.textInputs = make([]textinput.Model, 8)
	var t textinput.Model

	// 名称
	t = textinput.New()
	t.Placeholder = "任务名称"
	t.Focus()
	t.CharLimit = 50
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(m.currentTask.Name)
	}
	m.textInputs[0] = t

	// 类型
	t = textinput.New()
	t.Placeholder = "任务类型 (lua/shell)"
	t.CharLimit = 10
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(string(m.currentTask.Type))
	} else {
		t.SetValue("lua")
	}
	m.textInputs[1] = t

	// 内容
	t = textinput.New()
	t.Placeholder = "任务内容 (脚本内容或命令)"
	t.CharLimit = 1000
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(m.currentTask.Content)
	}
	m.textInputs[2] = t

	// 间隔
	t = textinput.New()
	t.Placeholder = "重复间隔 (秒)"
	t.CharLimit = 10
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(fmt.Sprintf("%d", m.currentTask.Interval))
	} else {
		t.SetValue("0")
	}
	m.textInputs[3] = t

	// 最大运行次数
	t = textinput.New()
	t.Placeholder = "最大运行次数 (0表示无限)"
	t.CharLimit = 10
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(fmt.Sprintf("%d", m.currentTask.MaxRuns))
	} else {
		t.SetValue("1")
	}
	m.textInputs[4] = t

	// 重试次数
	t = textinput.New()
	t.Placeholder = "重试次数"
	t.CharLimit = 10
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(fmt.Sprintf("%d", m.currentTask.RetryTimes))
	} else {
		t.SetValue("0")
	}
	m.textInputs[5] = t

	// 超时
	t = textinput.New()
	t.Placeholder = "超时 (秒)"
	t.CharLimit = 10
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(fmt.Sprintf("%d", m.currentTask.Timeout))
	} else {
		t.SetValue("0")
	}
	m.textInputs[6] = t

	// 描述
	t = textinput.New()
	t.Placeholder = "描述"
	t.CharLimit = 200
	t.Width = 40
	if m.currentTask != nil {
		t.SetValue(m.currentTask.Description)
	}
	m.textInputs[7] = t

	m.focusIndex = 0
}
