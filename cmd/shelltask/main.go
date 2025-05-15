// cmd/shelltask/main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/UserLeeZJ/shell-task/lua"
	"github.com/UserLeeZJ/shell-task/manager"
	"github.com/UserLeeZJ/shell-task/storage"
)

// Version 由构建时的 -ldflags 设置
var Version = "dev"

func main() {
	// 解析命令行参数
	var (
		dbPath    string
		scriptDir string
		noUI      bool
		help      bool
		version   bool
	)

	flag.StringVar(&dbPath, "db", "", "SQLite 数据库路径")
	flag.StringVar(&scriptDir, "scripts", "", "Lua 脚本目录")
	flag.BoolVar(&noUI, "no-ui", false, "不启动 UI 界面")
	flag.BoolVar(&help, "help", false, "显示帮助信息")
	flag.BoolVar(&version, "version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if version {
		fmt.Printf("Shell Task 版本: %s\n", Version)
		return
	}

	// 显示帮助信息
	if help {
		fmt.Println("Shell Task - 任务调度器")
		fmt.Println("用法: shelltask [选项]")
		fmt.Println("选项:")
		flag.PrintDefaults()
		return
	}

	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Shell Task 版本: %s", Version)

	// 如果未指定数据库路径，使用默认路径
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dbDir := filepath.Join(homeDir, ".shelltask")
			os.MkdirAll(dbDir, 0755)
			dbPath = filepath.Join(dbDir, "tasks.db")
		} else {
			dbPath = "tasks.db"
		}
	}

	// 创建 SQLite 存储
	sqliteStorage, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatalf("创建 SQLite 存储失败: %v", err)
	}
	defer sqliteStorage.Close()

	// 创建 Lua 执行器
	luaExecutor := lua.NewExecutor(scriptDir)

	// 创建任务管理器
	taskManager := manager.NewTaskManager(sqliteStorage, luaExecutor)

	// 启动任务管理器
	if err := taskManager.Start(); err != nil {
		log.Fatalf("启动任务管理器失败: %v", err)
	}
	defer taskManager.Stop()

	// 如果不启动 UI 界面，则进入守护模式
	if noUI {
		log.Println("进入守护模式，按 Ctrl+C 退出")

		// 等待中断信号
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("收到中断信号，正在退出...")
		return
	}

	// 使用简单的命令行界面
	runCLI(sqliteStorage, taskManager, luaExecutor)
}

// runCLI 运行命令行界面
func runCLI(storage *storage.SQLiteStorage, manager *manager.TaskManager, executor *lua.Executor) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== Shell Task 命令行界面 ===")
		fmt.Println("1. 列出所有任务")
		fmt.Println("2. 查看任务详情")
		fmt.Println("3. 创建新任务")
		fmt.Println("4. 编辑任务")
		fmt.Println("5. 删除任务")
		fmt.Println("6. 运行任务")
		fmt.Println("7. 停止任务")
		fmt.Println("8. 列出 Lua 脚本")
		fmt.Println("9. 创建 Lua 脚本")
		fmt.Println("0. 退出")
		fmt.Print("\n请选择操作: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			listTasks(storage)
		case "2":
			viewTask(storage)
		case "3":
			createTask(storage)
		case "4":
			editTask(storage)
		case "5":
			deleteTask(storage)
		case "6":
			runTask(storage, manager)
		case "7":
			stopTask(storage, manager)
		case "8":
			listScripts(executor)
		case "9":
			createScript(executor)
		case "0":
			fmt.Println("正在退出...")
			return
		default:
			fmt.Println("无效的选择，请重试")
		}
	}
}

