package fishing

import (
	"context"
	"errors"
	"fish/utils"
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	"github.com/nfnt/resize"

	"github.com/go-vgo/robotgo"
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
	//f.context, f.cancelFunc = context.WithCancel(context.Background())
	f.screenInfo = new(ScreenInfo)
	f.screenInfo.ScreenWidth, f.screenInfo.ScreenHeight = robotgo.GetScreenSize()
	bitmapRef := robotgo.CaptureScreen(0, 0, 10, 10)
	img := robotgo.ToImage(bitmapRef)
	displayWidth := img.Bounds().Size().X
	f.screenInfo.DisplayZoom = float64(10) / float64(displayWidth)
	fmt.Println(f.screenInfo)
	return f
}

func (f *Fishing) Run() error {
	go f.watchSwitch()
	go f.watchTask()
	select {}
}

func (f *Fishing) ColorPicker() {
	if ok := robotgo.AddEvent(f.Config.ColorPickerButton); ok {
		x, y := robotgo.GetMousePos()
		f.Config.FloatColor = utils.StrToRGBA(robotgo.GetPixelColor(x, y))
	}
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

func (f *Fishing) stepWaitPullHook(t *Task) bool {
	time.Sleep(2 * time.Second)
	f.Info("Active coordinate x:", f.activeX, "y:", f.activeY)
	width := f.Config.CompareCoordinate
	x := f.activeX - width/2
	y := f.activeY - width/2
	oldImg := robotgo.ToImage(robotgo.CaptureScreen(x, y, width, width))
	var okTimes int
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
			distance := utils.Compared(oldImg, newImg)
			f.Info("Compared distance value:", distance)
			if distance >= f.Config.Distance {
				okTimes++
				if okTimes > 1 {
					robotgo.Move(f.activeX, f.activeY)
					robotgo.MouseClick("right")
					robotgo.MouseClick("f1")
					robotgo.Move(0, 0)
					return true
				}
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
	robotgo.KeyTap("1")
	time.Sleep(4 * time.Second)
	// 截屏
	screen := robotgo.ToImage(robotgo.CaptureScreen())
	// 缩放
	screen = resize.Resize(uint(f.screenInfo.ScreenWidth), uint(f.screenInfo.ScreenHeight), screen, resize.NearestNeighbor)
	var maxRadius int
	if f.screenInfo.ScreenWidth > f.screenInfo.ScreenHeight {
		maxRadius = int(float64(f.screenInfo.ScreenHeight) / 4)
	} else {
		maxRadius = int(float64(f.screenInfo.ScreenWidth) / 4)
	}
	centerX := f.screenInfo.ScreenWidth / 2
	centerY := f.screenInfo.ScreenHeight / 2
	store := map[float64][]DiffColorToXY{}
	var diffKeys []float64
	for xx := -maxRadius; xx <= maxRadius; xx += 10 {
		for yy := -maxRadius; yy <= maxRadius; yy += 10 {
			// 色差比较
			var data DiffColorToXY
			data.X = xx + centerX
			data.Y = yy + centerY
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
	if f.activeY >= f.screenInfo.ScreenHeight/2 {
		return false
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
