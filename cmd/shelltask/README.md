# Shell Task

Shell Task 是一个基于 Go 语言开发的任务调度器，支持 Lua 脚本和 Shell 命令的执行，并提供了 TUI（终端用户界面）进行管理。

## 功能特点

- 支持 Lua 脚本任务
- 支持 Shell 命令任务
- SQLite 持久化存储任务
- 基于 Bubbletea 的终端用户界面
- 支持任务的创建、编辑、删除、运行和停止
- 支持任务重试、超时和定时执行

## 安装

### 从源码编译

```bash
git clone https://github.com/UserLeeZJ/shell-task.git
cd shell-task
go build -o shelltask.exe cmd/shelltask/main.go
```

## 使用方法

### 命令行参数

```
Shell Task - 任务调度器
用法: shelltask [选项]
选项:
  -db string
        SQLite 数据库路径
  -help
        显示帮助信息
  -no-ui
        不启动 UI 界面
  -scripts string
        Lua 脚本目录
  -version
        显示版本信息
```

### 启动 UI 界面

```bash
shelltask.exe
```

### 守护模式（不启动 UI）

```bash
shelltask.exe -no-ui
```

### 指定数据库和脚本目录

```bash
shelltask.exe -db C:\path\to\tasks.db -scripts C:\path\to\scripts
```

## UI 界面操作

### 任务列表

- `↑/k` - 上移
- `↓/j` - 下移
- `Enter` - 查看任务详情
- `c` - 创建新任务
- `f5` - 刷新任务列表
- `q` - 退出程序

### 任务详情

- `e` - 编辑任务
- `d` - 删除任务
- `r` - 运行任务
- `s` - 停止任务
- `Esc` - 返回任务列表

### 任务编辑

- `Tab` - 下一个字段
- `Shift+Tab` - 上一个字段
- `Ctrl+s` - 保存任务
- `Esc` - 取消编辑

## Lua 脚本示例

```lua
-- hello.lua
print("Hello from Lua!")
sleep(1)
print("This is a Lua script task")
```

## Shell 命令示例

```
echo Hello from Shell!
ping 127.0.0.1 -n 3
```

## 任务类型

### Lua 任务

Lua 任务使用内置的 Lua 解释器执行脚本。支持以下内置函数：

- `print(...)` - 打印信息
- `sleep(seconds)` - 休眠指定秒数

### Shell 任务

Shell 任务使用系统的命令行解释器执行命令。在 Windows 上使用 `cmd /C`，在其他系统上使用 `/bin/sh -c`。

## 数据存储

任务数据存储在 SQLite 数据库中，默认位置为：

- Windows: `%USERPROFILE%\.shelltask\tasks.db`
- Linux/macOS: `$HOME/.shelltask/tasks.db`

## Lua 脚本存储

Lua 脚本默认存储在以下位置：

- Windows: `%USERPROFILE%\.shelltask\scripts`
- Linux/macOS: `$HOME/.shelltask/scripts`

## 依赖项

- [github.com/yuin/gopher-lua](https://github.com/yuin/gopher-lua) - Go 语言的 Lua 解释器
- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite 数据库驱动
- [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - 终端 UI 框架
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - 终端样式库

## 许可证

MIT
