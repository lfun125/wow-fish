package main

import (
	"encoding/json"
	"fish/config"
	"fish/fishing"
	"fish/operation"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
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

// func ndpTime() time.Time {
// 	const ntpEpochOffset = 2208988800
// 	type packet struct {
// 		Settings       uint8
// 		Stratum        uint8
// 		Poll           int8
// 		Precision      int8
// 		RootDelay      uint32
// 		RootDispersion uint32
// 		ReferenceID    uint32
// 		RefTimeSec     uint32
// 		RefTimeFrac    uint32
// 		OrigTimeSec    uint32
// 		OrigTimeFrac   uint32
// 		RxTimeSec      uint32
// 		RxTimeFrac     uint32
// 		TxTimeSec      uint32
// 		TxTimeFrac     uint32
// 	}
// 	host := "ntp.aliyun.com:123"
// 	conn, err := net.Dial("udp", host)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()
// 	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
// 		log.Fatalf("failed to set deadline: %v", err)
// 	}

// 	req := &packet{Settings: 0x1B}

// 	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
// 		log.Fatalf("failed to send request: %v", err)
// 	}

// 	rsp := &packet{}
// 	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
// 		log.Fatalf("failed to read server response: %v", err)
// 	}

// 	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
// 	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32

// 	showtime := time.Unix(int64(secs), nanos)
// 	return showtime
// }
