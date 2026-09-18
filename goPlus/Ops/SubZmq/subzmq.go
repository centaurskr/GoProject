package main
import (
	"fmt"
	"os"
	"time"
	zmq "github.com/pebbe/zmq4"
)

func main(){
	if len(os.Args) < 2 {
		fmt.Printf("Zmq Sub Program\n")
		fmt.Printf("goSubZmq <endpoint> [topic]\n")
		fmt.Printf("  <endpoint> : Connect End point. tcp://127.0.0.1:23423\n")
		fmt.Printf("  [topic] : 수신 할 topic, 없으면 모두\n")
		os.Exit(0)
	}
	socket , err:= zmq.NewSocket(zmq.SUB)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	socket.Connect(os.Args[1])
	time.Sleep(1000 * time.Millisecond)
	defer socket.Close()
	if len(os.Args) == 3 {
		socket.SetSubscribe(os.Args[2])
	}else{
		socket.SetSubscribe("")
	}

	for{
		msg, _ := socket.RecvMessage(0)
		fmt.Printf("%s-%s\n", msg[0], msg[1])
	}
}
