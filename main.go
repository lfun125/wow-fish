package main

import (
	"encoding/json"
	"fish/fishing"
	"fmt"
	"log"
)

func main() {
	c := fishing.NewDefaultConfig()
	if importCfg := c.ParseParams(); importCfg {
		bts, _ := json.MarshalIndent(c, "", "    ")
		fmt.Println(string(bts))
		return
	}
	f := fishing.NewFishing(c)
	log.Fatalln(f.Run())
}
