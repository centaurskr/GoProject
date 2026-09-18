package main
import (
	//"strings"
	//"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"time"
)
var Exrate  map[string]string
type ExRateJson struct {
    currency  string `json:"CURRENCY"`
    basetime  string `json:"TIME"`
    rate      string `json:"RATE"`
}

func main(){
Exrate = map[string]string{
    "USDT":"1.0",
    "KRW" :"1.0",
    "BTC" :"1.0",
}
	sub,_ := zmq.NewSocket(zmq.SUB)
	sub.Connect("tcp://192.168.0.99:31102")
	sub.SetSubscribe("EXRATE")

	poller := zmq.NewPoller()
	poller.Add(sub, zmq.POLLIN)

	for {
		polls, _ := poller.Poll(5000 * time.Millisecond)
		if len(polls) == 0{continue;}
		for _, polled := range(polls){
			if polled.Socket != sub {continue;}

			msg, _ := sub.RecvMessage(0)
			exrate := ExRateJson{}
			/*
			err := json.Unmarshal(msg[1], &exrate)
			if err != nil {continue;}
			*/
			
			fmt.Println(exrate)

			fmt.Printf("%+v\n", msg)
		}
	}
}
