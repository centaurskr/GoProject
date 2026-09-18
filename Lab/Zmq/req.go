package main

import (
	"log"
	zmq "github.com/pebbe/zmq4"
	"os"
	"fmt"
	"time"
)

func main(){
    if len(os.Args) != 2{
        fmt.Printf("REQ endpoint가 필요합니다.\n\tex)tcp://127.0.0.1:5555\n")
        os.Exit(0)
    }

	context, _ := zmq.NewContext()
	defer context.Term()
	
	s, _ := context.NewSocket(zmq.REQ)
	defer s.Close()

	//err := s.Connect("tcp://127.0.0.1:5555")
	err := s.Connect(os.Args[1])
	if err != nil { log.Fatalln(err); }


	s.SetRcvtimeo(5 * time.Second) // 5-second receive timeout
	s.SetSndtimeo(5 * time.Second) // 5-second send timeout

	for i := 0; i < 10000; i++{
		msg := fmt.Sprintf("%d Message (%d)", os.Getpid(), i)
		//rtn, err := s.SendMessage([]byte(msg), 0)
		rtn, err := s.Send(msg, 0)
		if err != nil { 
			log.Printf("Send : %v", err)
			break
		}
		fmt.Printf("Send %d , %s ", rtn, msg); 
		//rmsg, _ := s.RecvMessageBytes(0)
		//fmt.Printf("from server : %s \n", string(rmsg[0][:]))
		rmsg, err := s.RecvMessage(0)
		if err != nil{
			log.Printf("Recv : %v ", err)
			break
		}
		fmt.Printf("from server : %v \n", rmsg)
		//rmsg, _ := s.Recv(0) -- 이것은안됨
		//fmt.Printf("from server : %s \n", rmsg)
		time.Sleep(500 * time.Millisecond)
	}

}
