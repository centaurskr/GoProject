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
func readStream(stream *zmq.Socket)(string, string, string, error){
	id, err := stream.Recv(0)
	if err != nil{ return "", "", "", err;}
	/*
	size, err := stream.Recv(0) 
	if err != nil{ return "", "",  "", err;}
	*/
	data, err := stream.Recv(0) 
	return id, "", data, nil
}
func writeStream(stream *zmq.Socket, id, data string)(int, error){
	// Echo Back
	stream.Send(id, zmq.SNDMORE)
	rtn, err :=stream.Send(data, 0)
	return rtn, err	
}
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
					id, size, data, err:=readStream(stream)
					if err != nil {
						fmt.Printf("ER[%x] : %s\n", id, err)
						continue
					}
					if len(data) == 0{
						fmt.Printf("ID [%x] Connection Close\n", id)
						continue
					}

					fmt.Printf("ID [%x] len[%s]  Data : %s\n", id, size, data)
					/*
					rtn, err := writeStream(stream, id, data)
					if err != nil{
						fmt.Printf("Write error : %v\n", err)
						continue
					}
					fmt.Printf("Write Ok! rtn [%d]\n", rtn)
					*/
			} // SWITCH
		} // range  polled
	} // main loop
}

