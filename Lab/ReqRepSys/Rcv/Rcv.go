package main

import (
	Log "centolib/logout"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"os"
)
func monitorFunction(soc *zmq.Socket){
	for {
		a, b, c, err := soc.RecvEvent(0)
		if err != nil {
			Log.Error("%v", err)
			continue
		}
		Log.Infor("%v %v %v", a, b, c)
	}
}
func main(){
	if(len(os.Args) != 4){
		fmt.Printf("%s <index> <REP endpoint> <REP endpoint>\n", os.Args[0])
		os.Exit(0)
	}

	Log.InitLogger(os.Args[1], Log.CONSOLE_MODE);defer Log.Close()

	context, _ := zmq.NewContext()
	defer context.Term()

	s, _:= context.NewSocket(zmq.REP)
	defer s.Close()
	s.Connect(os.Args[2])

	err := s.Monitor("inproc://monitor.rep", zmq.EVENT_ALL)
	if err != nil {
		Log.Error("%v", err)
		os.Exit(0)
	}
	monsocket, _:= context.NewSocket(zmq.PAIR);
	monsocket.Connect("inproc://monitor.rep");

	go monitorFunction(monsocket)

	for{
		msg, _ := s.RecvBytes(0)
		Log.Infor("Got : %s", string(msg[:]))

		// Using Check REQ socket Create
		reqsock, err :=context.NewSocket(zmq.REQ)
		if err != nil{
			Log.Error("zmq.NewSocket(REQ)\n%v", err)
			os.Exit(0)
		}
		// Bind 
		err = reqsock.Connect(os.Args[3])
		if err != nil{
			Log.Error("reqsock.Connect(%s)\n%v", os.Args[3], err)
			os.Exit(0)
		}
		// Send message to Using Process
		_, err = reqsock.SendBytes(msg, 0)
		if err != nil{
			Log.Error("SEND REQ : %v", err)
			os.Exit(0)
		}
		// Receive mesasge from Using Process
		rcv, err := reqsock.RecvBytes(0)	
		if err != nil{
			Log.Error("RECV REQ : %v", err)
			os.Exit(0)
		}
		reqsock.Close()

		// Return Message
		s.SendBytes(rcv, 0)
	}
	
}
