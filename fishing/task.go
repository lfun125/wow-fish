package fishing

import (
	"context"
	"fmt"
)

var (
	ErrOutOfBounds = fmt.Errorf("Out of bounds ")
)

type TaskType int

const (
	TaskTypeNil TaskType = iota
	// 寻找鱼漂
	TaskTypeThrowFishingRod
	// 等鱼上钩
	TaskTypeWait
)

type Task struct {
	Type    TaskType
	Context context.Context
}
