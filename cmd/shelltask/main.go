// cmd/shelltask/shelltask.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
