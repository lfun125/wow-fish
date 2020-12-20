package fishing

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-vgo/robotgo"
)

type Button struct {
	Key string
	Pos struct{ X, Y int }
}

func (b Button) Tap() {
	if b.Pos.X > 0 && b.Pos.Y > 0 {
		robotgo.Move(b.Pos.X, b.Pos.Y)
		robotgo.MouseClick("right")
	} else if b.Key != "" {
		robotgo.KeyTap(b.Key)
	}
}

func (b *Button) String() string {
	if b.Pos.X > 0 && b.Pos.Y > 0 {
		return fmt.Sprintf("%d,%d", b.Pos.X, b.Pos.Y)
	} else {
		return b.Key
	}
}

func (b *Button) Set(s string) error {
	ary := strings.Split(strings.ToLower(s), ",")
	if n := len(ary); n == 1 {
		b.Key = s
	} else if n == 2 {
		var err error
		if b.Pos.X, err = strconv.Atoi(ary[0]); err != nil {
			return err
		}
		if b.Pos.Y, err = strconv.Atoi(ary[1]); err != nil {
			return err
		}
	} else {
		return errors.New("Button format error. ")
	}
	return nil
}
