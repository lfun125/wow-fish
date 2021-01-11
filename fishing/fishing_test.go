package fishing

import (
	"context"
	"testing"
	"time"
)

func TestNewFishing(t *testing.T) {
	f := NewFishing(1)
	_ = f.Run()
}

func TestFishing_stepThrow(t *testing.T) {
	f := NewFishing(1)
	task := new(Task)
	task.Timeout = time.After(30 * time.Second)
	task.Type = TaskTypeThrowFishingRod
	task.Context, f.cancelFunc = context.WithCancel(context.Background())
	f.stepThrow(task)
}
