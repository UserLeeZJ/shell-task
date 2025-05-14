// examples/simple_example.go
package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	task "github.com/UserLeeZJ/shell-task"
)

func main() {
	t := task.New(
		task.WithName("DemoTask"),
		task.WithJob(func(ctx context.Context) error {
			fmt.Println("Running job...")
			return fmt.Errorf("simulated error")
		}),
		task.WithMaxRuns(3),
		task.WithRepeat(1*time.Second),
		task.WithRetry(2),
		task.WithLogger(log.Printf),
		task.WithErrorHandler(func(err error) {
			log.Println("Error handled:", err)
		}),
		task.WithCancelOnFailure(true),
		task.WithMetricCollector(func(res task.JobResult) {
			log.Printf("Job '%s' took %v, success: %t", res.Name, res.Duration, res.Success)
		}),
		task.WithRecover(func(r interface{}) {
			log.Printf("Recovered from panic: %v\nStack:\n%s", r, debug.Stack())
		}),
	)

	t.Run()

	time.Sleep(10 * time.Second)
	t.Stop()
}
