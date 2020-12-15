package main

import (
	"fish/fishing"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	f := fishing.NewFishing(c)
	f.ColorPicker()
	log.Fatalln(f.Run())
}
