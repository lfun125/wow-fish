package utils

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"log"
	"math"
	"os"
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
