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
	"strconv"
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

func Compared(before, after image.Image) int {
	r1, g1, b1 := averageColor(before)
	r2, g2, b2 := averageColor(after)
	n := int(math.Abs(float64(r1)-float64(r2)) + math.Abs(float64(g1)-float64(g2)) + math.Abs(float64(b1)-float64(b2)))
	return n
}

func averageColor(img image.Image) (r, g, b uint32) {
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
	r = uint32(tr / tol)
	g = uint32(tg / tol)
	b = uint32(tb / tol)
	return
}

func RGBDifference(v1, v2 color.RGBA) float64 {
	transform := func(v float64) float64 {
		if v == 0 {
			v = 1
		}
		return v
	}
	r1 := transform(math.Abs(float64(v1.R)))
	g1 := transform(math.Abs(float64(v1.G)))
	b1 := transform(math.Abs(float64(v1.B)))
	r2 := transform(math.Abs(float64(v2.R)))
	g2 := transform(math.Abs(float64(v2.G)))
	b2 := transform(math.Abs(float64(v2.B)))
	var d1, d2 float64
	if r1 > g1 {
		d1 += r1 / g1
	} else {
		d1 += g1 / r1
	}
	if r1 > b1 {
		d1 += r1 / b1
	} else {
		d1 += b1 / r1
	}

	if r2 > g2 {
		d2 += r2 / g2
	} else {
		d2 += g2 / r2
	}
	if r2 > b2 {
		d2 += r2 / b2
	} else {
		d1 += b2 / r2
	}
	return math.Abs(d1 - d2)
}

func StrToRGBA(colorStr string) color.RGBA {
	var data color.RGBA
	if len(colorStr) == 6 {
		rs := colorStr[0:2]
		gs := colorStr[2:4]
		bs := colorStr[4:]
		rr, _ := strconv.ParseInt(rs, 16, 64)
		gg, _ := strconv.ParseInt(gs, 16, 64)
		bb, _ := strconv.ParseInt(bs, 16, 64)
		data.R = uint8(rr)
		data.G = uint8(gg)
		data.B = uint8(bb)
	}
	return data
}
