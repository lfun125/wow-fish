package config

import (
	"errors"
	"fish/operation"
	"flag"
	"image/color"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// 开关按键
	SwitchButton string
	// 取色器按钮
	ColorPickerButton string
	// 删除垃圾宏
	ClearMacro operation.Button
	// 开河蚌宏
	OpenMacro operation.Button
	// 钓鱼按键
	FishingButton operation.Button
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
	Key           operation.Button
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

type SplitList []int

func (split *SplitList) Set(s string) error {
	ary := strings.Split(s, ",")
	*split = SplitList{}
	for _, v := range ary {
		vv, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		*split = append(*split, vv)
	}
	return nil
}

func (split SplitList) String() string {
	var ary []string
	for _, v := range split {
		ary = append(ary, strconv.Itoa(v))
	}
	return strings.Join(ary, ",")
}

func NewDefaultConfig() *Config {
	return &Config{
		FishingButton:     operation.Button{Key: "1"},
		OpenMacro:         operation.Button{Key: "2"},
		ClearMacro:        operation.Button{},
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

var C = NewDefaultConfig()

func ParseParams() (importCfg bool, splitList SplitList) {
	splitList = append(splitList, 0)
	flag.Var(&C.FishingButton, "fb", "钓鱼按键，如果是坐标用逗号隔开")
	flag.Var(&C.ClearMacro, "cm", "清理垃圾宏按键，如果是坐标用逗号隔开")
	flag.Var(&C.OpenMacro, "om", "打开河蚌箱子宏按键，如果是坐标用逗号隔开")
	flag.Float64Var(&C.Luminance, "l", C.Luminance, "明亮度大于等于这个值就收杆")
	flag.Var(&C.ListKeyCycle, "cycle", "key,time,cycle")
	flag.BoolVar(&importCfg, "import", false, "导出配置")
	flag.Var(&splitList, "split", "设置分配数量")
	flag.Parse()
	return
}
