package main
import (
	"centolib/api"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"encoding/binary"
	"os"
)

func main(){
	val, err := api.GetPlatformEnv("AFTERMAT_STREAM")
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	fmt.Printf("Connection point [%s]\n", val)

	ctx , _ := zmq.NewContext()
	
	socket, _:= ctx.NewSocket(zmq.STREAM)
	socket.Connect(val)

	sockid, _ := socket.GetIdentity()

	msg := "CONNECT AFTERMATCH"
	pkSize := make([]byte, 4)
	binary.BigEndian.PutUint32(pkSize, uint32(len(msg)))
	packet := append(pkSize, msg...)
	socket.SendMessage(sockid, packet)
	
	poller := zmq.NewPoller()
	poller.Add(socket, zmq.POLLIN)

	for {
		polled, err := poller.Poll(-1)
		if err != nil {
			fmt.Printf("poller : %v\n", err)
			continue
		}
		for _, poll := range polled{
			if poll.Socket != socket{
				continue
			}
			msg, _ := socket.RecvMessage(0)
			fmt.Printf("id [%x] [%v]\n", msg[0], msg[1])
		}
	}

}
