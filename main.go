package main

import (
	"fish/fishing"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	f := fishing.NewFishing(c)
	go f.ColorPicker()
	log.Fatalln(f.Run())
}
