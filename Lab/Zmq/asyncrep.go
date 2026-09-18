package main

import (
	"log"
	zmq "github.com/pebbe/zmq4"
	"os"
	"fmt"
	"time"
)

func main(){
	context, _ := zmq.NewContext()
	defer context.Term()
	
	s, _ := context.NewSocket(zmq.REQ)
	defer s.Close()

	err := s.Connect("tcp://127.0.0.1:5555")
	if err != nil { log.Fatalln(err); }

	// REQ socket은 Send() Send() Recv() Recv()가 불가 하며
	// Error : "Operation cannot be accomplished in current state"
	// Send() Recv() Send() Recv()로 짝을 맞춰야 함
	for i := 0; i < 100; i++{
		msg := fmt.Sprintf("%d Message (%d)", os.Getpid(), i)
		rtn, err := s.Send(msg, 0)
		if err != nil { log.Fatalln(err);
		}else { fmt.Printf("Send %d , %s\n", rtn, msg); }
	}
	time.Sleep(5 * time.Millisecond)
	for i := 0; i < 100; i++{
		rmsg, _ := s.RecvMessage(0)
		fmt.Printf("from server : %s \n", rmsg[0])
	}

}
