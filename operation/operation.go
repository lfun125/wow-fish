package operation

import (
	"fish/screen"
	"time"

	"github.com/go-vgo/robotgo"
)

type Operation func() interface{}

type task struct {
	fn        Operation
	active    bool
	splitArea int
	result    chan interface{}
}

var operationQueue = make(chan *task)

func AddOperation(splitArea int, active bool, o Operation) chan interface{} {
	result := make(chan interface{})
	go func() {
		operationQueue <- &task{
			fn:        o,
			splitArea: splitArea,
			result:    result,
			active:    active,
		}
	}()
	return result
}

func Do() {
	for v := range operationQueue {
		var x, y int
		w, h := screen.Info.ScreenWidth/2, screen.Info.ScreenHeight/2
		switch v.splitArea {
		case 0, 1:
			x, y = w/4, h/4
		case 2:
			x, y = w+w/4, h/4
		case 3:
			x, y = w/4, h+h/4
		case 4:
			x, y = w+w/4, h+h/4
		}
		if v.active {
			robotgo.MoveMouseSmooth(x, y, 1.0, 0.5)
			if v.splitArea != 0 {
				robotgo.MouseClick("left")
				time.Sleep(5 * time.Millisecond)
			}
		}
		if v.fn != nil {
			v.result <- v.fn()
		} else {
			v.result <- nil
		}
	}
}
