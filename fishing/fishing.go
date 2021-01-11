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
	"sync"
	"time"

	hook "github.com/robotn/gohook"

	"github.com/go-vgo/robotgo"
	"github.com/nfnt/resize"
)

var (
	ErrOutOfBounds = errors.New("Out of bounds ")
)

var (
	screenInfo        *ScreenInfo
	setScreenInfoOnce sync.Once
)

type Fishing struct {
	// 分屏所在分配区域 1-4; 0不分屏
	SplitArea int
	// 鱼漂颜色
	FloatColor color.RGBA
	activeX    int
	activeY    int
	task       chan *Task
	cancelFunc context.CancelFunc
	times      int
}

func init() {
	screenInfo = new(ScreenInfo)
	screenInfo.ScreenWidth, screenInfo.ScreenHeight = robotgo.GetScreenSize()
}

func NewFishing(splitArea int) *Fishing {
	f := new(Fishing)
	f.FloatColor = C.FloatColor
	f.SplitArea = splitArea
	f.task = make(chan *Task)
	bitmapRef := robotgo.CaptureScreen(0, 0, 10, 10)
	img := robotgo.ToImage(bitmapRef)
	displayWidth := img.Bounds().Size().X
	screenInfo.DisplayZoom = float64(10) / float64(displayWidth)
	f.Info("Screen info", "width", screenInfo.ScreenWidth, "height", screenInfo.ScreenHeight, "zoom", screenInfo.DisplayZoom)
	f.Info("Config info", fmt.Sprintf("%+v", C))
	for _, v := range C.ListKeyCycle {
		f.Info("Key cycle", fmt.Sprintf("%+v", v))
	}
	return f
}

func (f *Fishing) Run() error {
	go f.watchTask()
	select {}
}

func (f *Fishing) watchTask() {
	for task := range f.task {
		select {
		case <-task.Context.Done():
			f.Info("User manual pause")
		default:
			// 执行任务
			typ := f.runTask(task)
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

func WatchKeyboard(list ...*Fishing) {
	var keyTime time.Time
	robotgo.EventHook(hook.KeyHold, []string{C.SwitchButton}, func(e hook.Event) {
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
	robotgo.EventHook(hook.KeyHold, []string{C.ColorPickerButton}, func(e hook.Event) {
		if e.When.Sub(keyTime) < 300*time.Millisecond {
			return
		}
		keyTime = e.When
		x, y := robotgo.GetMousePos()
		var area int
		if x < screenInfo.ScreenWidth/2 {
			if y < screenInfo.ScreenHeight/2 {
				area = 1
			} else {
				area = 3
			}
		} else {
			if y < screenInfo.ScreenHeight/2 {
				area = 2
			} else {
				area = 4
			}
		}
		floatColor := utils.StrToRGBA(robotgo.GetPixelColor(x, y))
		for _, f := range list {
			if f.SplitArea == 0 || f.SplitArea == area {
				f.FloatColor = floatColor
			}
		}
	})
	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}

func (f *Fishing) start() {
	var kc *KeyCycle
	task := new(Task)
	// 判断是否有按键需要循环操作
	for _, v := range C.ListKeyCycle {
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
	// 按以下清楚垃圾开河蚌的宏
	C.OpenMacro.Tap()
	time.Sleep(2 * time.Second)
	f.Info("Active coordinate x:", f.activeX, "y:", f.activeY)
	compareCoordinate := C.CompareCoordinate
	if f.SplitArea > 0 {
		compareCoordinate /= 4
	}
	width := compareCoordinate
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
			if diff >= C.Luminance {
				robotgo.Move(f.activeX, f.activeY)
				robotgo.MouseClick("right")
				robotgo.Move(0, 0)
				return true
			} else if diff < C.Luminance/4 {
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
	C.FishingButton.Tap()
	time.Sleep(3 * time.Second)
	// 清楚垃圾
	C.ClearMacro.Tap()
	// 截屏
	var x, y, w, h int
	w, h = screenInfo.ScreenWidth, screenInfo.ScreenHeight
	switch f.SplitArea {
	case 1:
		w, h = screenInfo.ScreenWidth/2, screenInfo.ScreenHeight/2
	case 2:
		w, h = screenInfo.ScreenWidth/2, screenInfo.ScreenHeight/2
		x = w
	case 3:
		w, h = screenInfo.ScreenWidth/2, screenInfo.ScreenHeight/2
		y = h
	case 4:
		w, h = screenInfo.ScreenWidth/2, screenInfo.ScreenHeight/2
		x = w
		y = h
	}
	screen := robotgo.ToImage(robotgo.CaptureScreen(x, y, w, h))
	// 缩放
	screen = resize.Resize(uint(w), uint(h), screen, resize.NearestNeighbor)
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
	centerX, centerY := screenInfo.ScreenWidth/2, screenInfo.ScreenHeight*3/8
	switch f.SplitArea {
	case 1:
		centerX = screenInfo.ScreenWidth / 4
		centerY = screenInfo.ScreenHeight * 3 / 16
	case 2:
		centerX = screenInfo.ScreenWidth * 3 / 4
		centerY = screenInfo.ScreenHeight * 3 / 16
	case 3:
		centerX = screenInfo.ScreenWidth / 4
		centerY = screenInfo.ScreenHeight * 11 / 16
	case 4:
		centerX = screenInfo.ScreenWidth * 3 / 4
		centerY = screenInfo.ScreenHeight * 11 / 16
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
		r, g, b, a := screen.At(data.X, data.Y).RGBA()
		rgba := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
		data.Diff = utils.RGBDistance(C.FloatColor, rgba)
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
	compareCoordinate := C.CompareCoordinate
	if f.SplitArea > 0 {
		compareCoordinate /= 4
	}
	cutX := x - compareCoordinate/2
	cutY := y - compareCoordinate/2
	bitmapRef := robotgo.CaptureScreen(cutX, cutY, compareCoordinate, compareCoordinate)
	oldImg := robotgo.ToImage(bitmapRef)
	robotgo.Move(x, y)
	time.Sleep(20 * time.Millisecond)
	// 移动后对比图片
	bitmapRef = robotgo.CaptureScreen(cutX, cutY, compareCoordinate, compareCoordinate)
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
