package main
import (
	zmq "github.com/pebbe/zmq4"
	"time"
)

func main(){
	pub, _ := zmq.NewSocket(zmq.PUB)
	defer pub.Close()

	pub.Connect("tcp://127.0.0.1:32768")

	time.Sleep(1000 * time.Microsecond)
	
	// Topic AAA + 1 Message
	pub.SendMessage("AAA", "Message1")

	pub.SendMessage("BBB", "Message1", "Message2", "Message3")
}
