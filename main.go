package main

import (
	"fish/fishing"
	"fmt"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	c.ParseParams()
	f := fishing.NewFishing(c)
	f.ColorPicker()
	f.Info(fmt.Sprintf("Config: %+v\n", c))
	log.Fatalln(f.Run())
}
