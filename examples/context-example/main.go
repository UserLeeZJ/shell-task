// examples/context-example/main.go
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
	log.Println("上下文传递示例")

	// 示例1：使用简化API创建带上下文的任务
	log.Println("\n=== 示例1：简化API ===")
	simpleAPIExample()

	// 示例2：使用任务链传递上下文
	log.Println("\n=== 示例2：任务链 ===")
	taskChainExample()

	// 示例3：使用任务构建器API
	log.Println("\n=== 示例3：任务构建器 ===")
	taskBuilderExample()
}

// 示例1：使用上下文映射API创建带上下文的任务
func simpleAPIExample() {
	// 创建一个简单的任务，使用上下文传递数据
	task1 := task.TaskWithContextMap("数据准备", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("任务1：准备数据")

		// 将数据存储到上下文中
		data["result"] = "准备好的数据"
		data["timestamp"] = time.Now().Format(time.RFC3339)

		log.Printf("任务1：设置上下文数据 result=%v, timestamp=%v",
			data["result"], data["timestamp"])

		return nil
	})

	// 运行任务
	task1.Run()

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	// 获取上下文数据
	ctx := task1.GetContext()
	result, _ := ctx.GetString("result")
	timestamp, _ := ctx.GetString("timestamp")

	log.Printf("任务1完成，上下文数据：result=%v, timestamp=%v", result, timestamp)
}

// 示例2：使用任务链传递上下文
func taskChainExample() {
	// 创建第一个任务
	task1 := task.TaskWithContextMap("数据准备", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("任务链 - 任务1：准备数据")

		// 将数据存储到上下文中
		data["result"] = "准备好的数据"
		data["timestamp"] = time.Now().Format(time.RFC3339)

		log.Printf("任务链 - 任务1：设置上下文数据 result=%v, timestamp=%v",
			data["result"], data["timestamp"])

		return nil
	})

	// 创建第二个任务，处理第一个任务的结果
	task2 := task.TaskWithContextMap("数据处理", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("任务链 - 任务2：处理数据")

		// 获取上一个任务传递的数据
		if result, ok := data["result"].(string); ok {
			log.Printf("任务链 - 任务2：处理数据 '%s'", result)

			// 更新上下文数据
			data["processed"] = fmt.Sprintf("已处理: %s", result)
			data["processTime"] = time.Now().Format(time.RFC3339)
		} else {
			return fmt.Errorf("未找到上一个任务的结果")
		}

		return nil
	})

	// 创建第三个任务，使用前两个任务的结果
	task3 := task.TaskWithContextMap("结果输出", func(ctx context.Context, data map[string]interface{}) error {
		log.Println("任务链 - 任务3：输出结果")

		// 获取前两个任务传递的数据
		for k, v := range data {
			log.Printf("任务链 - 任务3：上下文数据 %s = %v", k, v)
		}

		return nil
	})

	// 创建任务链，自动传递上下文
	tasks := task.ChainTasks(task1, task2, task3)

	// 运行任务链
	for _, t := range tasks {
		t.Run()
		time.Sleep(100 * time.Millisecond) // 等待任务完成
	}
}

// 示例3：使用任务构建器API
func taskBuilderExample() {
	// 使用任务构建器创建任务
	builder := task.NewTaskBuilder("构建器任务")

	// 配置任务
	task1 := builder.
		WithMapContextJob(func(ctx context.Context, data map[string]interface{}) error {
			log.Println("构建器任务：执行中")

			// 设置上下文数据
			data["builder"] = "success"
			data["time"] = time.Now().Format(time.RFC3339)

			log.Printf("构建器任务：设置上下文数据 builder=%v, time=%v",
				data["builder"], data["time"])

			return nil
		}).
		WithContextValue("initial", "value"). // 设置初始上下文值
		WithContextPrep(func(ctx *task.TaskContext) {
			// 准备上下文
			ctx.Set("prep", "done")
			log.Println("构建器任务：上下文准备完成")
		}).
		WithContextClean(func(ctx *task.TaskContext) {
			// 清理上下文
			log.Println("构建器任务：上下文清理")

			// 获取所有上下文数据
			allData := ctx.GetAll()
			log.Printf("构建器任务：最终上下文数据 %v", allData)
		}).
		Run() // 构建并运行任务

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	// 获取上下文数据
	ctx := task1.GetContext()
	builderValue, _ := ctx.GetString("builder")
	initial, _ := ctx.GetString("initial")

	log.Printf("构建器任务完成，上下文数据：builder=%v, initial=%v", builderValue, initial)
}
