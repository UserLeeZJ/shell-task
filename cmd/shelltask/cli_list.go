// cmd/shelltask/cli_list.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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
