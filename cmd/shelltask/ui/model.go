// cmd/shelltask/ui/model.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/UserLeeZJ/shell-task/cmd/shelltask/storage"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 定义视图模式
type viewMode int

const (
	taskListMode viewMode = iota
	taskDetailMode
	taskEditMode
	taskCreateMode
	taskRunMode
	helpMode
)

// 定义样式
var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	statusBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#666666")).Padding(0, 1)
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#25A065"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	infoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
)

// 定义键盘映射
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Enter     key.Binding
	Back      key.Binding
	Quit      key.Binding
	Help      key.Binding
	Create    key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Run       key.Binding
	Stop      key.Binding
	Refresh   key.Binding
	Save      key.Binding
	Cancel    key.Binding
	NextField key.Binding
	PrevField key.Binding
}

// 创建默认键盘映射
func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "上移"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "下移"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "左移"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "右移"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "选择"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "返回"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "退出"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "帮助"),
		),
		Create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "创建"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "编辑"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "删除"),
		),
		Run: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "运行"),
		),
		Stop: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "停止"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("f5"),
			key.WithHelp("f5", "刷新"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "保存"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "取消"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "下一项"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "上一项"),
		),
	}
}

// ShellTaskModel 是应用程序的主模型
type ShellTaskModel struct {
	keys       keyMap
	help       help.Model
	table      table.Model
	textInputs []textinput.Model
	storage    *storage.SQLiteStorage
	tasks      []*storage.TaskInfo
	currentTask *storage.TaskInfo
	mode       viewMode
	width      int
	height     int
	err        error
	statusMsg  string
	focusIndex int
}

// NewModel 创建一个新的模型
func NewModel(storage *storage.SQLiteStorage) *ShellTaskModel {
	keys := newKeyMap()
	helpModel := help.New()
	helpModel.ShowAll = false

	// 创建表格
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "名称", Width: 20},
		{Title: "类型", Width: 10},
		{Title: "状态", Width: 10},
		{Title: "间隔", Width: 10},
		{Title: "运行次数", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// 设置表格样式
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#25A065")).
		Background(lipgloss.Color("0")).
		Bold(true)
	t.SetStyles(s)

	return &ShellTaskModel{
		keys:    keys,
		help:    helpModel,
		table:   t,
		storage: storage,
		mode:    taskListMode,
	}
}

// Init 初始化模型
func (m *ShellTaskModel) Init() tea.Cmd {
	return m.loadTasks
}

// loadTasks 加载任务列表
func (m *ShellTaskModel) loadTasks() tea.Msg {
	tasks, err := m.storage.ListTasks()
	if err != nil {
		return errMsg{err}
	}

	return tasksLoadedMsg{tasks: tasks}
}

// 定义消息类型
type (
	errMsg struct {
		err error
	}

	tasksLoadedMsg struct {
		tasks []*storage.TaskInfo
	}

	statusMsg struct {
		msg string
	}
)

// Update 更新模型
func (m *ShellTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// 处理全局键盘事件
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		// 根据当前模式处理键盘事件
		switch m.mode {
		case taskListMode:
			return m.updateTaskListMode(msg)
		case taskDetailMode:
			return m.updateTaskDetailMode(msg)
		case taskEditMode, taskCreateMode:
			return m.updateTaskEditMode(msg)
		case taskRunMode:
			return m.updateTaskRunMode(msg)
		case helpMode:
			return m.updateHelpMode(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(m.height - 10) // 留出空间给标题、状态栏和帮助
		m.help.Width = msg.Width

	case errMsg:
		m.err = msg.err
		m.statusMsg = errorStyle.Render(fmt.Sprintf("错误: %v", msg.err))
		return m, nil

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.updateTaskTable()
		return m, nil

	case statusMsg:
		m.statusMsg = msg.msg
		return m, nil
	}

	// 更新子组件
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	// 更新文本输入
	if m.mode == taskEditMode || m.mode == taskCreateMode {
		for i := range m.textInputs {
			if i == m.focusIndex {
				m.textInputs[i].Focus()
			} else {
				m.textInputs[i].Blur()
			}
			m.textInputs[i], cmd = m.textInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View 渲染视图
func (m *ShellTaskModel) View() string {
	switch m.mode {
	case taskListMode:
		return m.taskListView()
	case taskDetailMode:
		return m.taskDetailView()
	case taskEditMode, taskCreateMode:
		return m.taskEditView()
	case taskRunMode:
		return m.taskRunView()
	case helpMode:
		return m.helpView()
	default:
		return "未知视图模式"
	}
}

// 更新任务表格
func (m *ShellTaskModel) updateTaskTable() {
	rows := []table.Row{}
	for _, task := range m.tasks {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", task.ID),
			task.Name,
			string(task.Type),
			string(task.Status),
			fmt.Sprintf("%ds", task.Interval),
			fmt.Sprintf("%d/%d", task.RunCount, task.MaxRuns),
		})
	}
	m.table.SetRows(rows)
}
