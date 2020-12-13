package main

import (
	"fish/fishing"
	"log"
)

func main() {
	f := fishing.NewFishing(fishing.NewDefaultConfig())
	log.Fatalln(f.Run())
}
