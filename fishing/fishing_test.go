package fishing

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewFishing(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	_ = f.Run()
	fmt.Println(f.screenInfo)
}

func TestFishing_stepThrow(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	task := new(Task)
	task.Timeout = time.After(30 * time.Second)
	task.Type = TaskTypeThrowFishingRod
	task.Context, f.cancelFunc = context.WithCancel(context.Background())
	f.stepThrow(task)
}
