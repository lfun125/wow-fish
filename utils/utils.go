package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"sync"
)

func CutImage(src image.Image, x, y, w, h int) (image.Image, error) {
	var subImg image.Image

	if rgbImg, ok := src.(*image.YCbCr); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.YCbCr) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.RGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.RGBA) //图片裁剪x0 y0 x1 y1
	} else if rgbImg, ok := src.(*image.NRGBA); ok {
		subImg = rgbImg.SubImage(image.Rect(x, y, x+w, y+h)).(*image.NRGBA) //图片裁剪x0 y0 x1 y1
	} else {
		return subImg, errors.New("图片解码失败")
	}
	return subImg, nil
}

func SavePng(f string, img image.Image) {
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

func Round(val float64, places int) float64 {
	f := math.Pow10(places)
	return float64(int64(val*f+0.5)) / f
}

func ComparedLuminance(before, after image.Image) float64 {
	y1 := averageLuminance(before)
	y2 := averageLuminance(after)
	//return math.Abs(float64(rgb1.R)-float64(rgb2.R)) + math.Abs(float64(rgb1.G)-float64(rgb2.G)) + math.Abs(float64(rgb1.B)-float64(rgb2.B))
	return y2 - y1
}

func averageLuminance(img image.Image) (luminance float64) {
	rect := img.Bounds()
	with := rect.Size().X
	height := rect.Size().Y
	//wn := with / 10
	//hn := height / 10
	chCoordinate := make(chan [2]int, 1)
	go func() {
		for x := 1; x < with; x += 1 {
			for y := 1; y < height; y += 1 {
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
	var y, tol float64
	for v := range chResult {
		tol++
		y += 0.299*float64(v.R) + 0.587*float64(v.G) + 0.114*float64(v.B)
	}
	luminance = y / tol
	return
}

func Compared(before, after image.Image) float64 {
	rgb1 := averageColor(before)
	rgb2 := averageColor(after)
	//return math.Abs(float64(rgb1.R)-float64(rgb2.R)) + math.Abs(float64(rgb1.G)-float64(rgb2.G)) + math.Abs(float64(rgb1.B)-float64(rgb2.B))
	return RGBDistance(rgb1, rgb2)
}

func averageColor(img image.Image) (rgba color.RGBA) {
	rect := img.Bounds()
	with := rect.Size().X
	height := rect.Size().Y
	//wn := with / 10
	//hn := height / 10
	chCoordinate := make(chan [2]int, 1)
	go func() {
		for x := 1; x < with; x += 1 {
			for y := 1; y < height; y += 1 {
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
	rgba.R = uint8(tr / tol)
	rgba.G = uint8(tg / tol)
	rgba.B = uint8(tb / tol)
	return
}
