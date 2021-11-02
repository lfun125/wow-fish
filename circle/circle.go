package circle

import (
	"math"
)

type Coordinate struct {
	// 半径
	Radius float64
	// 角度
	Radian float64
	// X 坐标
	X int
	// Y 坐标
	Y int
}

type Circle struct {
	// 半径
	Radius float64
	// 中心点坐标
	Center [2]int
	// 单次弧长
	ArcLen float64
}

// NewCircle
// radius 周长
// arc 每次移动弧度
// x 中心点坐标
// y 中心点坐标
func NewCircle(radius, arc float64, x, y int) *Circle {
	c := &Circle{
		Radius: radius,
	}
	c.Center[0] = x
	c.Center[1] = y
	c.ArcLen = arc
	return c
}

// 获取一个圆周坐标
func (c Circle) ListCoordinates() (list []Coordinate) {
	cir := 2 * c.Radius * math.Pi
	for arc := 0.0; arc <= cir; arc += c.ArcLen {
		// 幅度
		radian := arc / c.Radius
		list = append(list, Coordinate{
			Radius: c.Radius,
			Radian: radian,
			X:      round(math.Sin(radian)*c.Radius) + c.Center[0],
			Y:      round(math.Cos(radian)*c.Radius) + c.Center[1],
		})
	}
	return
}

func round(x float64) int {
	// return int(math.Floor(x + 0/5))
	return int(math.Floor(x))
}
