package main

import (
	"bytes"
	"fish/screen"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	STEP               = 30
	StartFindFishFloat bool
)

func main() {
	go T()
	findFishFloat()
	select {}
}

func T() {
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
	STOP:
		if !StartFindFishFloat {
			continue
		}
		for _, c := range screen.ListScreenPoints(step) {
			if !StartFindFishFloat {
				continue STOP
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
					continue STOP
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

func savePng(f string, img image.Image) {
	file, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	data := bytes.NewBuffer(nil)
	err = png.Encode(data, img)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file.Write(data.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}
