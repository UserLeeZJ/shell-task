// examples/simple_example.go
package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/UserLeeZJ/shell-task"
)

func main() {
	task := scheduler.NewTask(
		scheduler.WithName("DemoTask"),
		scheduler.WithJob(func(ctx context.Context) error {
			fmt.Println("Running job...")
			return fmt.Errorf("simulated error")
		}),
		scheduler.WithMaxRuns(3),
		scheduler.WithRepeat(1*time.Second),
		scheduler.WithRetry(2),
		scheduler.WithLogger(log.Printf),
		scheduler.WithErrorHandler(func(err error) {
			log.Println("Error handled:", err)
		}),
		scheduler.WithCancelOnFailure(true),
		scheduler.WithMetricCollector(func(res scheduler.JobResult) {
			log.Printf("Job '%s' took %v, success: %t", res.Name, res.Duration, res.Success)
		}),
		scheduler.WithRecover(func(r interface{}) {
			log.Printf("Recovered from panic: %v\nStack:\n%s", r, debug.Stack())
		}),
	)

	task.Run()

	time.Sleep(10 * time.Second)
	task.Stop()
}
