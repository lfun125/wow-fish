package fishing

import (
	"fmt"
	"testing"
)

func TestNewFishing(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	_ = f.Run()
	fmt.Println(f.screenInfo)
}

func TestFishing_find(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	//originalImg := robotgo.ToImage(bitmapRef)
	_, _ = f.find(611, 99)
	_, _ = f.find(144, 70)
}
