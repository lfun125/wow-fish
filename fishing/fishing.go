package fishing

import (
	"context"
	"errors"
	"fish/circle"
	"fish/utils"
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	hook "github.com/robotn/gohook"

	"github.com/go-vgo/robotgo"
	"github.com/nfnt/resize"
)

var (
	ErrOutOfBounds = errors.New("Out of bounds ")
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
	f.screenInfo = new(ScreenInfo)
	f.screenInfo.ScreenWidth, f.screenInfo.ScreenHeight = robotgo.GetScreenSize()
	bitmapRef := robotgo.CaptureScreen(0, 0, 10, 10)
	img := robotgo.ToImage(bitmapRef)
	displayWidth := img.Bounds().Size().X
	f.screenInfo.DisplayZoom = float64(10) / float64(displayWidth)
	f.Info("Screen info", "width", f.screenInfo.ScreenWidth, "height", f.screenInfo.ScreenHeight, "zoom", f.screenInfo.DisplayZoom)
	f.Info("Config info", fmt.Sprintf("%+v", f.Config))
	for _, v := range f.Config.ListKeyCycle {
		f.Info("Key cycle", fmt.Sprintf("%+v", v))
	}
	return f
}

func (f *Fishing) Run() error {
	go f.watchKeyboard()
	go f.watchTask()
	select {}
}

func (f *Fishing) watchTask() {
	for task := range f.task {
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

func (f *Fishing) watchKeyboard() {
	var keyTime time.Time
	robotgo.EventHook(hook.KeyHold, []string{f.Config.SwitchButton}, func(e hook.Event) {
		if e.When.Sub(keyTime) < 300*time.Millisecond {
			return
		}
		keyTime = e.When
		if f.cancelFunc != nil {
			f.stop()
		} else {
			f.start()
		}
	})
	robotgo.EventHook(hook.KeyHold, []string{f.Config.ColorPickerButton}, func(e hook.Event) {
		if e.When.Sub(keyTime) < 300*time.Millisecond {
			return
		}
		keyTime = e.When
		x, y := robotgo.GetMousePos()
		f.Config.FloatColor = utils.StrToRGBA(robotgo.GetPixelColor(x, y))
		f.Info(fmt.Sprintf("Set fish float color: %v", f.Config.FloatColor))
	})
	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}

func (f *Fishing) start() {
	var kc *KeyCycle
	task := new(Task)
	for _, v := range f.Config.ListKeyCycle {
		if time.Since(v.ExecTime) > v.CycleDuration {
			kc = v
			break
		}
	}
	task.Timeout = time.After(30 * time.Second)
	if kc != nil {
		task.Type = TaskKeyboard
		ctx := context.WithValue(context.Background(), "KeyCycle", kc)
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
		kc := t.Context.Value("KeyCycle").(*KeyCycle)
		kc.Key.Tap()
		time.Sleep(kc.WaitTime)
		kc.ExecTime = time.Now()
		return TaskTypeThrowFishingRod
	case TaskTypeThrowFishingRod:
		f.Info("Start looking for fish floats")
		//if isFind := f.stepThrowFishingRod(t); isFind {
		if isFind := f.stepThrow(t); isFind {
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

// 等待鱼上钩
func (f *Fishing) stepWaitPullHook(t *Task) bool {
	// 按以下清楚垃圾开河蚌的宏
	f.Config.OpenMacro.Tap()
	time.Sleep(2 * time.Second)
	f.Info("Active coordinate x:", f.activeX, "y:", f.activeY)
	width := f.Config.CompareCoordinate
	x := f.activeX - width/2
	y := f.activeY - width/2
	oldImg := robotgo.ToImage(robotgo.CaptureScreen(x, y, width, width))
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
			newLuminance := utils.AverageLuminance(newImg)
			diff := newLuminance - oldLuminance
			f.Info(fmt.Sprintf("Compared luminance: %0.4f", diff))
			if diff >= f.Config.Luminance {
				robotgo.Move(f.activeX, f.activeY)
				robotgo.MouseClick("right")
				robotgo.Move(0, 0)
				return true
			} else if diff < f.Config.Luminance/4 {
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
	robotgo.Move(0, 0)
	// 按下下竿按键
	f.Config.FishingButton.Tap()
	time.Sleep(3 * time.Second)
	// 清楚垃圾
	f.Config.ClearMacro.Tap()
	// 截屏
	screen := robotgo.ToImage(robotgo.CaptureScreen())
	// 缩放
	screen = resize.Resize(uint(f.screenInfo.ScreenWidth), uint(f.screenInfo.ScreenHeight), screen, resize.NearestNeighbor)
	var maxRadius int
	if f.screenInfo.ScreenWidth > f.screenInfo.ScreenHeight {
		maxRadius = int(float64(f.screenInfo.ScreenHeight) / 32 * 9)
	} else {
		maxRadius = int(float64(f.screenInfo.ScreenWidth) / 32 * 9)
	}
	var circleList []circle.Coordinate
	for radius := 5; radius <= maxRadius; radius += 5 {
		cir := circle.NewCircle(float64(radius), 5, f.screenInfo.ScreenWidth, f.screenInfo.ScreenHeight)
		circleList = append(circleList, cir.ListCoordinates()...)
	}
	store := map[float64][]DiffColorToXY{}
	var diffKeys []float64
	for _, v := range circleList {
		// 色差比较
		var data DiffColorToXY
		data.X = v.X
		data.Y = v.Y
		r, g, b, a := screen.At(data.X, data.Y).RGBA()
		rgba := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
		data.Diff = utils.RGBDistance(f.Config.FloatColor, rgba)
		if len(store[data.Diff]) == 0 {
			diffKeys = append(diffKeys, data.Diff)
		}
		store[data.Diff] = append(store[data.Diff], data)
	}
	sort.Float64s(diffKeys)
	// 最大波动值
	var maxOscillation float64
	var number int
	if number = len(diffKeys); number > 3 {
		number = 3
	}
	for _, v := range diffKeys[0:number] {
		list := store[v]
		var number int
		for _, xy := range list {
			number++
			if number > 3 {
				continue
			}
			select {
			case <-t.Timeout:
				f.Info("Time out")
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
	}
	if f.activeX > 0 {
		robotgo.Move(f.activeX, f.activeY)
		return true
	}
	return false
}

func (f *Fishing) getRGBDistance(x, y int) (float64, error) {
	cutX := x - f.Config.CompareCoordinate/2
	cutY := y - f.Config.CompareCoordinate/2
	bitmapRef := robotgo.CaptureScreen(cutX, cutY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
	oldImg := robotgo.ToImage(bitmapRef)
	robotgo.Move(x, y)
	time.Sleep(20 * time.Millisecond)
	// 移动后对比图片
	bitmapRef = robotgo.CaptureScreen(cutX, cutY, f.Config.CompareCoordinate, f.Config.CompareCoordinate)
	if bitmapRef == nil {
		return 0, ErrOutOfBounds
	}
	resultImg := robotgo.ToImage(bitmapRef)
	n := utils.Compared(resultImg, oldImg)
	return n, nil
}

func (f Fishing) Info(args ...interface{}) {
	var data []interface{}
	data = append(data, fmt.Sprintf("Try: [%d]", f.times))
	data = append(data, args...)
	log.Println(data)
}
