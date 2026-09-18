package main

import (
    "fmt"
    "github.com/pebbe/zmq4"
	//"time"
)

func main() {
    router, _ := zmq4.NewSocket(zmq4.ROUTER)
    //router, _ := zmq4.NewSocket(zmq4.REP)
    defer router.Close()
    router.Bind("tcp://*:5555")

    for {
        msg, _ := router.RecvMessageBytes(0)
        fmt.Printf("Received from : %x\n\t%s\n", 
			 msg, string(msg[2][:]))
		bt := fmt.Sprintf("%s SVR", string(msg[2][:]))
		msg[2] = []byte(bt)
        router.SendMessage(msg, 0)
    }
}

