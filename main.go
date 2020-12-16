package main

import (
	"fish/fishing"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	c.ParseParams()
	f := fishing.NewFishing(c)
	f.ColorPicker()
	f.Info("Config: %+v\n", c)
	log.Fatalln(f.Run())
}
