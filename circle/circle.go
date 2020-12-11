package circle

import (
	"math"

	"github.com/go-vgo/robotgo"
)

type Coordinate struct {
	// 角度
	Angle float64
	X     int
	Y     int
}

type Action func(x, y int) error

func (c Coordinate) Move(fn Action) (err error) {
	robotgo.MoveMouse(c.X, c.Y)
	err = fn(c.X, c.Y)
	return
}

type Circle struct {
	// 半径
	Radius int
	// 中心点坐标
	Center [2]int
}

func NewCircle(radius, x, y int) *Circle {
	c := &Circle{
		Radius: radius,
	}
	c.Center[0] = x / 2
	c.Center[1] = y / 2
	return c
}

// 获取一个圆周坐标
func (c Circle) ListCoordinates() (list []Coordinate) {
	// 周长
	for angle := 0.0; angle < 360; angle++ {
		// 幅度
		radian := angle * (math.Pi / 180)
		list = append(list, Coordinate{
			Angle: angle,
			X:     round(math.Sin(radian)*float64(c.Radius)) + c.Center[0],
			Y:     round(math.Cos(radian)*float64(c.Radius)) + c.Center[1],
		})
	}
	return
}

func round(x float64) int {
	return int(math.Floor(x + 0/5))
}
