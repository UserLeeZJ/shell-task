// lua/executor.go
package lua

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Executor 是 Lua 脚本执行器
type Executor struct {
	scriptDir string
	modules   map[string]lua.LGFunction
	mutex     sync.Mutex
}

// NewExecutor 创建一个新的 Lua 执行器
func NewExecutor(scriptDir string) *Executor {
	if scriptDir == "" {
		// 如果未指定脚本目录，使用默认目录
		homeDir, err := os.UserHomeDir()
		if err == nil {
			scriptDir = filepath.Join(homeDir, ".shelltask", "scripts")
		} else {
			scriptDir = "scripts"
		}
	}

	// 确保脚本目录存在
	os.MkdirAll(scriptDir, 0755)

	return &Executor{
		scriptDir: scriptDir,
		modules:   make(map[string]lua.LGFunction),
	}
}

// RegisterModule 注册一个 Lua 模块
func (e *Executor) RegisterModule(name string, loader lua.LGFunction) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.modules[name] = loader
}

// ExecuteString 执行 Lua 脚本字符串
func (e *Executor) ExecuteString(ctx context.Context, script string) error {
	L := e.newState()
	defer L.Close()

	// 设置上下文
	L.SetContext(ctx)

	// 执行脚本
	return L.DoString(script)
}

// ExecuteFile 执行 Lua 脚本文件
func (e *Executor) ExecuteFile(ctx context.Context, filename string) error {
	// 如果文件名不是绝对路径，则在脚本目录中查找
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(e.scriptDir, filename)
	}

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", filename)
	}

	// 读取脚本文件
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// 执行脚本
	return e.ExecuteString(ctx, string(content))
}

// ListScripts 列出脚本目录中的所有 Lua 脚本
func (e *Executor) ListScripts() ([]string, error) {
	files, err := os.ReadDir(e.scriptDir)
	if err != nil {
		return nil, err
	}

	var scripts []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".lua") {
			scripts = append(scripts, file.Name())
		}
	}

	return scripts, nil
}

// SaveScript 保存 Lua 脚本到文件
func (e *Executor) SaveScript(name string, content string) error {
	if !strings.HasSuffix(name, ".lua") {
		name = name + ".lua"
	}

	filename := filepath.Join(e.scriptDir, name)
	return os.WriteFile(filename, []byte(content), 0644)
}

// DeleteScript 删除 Lua 脚本文件
func (e *Executor) DeleteScript(name string) error {
	if !strings.HasSuffix(name, ".lua") {
		name = name + ".lua"
	}

	filename := filepath.Join(e.scriptDir, name)
	return os.Remove(filename)
}

// newState 创建一个新的 Lua 状态
func (e *Executor) newState() *lua.LState {
	L := lua.NewState()

	// 注册模块
	for name, loader := range e.modules {
		L.PreloadModule(name, loader)
	}

	// 注册全局函数
	e.registerGlobalFunctions(L)

	return L
}

// registerGlobalFunctions 注册全局函数
func (e *Executor) registerGlobalFunctions(L *lua.LState) {
	// 注册 print 函数
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			fmt.Print(L.Get(i).String())
			if i != top {
				fmt.Print(" ")
			}
		}
		fmt.Println()
		return 0
	}))

	// 注册 sleep 函数
	L.SetGlobal("sleep", L.NewFunction(func(L *lua.LState) int {
		// 获取参数
		seconds := L.CheckNumber(1)

		// 获取上下文
		ctx := L.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// 创建定时器
		timer := time.NewTimer(time.Duration(seconds) * time.Second)
		defer timer.Stop()

		// 等待定时器或上下文取消
		select {
		case <-timer.C:
			// 正常返回
			return 0
		case <-ctx.Done():
			// 上下文取消
			L.RaiseError("execution canceled")
			return 0
		}
	}))
}

// CreateLuaJob 创建一个执行 Lua 脚本的任务函数
func (e *Executor) CreateLuaJob(script string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := e.ExecuteString(ctx, script)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("lua script error: %w", err)
		}
		return nil
	}
}

// CreateLuaFileJob 创建一个执行 Lua 脚本文件的任务函数
func (e *Executor) CreateLuaFileJob(filename string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := e.ExecuteFile(ctx, filename)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("lua script error: %w", err)
		}
		return nil
	}
}
