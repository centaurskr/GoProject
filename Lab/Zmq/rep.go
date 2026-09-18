package main

import (
	"log"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"os"
)

func main(){
	if len(os.Args) != 2{
		fmt.Printf("REP endpoint가 필요합니다.\n\tex)tcp://127.0.0.1:5556\n")
		os.Exit(0)
	}
	context, _ := zmq.NewContext()
	defer context.Term()
	
	s, _ := context.NewSocket(zmq.REP)
	defer s.Close()

	//err := s.Connect("tcp://127.0.0.1:5556")
	err := s.Connect(os.Args[1])
	if err != nil { log.Fatalln(err); }


	for {
		//var rmsg []byte	
		msg, _ := s.RecvMessage(0)
		//fmt.Printf("recv : %s ", string(msg[0][:]))
		rmsg := fmt.Sprintf("%s --- REP Message", string(msg[0][:]))
		s.SendMessage( rmsg) // rmsg is string
	}

}
