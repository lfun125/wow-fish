package main

import "fish/fishing"

func main() {
	f := fishing.NewFishing(fishing.NewDefaultConfig())
	f.Run()
}
