package fishing

import "fish/utils"

type ScreenInfo struct {
	// 屏幕宽度
	ScreenWidth int
	// 屏幕高度
	ScreenHeight int
	// 显示缩放比例
	DisplayZoom float64
}

// 坐标转像素
func (s ScreenInfo) CoordinateToPixel(v int) int {
	return int(utils.Round(float64(v)/s.DisplayZoom, 0))
}
