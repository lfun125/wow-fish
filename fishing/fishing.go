package fishing

import (
	"context"
	"errors"
	"fish/circle"
	"fish/config"
	"fish/operation"
	"fish/screen"
	"fish/utils"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"time"

	hook "github.com/robotn/gohook"

	"github.com/go-vgo/robotgo"
	"github.com/nfnt/resize"
)

type contextValue int

const (
	KeyCycle contextValue = iota + 1
)

var (
	ErrOutOfBounds = errors.New("Out of bounds ")
	FishingTime60  = 30 * time.Second
	FishingTime70  = 20 * time.Second
)

type Fishing struct {
	// 分屏所在分配区域 1-4; 0不分屏
	SplitArea int
	// 鱼漂颜色
	FloatColor   color.RGBA
	activeX      int
	activeY      int
	task         chan *Task
	cancelFunc   context.CancelFunc
	times        int
	listKeyCycle config.ListKeyCycle
}

func NewFishing(splitArea int) *Fishing {
	f := new(Fishing)
	f.FloatColor = config.C.FloatColor
	f.SplitArea = splitArea
	f.task = make(chan *Task)
	bitmapRef := robotgo.CaptureScreen(0, 0, 10, 10)
	img := robotgo.ToImage(bitmapRef)
	robotgo.FreeBitmap(bitmapRef)
	displayWidth := img.Bounds().Size().X
	screen.Info.DisplayZoom = float64(10) / float64(displayWidth)
	f.Info("Screen info", "width", screen.Info.ScreenWidth, "height", screen.Info.ScreenHeight, "zoom", screen.Info.DisplayZoom)
	f.Info("Config info", fmt.Sprintf("%+v", config.C))
	f.listKeyCycle = []*config.KeyCycle{}
	for _, v := range config.C.ListKeyCycle {
		ok := v.SplitArea == "0" || strings.Contains(v.SplitArea, fmt.Sprintf("_%d_", splitArea))
		if !ok {
			continue
		}
		f.listKeyCycle = append(f.listKeyCycle, v)
		f.Info("Key cycle", v.String())
	}
	return f
}

func (f *Fishing) Run() error {
	go f.watchTask()
	select {}
}

func (f *Fishing) watchTask() {
	for task := range f.task {
		// 执行任务
		typ := f.runTask(task)
		select {
		case <-task.Context.Done():
			f.Info("User manual pause")
		default:
			go func(task *Task) {
				switch typ {
				case TaskTypeThrowFishingRod:
					time.Sleep(2 * time.Second)
					f.start()
				case TaskTypeWait:
					f.waitPullHook(task.Context, task.Timeout)
				}
			}(task)
		}
	}
}

