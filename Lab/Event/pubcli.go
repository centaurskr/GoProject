package main
import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"os"
	"time"
)
func main(){
	if len(os.Args) < 2 {
		fmt.Println("zmq PUB client (connect)")
		fmt.Println("pubcli <connect point>")
		os.Exit(0)
	}
	socket, _ := zmq.NewSocket(zmq.PUB)
	err := socket.Connect(os.Args[1])
	if err != nil{
		fmt.Printf("socket.Bind(%s) : %v\n", os.Args[1], err)
		os.Exit(0)
	}
	defer socket.Close()

	time.Sleep(1000 * time.Millisecond) // For good socket connection

	for i:= 0; i < 100; i++{
		msg := fmt.Sprintf("PUB DATA ---- %d", i)
		size, err := socket.SendMessage("PUBDATA", msg)
		if err != nil{
			fmt.Printf("SendMessage : %v\n", err)
			break
		}
		fmt.Printf("Send size[%d] [%s]\n", size, msg)
		time.Sleep(500 * time.Millisecond)
	}
}
