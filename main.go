package main

import (
	"encoding/json"
	"fish/fishing"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	checkMac("00-FF-BD-F6-93-F1")
	c := fishing.NewDefaultConfig()
	if importCfg := c.ParseParams(); importCfg {
		bts, _ := json.MarshalIndent(c, "", "    ")
		fmt.Println(string(bts))
		return
	}
	f := fishing.NewFishing(c)
	log.Fatalln(f.Run())
}

func checkMac(s string) {
	s = strings.ToLower(strings.ReplaceAll(s, "-", ":"))
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	var ok bool
	for _, inter := range interfaces {
		ss := strings.ToLower(inter.HardwareAddr.String())
		ss = strings.ReplaceAll(ss, "-", ":")
		if ss == s {
			ok = true
			return
		}
	}
	if !ok {
		log.Fatalln("机器不匹配")
	}
}
