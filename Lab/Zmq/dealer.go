package main

import (
    "fmt"
    "github.com/pebbe/zmq4"
	"os"
)
func Send(){
	for i:=0; i < 10;i++ {
		go func(){
   			//dealer, _ := zmq4.NewSocket(zmq4.DEALER)
   			dealer, _ := zmq4.NewSocket(zmq4.REQ)
			msg:= fmt.Sprintf("Request[%d] : %d", i, os.Getpid())
			//identity := fmt.Sprintf("%d-%d", i, os.Getpid())
   			dealer.Connect("tcp://localhost:5555")

   			//dealer.Send(identity, zmq4.SNDMORE)
   			dealer.Send(msg, 0)
   			reply, _ := dealer.Recv(0)
   			fmt.Println("Received:", reply)
   			dealer.Close()
		}()
	}

}
func main() {
	go func(){
		for i:= 0; i < 10; i++{
			fmt.Println(i)
			go Send()
		}
	}()
	go func (){
		for i:= 0; i < 10; i++{
			fmt.Println(i)
			go Send()
		}
	}()
	go func (){
		for i:= 0; i < 10; i++{
			fmt.Println(i)
			go Send()
		}
	}()
	select{}
}