func (f *Fishing) start() {
	var kc *config.KeyCycle
	task := new(Task)
	// 判断是否有按键需要循环操作
	for _, v := range f.listKeyCycle {
		if time.Since(v.ExecTime) > v.CycleDuration {
			kc = v
			break
		}
	}
	dur := FishingTime70
	if config.C.WOWVersion == "60" {
		dur = FishingTime60
	}
	task.Timeout = time.After(dur)
	if kc != nil {
		task.Type = TaskKeyboard
		ctx := context.WithValue(context.Background(), KeyCycle, kc)
		task.Context, f.cancelFunc = context.WithCancel(ctx)
	} else {
		f.times++
		task.Type = TaskTypeThrowFishingRod
		task.Context, f.cancelFunc = context.WithCancel(context.Background())
	}
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
	case TaskKeyboard:
		kc := t.Context.Value(KeyCycle).(*config.KeyCycle)
		<-operation.AddOperation(f.SplitArea, true, func() interface{} {
			kc.Key.Tap()
			return nil
		})
		time.Sleep(kc.WaitTime)
		kc.ExecTime = time.Now()
		return TaskTypeThrowFishingRod
	case TaskTypeThrowFishingRod:
		f.Info("Start looking for fish floats")
		if isFind := f.stepThrow(t); isFind {
			// 找到鱼漂
			f.Info("Found a fishing float")
			return TaskTypeWait
		}
		f.Info("Not found fishing float")
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

// 等待鱼上钩
func (f *Fishing) stepWaitPullHook(t *Task) bool {
	time.Sleep(2 * time.Second)
	f.Info("Active coordinate x:", f.activeX, "y:", f.activeY)
	compareCoordinate := config.C.CompareCoordinate
	if f.SplitArea > 0 {
		compareCoordinate /= 2
	}
	width := compareCoordinate
	x := f.activeX - width/2
	y := f.activeY - width/2
	bitmapRef := robotgo.CaptureScreen(x, y, width, width)
	oldImg := robotgo.ToImage(bitmapRef)
	robotgo.FreeBitmap(bitmapRef)
	// 图片明亮度
	oldLuminance := utils.AverageLuminance(oldImg)
	for {
		select {
		case <-t.Timeout:
			f.Info("Time out")
			return false
		case <-t.Context.Done():
			return false
		default:
			bitmapRef := robotgo.CaptureScreen(x, y, width, width)
			newImg := robotgo.ToImage(bitmapRef)
			robotgo.FreeBitmap(bitmapRef)
			newLuminance := utils.AverageLuminance(newImg)
			diff := newLuminance - oldLuminance
			if diff >= 1 {
				f.Info(fmt.Sprintf("Compared luminance: %0.4f", diff))
			}
			if diff >= config.C.Luminance {
				// 上钩收杆
				<-operation.AddOperation(f.SplitArea, false, func() interface{} {
					area := getMouseSplitArea()
					wait := f.SplitArea > 0 && f.SplitArea != area
					robotgo.MoveMouse(f.activeX, f.activeY)
					if wait {
						time.Sleep(100 * time.Millisecond)
					}
					robotgo.MouseClick("right")
					if wait {
						time.Sleep(100 * time.Millisecond)
					}
					robotgo.MouseClick("right", true)
					return nil
				})
				<-operation.AddOperation(f.SplitArea, true, nil)
				return true
			} else if diff < config.C.Luminance/4 {
				oldLuminance = newLuminance
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

type DiffColorToXY struct {
	Diff float64
	X, Y int
}

// 下竿
func (f *Fishing) stepThrow(t *Task) bool {
	f.activeX = 0
	f.activeY = 0
	// 甩杆
	<-operation.AddOperation(f.SplitArea, true, func() interface{} {
		config.C.FishingButton.Tap()
		// 按以下清楚垃圾开河蚌的宏
		config.C.OpenMacro.Tap()
		config.C.ClearMacro.Tap()
		return nil
	})
	time.Sleep(3 * time.Second)
	// 截屏
	var w, h int
	w, h = screen.Info.ScreenWidth, screen.Info.ScreenHeight
	if f.SplitArea > 0 {
		w, h = screen.Info.ScreenWidth/2, screen.Info.ScreenHeight/2
	}
	bitmapRef := robotgo.CaptureScreen()
	screenImg := robotgo.ToImage(bitmapRef)
	robotgo.FreeBitmap(bitmapRef)
	// 缩放
	screenImg = resize.Resize(uint(screen.Info.ScreenWidth), uint(screen.Info.ScreenHeight), screenImg, resize.NearestNeighbor)
	var maxRadius int
	if w > h {
		maxRadius = int(float64(h) / 32 * 8)
	} else {
		maxRadius = int(float64(w) / 32 * 8)
	}
	var circleList []circle.Coordinate
	step := 5
	if f.SplitArea > 0 {
		step = 2
	}
	centerX, centerY := screen.Info.ScreenWidth/2, screen.Info.ScreenHeight*3/8
	switch f.SplitArea {
	case 1:
		centerX = screen.Info.ScreenWidth / 4
		centerY = screen.Info.ScreenHeight * 3 / 16
	case 2:
		centerX = screen.Info.ScreenWidth * 3 / 4
		centerY = screen.Info.ScreenHeight * 3 / 16
	case 3:
		centerX = screen.Info.ScreenWidth / 4
		centerY = screen.Info.ScreenHeight * 11 / 16
	case 4:
		centerX = screen.Info.ScreenWidth * 3 / 4
		centerY = screen.Info.ScreenHeight * 11 / 16
	}
	for radius := step; radius <= maxRadius; radius += step {
		cir := circle.NewCircle(float64(radius), 5, centerX, centerY)
		circleList = append(circleList, cir.ListCoordinates()...)
	}
	store := map[float64][]DiffColorToXY{}
	var diffKeys []float64
	for _, v := range circleList {
		// 色差比较
		var data DiffColorToXY
		data.X = v.X
		data.Y = v.Y
		r, g, b, a := screenImg.At(data.X, data.Y).RGBA()
		rgba := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
		data.Diff = utils.RGBDistance(f.FloatColor, rgba)
		if len(store[data.Diff]) == 0 {
			diffKeys = append(diffKeys, data.Diff)
		}
		store[data.Diff] = append(store[data.Diff], data)
	}
	sort.Float64s(diffKeys)
	// 最大波动值
	var maxOscillation float64 = -1
	var number int
	if number = len(diffKeys); number > 3 {
		number = 3
	}
	var ary []DiffColorToXY
	for _, v := range diffKeys[0:number] {
		list := store[v]
		var number int
		for _, xy := range list {
			number++
			if number > 3 {
				continue
			}
			ary = append(ary, xy)
		}
	}
	result := operation.AddOperation(f.SplitArea, false, func() interface{} {
		for _, xy := range ary {
			select {
			case <-t.Timeout:
				return false
			case <-t.Context.Done():
				return false
			default:
				distance, err := f.getRGBDistance(xy.X, xy.Y)
				if err == ErrOutOfBounds {
					return false
				} else if err != nil {
					f.Info(err)
					return false
				} else if distance > maxOscillation {
					maxOscillation = distance
					f.activeX = xy.X
					f.activeY = xy.Y
				}
			}
		}
		return true
	})
	if ok := <-result; !ok.(bool) {
		return false
	}
	if f.activeX > 0 {
		robotgo.Move(f.activeX, f.activeY)
		return true
	}
	return false
}

func (f *Fishing) getRGBDistance(x, y int) (float64, error) {
	compareCoordinate := config.C.CompareCoordinate
	if f.SplitArea > 0 {
		compareCoordinate /= 2
	}
	cutX := x - compareCoordinate/2
	cutY := y - compareCoordinate/2
	bitmapRef := robotgo.CaptureScreen(cutX, cutY, compareCoordinate, compareCoordinate)
	defer robotgo.FreeBitmap(bitmapRef)
	oldImg := robotgo.ToImage(bitmapRef)
	robotgo.Move(x, y)
	time.Sleep(20 * time.Millisecond)
	// 移动后对比图片
	bitmapRef = robotgo.CaptureScreen(cutX, cutY, compareCoordinate, compareCoordinate)
	defer robotgo.FreeBitmap(bitmapRef)
	if bitmapRef == nil {
		return 0, ErrOutOfBounds
	}
	resultImg := robotgo.ToImage(bitmapRef)
	n := utils.Compared(resultImg, oldImg)
	return n, nil
}

func (f Fishing) Info(args ...interface{}) {
	var data []interface{}
	data = append(data, fmt.Sprintf("[split->%d]", f.SplitArea))
	data = append(data, fmt.Sprintf("Try: [%d]", f.times))
	data = append(data, args...)
	log.Println(data...)
}

func getMouseSplitArea() int {
	x, y := robotgo.GetMousePos()
	var area int
	if x < screen.Info.ScreenWidth/2 {
		if y < screen.Info.ScreenHeight/2 {
			area = 1
		} else {
			area = 3
		}
	} else {
		if y < screen.Info.ScreenHeight/2 {
			area = 2
		} else {
			area = 4
		}
	}
	return area
}

func WatchKeyboard(list ...*Fishing) {
	var keyTime time.Time
	robotgo.EventHook(hook.KeyHold, []string{config.C.SwitchButton}, func(e hook.Event) {
		if e.When.Sub(keyTime) < 300*time.Millisecond {
			return
		}
		keyTime = e.When
		for _, f := range list {
			if f.cancelFunc != nil {
				f.stop()
			} else {
				f.start()
			}
		}
	})
	robotgo.EventHook(hook.KeyHold, []string{config.C.ColorPickerButton}, func(e hook.Event) {
		if e.When.Sub(keyTime) < 300*time.Millisecond {
			return
		}
		keyTime = e.When
		x, y := robotgo.GetMousePos()
		area := getMouseSplitArea()
		floatColor := utils.StrToRGBA(robotgo.GetPixelColor(x, y))
		for _, f := range list {
			if f.SplitArea == 0 || f.SplitArea == area {
				f.FloatColor = floatColor
				fmt.Println(floatColor)
			}
		}
	})
	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}
