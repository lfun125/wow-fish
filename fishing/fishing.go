package fishing

import (
	"context"
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
)

type Fishing struct {
	Config     *Config
	screenInfo *ScreenInfo
	task       chan *Task
	cancelFunc context.CancelFunc
}

func NewFishing(c *Config) *Fishing {
	f := new(Fishing)
	f.Config = c
	f.task = make(chan *Task)
	//f.context, f.cancelFunc = context.WithCancel(context.Background())
	f.screenInfo = new(ScreenInfo)
	f.screenInfo.ScreenWidth, f.screenInfo.ScreenHeight = robotgo.GetScreenSize()
	bitmapRef := robotgo.CaptureScreen(0, 0, 10, 10)
	img := robotgo.ToImage(bitmapRef)
	displayWidth := img.Bounds().Size().X
	f.screenInfo.DisplayZoom = float64(10) / float64(displayWidth)
	return f
}

func (f *Fishing) Run() error {
	go f.watchSwitch()
	go f.watchTask()
	select {}
	return nil
}

func (f *Fishing) watchTask() {
	for {
		select {
		case task := <-f.task:
			task.Do(f)
		}
		fmt.Println("for")
	}
}

func (f *Fishing) watchSwitch() {
	for {
		ok := robotgo.AddEvent(f.Config.SwitchButton)
		if ok {
			if f.cancelFunc != nil {
				f.stop()
			} else {
				f.start()
			}
		}
		time.Sleep(time.Second)
	}
}

func (f *Fishing) start() {
	task := new(Task)
	task.Type = TaskTypeThrowFishingRod
	task.Context, f.cancelFunc = context.WithCancel(context.Background())
	f.task <- task
}

func (f *Fishing) stop() {
	if f.cancelFunc != nil {
		f.cancelFunc()
		f.cancelFunc = nil
	}
}
