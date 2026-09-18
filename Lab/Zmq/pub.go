///
/// @brief  : Glang zmq pub test
/// @file   : pub.go
/// @date   : 2024. 05. 02. (목) 10:40:16 KST
/// @author : Cento
///
package main

import (
	"fmt"
	"time"
	//zmq "github.com/pebbe/zmq4/draft"
	zmq "github.com/pebbe/zmq4"
	"os"
)

func main() {
	if len(os.Args) != 2{
		fmt.Println("PUB test Program using Zmq. Topic is \"TOPIC_B\"")
		fmt.Printf("Using) pub.go <endpoint>\n")
		os.Exit(0)
	}
	pub, _:= zmq.NewSocket(zmq.PUB)
	defer pub.Close()
	
	pub.Connect(os.Args[1])
	fmt.Printf("Pub endpoint: %s. Connect Ok\n", os.Args[1])

	// 완전한 socket이 설정될 시간(1초 정도)
	//  0MQ is so fast, we need to wait a while..
	time.Sleep(time.Second * 1)


	topic := "Topic-B"

	for i:= 0 ;i<1000; i++{
		msg:= fmt.Sprintf("Message %d", i)
		sz, _ :=pub.SendMessage(topic, msg)
		fmt.Println(sz, topic, msg)

		time.Sleep(time.Second)
	}
}
