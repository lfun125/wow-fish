package fishing

import "C"
import (
	"context"
	"errors"
	"fish/utils"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	ErrOutOfBounds = errors.New("Out of bounds ")
	ErrClose       = errors.New("user close")
)

type Fishing struct {
	Config     *Config
	screenInfo *ScreenInfo
	activeX    int
	activeY    int
	task       chan *Task
	cancelFunc context.CancelFunc
	times      int
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
			typ := f.runTask(task)
			select {
			case <-task.Context.Done():
				f.Info("User manual pause")
			default:
				go func(task *Task) {
					switch typ {
					case TaskTypeThrowFishingRod:
						f.start()
					case TaskTypeWait:
						f.waitPullHook(task.Context, task.Timeout)
					}
				}(task)
			}
		}
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
	f.times++
	task := new(Task)
	task.Timeout = time.After(30 * time.Second)
	task.Type = TaskTypeThrowFishingRod
	task.Context, f.cancelFunc = context.WithCancel(context.Background())
	f.task <- task
}

// 等待拉钩
func (f *Fishing) waitPullHook(ctx context.Context, timeout <-chan time.Time) {
	task := new(Task)
	task.Timeout = timeout
	task.Type = TaskTypeWait
	task.Context = ctx
	f.task <- task
}

func (f *Fishing) stop() {
	if f.cancelFunc != nil {
		f.cancelFunc()
		f.cancelFunc = nil
	}
}

func (f *Fishing) runTask(t *Task) TaskType {
	switch t.Type {
	case TaskTypeThrowFishingRod:
		f.Info("Start looking for fish floats")
		if isFind := f.stepThrowFishingRod(t); isFind {
			// 找到鱼漂
			f.Info("Found a fishing float")
			return TaskTypeWait
		}
		f.Info("No fishing float found")
		return TaskTypeThrowFishingRod
	case TaskTypeWait:
		f.Info("Start waiting for the hook")
		if ok := f.stepWaitPullHook(t); ok {
			f.Info("Success")
		} else {
			f.Info("Fail")
		}
		time.Sleep(100 * time.Millisecond)
		return TaskTypeThrowFishingRod
	}
	return TaskTypeThrowFishingRod
}

func (f *Fishing) stepWaitPullHook(t *Task) bool {
	f.Info("Active coordinate x:", f.activeX, "y:", f.activeY)
	bitmapRef := robotgo.CaptureScreen(f.activeX, f.activeY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
	oldImg := robotgo.ToImage(bitmapRef)
	for {
		select {
		case <-t.Timeout:
			f.Info("Time out")
			return false
		case <-t.Context.Done():
			return false
		default:
			bitmapRef := robotgo.CaptureScreen(f.activeX, f.activeY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
			newImg := robotgo.ToImage(bitmapRef)
			n := utils.Compared(oldImg, newImg)
			if n > f.Config.FindThreshold {
				robotgo.KeyTap("right")
				return true
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// 下竿
func (f *Fishing) stepThrowFishingRod(t *Task) bool {
	// 按下下竿按键
	robotgo.KeyTap("1")
	time.Sleep(300 * time.Millisecond)
	// 最小半径 最大半径
	var minRadius, maxRadius int
	minRadius = f.Config.InitialRadius
	if f.screenInfo.ScreenWidth > f.screenInfo.ScreenHeight {
		maxRadius = int(float64(f.screenInfo.ScreenHeight) / 2.5)
	} else {
		maxRadius = int(float64(f.screenInfo.ScreenWidth) / 2.5)
	}
	for radius := minRadius; radius <= maxRadius; radius += f.Config.StepPixel {
		nextRadius := radius + f.Config.StepPixel
		// 周长
		cir := 2 * math.Pi * float64(nextRadius)
		// 转一圈需要几次
		n := cir / float64(f.Config.StepPixel)
		// 半径递增
		incR := float64(f.Config.StepPixel) / n
		// 单次弧度
		radianItem := utils.Round(2*math.Pi/n, 5)
		r := float64(radius)
		for radian := 0.0; radian <= utils.Round(2*math.Pi, 5); radian += radianItem {
			select {
			case <-t.Timeout:
				f.Info("Time out")
				return false
			case <-t.Context.Done():
				return false
			default:
				x := int(math.Cos(radian)*r) + f.screenInfo.ScreenWidth/2
				y := int(math.Sin(radian)*r) + f.screenInfo.ScreenHeight/2
				r += incR
				isFind, err := f.find(x, y)
				if err == ErrOutOfBounds {
					return false
				} else if err != nil {
					f.Info(err)
					return false
				} else if isFind {
					f.activeX = x
					f.activeY = y
					return true
				}
			}
		}
	}
	return false
}

func (f *Fishing) find(x, y int) (bool, error) {
	cutX := x - f.Config.CompareCoordinate/2
	cutY := y - f.Config.CompareCoordinate/2
	bitmapRef := robotgo.CaptureScreen(cutX, cutY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
	oldImg := robotgo.ToImage(bitmapRef)
	robotgo.Move(x, y)
	time.Sleep(10 * time.Millisecond)
	// 移动后对比图片
	bitmapRef = robotgo.CaptureScreen(cutX, cutY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
	if bitmapRef == nil {
		return false, ErrOutOfBounds
	}
	resultImg := robotgo.ToImage(bitmapRef)
	n := utils.Compared(resultImg, oldImg)
	if n > f.Config.FindThreshold {
		return true, nil
	}
	return false, nil
}

func (f Fishing) Info(args ...interface{}) {
	var data []interface{}
	data = append(data, fmt.Sprintf("Try: [%d]", f.times))
	data = append(data, args...)
	log.Println(data)
}
