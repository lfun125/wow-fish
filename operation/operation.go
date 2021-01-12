package operation

import (
	"fish/screen"
	"time"

	"github.com/go-vgo/robotgo"
)

type Operation func() interface{}

type task struct {
	fn        Operation
	splitArea int
	result    chan interface{}
}

var operationQueue = make(chan *task)

func AddOperation(splitArea int, o Operation) chan interface{} {
	result := make(chan interface{})
	go func() {
		operationQueue <- &task{
			fn:        o,
			splitArea: splitArea,
			result:    result,
		}
	}()
	return result
}

func Do() {
	for v := range operationQueue {
		var x, y int
		w, h := screen.Info.ScreenWidth/2, screen.Info.ScreenHeight/2
		switch v.splitArea {
		case 1:
			x, y = w/4, h/4
		case 2:
			x, y = w+w/4, h/4
		case 3:
			x, y = w/4, h+h/4
		case 4:
			x, y = w+w/4, h+h/4
		}
		if v.splitArea != 0 {
			robotgo.Move(x, y)
			robotgo.MouseClick("left")
			time.Sleep(5 * time.Millisecond)
		}
		v.result <- v.fn()
	}
}
