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
	// 寻找鱼标阀值
	FindThreshold int
	// 半径
	InitialRadius int
	// 抛竿按键
	ThrowButton string
	// 鱼漂颜色
	FloatColor color.RGBA
	// 差异度大于就收杆
	Distance float64
}

func NewDefaultConfig() *Config {
	return &Config{
		SwitchButton:      "f3",
		ColorPickerButton: "f4",
		CompareCoordinate: 100,
		FindThreshold:     20,
		InitialRadius:     40,
		ThrowButton:       "1",
		Distance:          0.1,
		FloatColor: color.RGBA{
			R: 255,
			G: 243,
			B: 167,
		},
	}
}

func (c *Config) ParseParams() {
	flag.Float64Var(&c.Distance, "dis", c.Distance, "差异度大于此值就收杆")
	flag.Parse()
}
