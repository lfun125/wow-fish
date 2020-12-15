package main

import (
	"fish/fishing"
	"fish/utils"
	"fmt"
	"log"

	"github.com/go-vgo/robotgo"
)

func main() {
	c := fishing.NewDefaultConfig()
	ok := robotgo.AddEvent("f4")
	if ok {
		x, y := robotgo.GetMousePos()
		c.FloatColor = utils.StrToRGBA(robotgo.GetPixelColor(x, y))
	}
	f := fishing.NewFishing(c)
	fmt.Println(c)
	log.Fatalln(f.Run())
}
