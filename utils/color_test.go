package utils

import (
	"fmt"
	"image/color"
	"testing"
)

func TestRGBDistance(t *testing.T) {
	c1 := color.RGBA{
		R: 255,
		G: 0,
		B: 0,
	}
	c2 := color.RGBA{
		R: 255,
		G: 255,
		B: 255,
	}
	fmt.Println(RGBDistance(c1, c2))
}
