package fishing

import (
	"context"
	"fish/utils"
	"fmt"
	"image"
	"log"
	"math"
	"time"

	"github.com/go-vgo/robotgo"
)

type TaskType int

const (
	TaskTypeNil TaskType = iota
	// 寻找鱼漂
	TaskTypeThrowFishingRod
	// 等鱼上钩
	TaskTypeWait
)

type Task struct {
	Type    TaskType
	Context context.Context
}

func (t *Task) Do(f *Fishing) {
	fmt.Println("type", t.Type)
	switch t.Type {
	case TaskTypeThrowFishingRod:
		select {
		case <-t.Context.Done():
			log.Println("关闭任务")
			return
		default:
			if isFind := t.throwFishingRod(f); isFind {
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
func (t *Task) throwFishingRod(f *Fishing) bool {
	// 按下下竿按键
	robotgo.KeyTap("1")
	// 下竿后截取屏幕
	bitmapRef := robotgo.CaptureScreen(0, 0)
	originalImg := robotgo.ToImage(bitmapRef)
	_ = originalImg
	time.Sleep(300 * time.Millisecond)
	// 最小半径 最大半径
	var minRadius, maxRadius int
	minRadius = f.Config.InitialRadius
	if f.screenInfo.ScreenWidth > f.screenInfo.ScreenHeight {
		maxRadius = int(float64(f.screenInfo.ScreenHeight) / 2.5)
	} else {
		maxRadius = int(float64(f.screenInfo.ScreenWidth) / 2.5)
	}
	var radiusList []int
	var cutWidth, cutHeight int
	cutWidth = int(utils.Round(float64(f.Config.ComparePixel)*f.screenInfo.DisplayZoom, 0))
	cutHeight = cutWidth
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
				robotgo.Move(x, y)
				time.Sleep(30 * time.Millisecond)
				// 移动后对比图片
				cutX := x - cutWidth/2
				cutY := x - cutHeight/2
				bitmapRef := robotgo.CaptureScreen(cutX, cutY, cutWidth, cutHeight)
				resultImg := robotgo.ToImage(bitmapRef)
				utils.SavePng(fmt.Sprintf("%d_%d.png", x, y), resultImg)
				fmt.Println(cutWidth, cutHeight)
			}
		}
		// 每次的角度
		radiusList = append(radiusList, radius)
	}
	return false
}

func compared(img *image.Image) {

}
