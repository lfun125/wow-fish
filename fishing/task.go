package fishing

import (
	"context"
	"time"
)

type TaskType int

const (
	TaskTypeNil TaskType = iota
	// 寻找鱼漂
	TaskTypeThrowFishingRod
	// 等鱼上钩
	TaskTypeWait
	// 按键
	TaskKeyboard
)

type Task struct {
	Type    TaskType
	Context context.Context
	Timeout <-chan time.Time
}
