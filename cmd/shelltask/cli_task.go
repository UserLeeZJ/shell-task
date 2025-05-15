// cmd/shelltask/cli_task.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/UserLeeZJ/shell-task/storage"
)

// createTask 创建新任务
func createTask(s *storage.SQLiteStorage) {
	scanner := bufio.NewScanner(os.Stdin)

	// 创建任务
	task := new(storage.TaskInfo)
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

	if err := s.SaveTask(task); err != nil {
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