// listTasks 列出所有任务
func listTasks(storage *storage.SQLiteStorage) {
	tasks, err := storage.ListTasks()
	if err != nil {
		fmt.Printf("获取任务列表失败: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("没有任务")
		return
	}

	fmt.Println("\n=== 任务列表 ===")
	fmt.Printf("%-5s %-20s %-10s %-10s %-10s %-10s\n", "ID", "名称", "类型", "状态", "间隔", "运行次数")
	fmt.Println(strings.Repeat("-", 70))

	for _, task := range tasks {
		fmt.Printf("%-5d %-20s %-10s %-10s %-10d %-10d\n",
			task.ID, task.Name, task.Type, task.Status, task.Interval, task.RunCount)
	}
}

// viewTask 查看任务详情
func viewTask(storage *storage.SQLiteStorage) {
	fmt.Print("请输入任务 ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的 ID: %v\n", err)
		return
	}

	task, err := storage.GetTask(id)
	if err != nil {
		fmt.Printf("获取任务失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 任务详情 ===")
	fmt.Printf("ID: %d\n", task.ID)
	fmt.Printf("名称: %s\n", task.Name)
	fmt.Printf("类型: %s\n", task.Type)
	fmt.Printf("状态: %s\n", task.Status)
	fmt.Printf("间隔: %d 秒\n", task.Interval)
	fmt.Printf("最大运行次数: %d\n", task.MaxRuns)
	fmt.Printf("重试次数: %d\n", task.RetryTimes)
	fmt.Printf("超时: %d 秒\n", task.Timeout)
	fmt.Printf("创建时间: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("更新时间: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))

	if !task.LastRunAt.IsZero() {
		fmt.Printf("上次运行: %s\n", task.LastRunAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("上次运行: 从未运行")
	}

	fmt.Printf("运行次数: %d\n", task.RunCount)

	if task.LastError != "" {
		fmt.Printf("上次错误: %s\n", task.LastError)
	} else {
		fmt.Println("上次错误: 无")
	}

	if task.Description != "" {
		fmt.Printf("描述: %s\n", task.Description)
	} else {
		fmt.Println("描述: 无")
	}

	if len(task.Tags) > 0 {
		fmt.Printf("标签: %s\n", strings.Join(task.Tags, ", "))
	} else {
		fmt.Println("标签: 无")
	}

	fmt.Println("\n内容:")
	fmt.Println(task.Content)
}

// createTask 创建新任务
func createTask(storage *storage.SQLiteStorage) {
	scanner := bufio.NewScanner(os.Stdin)

	task := &storage.TaskInfo{}
	task.Status = "idle"

	fmt.Print("任务名称: ")
	scanner.Scan()
	task.Name = scanner.Text()
	if task.Name == "" {
		fmt.Println("任务名称不能为空")
		return
	}

	fmt.Print("任务类型 (lua/shell): ")
	scanner.Scan()
	taskType := scanner.Text()
	switch taskType {
	case "lua":
		task.Type = "lua"
	case "shell":
		task.Type = "shell"
	default:
		fmt.Println("无效的任务类型")
		return
	}

	fmt.Print("任务内容 (脚本内容或命令): ")
	scanner.Scan()
	task.Content = scanner.Text()
	if task.Content == "" {
		fmt.Println("任务内容不能为空")
		return
	}

	fmt.Print("重复间隔 (秒): ")
	scanner.Scan()
	interval, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		fmt.Printf("无效的间隔: %v\n", err)
		return
	}
	task.Interval = interval

	fmt.Print("最大运行次数 (0表示无限): ")
	scanner.Scan()
	maxRuns, err := strconv.Atoi(scanner.Text())
	if err != nil {
		fmt.Printf("无效的最大运行次数: %v\n", err)
		return
	}
	task.MaxRuns = maxRuns

	fmt.Print("重试次数: ")
	scanner.Scan()
	retryTimes, err := strconv.Atoi(scanner.Text())
	if err != nil {
		fmt.Printf("无效的重试次数: %v\n", err)
		return
	}
	task.RetryTimes = retryTimes

	fmt.Print("超时 (秒): ")
	scanner.Scan()
	timeout, err := strconv.ParseInt(scanner.Text(), 10, 64)
	if err != nil {
		fmt.Printf("无效的超时: %v\n", err)
		return
	}
	task.Timeout = timeout

	fmt.Print("描述: ")
	scanner.Scan()
	task.Description = scanner.Text()

	fmt.Print("标签 (用逗号分隔): ")
	scanner.Scan()
	tagsStr := scanner.Text()
	if tagsStr != "" {
		task.Tags = strings.Split(tagsStr, ",")
		for i := range task.Tags {
			task.Tags[i] = strings.TrimSpace(task.Tags[i])
		}
	}

	if err := storage.SaveTask(task); err != nil {
		fmt.Printf("保存任务失败: %v\n", err)
		return
	}

	fmt.Printf("任务已创建，ID: %d\n", task.ID)
}

