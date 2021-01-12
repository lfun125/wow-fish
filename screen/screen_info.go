package screen

import (
	"fish/utils"

	"github.com/go-vgo/robotgo"
)

var Info *info

func init() {
	Info = new(info)
	Info.ScreenWidth, Info.ScreenHeight = robotgo.GetScreenSize()
}

type info struct {
	// 屏幕宽度
	ScreenWidth int
	// 屏幕高度
	ScreenHeight int
	// 显示缩放比例
	DisplayZoom float64
}

// 坐标转像素
func (s info) CoordinateToPixel(v int) int {
	return int(utils.Round(float64(v)/s.DisplayZoom, 0))
}
