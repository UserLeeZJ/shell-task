// cmd/shelltask/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/UserLeeZJ/shell-task/cmd/shelltask/lua"
	"github.com/UserLeeZJ/shell-task/cmd/shelltask/manager"
	"github.com/UserLeeZJ/shell-task/cmd/shelltask/storage"
	"github.com/UserLeeZJ/shell-task/cmd/shelltask/ui"
	tea "github.com/charmbracelet/bubbletea"
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

	// 创建 UI 模型
	model := ui.NewModel(sqliteStorage)

	// 创建 Bubbletea 程序
	p := tea.NewProgram(model, tea.WithAltScreen())

	// 运行 UI
	if _, err := p.Run(); err != nil {
		log.Fatalf("启动 UI 失败: %v", err)
	}
}
