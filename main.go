package main

import (
	"fish/fishing"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	c.ParseParams()
	f := fishing.NewFishing(c)
	log.Fatalln(f.Run())
}
