// examples/context-enhanced-example/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("增强上下文传递示例")

	// 示例1：任务组上下文共享
	log.Println("\n=== 示例1：任务组上下文共享 ===")
	taskGroupExample()

	// 示例2：上下文过滤和转换
	log.Println("\n=== 示例2：上下文过滤和转换 ===")
	contextFilterTransformExample()

	// 示例3：上下文验证
	log.Println("\n=== 示例3：上下文验证 ===")
	contextValidationExample()

	// 示例4：依赖任务之间的上下文传递
	log.Println("\n=== 示例4：依赖任务之间的上下文传递 ===")
	dependencyContextExample()
}

// 示例1：任务组上下文共享
func taskGroupExample() {
	// 创建任务组
	group := task.NewDefaultTaskGroup("数据处理组")

	// 设置组共享上下文
	group.SetContextValue("config.timeout", 5000)
	group.SetContextValue("config.retries", 3)
	group.SetContextValue("data.source", "database")

	// 创建第一个任务
	task1 := task.TaskWithContextMap("数据读取", func(ctx context.Context, data map[string]interface{}) error {
		// 获取组共享配置
		timeout := data["config.timeout"]
		retries := data["config.retries"]
		source := data["data.source"]

		log.Printf("任务1：使用配置 timeout=%v, retries=%v, source=%v", timeout, retries, source)

		// 设置任务特定的上下文值
		data["result"] = "读取的数据"
		data["timestamp"] = time.Now().Format(time.RFC3339)

		return nil
	})

	// 创建第二个任务
	task2 := task.TaskWithContextMap("数据处理", func(ctx context.Context, data map[string]interface{}) error {
		// 获取组共享配置
		timeout := data["config.timeout"]

		log.Printf("任务2：使用配置 timeout=%v", timeout)

		// 获取上一个任务的结果
		if result, ok := data["result"].(string); ok {
			log.Printf("任务2：处理数据 '%s'", result)

			// 更新上下文数据
			data["processed"] = fmt.Sprintf("已处理: %s", result)
		}

		return nil
	})

	// 将任务添加到组
	group.AddTasks(task1, task2)

	// 运行所有任务
	group.RunAll()

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	// 获取组上下文数据
	ctx := group.GetContext()
	allData := ctx.GetAll()

	log.Printf("任务组完成，共享上下文数据：%v", allData)
}

// 示例2：上下文过滤和转换
func contextFilterTransformExample() {
	// 创建一个任务，使用上下文过滤
	task1 := task.NewTaskBuilder("过滤任务").
		WithMapContextJob(func(ctx context.Context, data map[string]interface{}) error {
			log.Println("过滤任务：开始执行")

			// 设置多种类型的上下文数据
			data["user.name"] = "张三"
			data["user.age"] = 30
			data["system.version"] = "1.0"
			data["app.name"] = "示例应用"

			log.Printf("过滤任务：设置了多种上下文数据 %v", data)

			return nil
		}).
		Run()

	// 等待任务完成
	time.Sleep(50 * time.Millisecond)

	// 获取任务上下文
	ctx := task1.GetContext()

	// 过滤用户相关的上下文数据
	userContext := ctx.Filter("user.")
	log.Printf("用户相关的上下文数据：%v", userContext)

	// 过滤系统相关的上下文数据
	systemContext := ctx.Filter("system.")
	log.Printf("系统相关的上下文数据：%v", systemContext)

	// 创建一个转换函数
	transformer := func(key string, value interface{}) (string, interface{}) {
		// 将所有键转换为大写
		newKey := "transformed." + key

		// 如果值是字符串，添加前缀
		if strValue, ok := value.(string); ok {
			return newKey, "转换后: " + strValue
		}

		return newKey, value
	}

	// 应用转换
	transformedContext := ctx.Transform(transformer)
	log.Printf("转换后的上下文数据：%v", transformedContext.GetAll())
}

// 示例3：上下文验证
func contextValidationExample() {
	// 创建一个任务，使用上下文验证
	task1 := task.NewTaskBuilder("验证任务").
		WithMapContextJob(func(ctx context.Context, data map[string]interface{}) error {
			log.Println("验证任务：开始执行")

			// 设置上下文数据
			data["name"] = "张三"
			data["age"] = 30
			data["email"] = "zhangsan@example.com"

			log.Printf("验证任务：设置了上下文数据 %v", data)

			return nil
		}).
		WithRequiredContextKeys("name", "age", "email"). // 设置必需的键
		Run()

	// 等待任务完成
	time.Sleep(50 * time.Millisecond)

	// 获取任务上下文
	ctx := task1.GetContext()

	// 创建验证器
	validators := map[string]task.Validator{
		"name": func(key string, value interface{}) error {
			strValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("name 必须是字符串")
			}
			if len(strValue) == 0 {
				return fmt.Errorf("name 不能为空")
			}
			return nil
		},
		"age": func(key string, value interface{}) error {
			intValue, ok := value.(int)
			if !ok {
				return fmt.Errorf("age 必须是整数")
			}
			if intValue < 0 || intValue > 120 {
				return fmt.Errorf("age 必须在 0-120 之间")
			}
			return nil
		},
	}

	// 验证上下文
	if err := ctx.Validate(validators); err != nil {
		log.Printf("验证失败：%v", err)
	} else {
		log.Println("验证成功")
	}
}

// 示例4：依赖任务之间的上下文传递
func dependencyContextExample() {
	// 创建工作池
	pool := task.NewWorkerPool(2, nil)
	pool.Start()
	defer pool.Stop()

	// 创建第一个任务
	task1 := task.TaskWithContextMap("数据准备", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("依赖示例 - 任务1：准备数据")

		// 设置上下文数据
		data["source"] = "数据库"
		data["records"] = 100
		data["timestamp"] = time.Now().Format(time.RFC3339)

		log.Printf("依赖示例 - 任务1：设置上下文数据 %v", data)

		return nil
	})

	// 创建第二个任务，依赖于第一个任务
	task2 := task.TaskWithContextMap("数据处理", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("依赖示例 - 任务2：处理数据")

		// 获取第一个任务传递的上下文数据
		source, sourceExists := data["source"]
		records, recordsExists := data["records"]
		timestamp, timestampExists := data["timestamp"]

		log.Printf("依赖示例 - 任务2：接收到上下文数据 source=%v(%v), records=%v(%v), timestamp=%v(%v)",
			source, sourceExists, records, recordsExists, timestamp, timestampExists)

		// 设置新的上下文数据
		data["processed"] = true
		data["result"] = fmt.Sprintf("处理了来自%v的%v条记录", source, records)

		return nil
	})

	// 设置依赖关系
	task2.DependsOn(task1)

	// 提交任务
	pool.Submit(task2) // 先提交依赖任务
	pool.Submit(task1) // 后提交被依赖任务

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 获取任务上下文数据
	ctx1 := task1.GetContext()
	ctx2 := task2.GetContext()

	log.Printf("依赖示例 - 任务1上下文：%v", ctx1.GetAll())
	log.Printf("依赖示例 - 任务2上下文：%v", ctx2.GetAll())
}
