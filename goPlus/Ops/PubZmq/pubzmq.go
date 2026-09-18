package main
import (
	"fmt"
	"os"
	"time"
	zmq "github.com/pebbe/zmq4"
)

func main(){

	if len(os.Args) < 4 {
		fmt.Printf("Zmq Pub Program\n")
		fmt.Printf("goPubZmq <endpoint> <topic> <message>\n")
		fmt.Printf("  <endpoint> : Connect End point. tcp://127.0.0.1:23423\n")
		fmt.Printf("  <topic>    : message topic\n")	
		fmt.Printf("  <message>  : message \n")	
		os.Exit(0)
	}
	socket , err:= zmq.NewSocket(zmq.PUB)
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer socket.Close()

	socket.Connect(os.Args[1])
	time.Sleep(1000 * time.Millisecond)

	msgIf := make ([]interface{}, len(os.Args) -2)
	for i:= 2; i < len(os.Args); i++{
		msgIf[i-2] = os.Args[i]
	}
	rtn,err := socket.SendMessage(msgIf...)
	if err != nil{
		fmt.Printf("Send Error : %v\n", err)
	}else{
		fmt.Println("Send Ok : ", msgIf,rtn) 
	}
}
