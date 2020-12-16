package utils

import (
	"image/color"
	"strconv"

	"github.com/lucasb-eyer/go-colorful"
)

//func RGBDifference(v1, v2 color.RGBA) float64 {
//	v := math.Pow(float64(v1.R-v2.R), 2) +
//		math.Pow(float64(v1.G-v2.G), 2) +
//		math.Pow(float64(v1.B-v2.B), 2)
//	return math.Sqrt(v)
//}

func RGBDistance(v1, v2 color.RGBA) float64 {
	c1 := colorful.Color{
		R: float64(v1.R),
		G: float64(v1.G),
		B: float64(v1.B),
	}
	c2 := colorful.Color{
		R: float64(v2.R),
		G: float64(v2.G),
		B: float64(v2.B),
	}
	return c1.DistanceCIEDE2000(c2)
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

type XYZ struct {
	X, Y, Z float64
}
