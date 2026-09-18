package main

import (
	"log"
	zmq "github.com/pebbe/zmq4"
	//"strconv"
	//"time"
	"os"
)
func monitorFunction(soc *zmq.Socket){
	for {
		a, b, c, err := soc.RecvEvent(0)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(a, b, c)
	}
}
func main(){
	if(len(os.Args) != 2){
		log.Println("Error ", os.Args[0:1], "endpoint")
		os.Exit(0)
	}
	context, _ := zmq.NewContext()
	defer context.Term()

	s, _:= context.NewSocket(zmq.REP)
	defer s.Close()
	s.Bind(os.Args[1])
	err := s.Monitor("inproc://monitor.rep", zmq.EVENT_ALL)
	if err != nil {
		log.Fatalln(err)
	}

	monsocket, _:= context.NewSocket(zmq.PAIR);
	monsocket.Connect("inproc://monitor.rep");

	go monitorFunction(monsocket)
	for{
		msg, _ := s.Recv(0)
		log.Println("Got : ", string(msg))
		s.Send(msg, 0)
	}
	
}
