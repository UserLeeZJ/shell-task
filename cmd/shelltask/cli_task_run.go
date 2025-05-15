// cmd/shelltask/cli_task_run.go
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