// editTask 编辑任务
func editTask(storage *storage.SQLiteStorage) {
	fmt.Print("请输入任务 ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的 ID: %v\n", err)
		return
	}

	task, err := storage.GetTask(id)
	if err != nil {
		fmt.Printf("获取任务失败: %v\n", err)
		return
	}

	fmt.Printf("编辑任务: %s (ID: %d)\n", task.Name, task.ID)
	fmt.Println("(直接按回车保持原值不变)")

	fmt.Printf("任务名称 [%s]: ", task.Name)
	scanner.Scan()
	if name := scanner.Text(); name != "" {
		task.Name = name
	}

	fmt.Printf("任务类型 [%s]: ", task.Type)
	scanner.Scan()
	if taskType := scanner.Text(); taskType != "" {
		switch taskType {
		case "lua":
			task.Type = "lua"
		case "shell":
			task.Type = "shell"
		default:
			fmt.Println("无效的任务类型，保持原值不变")
		}
	}

	fmt.Printf("任务内容 [%s...]: ", truncateString(task.Content, 20))
	scanner.Scan()
	if content := scanner.Text(); content != "" {
		task.Content = content
	}

	fmt.Printf("重复间隔 [%d]: ", task.Interval)
	scanner.Scan()
	if intervalStr := scanner.Text(); intervalStr != "" {
		interval, err := strconv.ParseInt(intervalStr, 10, 64)
		if err != nil {
			fmt.Printf("无效的间隔: %v，保持原值不变\n", err)
		} else {
			task.Interval = interval
		}
	}

	fmt.Printf("最大运行次数 [%d]: ", task.MaxRuns)
	scanner.Scan()
	if maxRunsStr := scanner.Text(); maxRunsStr != "" {
		maxRuns, err := strconv.Atoi(maxRunsStr)
		if err != nil {
			fmt.Printf("无效的最大运行次数: %v，保持原值不变\n", err)
		} else {
			task.MaxRuns = maxRuns
		}
	}

	fmt.Printf("重试次数 [%d]: ", task.RetryTimes)
	scanner.Scan()
	if retryTimesStr := scanner.Text(); retryTimesStr != "" {
		retryTimes, err := strconv.Atoi(retryTimesStr)
		if err != nil {
			fmt.Printf("无效的重试次数: %v，保持原值不变\n", err)
		} else {
			task.RetryTimes = retryTimes
		}
	}

	fmt.Printf("超时 [%d]: ", task.Timeout)
	scanner.Scan()
	if timeoutStr := scanner.Text(); timeoutStr != "" {
		timeout, err := strconv.ParseInt(timeoutStr, 10, 64)
		if err != nil {
			fmt.Printf("无效的超时: %v，保持原值不变\n", err)
		} else {
			task.Timeout = timeout
		}
	}

	fmt.Printf("描述 [%s]: ", task.Description)
	scanner.Scan()
	if description := scanner.Text(); description != "" {
		task.Description = description
	}

	fmt.Printf("标签 [%s]: ", strings.Join(task.Tags, ", "))
	scanner.Scan()
	if tagsStr := scanner.Text(); tagsStr != "" {
		task.Tags = strings.Split(tagsStr, ",")
		for i := range task.Tags {
			task.Tags[i] = strings.TrimSpace(task.Tags[i])
		}
	}

	if err := storage.SaveTask(task); err != nil {
		fmt.Printf("保存任务失败: %v\n", err)
		return
	}

	fmt.Println("任务已更新")
}

// deleteTask 删除任务
func deleteTask(storage *storage.SQLiteStorage) {
	fmt.Print("请输入任务 ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的 ID: %v\n", err)
		return
	}

	fmt.Print("确认删除? (y/n): ")
	scanner.Scan()
	confirm := scanner.Text()
	if confirm != "y" && confirm != "Y" {
		fmt.Println("已取消")
		return
	}

	if err := storage.DeleteTask(id); err != nil {
		fmt.Printf("删除任务失败: %v\n", err)
		return
	}

	fmt.Println("任务已删除")
}

// runTask 运行任务
func runTask(storage *storage.SQLiteStorage, manager *manager.TaskManager) {
	fmt.Print("请输入任务 ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的 ID: %v\n", err)
		return
	}

	if manager.IsTaskRunning(id) {
		fmt.Println("任务已经在运行中")
		return
	}

	if err := manager.StartTask(id); err != nil {
		fmt.Printf("启动任务失败: %v\n", err)
		return
	}

	fmt.Println("任务已启动")
}

// stopTask 停止任务
func stopTask(storage *storage.SQLiteStorage, manager *manager.TaskManager) {
	fmt.Print("请输入任务 ID: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	idStr := scanner.Text()

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		fmt.Printf("无效的 ID: %v\n", err)
		return
	}

	if !manager.IsTaskRunning(id) {
		fmt.Println("任务未在运行")
		return
	}

	if err := manager.StopTask(id); err != nil {
		fmt.Printf("停止任务失败: %v\n", err)
		return
	}

	fmt.Println("任务已停止")
}

// listScripts 列出 Lua 脚本
func listScripts(executor *lua.Executor) {
	scripts, err := executor.ListScripts()
	if err != nil {
		fmt.Printf("获取脚本列表失败: %v\n", err)
		return
	}

	if len(scripts) == 0 {
		fmt.Println("没有脚本")
		return
	}

	fmt.Println("\n=== Lua 脚本列表 ===")
	for i, script := range scripts {
		fmt.Printf("%d. %s\n", i+1, script)
	}
}

// createScript 创建 Lua 脚本
func createScript(executor *lua.Executor) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("脚本名称: ")
	scanner.Scan()
	name := scanner.Text()
	if name == "" {
		fmt.Println("脚本名称不能为空")
		return
	}

	if !strings.HasSuffix(name, ".lua") {
		name = name + ".lua"
	}

	fmt.Println("请输入脚本内容 (输入 EOF 结束):")
	var contentBuilder strings.Builder
	for {
		scanner.Scan()
		line := scanner.Text()
		if line == "EOF" {
			break
		}
		contentBuilder.WriteString(line)
		contentBuilder.WriteString("\n")
	}

	content := contentBuilder.String()
	if content == "" {
		fmt.Println("脚本内容不能为空")
		return
	}

	if err := executor.SaveScript(name, content); err != nil {
		fmt.Printf("保存脚本失败: %v\n", err)
		return
	}

	fmt.Printf("脚本 %s 已保存\n", name)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
