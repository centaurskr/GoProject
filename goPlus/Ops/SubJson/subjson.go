package main
import (
	"fmt"
	"os"
	"time"
	"encoding/json"
	"bytes"
	"strconv"
	zmq "github.com/pebbe/zmq4"
)

func main(){
	if len(os.Args) < 3 {
		fmt.Printf("Zmq Sub Program\n")
		fmt.Printf("goSubZmq <endpoint> <json index> [topic]\n")
		fmt.Printf("  <endpoint> : Connect End point. tcp://127.0.0.1:23423\n")
		fmt.Printf("  <json index> : json packet의 위치\n")
		fmt.Printf("     Topic + JSON       : json index = 1\n")
		fmt.Printf("     Topic + MSG + JSON : json index = 2\n")
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
	if len(os.Args) == 4 {
		socket.SetSubscribe(os.Args[3])
	}else{
		socket.SetSubscribe("")
	}
	idx,_ := strconv.Atoi(os.Args[2])
	if idx == 0{
		fmt.Println("<json index> 오류. Index : ", idx)
		os.Exit(0)
	}
	for{
		var outBuff bytes.Buffer
		msg, _ := socket.RecvMessage(0)
		json.Indent(&outBuff, []byte(msg[idx]), "", "  ")
		fmt.Println(msg[0], outBuff.String())
	}
}
