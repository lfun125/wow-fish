package fishing

import "C"
import (
	"context"
	"fish/utils"
	"fmt"
	"log"
	"math"
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
			f.runTask(task)
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

func (f *Fishing) runTask(t *Task) {
	switch t.Type {
	case TaskTypeThrowFishingRod:
		select {
		case <-t.Context.Done():
			log.Println("关闭任务")
			return
		default:
			if isFind := f.throwFishingRod(t); isFind {
				// 找到鱼漂
			} else {
				// 未找到鱼漂
				log.Println("未找到鱼漂")
				go func() {
					f.task <- t
				}()
			}
		}
	}
}

// 下竿
func (f *Fishing) throwFishingRod(t *Task) bool {
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
			case <-t.Context.Done():
				log.Println("关闭寻找鱼漂")
				return false
			default:
				x := int(math.Cos(radian)*r) + f.screenInfo.ScreenWidth/2
				y := int(math.Sin(radian)*r) + f.screenInfo.ScreenHeight/2
				r += incR
				isFind, err := f.find(x, y)
				if err == ErrOutOfBounds {
					return false
				} else if err != nil {
					log.Println(err)
					return false
				} else if isFind {
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
	if n > f.Config.CompareCoordinate {
		return true, nil
	}
	return false, nil
}
