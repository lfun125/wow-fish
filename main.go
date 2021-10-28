package main

import (
	"encoding/json"
	"fish/config"
	"fish/fishing"
	"fish/operation"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

type User struct {
	Name string
}

func main() {
	//checkMac("00-FF-BD-F6-93-F1")
	var splitList []int
	var importCfg bool
	if importCfg, splitList = config.ParseParams(); importCfg {
		bts, _ := json.MarshalIndent(config.C, "", "    ")
		fmt.Println(string(bts))
		return
	}
	log.Println("splitList", splitList)
	var list []*fishing.Fishing
	for _, v := range splitList {
		list = append(list, fishing.NewFishing(v))
	}
	go debug()
	go fishing.WatchKeyboard(list...)
	go operation.Do()
	for _, f := range list {
		go func(f *fishing.Fishing) {
			log.Fatalln(f.Run())
		}(f)
	}
	select {}
}

func debug() {
	if config.C.Debug {
		log.Println(http.ListenAndServe(":8211", nil))
	}
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
