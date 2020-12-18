package fishing

import (
	"errors"
	"flag"
	"image/color"
	"strings"
	"time"
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
	// 按键循环
	ListKeyCycle ListKeyCycle
}

type KeyCycle struct {
	Key           string
	ExecTime      time.Time
	WaitTime      time.Duration
	CycleDuration time.Duration
}

type ListKeyCycle []*KeyCycle

func (*ListKeyCycle) String() string {
	return ""
}

func (list *ListKeyCycle) Set(s string) error {
	ary := strings.Split(s, ",")
	if len(ary) != 3 {
		return errors.New("Cycle wrong format. ")
	}
	timeDuration, err := time.ParseDuration(ary[1])
	if err != nil {
		return err
	}
	cycleDuration, err := time.ParseDuration(ary[2])
	if err != nil {
		return err
	}
	data := new(KeyCycle)
	data.Key = strings.ToLower(strings.TrimSpace(ary[0]))
	data.WaitTime = timeDuration
	data.CycleDuration = cycleDuration
	data.ExecTime = time.Now()
	*list = append(*list, data)
	return nil
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
	flag.Var(&c.ListKeyCycle, "cycle", "key,time,cycle")
	flag.Parse()
}
