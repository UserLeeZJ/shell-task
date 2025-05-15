// cmd/shelltask/cli_functions2.go
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
