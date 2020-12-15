package fishing

import "image/color"

type Config struct {
	// 开关按键
	SwitchButton string
	// 取色器按钮
	ColorPickerButton string
	// 每一步移动弧度
	StepPixel int
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
}

func NewDefaultConfig() *Config {
	return &Config{
		SwitchButton:      "f3",
		ColorPickerButton: "f4",
		StepPixel:         50,
		CompareCoordinate: 120,
		FindThreshold:     20,
		InitialRadius:     40,
		ThrowButton:       "1",
		FloatColor: color.RGBA{
			R: 255,
			G: 243,
			B: 167,
		},
	}
}
