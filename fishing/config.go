package fishing

import (
	"errors"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"github.com/klauspost/cpuid"
)

type Config struct {
	// 开关按键
	SwitchButton string
	// 取色器按钮
	ColorPickerButton string
	// 删除垃圾宏
	ClearMacro Button
	// 开河蚌宏
	OpenMacro Button
	// 钓鱼按键
	FishingButton Button
	// 对比区域坐标
	CompareCoordinate int
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
		OpenMacro:         Button{Key: "2"},
		ClearMacro:        Button{},
		SwitchButton:      "f3",
		ColorPickerButton: "f4",
		CompareCoordinate: 100,
		Luminance:         4,
		FloatColor: color.RGBA{
			R: 255,
			G: 243,
			B: 167,
		},
	}
}

func (c *Config) ParseParams() (importCfg bool) {
	flag.Var(&c.FishingButton, "fb", "钓鱼按键，如果是坐标用逗号隔开")
	flag.Var(&c.ClearMacro, "cm", "清理垃圾宏按键，如果是坐标用逗号隔开")
	flag.Var(&c.OpenMacro, "om", "打开河蚌箱子宏按键，如果是坐标用逗号隔开")
	flag.Float64Var(&c.Luminance, "l", c.Luminance, "明亮度大于等于这个值就收杆")
	flag.Var(&c.ListKeyCycle, "cycle", "key,time,cycle")
	flag.BoolVar(&importCfg, "import", false, "导出配置")
	var cpu bool
	flag.BoolVar(&cpu, "cpu", false, "")
	flag.Parse()
	if cpu {
		fmt.Println(cpuid.CPU.VendorString)
		os.Exit(0)
	}
	return
}
