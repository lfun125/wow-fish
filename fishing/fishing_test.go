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
