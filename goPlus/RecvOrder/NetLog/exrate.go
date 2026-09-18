// 환율 관련 Function
// 
package main

import (
	"encoding/json"
	"strings"
	zmq "github.com/pebbe/zmq4"
	"time"
)

// 환율
var Exrate =  map[string]string{
	"USDT":"1.0",
	"KRW" :"1.0",
	"BTC" :"1.0",
}
type ExRateJson struct {
	Currency  string `json:"CURRENCY"`
	Basetime  string `json:"TIME"`
	Rate      string `json:"RATE"`
}
////////////// 환율 수신 go Routine /////
func goRcvExrate(socket *zmq.Socket){
		
	socket.SetSubscribe("EXRATE")

	poller := zmq.NewPoller()
	poller.Add(socket, zmq.POLLIN)
	// 환율 읽기 

	for {
		polls, _ := poller.Poll(5000 * time.Millisecond) // 5초
		if len(polls) == 0 { continue; }
		for _, polled := range polls{
			if polled.Socket != socket{continue;}
			// msg[0] : EXRATE
			// msg[1] : {"CURRENCY":"USDT", "TIME":"%s", "RATE":"%s" }
			msg, _ := socket.RecvMessage(0)
			exrate := ExRateJson{}
			err := json.Unmarshal([]byte(msg[1]), &exrate)
			if err != nil {continue;}
			Exrate[exrate.Currency] = exrate.Rate;
		}
	}
}

// GetExrateOfCoin은 "BTC.USDT" 형식의 코인 문자열에서 거래통화의 현재 환율을 반환한다.
func GetExrateOfCoin(coin string)string{
	slice := strings.Split(coin, ".")
	return Exrate[slice[1]]
}
