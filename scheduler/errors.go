// scheduler/errors.go
package scheduler

import (
	"errors"
)

// 常见错误
var (
	ErrTaskNotFound = errors.New("task not found")
	ErrTimeout      = errors.New("operation timed out")
)
