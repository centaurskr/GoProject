package main
import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

func main(){
	sub, _ := zmq.NewSocket(zmq.SUB)
	defer sub.Close()
	sub.Bind("tcp://127.0.0.1:32768")
	sub.SetSubscribe("AAA")
	sub.SetSubscribe("BBB")

	for {
		msg, err := sub.RecvMessageBytes(0)
		if err != nil{
			fmt.Printf("Error RecvMesageBytes(0) : %v\n", err)
			return
		}
		fmt.Printf("How many messages : [%d]\n", len(msg))
		topic := string(msg[0][:])
		switch topic {
			case "AAA" :
				fmt.Printf("Topic AAA. msg-0 [%s]\n", string(msg[1][:]))
			case "BBB" :
				fmt.Printf("Topic BBB. msg-1 [%s]\n", string(msg[1][:]))
				fmt.Printf("Topic BBB. msg-2 [%s]\n", string(msg[2][:]))
				fmt.Printf("Topic BBB. msg-3 [%s]\n", string(msg[3][:]))
		}
	}
}
