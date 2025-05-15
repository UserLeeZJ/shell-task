// cmd/shelltask/cli_script.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/UserLeeZJ/shell-task/lua"
)

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
