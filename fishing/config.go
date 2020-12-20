package fishing

import (
	"errors"
	"flag"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"
)

type Button struct {
	Key string
	Pos struct{ X, Y int }
}

func (b Button) IsXY() bool {
	return b.Pos.X > 0 && b.Pos.Y > 0
}

type Config struct {
	// 开关按键
	SwitchButton string
	// 取色器按钮
	ColorPickerButton string
	// 删除垃圾宏
	ClearMacro string
	// 开河蚌宏
	OpenMacro string
	// 钓鱼按键
	FishingButton Button
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
	Key           Button
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
	data.Key.Key = strings.ToLower(strings.TrimSpace(ary[0]))
	data.WaitTime = timeDuration
	data.CycleDuration = cycleDuration
	data.ExecTime = time.Now()
	*list = append(*list, data)
	return nil
}

func NewDefaultConfig() *Config {
	return &Config{
		FishingButton:     Button{Key: "1"},
		OpenMacro:         "2",
		ClearMacro:        "6",
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

func (c *Config) ParseParams() (importCfg bool) {
	var fb string
	flag.StringVar(&fb, "fb", c.FishingButton.Key, "钓鱼按键，如果是坐标用逗号隔开")
	flag.Float64Var(&c.Luminance, "l", c.Luminance, "明亮度大于等于这个值就收杆")
	flag.Var(&c.ListKeyCycle, "cycle", "key,time,cycle")
	flag.BoolVar(&importCfg, "import", false, "导出配置")
	flag.Parse()
	if ary := strings.Split(fb, ","); len(ary) == 1 {
		c.FishingButton.Key = fb
	} else if len(ary) == 2 {
		var err error
		if c.FishingButton.Pos.X, err = strconv.Atoi(ary[0]); err != nil {
			flag.Usage()
			os.Exit(1)
		}
		if c.FishingButton.Pos.Y, err = strconv.Atoi(ary[1]); err != nil {
			flag.Usage()
			os.Exit(1)
		}
	}
	return
}
