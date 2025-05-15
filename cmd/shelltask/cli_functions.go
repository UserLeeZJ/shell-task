// cmd/shelltask/cli_functions.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/UserLeeZJ/shell-task/lua"
	"github.com/UserLeeZJ/shell-task/manager"
	"github.com/UserLeeZJ/shell-task/storage"
)

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
