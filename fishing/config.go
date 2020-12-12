package fishing

type Config struct {
	// 开关按键
	SwitchButton string
	// 每一步移动弧度
	StepPixel int
	// 对比区域
	ComparePixel int
	// 寻找鱼标阀值
	FindThreshold int
	// 半径
	InitialRadius int
	// 抛竿按键
	ThrowButton string
}

func NewDefaultConfig() *Config {
	return &Config{
		SwitchButton:  "f10",
		StepPixel:     30,
		ComparePixel:  30,
		FindThreshold: 20,
		InitialRadius: 20,
		ThrowButton:   "1",
	}
}
