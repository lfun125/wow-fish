package screen

import (
	"fish/circle"
	"image"
	"sync"

	"github.com/go-vgo/robotgo"
)

func ListScreenPoints(step int) (list []circle.Coordinate) {
	with, height := robotgo.GetScreenSize()
	var max int
	if with > height {
		max = height
	} else {
		max = with
	}
	max = int(float64(max) / 2.5)
	for radius := 20; radius <= max; radius += step {
		c := circle.NewCircle(float64(radius), float64(step), with, height)
		list = append(list, c.ListCoordinates()...)
	}
	return
}

func CaptureScreen(x, y int, step int) image.Image {
	l := step / 2
	bitmapRef := robotgo.CaptureScreen(x-l, y-l, step, step)
	img := robotgo.ToImage(bitmapRef)
	return img
}

func AverageColor(img image.Image) (r, g, b uint32) {
	rect := img.Bounds()
	with := rect.Size().X
	height := rect.Size().Y
	wn := with / 10
	hn := height / 10
	chCoordinate := make(chan [2]int, 1)
	go func() {
		for x := 1; x < with; x += wn {
			for y := 1; y < height; y += hn {
				chCoordinate <- [2]int{x, y}
			}
		}
		close(chCoordinate)
	}()
	type Color struct {
		R, G, B uint32
	}
	chResult := make(chan Color)
	var wt sync.WaitGroup
	for i := 0; i <= 100; i++ {
		wt.Add(1)
		go func() {
			for v := range chCoordinate {
				var mc Color
				mc.R, mc.G, mc.B, _ = img.At(v[0], v[1]).RGBA()
				mc.R >>= 8
				mc.G >>= 8
				mc.B >>= 8
				chResult <- mc
			}
			wt.Done()
		}()
	}
	go func() {
		wt.Wait()
		close(chResult)
	}()
	var tr, tg, tb, tol uint64
	for v := range chResult {
		tol++
		tr += uint64(v.R)
		tg += uint64(v.G)
		tb += uint64(v.B)
	}
	r = uint32(tr / tol)
	g = uint32(tg / tol)
	b = uint32(tb / tol)
	return
}
