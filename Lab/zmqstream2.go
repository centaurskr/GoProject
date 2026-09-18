///
/// zmq.STREAM을 이용한  
/// tcp server TEST
///
package main
import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	//"centolib/api"
	"time"
)
func main(){

	stream,_ := zmq.NewSocket(zmq.STREAM)
	stream.Bind("tcp://*:32767")	

	poller := zmq.NewPoller()
	poller.Add(stream, zmq.POLLIN)
	for{
		polled, _ := poller.Poll(1000 * time.Millisecond)
		for _, socket := range polled{
			switch socket.Socket {
				case stream :
					msg, err:=stream.RecvMessage(0)
					if err != nil {
						fmt.Printf("%v\n",  err)
						continue
					}
					id := msg[0]
					data := []byte(msg[1])
					if len(data) == 0{
						fmt.Printf("ID [%x] Connection Close\n", id)
						continue
					}

					fmt.Printf("ID [%x] Data : %s\n", id, string(data))
			} // SWITCH
		} // range  polled
	} // main loop
}

