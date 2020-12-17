package fishing

import (
	"flag"
	"image/color"
)

type Config struct {
	// 开关按键
	SwitchButton string
	// 取色器按钮
	ColorPickerButton string
	// 对比区域坐标
	CompareCoordinate int
	// 抛竿按键
	ThrowButton string
	// 鱼漂颜色
	FloatColor color.RGBA
	// 明亮度大于等于这个值就收杆
	Luminance float64
}

func NewDefaultConfig() *Config {
	return &Config{
		SwitchButton:      "f3",
		ColorPickerButton: "f4",
		CompareCoordinate: 100,
		ThrowButton:       "1",
		Luminance:         4,
		FloatColor: color.RGBA{
			R: 255,
			G: 243,
			B: 167,
		},
	}
}

func (c *Config) ParseParams() {
	flag.Float64Var(&c.Luminance, "l", c.Luminance, "明亮度大于等于这个值就收杆")
	flag.Parse()
}
