package fishing

import (
	"fmt"
	"testing"
)

func TestNewFishing(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	f.Run()
	fmt.Println(f.screenInfo)
}

func TestFishing_find(t *testing.T) {
	f := NewFishing(NewDefaultConfig())
	//originalImg := robotgo.ToImage(bitmapRef)
	f.find(611, 99)
	f.find(144, 70)
}
