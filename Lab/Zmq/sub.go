///
/// @brief  : Glang zmq sub test
/// @file   : sub.go
/// @date   : 2024. 05. 02. (목) 10:40:16 KST
/// @author : Cento
///
package main

import (
	"fmt"
	//"time"
	//zmq "github.com/pebbe/zmq4/draft"
	zmq "github.com/pebbe/zmq4"
	"os"
)

func main() {
	
	if len(os.Args) != 2{
		fmt.Printf("SUB endpoint가 필요합니다.ex)tcp://127.0.0.1:31004\n")
		os.Exit(0)
	}
	endpoint := os.Args[1]

	sub, _:= zmq.NewSocket(zmq.SUB)
	defer sub.Close()
	
	sub.Connect(endpoint)
	fmt.Printf("Sub End point : %s\n", endpoint)

	sub.SetSubscribe("")
	// 완전한 socket이 설정될 시간(1초 정도)
	//  0MQ is so fast, we need to wait a while..
	//time.Sleep(time.Second * 1)


	for {
		
		//topic, _= sub.Recv(0)
		//msg,_ =sub.Recv(0)
		msgs, _:= sub.RecvMessage(0)	
		fmt.Println(msgs[0], msgs[1])
	}
}
