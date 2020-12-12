package main

import (
	"fish/screen"
	"fmt"
	"image"
	"math"
	"time"

	"github.com/go-vgo/robotgo"
)

type Task int

const (
	// 寻找鱼漂
	TaskFind Task = iota
	// 等鱼上钩
	TaskWait
)

var (
	STEP               = 30
	StartFindFishFloat bool
)

func main() {
	go monitorStart()
	findFishFloat()
	select {}
}

func monitorStart() {
	for {
		ok := robotgo.AddEvent("f10")
		if ok {
			StartFindFishFloat = !StartFindFishFloat
		}
		time.Sleep(time.Second)
	}
}

func findFishFloat() {
	step := STEP
	for {
		time.Sleep(100 * time.Millisecond)
		if !StartFindFishFloat {
			continue
		}
		for _, c := range screen.ListScreenPoints(step) {
			if !StartFindFishFloat {
				break
			}
			before := screen.CaptureScreen(c.X, c.Y, step)
			robotgo.MoveMouse(c.X, c.Y)
			time.Sleep(30 * time.Millisecond)
			after := screen.CaptureScreen(c.X, c.Y, step)
			n := compared(before, after)
			if n > 0 {
				fmt.Println(n)
				if n >= 20 {
					StartFindFishFloat = false
					break
				}
			}
		}
	}
}

func compared(before, after image.Image) int {
	r1, g1, b1 := screen.AverageColor(before)
	r2, g2, b2 := screen.AverageColor(after)
	n := int(math.Abs(float64(r1)-float64(r2)) + math.Abs(float64(g1)-float64(g2)) + math.Abs(float64(b1)-float64(b2)))
	return n
}
