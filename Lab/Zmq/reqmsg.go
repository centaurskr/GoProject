package main

import (
	"log"
	zmq "github.com/pebbe/zmq4"
	"os"
	"fmt"
	"time"
)

func main(){
    if len(os.Args) != 3{
		fmt.Printf("단문 REQ Message\n") 
		fmt.Printf("Using) %s <endpoint> <Message>\n", os.Args[0])
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

	_, err = s.SendMessage(os.Args[2])
	if err != nil { 
		log.Printf("Send : %v", err)
		os.Exit(0)
	}
	msg, _ := s.RecvMessage(0)
	fmt.Printf("%v\n", msg)
}
